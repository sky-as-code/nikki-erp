package cqrs

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	"github.com/sky-as-code/nikki-erp/common/logging"
	cqrsUtil "github.com/sky-as-code/nikki-erp/common/util/cqrs"
)

func NewWaterMillCqrsBus() (*WaterMillCqrsBus, error) {
	logger := watermillLogger()
	pubSub := goChannelPubSub(logger)

	queryBus, err := createQueryBus(pubSub, logger)
	if err != nil {
		return nil, err
	}

	cmdBus, err := createCommandBus(pubSub, logger)
	if err != nil {
		return nil, err
	}

	queryProcessor, err := createQueryProcessor(pubSub, logger)
	if err != nil {
		return nil, err
	}

	cmdProcessor, err := createCommandProcessor(pubSub, logger)
	if err != nil {
		return nil, err
	}

	return &WaterMillCqrsBus{
		queryBus:       queryBus,
		cmdBus:         cmdBus,
		queryProcessor: queryProcessor,
		cmdProcessor:   cmdProcessor,
	}, nil
}

func createQueryBus(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) (*cqrsUtil.QueryBus, error) {
	config := cqrsUtil.QueryBusConfig{
		Publisher:  pubSub,
		Subscriber: pubSub,
		GenerateQueryTopic: func(params cqrsUtil.QueryBusGeneratePublishTopicParams) (string, error) {
			return genQueryTopic(params.QueryName), nil
		},
		Marshaler: messageMarshaler(),
		Logger:    logger,
	}
	return cqrsUtil.NewQueryBusWithConfig(config)
}

func createQueryProcessor(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) (*cqrsUtil.QueryProcessor, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	config := cqrsUtil.QueryProcessorConfig{
		GenerateSubscribeTopic: func(params cqrsUtil.QueryProcessorGenerateSubscribeTopicParams) (string, error) {
			return genQueryTopic(params.QueryName), nil
		},
		SubscriberConstructor: func(params cqrsUtil.QueryProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return pubSub, nil
		},
		Publisher: pubSub,
		Marshaler: messageMarshaler(),
		Logger:    logger,
	}

	return cqrsUtil.NewQueryProcessorWithConfig(router, config)
}

func createCommandBus(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) (*cqrs.CommandBus, error) {
	return cqrs.NewCommandBusWithConfig(pubSub, cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return genCommandTopic(params.CommandName), nil
		},
		Marshaler: messageMarshaler(),
		Logger:    logger,
	})
}

func createCommandProcessor(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) (*cqrs.CommandProcessor, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	config := cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return genCommandTopic(params.CommandName), nil
		},
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return pubSub, nil
		},
		Marshaler: messageMarshaler(),
		Logger:    logger,
	}

	return cqrs.NewCommandProcessorWithConfig(router, config)
}

func watermillLogger() watermill.LoggerAdapter {
	slogger := logging.Logger().InnerLogger().(*slog.Logger)

	watermill.NewSlogLoggerWithLevelMapping(slogger, map[slog.Level]slog.Level{
		// Watermill does not have a trace level, so we map it to warn,
		// so that we will call watermillLogger().Trace() to print warnings.
		watermill.LevelTrace: slog.LevelWarn,
	})
	return watermill.NewSlogLogger(slogger)
}

func goChannelPubSub(logger watermill.LoggerAdapter) *gochannel.GoChannel {
	return gochannel.NewGoChannel(
		gochannel.Config{},
		logger,
	)
}

func messageMarshaler() cqrs.ProtoMarshaler {
	return cqrs.ProtoMarshaler{
		GenerateName: cqrs.NamedStruct(cqrs.StructName),
	}
}

type Namer interface {
	Name() string
}

type WaterMillCqrsBus struct {
	cmdBus         *cqrs.CommandBus
	cmdProcessor   *cqrs.CommandProcessor
	queryBus       *cqrsUtil.QueryBus
	queryProcessor *cqrsUtil.QueryProcessor
}

func (this *WaterMillCqrsBus) ExecCommand(ctx context.Context, command Namer) error {
	return this.cmdBus.Send(ctx, command)
}

func (this *WaterMillCqrsBus) AddCommandHandlers(commandHandlers ...cqrs.CommandHandler) error {
	return this.cmdProcessor.AddHandlers(commandHandlers...)
}

func (this *WaterMillCqrsBus) ExecQuery(ctx context.Context, query Namer) (_ <-chan cqrsUtil.Reply, err error) {
	return this.queryBus.Send(ctx, query)
}

func (this *WaterMillCqrsBus) AddQueryHandlers(queryHandlers ...cqrsUtil.QueryHandler) error {
	return this.queryProcessor.AddHandlers(queryHandlers...)
}

func genCommandTopic(commandName string) string {
	return "command:" + commandName
}

func genQueryTopic(commandName string) string {
	return "query:" + commandName
}
