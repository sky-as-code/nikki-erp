package cqrs

import (
	"context"
)

type CqrsBus interface {
	SubscribeRequests(ctx context.Context, handlers ...RequestHandler) (err error)
	RequestNoReply(ctx context.Context, request Request) (err error)
	Request(ctx context.Context, request Request) (_ <-chan Reply, err error)
}

// Deprecated: Not used
type Namer interface {
	Name() string
}
