package cqrs

import (
	"context"
	stdErrors "errors"
	"fmt"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

const DefaultQueryTimeoutSecs = 50 * time.Second
const MetaCorrelationId = "correlation_id"
const MetaResponseTopic = "response_topic"

type QueryBusConfig struct {
	// MessagePublisher message.Publisher
	// CommandBus       *cqrs.CommandBus
	// EventProcessor   *cqrs.EventProcessor
	// Logger           watermill.LoggerAdapter

	// Publisher is used to publish queries.
	Publisher message.Publisher

	// Subscriber is used to subscribe to responses.
	Subscriber message.Subscriber

	// GenerateQueryTopic is used to generate topic for publishing query and subscribing response.
	GenerateQueryTopic QueryBusGenerateQueryTopicFn

	// Marshaler is used to marshal and unmarshal commands.
	// It is required.
	Marshaler QueryEventMarshaler

	// MaxTimeout is the maximum time to wait for a response.
	// Each Send() invocation can accept a context.WithTimeout with shorter wait time.
	//
	// This option is not required. Default is 50 seconds.
	MaxTimeout time.Duration

	// OnSend is called before publishing the query.
	// This option is not required.
	OnSend QueryBusOnSendFn

	// OnRespond is called before handling response.
	// OnRespond works in a similar way to middlewares: you can inject additional logic before and after handling a response.
	//
	// Because of that, you need to explicitly call params.Handler.Handle() to handle the command.
	//  func(params QueryBusOnRespondParams) (err error) {
	//      // logic before handle
	//      //  (...)
	//
	//      err := params.Handler.Handle(params.Message.Context(), params.Response)
	//
	//      // logic after handle
	//      //  (...)
	//
	//      return err
	//  }
	//
	// This option is not required.
	OnRespond QueryBusOnRespondFn

	// Logger instance used to log.
	// If not provided, watermill.NopLogger is used.
	Logger watermill.LoggerAdapter
}

func (this *QueryBusConfig) setDefaults() {
	if this.MaxTimeout == 0 {
		this.MaxTimeout = DefaultQueryTimeoutSecs
	}
}

func (this QueryBusConfig) Validate() error {
	var err error

	if this.Publisher == nil {
		err = stdErrors.Join(err, errors.New("missing Publisher"))
	}
	if this.Subscriber == nil {
		err = stdErrors.Join(err, errors.New("missing Subscriber"))
	}

	if this.Marshaler == nil {
		err = stdErrors.Join(err, errors.New("missing Marshaler"))
	}

	if this.GenerateQueryTopic == nil {
		err = stdErrors.Join(err, errors.New("missing GenerateQueryTopic"))
	}

	return err
}

type QueryBusGenerateQueryTopicFn func(QueryBusGeneratePublishTopicParams) (string, error)

type QueryBusGeneratePublishTopicParams struct {
	QueryName string
}

type QueryBusOnSendFn func(params QueryBusOnSendParams) error
type QueryBusOnSendParams struct {
	QueryName string
	Query     any

	// Message is the raw message before marshaling to Query.
	// It is never nil and can be modified.
	Message *message.Message
}

type QueryBusOnRespondFn func(params QueryBusOnRespondParams) error
type QueryBusOnRespondParams struct {
	ResponseName string
	Response     any

	// Message is the raw message before unmarshaling to Response.
	// Message is never nil and can be modified.
	Message *message.Message
}

// NewQueryBusWithConfig creates a new QueryBus.
func NewQueryBusWithConfig(config QueryBusConfig) (*QueryBus, error) {
	config.setDefaults()
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	queryBus := QueryBus{
		config: config,
	}
	err := queryBus.setupInternalCmdBus()
	if err != nil {
		return nil, err
	}

	return &queryBus, nil
}

type queryInvocation struct {
	correlationId     string
	responseChan      chan Reply
	responseTopic     string
	responseProcessor *cqrs.CommandProcessor
	responseRouter    *message.Router
}

type QueryBus struct {
	commandBus *cqrs.CommandBus
	config     QueryBusConfig

	// invocations keeps track of all sent queries, keyed by correlation ID
	invocations sync.Map
}

func (this *QueryBus) InvocationsCount() int {
	count := 0
	this.invocations.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// Send sends command to the command bus.
func (this QueryBus) Send(ctx context.Context, query any) (_ <-chan Reply, err error) {
	// func (this QueryBus) Send(ctx context.Context, query any) (any, error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "QueryBus.Send")
		}
	}()

	packet := NewQueryPacket(query)
	corId := packet.CorrelationId()
	queryName := this.config.Marshaler.Name(query)
	respChan := make(chan Reply, 1)
	invocation := queryInvocation{
		correlationId: corId,
		responseChan:  respChan,
		responseTopic: responseTopic(queryName, corId),
	}
	this.invocations.Store(corId, invocation)
	outputCh := make(chan Reply, 1)

	defer func() {
		close(respChan)
		close(outputCh)
		this.invocations.Delete(corId)
	}()

	router, respProcessor, err := this.setupInternalCmdProcessor(invocation)
	ft.PanicOnErr(err)

	invocation.responseProcessor = respProcessor
	invocation.responseRouter = router

	err = this.subscribeToResponse(ctx, invocation)
	ft.PanicOnErr(err)

	err = this.commandBus.Send(ctx, packet)
	ft.PanicOnErr(err)

	go func() {
		var reply Reply
		// Wait for response until timeout
		select {
		case reply = <-respChan:
			err = nil
		case <-ctx.Done():
			err = ctx.Err()
		case <-time.After(this.config.MaxTimeout):
			err = errors.Errorf("timeout waiting for query response")
		}
		// Destroy subscription
		invocation.responseRouter.Close()
		outputCh <- reply
	}()

	return outputCh, err
}

