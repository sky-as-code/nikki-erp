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
	"go.bryk.io/pkg/ulid"
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

// Verify RedisEventBus implements EventBus interface

func (bus *RedisEventBus) PublishNoReply(ctx context.Context, packet *EventPacket) (err error) {
	defer func() {
		err = ft.RecoverPanic(recover(), "failed to publish event")
	}()

	err = bus.publisher.Publish(packet.eventTopic, packet.message)
	ft.PanicOnErr(err)

	return nil
}

func (bus *RedisEventBus) PublishWaitReply(ctx context.Context, packet *EventPacket, result any) (err error) {
	ctx, cancelSubscription := context.WithCancel(ctx)

	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to publish event of type %s", err); e != nil {
			err = e
			cancelSubscription()
		}
	}()

	replyChan, errChan := bus.subscribeReply(ctx, packet, result, cancelSubscription)
	ft.PanicOnErr(err)

	err = bus.publisher.Publish(packet.eventTopic, packet.message)
	ft.PanicOnErr(err)

	select {
	case reply := <-replyChan:
		if reply.Error != nil {
			return errors.New(*reply.Error)
		}
		return nil
	case err := <-errChan:
		return err
	}
}

func (bus *RedisEventBus) subscribeReply(ctx context.Context, packet *EventPacket, result any, cancelSubscription context.CancelFunc) (<-chan *Reply[any], <-chan error) {
	replyChan := make(chan *Reply[any])
	errChan := make(chan error)

	handleErr := func() {
		if r := recover(); r != nil {
			err := errors.Wrap(r.(error), fmt.Sprintf("failed to subscribe for reply from topic %s", packet.replyTopic))
			errChan <- err
			close(errChan)
			close(replyChan)
		}
	}

	defer handleErr()

	msgChan, err := bus.subscriber.Subscribe(ctx, packet.replyTopic)
	if err != nil {
		errChan <- err
		return replyChan, errChan
	}

	go func() {
		defer cancelSubscription()
		defer handleErr()

		select {
		case msg := <-msgChan:
			msg.Ack()
			reply := &Reply[any]{
				Result: result,
			}
			err = bus.marshaler.Unmarshal(msg, reply)
			if err == nil {
				replyChan <- reply
				close(replyChan)
				close(errChan)
				return
			}
		case <-ctx.Done():
			err = ctx.Err()
		case <-time.After(bus.maxTimeout):
			err = errors.Errorf("timeout for event %s (%s)", packet.correlationId, packet.eventTopic)
		}

		// If we reach here, it means we have an error,
		// close error channel first to follow the failure path
		errChan <- err
		close(errChan)
		close(replyChan)
	}()

	return replyChan, errChan
}

// Subscribe to events on a topic with an event handler
func (bus *RedisEventBus) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in Subscribe for topic %s: %v", topic, r)
			bus.logger.Error("subscribe panic", err)
		}
	}()

	// Check if already subscribed to this topic
	if _, exists := bus.subscriptions.Load(topic); exists {
		return errors.Errorf("already subscribed to topic: %s", topic)
	}

	msgChan, err := bus.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to subscribe to topic: %s", topic))
	}

	// Store subscription
	bus.subscriptions.Store(topic, true)

	go func() {
		defer func() {
			bus.subscriptions.Delete(topic)
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in message handler for topic %s: %v", topic, r)
				bus.logger.Error("message handler panic", err)
			}
		}()

		for {
			select {
			case msg := <-msgChan:
				if msg == nil {
					bus.logger.Info("subscription closed for topic", topic)
					return
				}

				// Extract event packet information from message metadata
				correlationId := msg.Metadata.Get(MetaCorrelationId)
				eventTopic := msg.Metadata.Get(MetaEventTopic)
				replyTopic := msg.Metadata.Get(MetaReplyTopic)

				packet := &EventPacket{
					correlationId: correlationId,
					eventTopic:    eventTopic,
					replyTopic:    replyTopic,
					message:       msg,
				}

				// Handle the event
				if err := handler.Handle(ctx, packet); err != nil {
					bus.logger.Error("error handling event on topic "+topic, err)

					// Send error reply if reply topic is specified
					if replyTopic != "" && replyTopic != MetaNoReply {
						errorMsg := errors.Wrap(err, "error handling event").Error()
						errorReply := &Reply[any]{
							Result: nil, // No specific result to return
							Error:  &errorMsg,
						}
						if replyErr := bus.PublishReply(ctx, replyTopic, errorReply, correlationId); replyErr != nil {
							bus.logger.Error("failed to send error reply", replyErr)
						}
					}
				} else {
					// Send success reply if reply topic is specified
					if replyTopic != "" && replyTopic != MetaNoReply {
						successReply := &Reply[any]{
							Result: nil, // No specific result to return
							Error:  nil, // No error
						}
						if replyErr := bus.PublishReply(ctx, replyTopic, successReply, correlationId); replyErr != nil {
							bus.logger.Error("failed to send success reply", replyErr)
						}
					}
				}

				msg.Ack()

			case <-ctx.Done():
				bus.logger.Info("context cancelled for topic", topic)
				return
			}
		}
	}()

	return nil
}

