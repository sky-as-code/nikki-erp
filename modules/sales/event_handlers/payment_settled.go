// Package eventhandlers acts on the events Sales subscribes to.
//
// A handler here is the thin half of the work: it turns an event into a domain call and nothing
// more. The rules — what may move a payment, whether a bill is now settled — stay in
// domain/services, so that the same rules apply whether a verdict arrived on the bus or was found
// by the reconciliation sweep.
package eventhandlers

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itEvent "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/event"
)

// paymentSettledHandler applies one verdict to the payment awaiting it.
//
// One type serves all four outcomes, mapped rather than branched: the work is identical apart from
// the state the payment lands in, and four near-copies would be four places to fix a bug.
type paymentSettledHandler struct {
	outcome services.ConfirmPaymentOutcome
	logger  logging.LoggerService
}

func (this *paymentSettledHandler) Handle(
	ctx corectx.Context, event itEvent.PaymentSettledEvent,
) error {
	result, err := services.ConfirmPaymentAndSettle(ctx, services.ConfirmPaymentParams{
		PaymentOrderId:   event.OrderId,
		SalesPaymentId:   event.SalesPaymentIdFromMetadata(),
		Outcome:          this.outcome,
		RefTransactionId: event.RefTransactionId,
	})
	if err != nil {
		return err
	}

	if !result.Found {
		// A verdict for an order Sales never opened. Not an error — the same bus carries the vending
		// machines' orders, and Sales hears about those too — but worth a line when the money was
		// actually taken, because then it is money nothing here accounts for.
		if this.outcome == services.ConfirmPaymentPaid {
			this.logger.Warn("sales: a settled payment order matched no sales payment",
				logging.Attr{"order_id": event.OrderId, "type": string(event.Type)})
		}
		return nil
	}

	if !result.Applied {
		// A replay: the announcement and the sweep both reached the same payment, which is the
		// design working rather than a fault.
		return nil
	}

	this.logger.Info("sales: payment settled",
		logging.Attr{
			"sales_payment_id": result.SalesPaymentId,
			"sales_bill_id":    result.SalesBillId,
			"status":           result.Status,
		})
	return nil
}

// NewPaymentSettledHandlerRegistry maps each verdict to the handler that applies it.
func NewPaymentSettledHandlerRegistry(
	logger logging.LoggerService,
) itEvent.PaymentSettledHandlerRegistry {
	return itEvent.PaymentSettledHandlerRegistry{
		itEvent.PaymentSettledPaid: &paymentSettledHandler{
			outcome: services.ConfirmPaymentPaid, logger: logger,
		},
		itEvent.PaymentSettledFailed: &paymentSettledHandler{
			outcome: services.ConfirmPaymentFailed, logger: logger,
		},
		itEvent.PaymentSettledExpired: &paymentSettledHandler{
			outcome: services.ConfirmPaymentExpired, logger: logger,
		},
		itEvent.PaymentSettledCanceled: &paymentSettledHandler{
			outcome: services.ConfirmPaymentCanceled, logger: logger,
		},
	}
}

// InitHandlers registers the registry. It must run before the subscribers are wired, because a
// subscriber resolves its registry at construction.
func InitHandlers() error {
	return deps.Register(NewPaymentSettledHandlerRegistry)
}
