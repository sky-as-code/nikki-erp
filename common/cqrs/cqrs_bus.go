package cqrs

import (
	"context"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	"github.com/sky-as-code/nikki-erp/common/logging"
	cqrsUtil "github.com/sky-as-code/nikki-erp/common/util/cqrs"
)

func NewWaterMillCqrsBus() (*WaterMillCqrsBus, error) {
	logger := watermillLogger()
	pubSub := goChannelPubSub(logger)
	config := createQueryBusConfig(pubSub, logger)
	queryBus, err := cqrsUtil.NewQueryBusWithConfig(config)
	if err != nil {
		return nil, err
	}

	cmdBus, err := createCommandBus(pubSub, logger)
	if err != nil {
		return nil, err
	}

	return &WaterMillCqrsBus{
		queryBus: queryBus,
		cmdBus:   cmdBus,
	}, nil
}

func createQueryBusConfig(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) cqrsUtil.QueryBusConfig {
	return cqrsUtil.QueryBusConfig{
		Publisher:  pubSub,
		Subscriber: pubSub,
		GenerateQueryTopic: func(params cqrsUtil.QueryBusGeneratePublishTopicParams) (string, error) {
			return genQueryTopic(params.QueryName), nil
		},
		Marshaler: messageMarshaler(),
		Logger:    logger,
	}
}

func createCommandBus(pubSub *gochannel.GoChannel, logger watermill.LoggerAdapter) (*cqrs.CommandBus, error) {
	return cqrs.NewCommandBusWithConfig(pubSub, cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return genCommandTopic(params.CommandName), nil
		},
		OnSend:    createCommandBusOnSendHandler(logger),
		Marshaler: messageMarshaler(),
		Logger:    logger,
	})
}

func createCommandBusOnSendHandler(logger watermill.LoggerAdapter) func(params cqrs.CommandBusOnSendParams) error {
	return func(params cqrs.CommandBusOnSendParams) error {
		logger.Debug("Sending command", watermill.LogFields{
			"commandName": params.CommandName,
		})
		params.Message.Metadata.Set("sentAt", time.Now().String())
		return nil
	}
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
	cmdBus   *cqrs.CommandBus
	queryBus *cqrsUtil.QueryBus
}

func (this *WaterMillCqrsBus) ExecCommand(ctx context.Context, command Namer) error {
	return this.cmdBus.Send(ctx, command)
}

func (this *WaterMillCqrsBus) ExecQuery(ctx context.Context, query Namer) (_ <-chan cqrsUtil.Reply, err error) {
	return this.queryBus.Send(ctx, query)
}

func genCommandTopic(commandName string) string {
	return "command:" + commandName
}

func genQueryTopic(commandName string) string {
	return "query:" + commandName
}
