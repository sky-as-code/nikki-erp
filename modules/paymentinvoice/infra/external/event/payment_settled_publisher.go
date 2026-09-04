// Package event publishes this module's announcements on the internal Watermill bus.
//
// Nothing here reaches a gateway or an ordering system over the network. It announces "this order
// reached a verdict" so that other modules in the same build can act on it.
package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreEvent "github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	itEvent "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/event"
)

// publishAsyncTimeout bounds a fire-and-forget publish. Generous, because the goroutine has been
// detached from the request and nothing is waiting on it — but not unbounded, or a wedged broker
// would leak one goroutine per settled order.
const publishAsyncTimeout = time.Minute

type PaymentSettledEventPublisherImpl struct {
	bus    coreEvent.EventBus
	logger logging.LoggerService
	topic  string
}

// NewPaymentSettledEventPublisher wires the announcement onto the internal bus.
//
// The default topic comes from constants rather than a local literal, so the publisher and every
// subscriber read one value and cannot drift apart over where they meet.
func NewPaymentSettledEventPublisher(
	bus coreEvent.EventBus, cfg config.ConfigService, logger logging.LoggerService,
) itEvent.PaymentSettledEventPublisher {
	return &PaymentSettledEventPublisherImpl{
		bus:    bus,
		logger: logger,
		topic: cfg.GetStr(
			constants.PaymentSettledEventTopic, constants.DefaultPaymentSettledEventTopic),
	}
}

// PublishAsync announces a verdict without blocking the caller.
//
// It is called AFTER the order was settled in its own transaction, so a broker that is slow or down
// must not fail the work that already succeeded — the worst case is a subscriber that learns late,
// which its own reconciliation covers.
//
// context.WithoutCancel is what makes that work: the request context is cancelled the moment the
// response is written, so a goroutine inheriting it would be cancelled before it could publish.
// NewRequestContext keeps the values the request carried, the tenant above all, which is also
// copied onto the event itself for the subscriber's benefit.
func (this *PaymentSettledEventPublisherImpl) PublishAsync(
	ctx corectx.Context, event itEvent.PaymentSettledEvent,
) {
	if event.OrderId == "" {
		// An announcement naming no order tells a subscriber nothing it can act on.
		return
	}

	bgCtx := corectx.NewRequestContext(context.WithoutCancel(ctx.InnerContext()))

	go func() {
		defer func() {
			// A panic on this goroutine would take the process down with it, and the settlement that
			// spawned it has already been committed. It is contained and logged instead.
			if recovered := recover(); recovered != nil {
				this.logger.Errorf("paymentinvoice: announcing order '%s' panicked: %v",
					event.OrderId, recovered)
			}
		}()

		timedCtx, cancel := context.WithTimeout(bgCtx.InnerContext(), publishAsyncTimeout)
		defer cancel()

		if err := this.publish(timedCtx, event); err != nil {
			this.logger.Warn("paymentinvoice: payment settled event publish failed",
				logging.Attr{
					"error":    err.Error(),
					"order_id": event.OrderId,
					"type":     string(event.Type),
				})
		}
	}()
}

func (this *PaymentSettledEventPublisherImpl) publish(
	ctx context.Context, event itEvent.PaymentSettledEvent,
) error {
	// Marshaled here rather than handed to the bus as a struct: the bus base64-encodes a payload it
	// has to marshal itself, and a subscriber decoding this expects plain JSON.
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req := coreEvent.NewEventRequest("", this.topic, "", &message.Message{Payload: body})
	return this.bus.PublishRequest(ctx, *req)
}
