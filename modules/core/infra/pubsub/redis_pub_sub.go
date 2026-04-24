package pubsub

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type RedisPubSub struct {
	logger      logging.LoggerService
	redisClient *redis.Client
}

func NewRedisPubSub(logger logging.LoggerService, redisClient *redis.Client) (Publisher, Subcriber) {
	r := &RedisPubSub{
		logger:      logger,
		redisClient: redisClient,
	}

	return r, r
}

func (this *RedisPubSub) Publish(ctx context.Context, channel string, message any) error {
	this.logger.Debug("Publishing message to topic", logging.Attr{"topic": channel, "message": message})
	return this.redisClient.Publish(ctx, channel, message).Err()
}

func (this *RedisPubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
	pubSub := this.redisClient.Subscribe(ctx, channel)
	this.logger.Debug("Subscribing to topic", logging.Attr{"topic": channel})
	_, err := pubSub.Receive(ctx)
	if err != nil {
		this.logger.Error("Failed to subscribe to topic", err)
		return nil, err
	}

	out := make(chan []byte)

	go func() {
		defer close(out)
		for msg := range pubSub.Channel() {
			out <- []byte(msg.Payload)
		}
	}()

	return out, nil
}

func (this *RedisPubSub) Close() error {
	return this.redisClient.Close()
}
