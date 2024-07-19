package datadogsender

import (
	"context"
	"time"
)

type Event struct {
	Title     string
	Text      string
	Timestamp time.Time
	Tags      []string
}

type Sender interface {
	SendEvent(context.Context, *Event) (string, error)
}
