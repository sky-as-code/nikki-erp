package pubsub

import (
	"github.com/redis/go-redis/v9"
	"github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)


type RedisPubSub struct {
	logger logging.LoggerService
	redisClient *redis.Client
}

func NewRedisPubSub(logger logging.LoggerService, redisClient *redis.Client) *RedisPubSub {
	return &RedisPubSub{
		logger:      logger,
		redisClient: redisClient,
	}
}

func (this *RedisPubSub) Publish(ctx context.Context, topic string, message any) error {
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