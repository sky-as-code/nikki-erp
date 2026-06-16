package pubsub

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type RedisPubSub struct {
	logger      logging.LoggerService
	redisClient *redis.Client
}

func NewRedisPubSub(logger logging.LoggerService, cfg config.ConfigService) PubSubOut {
	r := &RedisPubSub{
		logger: logger,
		redisClient: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.GetStr(c.PubSubRedisHost), cfg.GetStr(c.PubSubRedisPort)),
			Password: cfg.GetStr(c.PubSubRedisPassword),
			DB:       cfg.GetInt(c.PubSubRedisDB),
		}),
	}

	return PubSubOut{
		PubSub:    r,
		Publisher: r,
		Subcriber: r,
	}
}

func (this *RedisPubSub) Publish(ctx context.Context, topic string, message []byte) error {
	this.logger.Debug("Publishing message to topic", logging.Attr{"topic": topic, "message": message})
	return this.redisClient.Publish(ctx, topic, message).Err()
}

func (this *RedisPubSub) Subscribe(ctx context.Context, topic string) (<-chan []byte, error) {
	pubSub := this.redisClient.Subscribe(ctx, topic)
	this.logger.Debug("Subscribing to topic", logging.Attr{"topic": topic})
	_, err := pubSub.Receive(ctx)
	if err != nil {
		this.logger.Error("Failed to subscribe to topic", err)
		return nil, err
	}

	out := make(chan []byte, 1000)

	go func() {
		defer close(out)
		defer pubSub.Unsubscribe(ctx, topic)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-pubSub.Channel():
				if !ok {
					return
				}

				out <- []byte(msg.Payload)
			}
		}
	}()

	return out, nil
}
