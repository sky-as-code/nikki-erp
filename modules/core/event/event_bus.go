package event

import (
	"context"
	stdErrors "errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const (
	MetaEventTopic          = "event_topic"
	DefaultEventTimeoutSecs = "30"
	MetaCorrelationId       = "correlation_id"
	MetaReplyTopic          = "reply_topic"
	MetaNoReply             = "no_reply"
)

type EventBusParams struct {
	dig.In

	Config config.ConfigService
	Logger logging.LoggerService
}

func NewRedisEventBus(params EventBusParams) (EventBus, error) {
	host := params.Config.GetStr(c.EventBusRedisHost)
	port := params.Config.GetStr(c.EventBusRedisPort)
	password := params.Config.GetStr(c.EventBusRedisPassword)
	db := params.Config.GetInt(c.EventBusRedisDB)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client: redisClient,
		},
		watermill.NewSlogLogger(params.Logger.InnerLogger().(*slog.Logger)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Redis publisher")
	}

	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        redisClient,
			ConsumerGroup: "event_bus_consumer_group",
		},
		watermill.NewSlogLogger(params.Logger.InnerLogger().(*slog.Logger)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Redis subscriber")
	}

	maxTimeoutSec := params.Config.GetInt(c.EventRequestTimeoutSecs, DefaultEventTimeoutSecs)

	return &RedisEventBus{
		logger:     params.Logger,
		publisher:  publisher,
		subscriber: subscriber,
		maxTimeout: time.Duration(maxTimeoutSec) * time.Second,
		marshaler:  cqrs.JSONMarshaler{GenerateName: cqrs.NamedStruct(cqrs.StructName)},
	}, nil
}

type RedisEventBus struct {
	logger        logging.LoggerService
	publisher     message.Publisher
	subscriber    message.Subscriber
	subscriptions sync.Map
	maxTimeout    time.Duration
	marshaler     cqrs.CommandEventMarshaler
}

func (bus *RedisEventBus) PublishRequest(ctx context.Context, request EventRequest) (err error) {
	defer func() {
		err = ft.RecoverPanicFailedTo(recover(), "publish event")
	}()

	// Marshal the event
	msg, err := bus.marshaler.Marshal(request.message.Payload)
	ft.PanicOnErr(err)

	// Set metadata
	msg.Metadata.Set(MetaEventTopic, request.eventTopic)
	msg.Metadata.Set(MetaCorrelationId, request.correlationId)
	msg.Metadata.Set(MetaReplyTopic, request.replyTopic)
	msg.Metadata.Set(MetaNoReply, "false")

	// Publish the event
	err = bus.publisher.Publish(request.eventTopic, msg)
	ft.PanicOnErr(err)

	return nil
}

func (bus *RedisEventBus) PublishRequestWaitReply(ctx context.Context, request EventRequest, DataReply any) (reply *Reply[any], err error) {
	ctx, cancelSubscription := context.WithCancel(ctx)

	defer func() {
		err = ft.RecoverPanicFailedTo(recover(), "publish event and wait reply")
	}()

	replyChan, errChan := bus.subscribeReply(ctx, request, DataReply, cancelSubscription)

	err = bus.PublishRequest(ctx, request)
	ft.PanicOnErr(err)

	select {
	case reply := <-replyChan:
		return reply, nil
	case err = <-errChan:
		return nil, err
	}
}

func (bus *RedisEventBus) subscribeReply(ctx context.Context, request EventRequest, result any, cancelSubscription context.CancelFunc) (<-chan *Reply[any], <-chan error) {
	replyChan := make(chan *Reply[any])
	errChan := make(chan error)

	msgChan, err := bus.subscriber.Subscribe(ctx, request.replyTopic)
	if err != nil {
		errChan <- err
		return replyChan, errChan
	}

	go func() {
		defer cancelSubscription()
		defer close(replyChan)
		defer close(errChan)

		for {
			select {
			case msg := <-msgChan:
				msg.Ack()
				reply := &Reply[any]{
					Result: result,
				}
				err = bus.marshaler.Unmarshal(msg, reply)
				if err == nil {
					replyChan <- reply
					return
				}
			case <-ctx.Done():
				errChan <- ctx.Err()
			case <-time.After(bus.maxTimeout):
				return
			}
		}

	}()

	return replyChan, errChan
}

func (bus *RedisEventBus) SubscribeRequest(ctx context.Context, request EventRequest, result any) (requestChan chan any, err error) {
	defer func() {
		err = ft.RecoverPanicFailedTo(recover(), "publish reply")
	}()

	if _, exists := bus.subscriptions.Load(request.eventTopic); exists {
		return nil, fmt.Errorf("already subscribed to topic: %s", request.eventTopic)
	}

	msgChan, err := bus.subscriber.Subscribe(ctx, request.eventTopic)
	ft.PanicOnErr(err)

	bus.subscriptions.Store(request.eventTopic, msgChan)

	requestChan = make(chan any, 10)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in message handler for topic %s: %v", request.eventTopic, r)
				bus.logger.Error("message handler panic", err)
			}
		}()

		for {
			select {
			case msg, ok := <-msgChan:
				if !ok {
					bus.subscriptions.Delete(request.eventTopic)
					close(requestChan)
					return
				}
				msg.Ack()
				err := bus.marshaler.Unmarshal(msg, result)
				if err != nil {
					return
				}

				requestChan <- result
			case <-ctx.Done():
				bus.subscriptions.Delete(request.eventTopic)
				close(requestChan)
				return
			}
		}
	}()

	return requestChan, nil
}

// PublishReply publishes a reply to the specified reply topic
func (bus *RedisEventBus) PublishReply(ctx context.Context, request EventRequest, reply *Reply[any]) (err error) {
	defer func() {
		err = ft.RecoverPanicFailedTo(recover(), "publish reply")
	}()

	// Marshal the reply
	msg, err := bus.marshaler.Marshal(reply)
	ft.PanicOnErr(err)

	// Set metadata
	msg.Metadata.Set(MetaCorrelationId, request.correlationId)
	msg.Metadata.Set(MetaReplyTopic, request.replyTopic)

	// Publish the reply
	err = bus.publisher.Publish(request.replyTopic, msg)
	ft.PanicOnErr(err)

	return nil
}

// Close closes the event bus and all its subscriptions
func (bus *RedisEventBus) Close() error {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in Close: %v", r)
			bus.logger.Error("close panic", err)
		}
	}()

	var errs []error

	// Close publisher
	if err := bus.publisher.Close(); err != nil {
		errs = append(errs, errors.Wrap(err, "failed to close publisher"))
	}

	// Close subscriber
	if err := bus.subscriber.Close(); err != nil {
		errs = append(errs, errors.Wrap(err, "failed to close subscriber"))
	}

	// Clear subscriptions
	bus.subscriptions.Range(func(key, value interface{}) bool {
		bus.subscriptions.Delete(key)
		return true
	})

	if len(errs) > 0 {
		return stdErrors.Join(errs...)
	}

	bus.logger.Info("event bus closed successfully", nil)
	return nil
}
