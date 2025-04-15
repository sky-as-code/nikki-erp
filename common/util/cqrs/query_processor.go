package cqrs

import (
	stdErrors "errors"
	"fmt"
	"reflect"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
)

type QueryProcessorConfig struct {
	// GenerateSubscribeTopic is used to generate topic for subscribing queries.
	GenerateSubscribeTopic QueryProcessorGenerateSubscribeTopicFn

	// SubscriberConstructor is used to create subscriber for QueryHandler.
	SubscriberConstructor QueryProcessorSubscriberConstructorFn

	// Publisher is used to publish replies.
	Publisher message.Publisher

	// OnHandle is called before handling query.
	// OnHandle works in a similar way to middlewares: you can inject additional logic before and after handling a query.
	//
	// Because of that, you need to explicitly call params.Handler.Handle() to handle the query.
	//  func(params QueryProcessorOnHandleParams) (err error) {
	//      // logic before handle
	//      //  (...)
	//
	//      err := params.Handler.Handle(params.Message.Context(), params.Query)
	//
	//      // logic after handle
	//      //  (...)
	//
	//      return err
	//  }
	//
	// This option is not required.
	OnHandle QueryProcessorOnHandleFn

	// Marshaler is used to marshal and unmarshal queries.
	// It is required.
	Marshaler QueryEventMarshaler

	// Logger instance used to log.
	// If not provided, watermill.NopLogger is used.
	Logger watermill.LoggerAdapter

	// If true, QueryProcessor will ack messages even if QueryHandler returns an error.
	// If RequestReplyBackend is not null and sending reply fails, the message will be nack-ed anyway.
	AckQueryHandlingErrors bool
}

func (c *QueryProcessorConfig) setDefaults() {
	if c.Logger == nil {
		c.Logger = watermill.NopLogger{}
	}
}

func (c QueryProcessorConfig) Validate() error {
	var err error

	if c.Marshaler == nil {
		err = stdErrors.Join(err, errors.New("missing Marshaler"))
	}
	if c.Publisher == nil {
		err = stdErrors.Join(err, errors.New("missing Publisher"))
	}

	if c.GenerateSubscribeTopic == nil {
		err = stdErrors.Join(err, errors.New("missing GenerateSubscribeTopic"))
	}
	if c.SubscriberConstructor == nil {
		err = stdErrors.Join(err, errors.New("missing SubscriberConstructor"))
	}

	return err
}

type QueryProcessorGenerateSubscribeTopicFn func(QueryProcessorGenerateSubscribeTopicParams) (string, error)

type QueryProcessorGenerateSubscribeTopicParams struct {
	QueryName    string
	QueryHandler QueryHandler
}

// QueryProcessorSubscriberConstructorFn creates subscriber for QueryHandler.
// It allows you to create a separate customized Subscriber for every query handler.
type QueryProcessorSubscriberConstructorFn func(QueryProcessorSubscriberConstructorParams) (message.Subscriber, error)

type QueryProcessorSubscriberConstructorParams struct {
	QueryName   string
	HandlerName string
	Handler     QueryHandler
}

type QueryProcessorOnHandleFn func(params QueryProcessorOnHandleParams) (Reply, error)

type QueryProcessorOnHandleParams struct {
	Handler QueryHandler

	QueryName string
	Query     any

	// Message is the raw message before marshaling to Query.
	// It is never nil and can be modified.
	Message *message.Message
}

// NewQueryProcessorWithConfig creates a new QueryProcessor
func NewQueryProcessorWithConfig(router *message.Router, config QueryProcessorConfig) (*QueryProcessor, error) {
	config.setDefaults()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	if router == nil {
		return nil, errors.New("missing router")
	}

	return &QueryProcessor{
		router: router,
		config: config,
	}, nil
}

// QueryProcessor determines which QueryHandler should handle the query received from the query bus.
type QueryProcessor struct {
	router *message.Router

	handlers []QueryHandler

	config QueryProcessorConfig
}

// AddHandlers adds a new QueryHandler to the QueryProcessor and adds it to the router.
func (this *QueryProcessor) AddHandlers(handlers ...QueryHandler) error {
	handledQueries := map[string]struct{}{}
	for _, handler := range handlers {
		queryName := this.config.Marshaler.Name(handler.NewQuery())
		if _, ok := handledQueries[queryName]; ok {
			return DuplicateQueryHandlerError{queryName}
		}

		handledQueries[queryName] = struct{}{}
	}

	for _, handler := range handlers {
		if _, err := this.addHandlerToRouter(this.router, handler); err != nil {
			return err
		}

		this.handlers = append(this.handlers, handler)
	}

	return nil
}

