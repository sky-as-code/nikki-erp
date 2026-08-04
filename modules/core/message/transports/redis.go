package transports

import (
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type RedisWatermillTransport struct {
	Transport *MessageTransport `name:"redis"`
	dig.Out
}

func NewRedisTransport(config config.ConfigService, logger logging.LoggerService) (RedisWatermillTransport, error) {
	host := config.GetStr(c.EventBusRedisHost)
	port := config.GetStr(c.EventBusRedisPort)
	password := config.GetStr(c.EventBusRedisPassword)
	db := config.GetInt(c.EventBusRedisDB)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: redisClient,
		},
		watermill.NewSlogLogger(logger.InnerLogger().(*slog.Logger)),
	)
	if err != nil {
		return RedisWatermillTransport{}, errors.Wrap(err, "failed to create Redis publisher")
	}

	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "event_bus_consumer_group",
		},
		watermill.NewSlogLogger(logger.InnerLogger().(*slog.Logger)),
	)
	if err != nil {
		return RedisWatermillTransport{}, errors.Wrap(err, "failed to create Redis subscriber")
	}

	return RedisWatermillTransport{
		Transport: &MessageTransport{
			Subscriber: subscriber,
			Publisher:  publisher,
		},
	}, nil
}
