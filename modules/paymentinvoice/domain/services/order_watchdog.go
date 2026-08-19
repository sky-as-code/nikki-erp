package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// WatchdogVerdict is what asking the gateway about one stale order came to.
type WatchdogVerdict struct {
	// Settled is false when the gateway has not reached an outcome. Such an order is left alone
	// rather than expired: a customer still at a card terminal has not failed to pay.
	Settled bool

	// Paid is the outcome, meaningful only when Settled.
	Paid bool

	// Applied is false when the order had already reached a verdict by the time this one landed —
	// which happens whenever the callback arrives while the sweep is in flight.
	Applied bool

	// OrderId is the identifier the ordering system holds, for the notification that follows.
	OrderId string
}

// ReconcileStaleOrder asks the gateway what became of an order no callback ever arrived for.
//
// This is the recovery path for a lost callback, and the reason it exists is that a gateway which
// cannot deliver its callback still took the customer's money. Asking closes the gap.
//
// The verdict is applied through ApplyGatewayResult, the same convergence point the callbacks use,
// so a callback arriving mid-sweep and this sweep cannot both apply: whichever reaches the
// transaction second sees the state the first committed and does nothing. That check has to happen
// inside the transaction, which is why this does not pre-check the status it was handed.
func (this *OrderDomainService) ReconcileStaleOrder(
	ctx corectx.Context, stale StaleOrder,
) (*WatchdogVerdict, error) {
	adapter, err := this.adapterForMethodId(ctx, stale.PaymentMethodId)
	if err != nil {
		return nil, err
	}
	if adapter == nil {
		// The method's gateway is not enabled on this deployment, so nothing here can ask about
		// the order. Expiring it on that basis would be wrong: the gateway may well have taken
		// the money, and this deployment simply cannot see it.
		return &WatchdogVerdict{}, nil
	}

	order, err := findOrderByBusinessId(ctx, stale.OrderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return &WatchdogVerdict{}, nil
	}

	// The question goes to the account that took the payment. A gateway asked about an order under
	// another merchant's credentials answers that it has never seen it, which the sweep would read
	// as "not settled" and eventually expire — closing an order the customer really did pay.
	profileConfig, err := this.profileConfigById(ctx, stale.PaymentProfileId)
	if err != nil {
		return nil, err
	}

	outcome, err := adapter.CheckOrder(ctx, itGateway.CheckOrderRequest{
		OrderCode:     stale.OrderCode,
		Amount:        derefDecimal(order.GetAmount()),
		Metadata:      stale.Metadata,
		ProfileConfig: profileConfig,
	})
	if err != nil {
		// The gateway could not be reached or refused the question. The order keeps its status and
		// the next sweep asks again — an unreachable gateway is not evidence about a payment.
		return nil, errors.Wrapf(err, "ReconcileStaleOrder(%s)", stale.OrderId)
	}
	if outcome == nil || !outcome.Settled {
		return &WatchdogVerdict{OrderId: stale.OrderId}, nil
	}

	applied, err := this.ApplyGatewayResult(ctx, GatewayResult{
		OrderCode:        stale.OrderCode,
		Paid:             outcome.Paid,
		RefTransactionId: outcome.RefTransactionId,
		RefPayload:       outcome.RawResponse,
	})
	if err != nil {
		return nil, err
	}

	return &WatchdogVerdict{
		Settled: true,
		Paid:    outcome.Paid,
		Applied: applied.Applied,
		OrderId: applied.OrderId,
	}, nil
}

// ExpireStaleOrder gives up on an order the gateway reached no verdict on.
//
// It is the end of the line for an order nobody paid: the customer walked away, and holding it
// open forever would leave the terminal's queue and the operator's list growing without bound.
//
// The write goes through the same terminal-status check the callbacks use, so an order settled
// between the sweep reading it and this call is not walked back to expired.
func (this *OrderDomainService) ExpireStaleOrder(ctx corectx.Context, stale StaleOrder) (bool, error) {
	expired := false
	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		order, err := findOrderByCode(tranxCtx, stale.OrderCode)
		if err != nil || order == nil {
			return err
		}
		if isTerminalPaymentStatus(derefString(order.GetStatus())) {
			return nil
		}

		if err := writeOrderFields(tranxCtx, derefString(order.GetId()), dmodel.DynamicFields{
			models.OrderFieldStatus: models.OrderStatusExpired,
		}); err != nil {
			return err
		}

		// The payment attempt is canceled rather than failed: nothing was refused, the customer
		// simply never completed it, and the distinction is what someone reconciling later reads.
		transaction, err := findPendingPaymentTransaction(tranxCtx, derefString(order.GetId()))
		if err != nil {
			return err
		}
		if transaction != nil {
			if err := writeTransactionFields(tranxCtx, derefString(transaction.GetId()), dmodel.DynamicFields{
				models.TransactionFieldStatus: models.TransactionStatusCanceled,
			}); err != nil {
				return err
			}
		}

		expired = true
		return nil
	})
	return expired, err
}

// adapterForMethodId resolves the gateway serving one payment method.
//
// Unlike the create path this tolerates an inactive method: a method withdrawn from use after an
// order was taken still has to be asked about that order, or every payment in flight at the moment
// of withdrawal is stranded.
func (this *OrderDomainService) adapterForMethodId(
	ctx corectx.Context, methodId string,
) (itGateway.PaymentGateway, error) {
	if methodId == "" {
		return nil, nil
	}

	engine, err := engineFor(models.PaymentMethodSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PaymentMethodFieldId: methodId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "adapterForMethodId")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	method := models.NewPaymentMethodFrom(found.Data)
	adapter, exists := this.registry.Get(derefString(method.GetAdapterCode()))
	if !exists {
		return nil, nil
	}
	return adapter, nil
}
