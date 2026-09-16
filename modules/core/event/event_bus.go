package event

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/message/transports"
)

const (
	MetaEventTopic          = "event_topic"
	DefaultEventTimeoutSecs = "30"
	MetaCorrelationId       = "correlation_id"
	MetaReplyTopic          = "reply_topic"
	MetaNoReply             = "no_reply"

	// MetaDomainConstraints carries the publisher's domain constraints to the subscriber, which
	// consumes on a goroutine that has no request behind it. The name and the JSON encoding match
	// what the CQRS bus already puts on its own messages (cqrs.MetaDomainConstraints), so the two
	// buses scope a handler the same way; see EventDelivery for why this belongs on the message
	// rather than inside each event's payload.
	MetaDomainConstraints = "domain_constraints"
)

type EventBusParams struct {
	dig.In

	Config config.ConfigService
	Logger logging.LoggerService

	Transport *transports.MessageTransport `name:"mqtt"`
}

func NewEventBus(params EventBusParams) (EventBus, error) {
	maxTimeoutSec := params.Config.GetInt(c.EventRequestTimeoutSecs, DefaultEventTimeoutSecs)

	return &EventBusImpl{
		logger:     params.Logger,
		publisher:  params.Transport.Publisher,
		subscriber: params.Transport.Subscriber,
		maxTimeout: time.Duration(maxTimeoutSec) * time.Second,
		marshaler:  cqrs.JSONMarshaler{GenerateName: cqrs.NamedStruct(cqrs.StructName)},
	}, nil
}

type EventBusImpl struct {
	logger        logging.LoggerService
	publisher     message.Publisher
	subscriber    message.Subscriber
	subscriptions sync.Map
	maxTimeout    time.Duration
	marshaler     cqrs.CommandEventMarshaler
}

func (bus *EventBusImpl) PublishRequest(ctx context.Context, request EventRequest) (err error) {
	defer func() {
		err = ft.RecoverPanicFailedTo(recover(), "publish event")
	}()

	var msg *message.Message
	if request.message != nil && len(request.message.Payload) > 0 {
		// Payload is already JSON bytes; do not json.Marshal([]byte) again (would base64-encode).
		msg = message.NewMessage(watermill.NewUUID(), request.message.Payload)
	} else {
		var err error
		msg, err = bus.marshaler.Marshal(request.message.Payload)
		ft.PanicOnErr(err)
	}

	// Set metadata
	msg.Metadata.Set(MetaEventTopic, request.eventTopic)
	msg.Metadata.Set(MetaCorrelationId, request.correlationId)
	msg.Metadata.Set(MetaReplyTopic, request.replyTopic)
	msg.Metadata.Set(MetaNoReply, "false")
	setDomainConstraints(ctx, msg)

	// Publish the event
	err = bus.publisher.Publish(request.eventTopic, msg)
	ft.PanicOnErr(err)

	return nil
}

func (bus *EventBusImpl) PublishRequestWaitReply(ctx context.Context, request EventRequest, DataReply any) (reply *Reply[any], err error) {
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

func (bus *EventBusImpl) subscribeReply(ctx context.Context, request EventRequest, result any, cancelSubscription context.CancelFunc) (<-chan *Reply[any], <-chan error) {
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

// SubscribeEvent delivers one EventDelivery per message on the topic.
//
// prototype is a TYPE, not a buffer: a fresh value of its type is allocated for every message, so a
// subscriber may hold on to what it receives and hand it to a handler. The method this replaced
// decoded every message into that one value and handed the same pointer out repeatedly, which is
// why each of its callers had to copy the struct — and its slices and maps — before using it.
func (bus *EventBusImpl) SubscribeEvent(
	ctx context.Context, request EventRequest, prototype any,
) (deliveryChan chan EventDelivery, err error) {
	// Only a panic is converted here. Assigning unconditionally, as the rest of this file does,
	// would overwrite the "already subscribed" error below with nil and return a nil channel to a
	// caller that was told nothing went wrong.
	defer func() {
		if recovered := recover(); recovered != nil {
			err = ft.RecoverPanicFailedTo(recovered, "subscribe to event topic")
		}
	}()

	if _, exists := bus.subscriptions.Load(request.eventTopic); exists {
		return nil, fmt.Errorf("already subscribed to topic: %s", request.eventTopic)
	}

	msgChan, err := bus.subscriber.Subscribe(ctx, request.eventTopic)
	ft.PanicOnErr(err)

	bus.subscriptions.Store(request.eventTopic, msgChan)

	deliveryChan = make(chan EventDelivery, 10)

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
					close(deliveryChan)
					return
				}
				msg.Ack()

				payload := newPayloadOf(prototype)
				if err := bus.marshaler.Unmarshal(msg, payload); err != nil {
					bus.logger.Error("event bus unmarshal failed",
						fmt.Errorf("topic %s: %w", request.eventTopic, err))
					continue
				}

				deliveryChan <- EventDelivery{
					Payload:     payload,
					Constraints: bus.domainConstraintsOf(msg, request.eventTopic),
				}
			case <-ctx.Done():
				bus.subscriptions.Delete(request.eventTopic)
				close(deliveryChan)
				return
			}
		}
	}()

	return deliveryChan, nil
}

// domainConstraintsOf reads the publisher's scope back off a message.
//
// Unreadable constraints are logged and dropped rather than failing the message: the event itself
// decoded fine, and a handler that runs unscoped fails loudly on its first query, whereas a message
// silently discarded here would look like an event that was never published.
func (bus *EventBusImpl) domainConstraintsOf(msg *message.Message, topic string) dmodel.DynamicFields {
	encoded := msg.Metadata.Get(MetaDomainConstraints)
	if encoded == "" {
		return nil
	}

	constraints := dmodel.DynamicFields{}
	if err := json.Unmarshal([]byte(encoded), &constraints); err != nil {
		bus.logger.Error("event bus domain constraints decode failed",
			fmt.Errorf("topic %s: %w", topic, err))
		return nil
	}
	return constraints
}

// setDomainConstraints stamps the publisher's scope onto an outgoing message.
//
// ctx is a plain context.Context because that is what the bus takes, but the constraints live in
// the inner context's values either way, so this reads them from a corectx.Context and from the
// detached background context a publisher derived from one.
func setDomainConstraints(ctx context.Context, msg *message.Message) {
	constraints, ok := ctx.Value(corectx.CtxKeyDomainConstraints).(dmodel.DynamicFields)
	if !ok || len(constraints) == 0 {
		return
	}

	encoded, err := json.Marshal(constraints)
	if err != nil {
		// Nothing useful to do with this: the constraints are a map of scalars, so a failure here
		// means a caller put something exotic in them. The subscriber is left unscoped and says so
		// on its first query, which is louder than failing the publish of an event describing a
		// write that already committed.
		return
	}
	msg.Metadata.Set(MetaDomainConstraints, string(encoded))
}

// newPayloadOf allocates a value of the prototype's type for one message to decode into.
func newPayloadOf(prototype any) any {
	protoType := reflect.TypeOf(prototype)
	if protoType == nil || protoType.Kind() != reflect.Ptr {
		// Not a pointer, so there is nothing to allocate into and nothing the caller could read back
		// either. Handed straight to the unmarshaler, which reports it far better than a panic here.
		return prototype
	}
	return reflect.New(protoType.Elem()).Interface()
}

// PublishReply publishes a reply to the specified reply topic
func (bus *EventBusImpl) PublishReply(ctx context.Context, request EventRequest, reply *Reply[any]) (err error) {
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
func (bus *EventBusImpl) Close() error {
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
