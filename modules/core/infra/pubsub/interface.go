package pubsub

import "context"

type Publisher interface {
	Publish(ctx context.Context, channel string, message any) error
}

type Subcriber interface {
	Subscribe(ctx context.Context, channel string) (<-chan []byte, error)
}
