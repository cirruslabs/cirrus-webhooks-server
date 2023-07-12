package datadogsender

import (
	"context"
	"errors"
	"fmt"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

var ErrAPISenderFailed = errors.New("API sender failed to send the event")

type APISender struct {
	apiClient *datadog.APIClient
	eventsAPI *datadogV1.EventsApi

	apiKey  string
	apiSite string
}

func NewAPISender(apiKey string, apiSite string) (*APISender, error) {
	apiClient := datadog.NewAPIClient(datadog.NewConfiguration())

	return &APISender{
		apiClient: apiClient,
		eventsAPI: datadogV1.NewEventsApi(apiClient),

		apiKey:  apiKey,
		apiSite: apiSite,
	}, nil
}

func (sender *APISender) SendEvent(ctx context.Context, event *Event) error {
	ctx = context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: sender.apiKey,
			},
		},
	)

	ctx = context.WithValue(ctx,
		datadog.ContextServerVariables,
		map[string]string{
			"site": sender.apiSite,
		})

	_, _, err := sender.eventsAPI.CreateEvent(ctx, datadogV1.EventCreateRequest{
		Title: event.Title,
		Text:  event.Text,
		Tags:  event.Tags,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPISenderFailed, err)
	}

	return nil
}