// AddHandler adds a new QueryHandler to the QueryProcessor and adds it to the router.
func (this *QueryProcessor) AddHandler(handler QueryHandler) (*message.Handler, error) {
	h, err := this.addHandlerToRouter(this.router, handler)
	if err != nil {
		return nil, err
	}

	this.handlers = append(this.handlers, handler)

	return h, nil
}

// DuplicateQueryHandlerError occurs when a handler with the same name already exists.
type DuplicateQueryHandlerError struct {
	QueryName string
}

func (this DuplicateQueryHandlerError) Error() string {
	return fmt.Sprintf("query handler for query %s already exists", this.QueryName)
}

func (this QueryProcessor) addHandlerToRouter(router *message.Router, handler QueryHandler) (*message.Handler, error) {
	handlerName := handler.HandlerName()
	queryName := this.config.Marshaler.Name(handler.NewQuery())

	topicName, err := this.config.GenerateSubscribeTopic(QueryProcessorGenerateSubscribeTopicParams{
		QueryName:    queryName,
		QueryHandler: handler,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot generate topic for query handler %s", handlerName)
	}

	logger := this.config.Logger.With(watermill.LogFields{
		"query_handler_name": handlerName,
		"topic":              topicName,
	})

	handlerFunc, err := this.routerHandlerFunc(handler, logger)
	if err != nil {
		return nil, err
	}

	logger.Debug("Adding CQRS query handler to router", nil)

	subscriber, err := this.config.SubscriberConstructor(QueryProcessorSubscriberConstructorParams{
		QueryName:   queryName,
		HandlerName: handlerName,
		Handler:     handler,
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot create subscriber for query processor")
	}

	return router.AddHandler(
		handlerName,
		topicName,
		subscriber,
		"",
		this.config.Publisher,
		handlerFunc,
	), nil
}

// Handlers returns the QueryProcessor's handlers.
func (this QueryProcessor) Handlers() []QueryHandler {
	return this.handlers
}

func (this QueryProcessor) routerHandlerFunc(handler QueryHandler, logger watermill.LoggerAdapter) (message.HandlerFunc, error) {
	query := handler.NewQuery()
	queryName := this.config.Marshaler.Name(query)

	if err := this.validateQuery(query); err != nil {
		return nil, err
	}

	return func(msg *message.Message) ([]*message.Message, error) {
		query := handler.NewQuery()
		messageQueryName := this.config.Marshaler.NameFromMessage(msg)

		if messageQueryName != queryName {
			logger.Trace("Received different query type than expected, ignoring", watermill.LogFields{
				"message_uuid":        msg.UUID,
				"expected_query_type": queryName,
				"received_query_type": messageQueryName,
			})
			return nil, nil
		}

		logger.Debug("Handling query", watermill.LogFields{
			"message_uuid":        msg.UUID,
			"received_query_type": messageQueryName,
		})

		ctx := cqrs.CtxWithOriginalMessage(msg.Context(), msg)
		msg.SetContext(ctx)

		if err := this.config.Marshaler.Unmarshal(msg, query); err != nil {
			return nil, err
		}

		handle := func(params QueryProcessorOnHandleParams) (Reply, error) {
			return params.Handler.Handle(ctx, params.Query)
		}
		if this.config.OnHandle != nil {
			handle = this.config.OnHandle
		}

		reply, err := handle(QueryProcessorOnHandleParams{
			Handler:   handler,
			QueryName: messageQueryName,
			Query:     query,
			Message:   msg,
		})

		if this.config.AckQueryHandlingErrors && err != nil {
			logger.Error("Error when handling query, acking (AckQueryHandlingErrors is enabled)", err, nil)
			return nil, nil
		}
		if err != nil {
			logger.Debug("Error when handling query, nacking", watermill.LogFields{"err": err})
			return nil, err
		}

		replyMsg, err := this.config.Marshaler.Marshal(reply)
		return message.Messages{replyMsg}, err
	}, nil
}

func (this QueryProcessor) validateQuery(query any) error {
	// QueryHandler's NewQuery must return a pointer, because it is used to unmarshal
	if err := isPointer(query); err != nil {
		return errors.Wrap(err, "query must be a non-nil pointer")
	}

	return nil
}

func isPointer(v any) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return NonPointerError{rv.Type()}
	}

	return nil
}

type NonPointerError struct {
	Type reflect.Type
}

func (this NonPointerError) Error() string {
	return "non-pointer query: " + this.Type.String() + ", handler.NewQuery() should return pointer to the query"
}
