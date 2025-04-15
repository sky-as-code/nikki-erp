package cqrs

import (
	"context"
)

// QueryHandler receives a query and handles it with the Handle method.
// If using DDD, QueryHandler may modify and persist the aggregate.
//
// In contrast to EventHandler, every Query must have only one QueryHandler.
//
// One instance of QueryHandler is used during handling messages.
// When multiple queries are delivered at the same time, Handle method can be executed multiple times at the same time.
// Because of that, Handle method needs to be thread safe!
type QueryHandler interface {
	// HandlerName is the name used in message.Router while creating handler.
	//
	// It will be also passed to CommandsSubscriberConstructor.
	// May be useful, for example, to create a consumer group per each handler.
	//
	// WARNING: If HandlerName was changed and is used for generating consumer groups,
	// it may result with **reconsuming all messages**!
	HandlerName() string

	NewQuery() any

	Handle(ctx context.Context, cmd any) (Reply, error)
}

// NewQueryHandler creates a new QueryHandler implementation based on provided function
// and query type inferred from function argument.
func NewQueryHandler[Query any](
	handlerName string,
	handleFunc func(ctx context.Context, query *Query) (Reply, error),
) QueryHandler {
	return &genericQueryHandler[Query]{
		handleFunc:  handleFunc,
		handlerName: handlerName,
	}
}

type genericQueryHandler[Query any] struct {
	handleFunc  func(ctx context.Context, query *Query) (Reply, error)
	handlerName string
}

func (this genericQueryHandler[Query]) HandlerName() string {
	return this.handlerName
}

func (this genericQueryHandler[Query]) NewQuery() any {
	tVar := new(Query)
	return tVar
}

func (this genericQueryHandler[Query]) Handle(ctx context.Context, query any) (Reply, error) {
	q := query.(*Query)
	return this.handleFunc(ctx, q)
}
