package pubsub

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, message any) error
}

type Subcriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan []byte, error)
}
