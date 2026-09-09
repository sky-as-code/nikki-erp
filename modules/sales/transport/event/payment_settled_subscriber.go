// Package eventtransport subscribes Sales to the internal event bus and dispatches each event to
// the handler registered for its type.
package eventtransport

import (
	"context"
	"time"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreEvent "github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	piconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"

	itEvent "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/event"
)

// handleTimeout bounds one event's handling. Applying a verdict reads a payment, writes it, and may
// settle a bill, so it is not instant — but a wedged handler must not hold the consumer loop
// forever, because the loop is sequential and every verdict behind it would stall.
const handleTimeout = 30 * time.Second

// NewPaymentSettledSubscriber wires the consumer.
//
// The topic is read from the publishing module's constant, not a local literal. This is the one
// place Sales imports paymentinvoice outside infra/external, and it is deliberate: two halves of a
// pub/sub pair that declare the topic separately drift apart silently, and the failure looks like
// payments simply never settling.
func NewPaymentSettledSubscriber(
	bus coreEvent.EventBus,
	cfg config.ConfigService,
	logger logging.LoggerService,
	handlers itEvent.PaymentSettledHandlerRegistry,
) *PaymentSettledSubscriber {
	return &PaymentSettledSubscriber{
		bus:    bus,
		logger: logger,
		topic: cfg.GetStr(
			piconstants.PaymentSettledEventTopic, piconstants.DefaultPaymentSettledEventTopic),
		handlers: handlers,
	}
}

type PaymentSettledSubscriber struct {
	bus      coreEvent.EventBus
	logger   logging.LoggerService
	topic    string
	handlers itEvent.PaymentSettledHandlerRegistry
}

// clonePaymentSettledEvent copies the event out of the shared decode buffer.
//
// SubscribeRequest decodes every message into ONE struct that it reuses, so a handler holding the
// pointer would see its fields change under it when the next message arrives. The metadata map is
// copied too, not just the struct: the shared struct's map header would otherwise still point at a
// map the next decode overwrites.
func clonePaymentSettledEvent(event *itEvent.PaymentSettledEvent) itEvent.PaymentSettledEvent {
	out := *event

	if event.Metadata != nil {
		metadata := make(map[string]any, len(event.Metadata))
		for key, value := range event.Metadata {
			metadata[key] = value
		}
		out.Metadata = metadata
	}
	return out
}

func (this *PaymentSettledSubscriber) Register(ctx context.Context) error {
	shared := &itEvent.PaymentSettledEvent{}
	req := coreEvent.NewEventRequest("", this.topic, "", nil)

	eventChan, err := this.bus.SubscribeRequest(ctx, *req, shared)
	if err != nil {
		return err
	}

	this.logger.Info("sales payment settled subscriber registered",
		logging.Attr{"topic": this.topic})

	go this.consume(ctx, shared, eventChan)

	return nil
}

func (this *PaymentSettledSubscriber) consume(
	rootCtx context.Context, shared *itEvent.PaymentSettledEvent, eventChan chan any,
) {
	for range eventChan {
		event := clonePaymentSettledEvent(shared)
		bg, cancel := context.WithTimeout(rootCtx, handleTimeout)
		this.handle(corectx.NewRequestContext(bg), event)
		cancel()
	}
}

// handle dispatches one verdict, and swallows a handler failure by design.
//
// The event describes a settlement that has ALREADY been recorded upstream. Returning the error
// would only stop the consumer loop, taking every later verdict with it; the payment is still
// reachable by the reconciliation sweep, which is exactly the case the sweep exists for.
func (this *PaymentSettledSubscriber) handle(
	ctx corectx.Context, event itEvent.PaymentSettledEvent,
) {
	handler, ok := this.handlers[event.Type]
	if !ok {
		// A verdict this build does not know. Logged rather than guessed at: forcing an unknown
		// outcome into one of the four would move a payment on a assumption.
		this.logger.Warn("no sales handler for payment settled type",
			logging.Attr{"type": string(event.Type), "order_id": event.OrderId})
		return
	}

	if err := handler.Handle(ctx, event); err != nil {
		this.logger.Warn("sales payment settled handler failed",
			logging.Attr{
				"error":    err.Error(),
				"type":     string(event.Type),
				"order_id": event.OrderId,
			})
	}
}

// InitEventSubscribers registers the subscriber and starts it consuming.
//
// It must run after InitHandlers, which registers the registry this resolves.
func InitEventSubscribers() error {
	if err := deps.Register(NewPaymentSettledSubscriber); err != nil {
		return err
	}
	return deps.Invoke(func(subscriber *PaymentSettledSubscriber) error {
		return subscriber.Register(context.Background())
	})
}
