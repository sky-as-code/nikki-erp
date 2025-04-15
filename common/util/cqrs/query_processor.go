package cqrs

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

type QueryProcessorConfig struct {
	// GenerateSubscribeTopic is used to generate topic for subscribing command.
	GenerateSubscribeTopic CommandProcessorGenerateSubscribeTopicFn

	// SubscriberConstructor is used to create subscriber for QueryHandler.
	SubscriberConstructor QueryProcessorSubscriberConstructorFn

	// OnHandle is called before handling command.
	// OnHandle works in a similar way to middlewares: you can inject additional logic before and after handling a command.
	//
	// Because of that, you need to explicitly call params.Handler.Handle() to handle the command.
	//  func(params CommandProcessorOnHandleParams) (err error) {
	//      // logic before handle
	//      //  (...)
	//
	//      err := params.Handler.Handle(params.Message.Context(), params.Command)
	//
	//      // logic after handle
	//      //  (...)
	//
	//      return err
	//  }
	//
	// This option is not required.
	OnHandle QueryProcessorOnHandleFn

	// Marshaler is used to marshal and unmarshal commands.
	// It is required.
	Marshaler cqrs.CommandEventMarshaler

	// Logger instance used to log.
	// If not provided, watermill.NopLogger is used.
	Logger watermill.LoggerAdapter

	// If true, CommandProcessor will ack messages even if CommandHandler returns an error.
	// If RequestReplyBackend is not null and sending reply fails, the message will be nack-ed anyway.
	//
	// Warning: It's not recommended to use this option when you are using requestreply component
	// (requestreply.NewCommandHandler or requestreply.NewCommandHandlerWithResult), as it may ack the
	// command when sending reply failed.
	//
	// When you are using requestreply, you should use requestreply.PubSubBackendConfig.AckCommandErrors.
	AckCommandHandlingErrors bool
}

type CommandProcessorGenerateSubscribeTopicFn func(CommandProcessorGenerateSubscribeTopicParams) (string, error)

type CommandProcessorGenerateSubscribeTopicParams struct {
	CommandName    string
	CommandHandler QueryHandler
}

// QueryProcessorSubscriberConstructorFn creates subscriber for CommandHandler.
// It allows you to create a separate customized Subscriber for every command handler.
type QueryProcessorSubscriberConstructorFn func(QueryProcessorSubscriberConstructorParams) (message.Subscriber, error)

type QueryProcessorSubscriberConstructorParams struct {
	QueryName   string
	HandlerName string
	Handler     QueryHandler
}

type QueryProcessorOnHandleFn func(params QueryProcessorOnHandleParams) error

type QueryProcessorOnHandleParams struct {
	Handler QueryHandler

	CommandName string
	Command     any

	// Message is never nil and can be modified.
	Message *message.Message
}
