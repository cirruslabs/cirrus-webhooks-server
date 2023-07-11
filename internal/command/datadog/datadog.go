package datadog

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/brpaz/echozap"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

var httpAddr string
var httpPath string
var eventType string
var secretToken string
var dogStatsdAddr string

var (
	ErrDatadogFailed               = errors.New("failed to stream Cirrus CI events to DataDog")
	ErrSignatureVerificationFailed = errors.New("event signature verification failed")
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datadog",
		Short: "Stream Cirrus CI webhook events to DataDog",
		RunE:  runDatadog,
	}

	cmd.PersistentFlags().StringVar(&httpAddr, "http-addr", ":8080",
		"address on which the HTTP server will listen on")
	cmd.PersistentFlags().StringVar(&httpPath, "http-path", "/",
		"HTTP path on which the webhook events will be expected")
	cmd.PersistentFlags().StringVar(&eventType, "event-type", "audit_event",
		"event type to process (for example \"build\", \"task\" or \"audit_event\")")
	cmd.PersistentFlags().StringVar(&secretToken, "secret-token", "",
		"if specified, this value will be used as a HMAC SHA-256 secret to verify the webhook events")
	cmd.PersistentFlags().StringVar(&dogStatsdAddr, "dogstatsd-addr", "127.0.0.1:8125",
		"DogStatsD address to send the events to")

	return cmd
}

func runDatadog(cmd *cobra.Command, args []string) error {
	// Initialize the logger
	logger := zap.Must(zap.NewProduction()).Sugar()

	// Initialize the DogStatsD client
	logger.Infof("connecting to DogStatsD on %s", dogStatsdAddr)
	dogStatsdClient, err := statsd.New(dogStatsdAddr)
	if err != nil {
		return fmt.Errorf("%w: failed to initialize DogStatsD client: %v",
			ErrDatadogFailed, err)
	}

	// Configure HTTP server
	e := echo.New()

	e.Use(echozap.ZapLogger(logger.Desugar()))

	e.POST(httpPath, func(ctx echo.Context) error {
		return processWebhookEvent(ctx, logger, dogStatsdClient)
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

func processWebhookEvent(ctx echo.Context, logger *zap.SugaredLogger, statsdClient *statsd.Client) error {
	// Make sure this is an event we're looking for
	presentedEventType := ctx.Request().Header.Get("X-Cirrus-Event")
	if presentedEventType != eventType {
		logger.Debugf("skipping event of type %q because we only process events of type %q",
			presentedEventType, eventType)

		return nil
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

	// Log this event into the DataDog
	evt := &statsd.Event{
		Title: fmt.Sprintf("Webhook event of type %s", eventType),
		Text:  string(body),
	}

	if err := evt.Check(); err != nil {
		return fmt.Errorf("%w: event validation failed: %v", ErrDatadogFailed, err)
	}

	if err := statsdClient.Event(evt); err != nil {
		return fmt.Errorf("%w: %v", ErrDatadogFailed, err)
	}

	return nil
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
