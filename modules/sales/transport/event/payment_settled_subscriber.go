// Package eventtransport subscribes Sales to the internal event bus and dispatches each event to
// the handler registered for its type.
package eventtransport

import (
	"context"
	"fmt"
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

func (this *PaymentSettledSubscriber) Register(ctx context.Context) error {
	req := coreEvent.NewEventRequest("", this.topic, "", nil)

	deliveries, err := this.bus.SubscribeEvent(ctx, *req, &itEvent.PaymentSettledEvent{})
	if err != nil {
		return err
	}

	this.logger.Info("sales payment settled subscriber registered",
		logging.Attr{"topic": this.topic})

	go this.consume(ctx, deliveries)

	return nil
}

func (this *PaymentSettledSubscriber) consume(
	rootCtx context.Context, deliveries chan coreEvent.EventDelivery,
) {
	for delivery := range deliveries {
		event, ok := delivery.Payload.(*itEvent.PaymentSettledEvent)
		if !ok {
			// The bus decoded into the prototype this subscriber gave it, so this cannot happen
			// without the bus having changed underneath. Logged rather than asserted, because a
			// panic here would stop every later verdict.
			this.logger.Warn("sales payment settled event had an unexpected payload type", nil)
			continue
		}

		bg, cancel := context.WithTimeout(rootCtx, handleTimeout)
		this.dispatch(corectx.NewRequestContext(bg), delivery, *event)
		cancel()
	}
}

// dispatch re-establishes the publisher's scope before handling, and contains a panic that would
// otherwise take the whole process down with it.
//
// The context this consumer builds descends from context.Background(), set once when Register ran
// at boot, so it names no tenant and no organization. ScopeContext puts back what the publishing
// request was scoped by; it does nothing on a single-tenant build, where there was nothing to carry.
//
// The recover is a backstop, not the fix: a handler that still panics for some other reason must
// not stop every later verdict behind it in this sequential consumer loop.
func (this *PaymentSettledSubscriber) dispatch(
	ctx corectx.Context, delivery coreEvent.EventDelivery, event itEvent.PaymentSettledEvent,
) {
	defer func() {
		if recovered := recover(); recovered != nil {
			this.logger.Error("sales payment settled handler panicked",
				fmt.Errorf("%v", recovered))
		}
	}()

	delivery.ScopeContext(ctx)

	this.handle(ctx, event)
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
