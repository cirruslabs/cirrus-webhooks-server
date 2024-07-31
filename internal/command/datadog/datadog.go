package datadog

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brpaz/echozap"
	payloadpkg "github.com/cirruslabs/cirrus-webhooks-server/internal/command/datadog/payload"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

var debug bool
var httpAddr string
var httpPath string
var eventTypes []string
var secretToken string
var dogstatsdAddr string
var apiKey string
var apiSite string

var (
	ErrDatadogFailed               = errors.New("failed to stream Cirrus CI events to Datadog")
	ErrSignatureVerificationFailed = errors.New("event signature verification failed")
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datadog",
		Short: "Stream Cirrus CI webhook events to Datadog",
		RunE:  runDatadog,
	}

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	cmd.PersistentFlags().StringVar(&httpAddr, "http-addr", ":8080",
		"address on which the HTTP server will listen on")
	cmd.PersistentFlags().StringVar(&httpPath, "http-path", "/",
		"HTTP path on which the webhook events will be expected")
	cmd.PersistentFlags().StringSliceVar(&eventTypes, "event-types", []string{},
		"comma-separated list of the event types to limit processing to "+
			"(for example, --event-types=audit_event or --event-types=build,task")
	cmd.PersistentFlags().StringVar(&secretToken, "secret-token", "",
		"if specified, this value will be used as a HMAC SHA-256 secret to verify the webhook events")
	cmd.PersistentFlags().StringVar(&dogstatsdAddr, "dogstatsd-addr", "",
		"enables sending webhook events as Datadog events via the DogStatsD protocol to the specified address "+
			"(for example, --dogstatsd-addr=127.0.0.1:8125)")
	cmd.PersistentFlags().StringVar(&apiKey, "api-key", "",
		"Enables sending webhook events as Datadog logs via the Datadog API using the specified API key")
	cmd.PersistentFlags().StringVar(&apiSite, "api-site", "datadoghq.com",
		"specifies the Datadog site to use when sending webhook events as Datadog logs via the Datadog API")

	return cmd
}

func runDatadog(cmd *cobra.Command, args []string) error {
	// Initialize the logger
	config := zap.NewProductionConfig()
	if debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger := zap.Must(config.Build()).Sugar()

	// Initialize a Datadog sender
	var sender datadogsender.Sender
	var err error

	switch {
	case dogstatsdAddr != "":
		sender, err = datadogsender.NewDogstatsdSender(dogstatsdAddr)
	case apiKey != "":
		sender, err = datadogsender.NewAPISender(apiKey, apiSite)
	default:
		return fmt.Errorf("%w: no sender configured, please specify either --api-key or --dogstatsd-addr",
			ErrDatadogFailed)
	}

	if err != nil {
		return err
	}

	// Convert event types to a set for faster lookup
	eventTypesSet := mapset.NewSet[string](eventTypes...)

	// Configure HTTP server
	e := echo.New()

	e.Use(echozap.ZapLogger(logger.Desugar()))

	e.POST(httpPath, func(ctx echo.Context) error {
		return processWebhookEvent(ctx, logger, sender, eventTypesSet)
	})

	server := &http.Server{
		Addr:              httpAddr,
		Handler:           e,
		ReadHeaderTimeout: 10 * time.Second,
	}

	logger.Infof("starting HTTP server on %s", httpAddr)

	httpServerErrCh := make(chan error, 1)

	go func() {
		httpServerErrCh <- server.ListenAndServe()
	}()

	select {
	case <-cmd.Context().Done():
		if err := server.Close(); err != nil {
			return err
		}
	case httpServerErr := <-httpServerErrCh:
		return httpServerErr
	}

	return <-httpServerErrCh
}

func processWebhookEvent(
	ctx echo.Context,
	logger *zap.SugaredLogger,
	sender datadogsender.Sender,
	eventTypesSet mapset.Set[string],
) error {
	// Make sure this is an event we're looking for
	presentedEventType := ctx.Request().Header.Get("X-Cirrus-Event")

	if eventTypesSet.Cardinality() != 0 && !eventTypesSet.Contains(presentedEventType) {
		logger.Debugf("skipping event of type %q because we only process events of types %s",
			presentedEventType, strings.Join(eventTypesSet.ToSlice(), ", "))

		return ctx.String(http.StatusOK, fmt.Sprintf("skipping event of type %q", presentedEventType))
	}

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		logger.Warnf("failed to read request's body: %v", err)

		return ctx.NoContent(http.StatusBadRequest)
	}

	// Verify that this event comes from the Cirrus CI
	if err := verifyEvent(ctx, body); err != nil {
		logger.Warnf("%v", err)

		return ctx.NoContent(http.StatusBadRequest)
	}

	// Log this event into the Datadog
	evt := &datadogsender.Event{
		Title: "Webhook event",
		Text:  string(body),
		Tags:  []string{fmt.Sprintf("webhook_event_type:%s", presentedEventType)},
	}

	// Enrich the event with tags
	var payload payloadpkg.Payload

	switch presentedEventType {
	case "audit_event":
		payload = &payloadpkg.AuditEvent{}
	default:
		payload = &payloadpkg.Common{}
	}

	if err = json.Unmarshal(body, payload); err != nil {
		logger.Warnf("failed to enrich Datadog event with tags: "+
			"failed to parse the webhook event of type %q as JSON: %v", presentedEventType, err)
	} else {
		payload.Enrich(ctx.Request().Header, evt, logger)
	}

	// Datadog silently discards log events submitted with a
	// timestamp that is more than 18 hours in the past, sigh.
	//
	// [1]: https://docs.datadoghq.com/api/latest/logs/#send-logs
	if !evt.Timestamp.IsZero() && time.Since(evt.Timestamp) >= 18*time.Hour {
		logger.Warnf("submitting an event of type %q with a timestamp that is more than "+
			"18 hours in the past, it'll likely going to be discarded", presentedEventType)
	}

	message, err := sender.SendEvent(ctx.Request().Context(), evt)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatadogFailed, err)
	}

	return ctx.String(http.StatusCreated, message)
}

func verifyEvent(ctx echo.Context, body []byte) error {
	// Nothing to do
	if secretToken == "" {
		return nil
	}

	// Calculate the expected signature
	hmacSHA256 := hmac.New(sha256.New, []byte(secretToken))
	hmacSHA256.Write(body)
	expectedSignature := hmacSHA256.Sum(nil)

	// Prepare the presented signature
	presentedSignatureRaw := ctx.Request().Header.Get("X-Cirrus-Signature")
	presentedSignature, err := hex.DecodeString(presentedSignatureRaw)
	if err != nil {
		return fmt.Errorf("%w: failed to hex-decode the signature %q: %v",
			ErrSignatureVerificationFailed, presentedSignatureRaw, err)
	}

	// Compare signatures
	if !hmac.Equal(expectedSignature, presentedSignature) {
		return fmt.Errorf("%w: signature is not valid", ErrSignatureVerificationFailed)
	}

	return nil
}
