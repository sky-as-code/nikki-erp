package pubsub

import (
	"context"

	"go.uber.org/dig"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, message []byte) error
}

type Subcriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan []byte, error)
}

type PubSub interface {
	Publisher
	Subcriber
}

type PubSubOut struct {
	dig.Out
	PubSub
	Publisher
	Subcriber
}
