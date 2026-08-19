package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// sweepPageSize bounds one pass of a background sweep.
//
// A sweep that took every stale order at once would hold them all in memory and run one long
// transaction; taking a bounded page instead means a backlog is worked off over several runs. The
// watchdog runs every minute, so a page per minute drains any realistic backlog quickly.
const sweepPageSize = 200

// StaleOrder is one order a sweep found, reduced to what the sweep acts on.
//
// It carries the terminal id out of the metadata map because the mPOS adapter needs it to ask
// about an order, and reaching back into the map at the call site would put gateway-specific
// knowledge in the caller.
type StaleOrder struct {
	// Pk is this module's primary key, which the status write is keyed by.
	Pk string

	// OrderId is the identifier the ordering system holds, used when notifying it.
	OrderId string

	// OrderCode is the gateway's handle, used when asking it about the payment.
	OrderCode string

	Status          string
	PaymentMethodId string

	// PaymentProfileId is the merchant account the order was collected into, empty when it was
	// collected with the deployment's own credentials. The gateway has to be asked about the order
	// under the same account, or it answers that it has never seen it.
	PaymentProfileId string

	ReturnUrl string
	Metadata  map[string]any
}

// FindStaleOrders returns orders still awaiting a verdict past the point they should have had one.
//
// Only pending and processing qualify: every other status is a verdict already reached, and a
// sweep that re-examined one would be able to walk back a settled payment. The cut-off is passed
// in rather than read here so the caller owns the configuration and the tests own the clock.
func FindStaleOrders(ctx corectx.Context, olderThan time.Time) ([]StaleOrder, error) {
	return findOrdersOlderThan(ctx, olderThan, []any{
		models.OrderStatusPending,
		models.OrderStatusProcessing,
	})
}

// FindCleanableOrders returns orders old enough to delete.
//
// Expired joins the two unfinished states here: an expired order was never paid, so once it is
// old enough that nobody is going to ask about it, it is only taking up space. A paid, failed,
// canceled or refunded order is never in this set — those are the financial record.
func FindCleanableOrders(ctx corectx.Context, olderThan time.Time) ([]StaleOrder, error) {
	return findOrdersOlderThan(ctx, olderThan, []any{
		models.OrderStatusPending,
		models.OrderStatusProcessing,
		models.OrderStatusExpired,
	})
}

// findOrdersOlderThan runs the shared query behind both sweeps.
func findOrdersOlderThan(
	ctx corectx.Context, olderThan time.Time, statuses []any,
) ([]StaleOrder, error) {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.OrderFieldStatus, dmodel.In, statuses...),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldCreatedAt, dmodel.LessThan, olderThan),
	)
	// Oldest first, so a backlog larger than one page is drained in the order it accumulated
	// rather than the same page being re-read every run.
	graph.OrderBy(basemodel.FieldCreatedAt)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  sweepPageSize,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findOrdersOlderThan")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	stale := make([]StaleOrder, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		order := models.NewOrderFrom(item)
		stale = append(stale, StaleOrder{
			Pk:               derefString(order.GetId()),
			OrderId:          derefString(order.GetOrderId()),
			OrderCode:        derefString(order.GetOrderCode()),
			Status:           derefString(order.GetStatus()),
			PaymentMethodId:  derefString((*string)(order.GetPaymentMethodId())),
			PaymentProfileId: derefString((*string)(order.GetPaymentProfileId())),
			ReturnUrl:        derefString(order.GetReturnUrl()),
			Metadata:         order.GetMetadata(),
		})
	}
	return stale, nil
}

// DeleteOrder removes one order and the transactions filed against it.
//
// The transactions go first: they carry an edge to the order, so deleting the order while they
// still point at it either fails on the constraint or leaves rows referencing nothing, depending
// on how the edge was generated. Doing it in this order does not depend on which.
//
// Both deletions share one transaction, so a failure part-way through leaves the order intact
// rather than stripped of its evidence.
func DeleteOrder(ctx corectx.Context, orderPk string) error {
	return withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		transactions, err := findTransactionsOfOrder(tranxCtx, orderPk)
		if err != nil {
			return err
		}

		transactionEngine, err := engineFor(models.TransactionSchemaName)
		if err != nil {
			return err
		}
		for _, transactionPk := range transactions {
			_, err := transactionEngine.ResourceRepository().DeleteOne(tranxCtx, dmodel.DynamicFields{
				models.TransactionFieldId: transactionPk,
			})
			if err != nil {
				return errors.Wrapf(err, "DeleteOrder: transaction '%s'", transactionPk)
			}
		}

		orderEngine, err := engineFor(models.OrderSchemaName)
		if err != nil {
			return err
		}
		_, err = orderEngine.ResourceRepository().DeleteOne(tranxCtx, dmodel.DynamicFields{
			models.OrderFieldId: orderPk,
		})
		return errors.Wrapf(err, "DeleteOrder: order '%s'", orderPk)
	})
}

// findTransactionsOfOrder returns the primary keys of every transaction filed against one order.
func findTransactionsOfOrder(ctx corectx.Context, orderPk string) ([]string, error) {
	engine, err := engineFor(models.TransactionSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.TransactionFieldOrderId, dmodel.Equals, orderPk),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  sweepPageSize,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findTransactionsOfOrder")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	keys := make([]string, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		keys = append(keys, derefString(models.NewTransactionFrom(item).GetId()))
	}
	return keys, nil
}
