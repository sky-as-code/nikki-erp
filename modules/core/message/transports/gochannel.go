package transports

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type GoChannelTransport struct {
	Transport *MessageTransport `name:"go-channel"`
	dig.Out
}

func NewGoChannleTransport(logger logging.LoggerService) GoChannelTransport {
	pubsub := goChannelPubSub(logger)
	return GoChannelTransport{
		Transport: &MessageTransport{
			Subscriber: pubsub,
			Publisher:  pubsub,
		},
	}
}

func goChannelPubSub(logger logging.LoggerService) *gochannel.GoChannel {
	slogger := logger.InnerLogger().(*slog.Logger)

	watermill.NewSlogLoggerWithLevelMapping(slogger, map[slog.Level]slog.Level{
		// Watermill does not have a trace level, so we map it to warn,
		// so that we will call watermillLogger().Trace() to print warnings.
		watermill.LevelTrace: slog.LevelWarn,
	})

	return gochannel.NewGoChannel(gochannel.Config{}, watermill.NewSlogLogger(slogger))
}
