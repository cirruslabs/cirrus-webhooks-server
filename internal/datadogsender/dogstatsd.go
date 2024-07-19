package datadogsender

import (
	"context"
	"errors"
	"fmt"
	"github.com/DataDog/datadog-go/v5/statsd"
)

var ErrDogstatsdSenderFailed = errors.New("DogStatsD sender failed to send the event")

type DogstatsdSender struct {
	client *statsd.Client
}

func NewDogstatsdSender(addr string) (*DogstatsdSender, error) {
	client, err := statsd.New(addr)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to initialize DogStatsD client: %v",
			ErrDogstatsdSenderFailed, err)
	}

	return &DogstatsdSender{
		client: client,
	}, nil
}

func (sender *DogstatsdSender) SendEvent(ctx context.Context, event *Event) (string, error) {
	if err := sender.client.Event(&statsd.Event{
		Title:     event.Title,
		Text:      event.Text,
		Timestamp: event.Timestamp,
		Tags:      event.Tags,
	}); err != nil {
		return "", fmt.Errorf("%w: %v", ErrDogstatsdSenderFailed, err)
	}

	return "", nil
}
