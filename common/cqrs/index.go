package cqrs

// import (
// 	"context"
// 	"log/slog"
// 	"time"

// 	"github.com/ThreeDotsLabs/watermill"
// 	"github.com/ThreeDotsLabs/watermill/components/cqrs"
// 	"github.com/ThreeDotsLabs/watermill/components/requestreply"
// 	"github.com/ThreeDotsLabs/watermill/message"
// 	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
// 	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

// 	"github.com/sky-as-code/nikki-erp/common/logging"
// )

// func InitSubModule() error {
// 	logger := watermillLogger()
// 	pubSub := goChannelPubSub(logger)

// 	backend, err := requestreply.NewPubSubBackend[any](requestreply.PubSubBackendConfig{
// 		Publisher: pubSub,
// 		SubscriberConstructor: func(params requestreply.PubSubBackendSubscribeParams) (message.Subscriber, error) {
// 			return pubSub, nil
// 		},
// 	},

// 	commandBus, err := createCommandBus(pubSub, logger)
// 	if err != nil {
// 		return err
// 	}

// 	router, err := setupRouter(logger)
// 	if err != nil {
// 		return err
// 	}

// 	if err := setupCommandProcessor(router, pubSub, logger); err != nil {
// 		return err
// 	}

// 	eventBus, err := setupEventBus(pubSub, logger)
// 	if err != nil {
// 		return err
// 	}

// 	eventProcessor, err := setupEventProcessor(router, pubSub, logger)
// 	if err != nil {
// 		return err
// 	}

// 	return router.Run(context.Background())
// }

// func messageMarshaler() cqrs.ProtoMarshaler {
// 	return cqrs.ProtoMarshaler{
// 		GenerateName: cqrs.NamedStruct(cqrs.StructName),
// 		// GenerateName: cqrs.StructName,
// 	}
// }

// func setupCommandBus(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) (*cqrs.CommandBus, error) {
// 	return cqrs.NewCommandBusWithConfig(pubSub, cqrs.CommandBusConfig{
// 		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
// 			return generateCommandsTopic(params.CommandName), nil
// 		},
// 		OnSend:    createCommandBusOnSendHandler(logger),
// 		Marshaler: messageMarshaler(),
// 		Logger:    logger,
// 	})
// }

// func createCommandBusOnSendHandler(logger watermill.LoggerAdapter) func(params cqrs.CommandBusOnSendParams) error {
// 	return func(params cqrs.CommandBusOnSendParams) error {
// 		logger.Info("Sending command", watermill.LogFields{
// 			"commandName": params.CommandName,
// 		})
// 		params.Message.Metadata.Set("sentAt", time.Now().String())
// 		return nil
// 	}
// }

// func setupRouter(logger watermill.LoggerAdapter) (*message.Router, error) {
// 	router, err := message.NewRouter(message.RouterConfig{}, logger)
// 	if err != nil {
// 		return nil, err
// 	}
// 	router.AddMiddleware(middleware.Recoverer)
// 	return router, nil
// }

// func setupCommandProcessor(router *message.Router, pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) error {
// 	config := createCommandProcessorConfig(pubSub, logger)
// 	_, err := cqrs.NewCommandProcessorWithConfig(router, config)
// 	return err
// }

// func createCommandProcessorConfig(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) cqrs.CommandProcessorConfig {
// 	return cqrs.CommandProcessorConfig{
// 		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
// 			return generateCommandsTopic(params.CommandName), nil
// 		},
// 		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
// 			return pubSub, nil
// 		},
// 		OnHandle:  createCommandProcessorOnHandleHandler(logger),
// 		Marshaler: messageMarshaler(),
// 		Logger:    logger,
// 	}
// }

// func createCommandProcessorOnHandleHandler(logger watermill.LoggerAdapter) func(params cqrs.CommandProcessorOnHandleParams) error {
// 	return func(params cqrs.CommandProcessorOnHandleParams) error {
// 		start := time.Now()
// 		err := params.Handler.Handle(params.Message.Context(), params.Command)
// 		logger.Info("Command handled", watermill.LogFields{
// 			"commandName": params.CommandName,
// 			"duration":    time.Since(start),
// 			"err":         err,
// 		})
// 		return err
// 	}
// }

// func setupEventBus(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) (*cqrs.EventBus, error) {
// 	eventBus, err := cqrs.NewEventBusWithConfig(pubSub, cqrs.EventBusConfig{
// 		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
// 			return generateEventsTopic(params.EventName), nil
// 		},
// 		OnPublish: createEventBusOnPublishHandler(logger),
// 		Marshaler: messageMarshaler(),
// 		Logger:    logger,
// 	})
// 	return eventBus, err
// }

// func createEventBusOnPublishHandler(logger watermill.LoggerAdapter) func(params cqrs.OnEventSendParams) error {
// 	return func(params cqrs.OnEventSendParams) error {
// 		logger.Info("Publishing event", watermill.LogFields{
// 			"eventName": params.EventName,
// 		})
// 		params.Message.Metadata.Set("publishedAt", time.Now().String())
// 		return nil
// 	}
// }

// func setupEventProcessor(router *message.Router, pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) (*cqrs.EventProcessor, error) {
// 	config := createEventProcessorConfig(pubSub, logger)
// 	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, config)
// 	return eventProcessor, err
// }

// func createEventProcessorConfig(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) cqrs.EventProcessorConfig {
// 	return cqrs.EventProcessorConfig{
// 		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
// 			return generateEventsTopic(params.EventName), nil
// 		},
// 		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
// 			return pubSub, nil
// 		},
// 		OnHandle:  createEventProcessorOnHandleHandler(logger),
// 		Marshaler: messageMarshaler(),
// 		Logger:    logger,
// 	}
// }

// func createEventProcessorOnHandleHandler(logger watermill.LoggerAdapter) func(params cqrs.EventProcessorOnHandleParams) error {
// 	return func(params cqrs.EventProcessorOnHandleParams) error {
// 		start := time.Now()
// 		err := params.Handler.Handle(params.Message.Context(), params.Event)
// 		logger.Info("Event handled", watermill.LogFields{
// 			"eventName": params.EventName,
// 			"duration":  time.Since(start),
// 			"err":       err,
// 		})
// 		return err
// 	}
// }

// func generateEventsTopic(eventName string) string {
// 	return "events:" + eventName
// }

// func generateCommandsTopic(commandName string) string {
// 	return "commands:" + commandName
// }
