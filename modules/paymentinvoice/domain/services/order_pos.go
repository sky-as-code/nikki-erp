package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// posOrderClearer is the part of the card-terminal adapter that clears a terminal's queue.
//
// It is not on the gateway port: only a physical terminal has a queue of prompts waiting on it,
// and putting a method on the port that every other adapter would have to stub would say the
// opposite. The mPOS adapter is reached through this interface rather than by importing the
// package, so the domain layer still does not depend on any adapter.
type posOrderClearer interface {
	RemovePosOrders(
		ctx corectx.Context,
		orderCode string,
		posId string,
		amount decimal.Decimal,
		profileConfig map[string]any,
	) error
}

// RemovePosOrdersCommand names the terminal whose queued prompts are to be cleared.
type RemovePosOrdersCommand struct {
	PosId string
}

// RemovePosOrdersResult reports how many orders were withdrawn from the terminal.
type RemovePosOrdersResult struct {
	AffectedCount int
}

// RemovePosOrders withdraws the payment prompts still queued on one card terminal.
//
// A terminal holds prompts until a cashier acts on them, so an abandoned order leaves one on the
// screen indefinitely — the next customer is shown a prompt for someone else's purchase, and can
// pay it. Clearing them is what stops that.
//
// It replaces the unauthenticated DELETE /mpos/pos-orders/:posId of the service this module
// supersedes. That endpoint let anyone who could reach it cancel any terminal's queue; here it is
// an authenticated action with its own permission.
//
// Each order is cleared at the gateway and then expired locally, and one that the gateway refuses
// is left alone: the prompt is still live on the terminal, so recording it as expired here would
// have the two disagree about whether the customer can still pay.
func (this *OrderDomainService) RemovePosOrders(
	ctx corectx.Context, cmd RemovePosOrdersCommand,
) (*RemovePosOrdersResult, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	if cmd.PosId == "" {
		appendFieldViolation(vErrs, models.OrderMetaPosId,
			"paymentinvoice.pos_id_required", "the terminal must be identified")
		return nil, vErrs, nil
	}

	orders, err := this.findOpenOrdersOnTerminal(ctx, cmd.PosId)
	if err != nil {
		return nil, vErrs, err
	}

	cleared := 0
	for _, order := range orders {
		done, err := this.clearOneposOrder(ctx, order)
		if err != nil {
			return nil, vErrs, err
		}
		if done {
			cleared++
		}
	}

	return &RemovePosOrdersResult{AffectedCount: cleared}, vErrs, nil
}

// clearOneposOrder withdraws one order's prompt and expires it, reporting whether it was cleared.
func (this *OrderDomainService) clearOneposOrder(
	ctx corectx.Context, order models.Order,
) (bool, error) {
	method, err := this.readMethod(ctx, derefString(order.GetPaymentMethodId()))
	if err != nil || method == nil {
		return false, err
	}

	adapter, exists := this.registry.Get(derefString(method.GetAdapterCode()))
	if !exists {
		return false, nil
	}
	clearer, ok := adapter.(posOrderClearer)
	if !ok {
		// The order names a method whose adapter has no terminal queue, so there is nothing on
		// a terminal to withdraw. Not an error: the metadata match above is by pos_id alone.
		return false, nil
	}

	// The prompt is withdrawn by the account that queued it: the terminal holds it against that
	// merchant, and another account's credentials cannot cancel it.
	profileConfig, err := this.profileConfigForOrder(ctx, order)
	if err != nil {
		return false, err
	}

	posId, _ := order.GetMetadata()[models.OrderMetaPosId].(string)
	err = clearer.RemovePosOrders(ctx,
		derefString(order.GetOrderCode()), posId, derefDecimal(order.GetAmount()), profileConfig)
	if err != nil {
		// The prompt may still be live on the terminal, so the order is left payable rather
		// than marked expired against a gateway that still considers it open.
		return false, nil
	}

	err = withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		return writeOrderFields(tranxCtx, derefString(order.GetId()), dmodel.DynamicFields{
			models.OrderFieldStatus: models.OrderStatusExpired,
		})
	})
	return err == nil, err
}

// findOpenOrdersOnTerminal lists the orders still awaiting payment on one terminal.
//
// Terminated orders are excluded: an order already paid, failed or expired has no prompt left on
// the device, and withdrawing one would ask the gateway to cancel a payment it has taken.
func (this *OrderDomainService) findOpenOrdersOnTerminal(
	ctx corectx.Context, posId string,
) ([]models.Order, error) {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().Or(
			*dmodel.NewSearchNode().NewCondition(
				models.OrderFieldStatus, dmodel.Equals, models.OrderStatusPending),
			*dmodel.NewSearchNode().NewCondition(
				models.OrderFieldStatus, dmodel.Equals, models.OrderStatusProcessing),
		),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  posOrderScanLimit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findOpenOrdersOnTerminal")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	// The terminal id lives inside the metadata map, which the query builder cannot filter on,
	// so the match is made here. The status filter above is what keeps this set small.
	orders := make([]models.Order, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		order := models.NewOrderFrom(item)
		if stored, _ := order.GetMetadata()[models.OrderMetaPosId].(string); stored == posId {
			orders = append(orders, *order)
		}
	}
	return orders, nil
}

// posOrderScanLimit bounds the scan for orders queued on a terminal. A terminal serves one
// customer at a time, so a queue anywhere near this size means something else is wrong.
const posOrderScanLimit = 200

// readMethod fetches a payment method without judging whether it is still taking payments.
func (this *OrderDomainService) readMethod(
	ctx corectx.Context, methodId string,
) (*models.PaymentMethod, error) {
	engine, err := engineFor(models.PaymentMethodSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PaymentMethodFieldId: methodId,
	})
	if err != nil || found == nil || !found.HasData {
		return nil, err
	}
	return models.NewPaymentMethodFrom(found.Data), nil
}
