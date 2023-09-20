package datadogsender

import "context"

type Event struct {
	Title string
	Text  string
	Tags  []string
}

type Sender interface {
	SendEvent(context.Context, *Event) (string, error)
}