// SubscribeReply subscribes to reply messages on a specific topic
func (bus *RedisEventBus) SubscribeReply(ctx context.Context, replyTopic string, handler ReplyHandler) error {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in SubscribeReply for topic %s: %v", replyTopic, r)
			bus.logger.Error("subscribe reply panic", err)
		}
	}()

	// Check if already subscribed to this reply topic
	replyKey := fmt.Sprintf("reply_%s", replyTopic)
	if _, exists := bus.subscriptions.Load(replyKey); exists {
		return errors.Errorf("already subscribed to reply topic: %s", replyTopic)
	}

	msgChan, err := bus.subscriber.Subscribe(ctx, replyTopic)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to subscribe to reply topic: %s", replyTopic))
	}

	// Store subscription
	bus.subscriptions.Store(replyKey, true)

	go func() {
		defer func() {
			bus.subscriptions.Delete(replyKey)
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in reply handler for topic %s: %v", replyTopic, r)
				bus.logger.Error("reply handler panic", err)
			}
		}()

		for {
			select {
			case msg := <-msgChan:
				if msg == nil {
					bus.logger.Info("reply subscription closed for topic", replyTopic)
					return
				}

				correlationId := msg.Metadata.Get(MetaCorrelationId)

				// Unmarshal the reply
				var reply Reply[any]
				if err := bus.marshaler.Unmarshal(msg, &reply); err != nil {
					bus.logger.Error("failed to unmarshal reply on topic "+replyTopic, err)
					msg.Ack()
					continue
				}

				replyPacket := &ReplyPacket[any]{
					correlationId: correlationId,
					reply:         reply,
				}

				// Handle the reply
				if err := handler.Handle(ctx, replyPacket); err != nil {
					bus.logger.Error("error handling reply on topic "+replyTopic, err)
				}

				msg.Ack()

			case <-ctx.Done():
				bus.logger.Info("context cancelled for reply topic", replyTopic)
				return
			}
		}
	}()

	return nil
}

// PublishReply publishes a reply message to the specified reply topic
func (bus *RedisEventBus) PublishReply(ctx context.Context, replyTopic string, reply any, correlationId string) error {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in PublishReply for topic %s: %v", replyTopic, r)
			bus.logger.Error("publish reply panic", err)
		}
	}()

	// Marshal the reply
	msg, err := bus.marshaler.Marshal(reply)
	if err != nil {
		return errors.Wrap(err, "failed to marshal reply")
	}

	// Set metadata
	msg.Metadata.Set(MetaCorrelationId, correlationId)
	msg.Metadata.Set(MetaReplyTopic, replyTopic)

	// Generate unique message ID
	msgId, err := ulid.New()
	if err != nil {
		return errors.Wrap(err, "failed to generate message ID")
	}
	msg.UUID = msgId.String()

	// Publish the reply
	if err := bus.publisher.Publish(replyTopic, msg); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to publish reply to topic: %s", replyTopic))
	}

	bus.logger.Debug("published reply to topic", fmt.Sprintf("%s with correlation ID %s", replyTopic, correlationId))
	return nil
}

// PublishEvent is a helper method to publish events without waiting for a reply
func (bus *RedisEventBus) PublishEvent(ctx context.Context, topic string, eventData any) error {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in PublishEvent for topic %s: %v", topic, r)
			bus.logger.Error("publish event panic", err)
		}
	}()

	// Generate unique event ID
	eventId, err := ulid.New()
	if err != nil {
		return errors.Wrap(err, "failed to generate event ID")
	}

	// Marshal the event
	msg, err := bus.marshaler.Marshal(eventData)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
	}

	// Set metadata
	msg.Metadata.Set(MetaEventTopic, topic)
	msg.Metadata.Set(MetaCorrelationId, eventId.String())
	msg.Metadata.Set(MetaReplyTopic, MetaNoReply)
	msg.UUID = eventId.String()
	// Create event packet
	packet := &EventPacket{
		correlationId: eventId.String(),
		eventTopic:    topic,
		replyTopic:    MetaNoReply,
		message:       msg,
	}

	// Publish the event
	return bus.PublishNoReply(ctx, packet)
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