func (qb *QueryBus) subscribeToResponse(ctx context.Context, invocation queryInvocation) error {
	err := invocation.responseProcessor.AddHandlers(
		cqrs.NewCommandHandler(invocation.responseTopic, func(ctx context.Context, reply *Reply) error {
			invocation.responseChan <- *reply
			return nil
		}),
	)
	if err != nil {
		return err
	}
	if err := invocation.responseRouter.Run(ctx); err != nil {
		return err
	}
	return nil
}

func (this *QueryBus) setupInternalCmdBus() error {
	cmdBusConfig := cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return this.config.GenerateQueryTopic(QueryBusGeneratePublishTopicParams{
				QueryName: params.CommandName,
			})
		},
		OnSend: func(params cqrs.CommandBusOnSendParams) error {
			packet := params.Command.(QueryPacket)
			correlationId := packet.CorrelationId()

			// The response message must have this exact correlation_id
			// and be sent through this response_topic.
			params.Message.Metadata.Set(MetaCorrelationId, correlationId)
			params.Message.Metadata.Set(MetaResponseTopic, responseTopic(params.CommandName, correlationId))

			if this.config.OnSend != nil {
				queryParams := QueryBusOnSendParams{
					QueryName: params.CommandName,
					Query:     packet,
					Message:   params.Message,
				}
				return this.config.OnSend(queryParams)
			}
			return nil
		},
		Marshaler: this.config.Marshaler,
		Logger:    this.config.Logger,
	}
	cmdBus, err := cqrs.NewCommandBusWithConfig(this.config.Publisher, cmdBusConfig)
	if err != nil {
		return err
	}
	this.commandBus = cmdBus
	return nil
}

func (this *QueryBus) setupInternalCmdProcessor(invocation queryInvocation) (*message.Router, *cqrs.CommandProcessor, error) {
	router, err := message.NewRouter(message.RouterConfig{}, this.config.Logger)
	if err != nil {
		return nil, nil, err
	}
	processorConfig := cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return invocation.responseTopic, nil
		},
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return this.config.Subscriber, nil
		},
		OnHandle: func(responseParams cqrs.CommandProcessorOnHandleParams) error {
			responseTopic := invocation.responseTopic
			queryCorId := invocation.correlationId
			responseCorId := responseParams.Message.Metadata.Get(MetaCorrelationId)
			if queryCorId != responseCorId {
				this.config.Logger.Error("Response correlation_id mismatch", errors.New("response correlation_id mismatch"), watermill.LogFields{
					"queryCorrelationId":    queryCorId,
					"responseCorrelationId": responseCorId,
					"responseTopic":         responseTopic,
				})
				return nil
			}
			if this.config.OnRespond != nil {
				return this.config.OnRespond(QueryBusOnRespondParams{
					ResponseName: responseParams.CommandName,
					Response:     responseParams.Command,
					Message:      responseParams.Message,
				})
			}
			return responseParams.Handler.Handle(responseParams.Message.Context(), responseParams.Command)
		},
		Marshaler: this.config.Marshaler,
		Logger:    this.config.Logger,
	}
	respProcessor, err := cqrs.NewCommandProcessorWithConfig(router, processorConfig)
	return router, respProcessor, err
}

func responseTopic(queryTopic string, correlationId string) string {
	return fmt.Sprintf("%s:response-%s", queryTopic, correlationId)
}
