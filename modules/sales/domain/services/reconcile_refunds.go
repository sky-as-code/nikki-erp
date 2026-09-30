package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

type ReconcileRefundsResult struct {
	Retried          int
	RetryFailed      int
	CancelledRefunds int
}

func ReconcileStaleRefunds(
	ctx corectx.Context,
	dLock lock.DistributedLock,
	deps RefundProcessingDeps,
	now time.Time,
	olderThan time.Duration,
	limit int,
) (*ReconcileRefundsResult, error) {
	result := &ReconcileRefundsResult{}
	cutoff := model.ModelDateTime(now.Add(-olderThan))

	legs, err := searchOlderThan(ctx, models.SalesRefundPaymentSchemaName, cutoff, limit,
		*dmodel.NewSearchNode().NewCondition(models.SalesRefundPaymentFieldStatus, dmodel.Equals,
			string(models.SalesRefundPaymentStatusPending)))
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, leg := range legs {
		returnId := stringOf(leg, models.SalesRefundPaymentFieldSalesReturnId)
		if returnId == "" || seen[returnId] {
			continue
		}
		seen[returnId] = true
		_, vErrs, err := ProcessReturn(ctx, ProcessReturnParams{SalesReturnId: returnId}, dLock,
			deps.Fulfillment, deps.Invoicing, deps.PaymentOrders)
		if err != nil || vErrs != nil {
			result.RetryFailed++
			continue
		}
		result.Retried++
	}

	orders, err := searchOlderThan(ctx, models.SalesOrderSchemaName, cutoff, limit,
		*dmodel.NewSearchNode().NewCondition(models.SalesOrderFieldStatus, dmodel.Equals,
			string(models.SalesOrderStatusCancelled)),
		*dmodel.NewSearchNode().NewCondition(models.SalesOrderFieldPaymentStatus, dmodel.In,
			string(models.SalesOrderPaymentStatusPaid), string(models.SalesOrderPaymentStatusOverpaid)))
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		refund, err := withLockedCancelledRefund(ctx, dLock, stringOf(order, models.SalesOrderFieldId))
		if err != nil {
			result.RetryFailed++
			continue
		}
		if refund != nil {
			result.CancelledRefunds++
		}
	}
	return result, nil
}

func withLockedCancelledRefund(
	ctx corectx.Context, dLock lock.DistributedLock, orderId string,
) (*CancelledOrderRefund, error) {
	release, err := acquireOrderLock(ctx, dLock, orderId)
	if err != nil {
		return nil, err
	}
	defer release()
	return refundCancelledOrderUnderLock(ctx, orderId)
}

func searchOlderThan(
	ctx corectx.Context, schemaName string, cutoff model.ModelDateTime, limit int, conditions ...dmodel.SearchNode,
) ([]dmodel.DynamicFields, error) {
	engineRepo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	size := limit
	if size <= 0 {
		size = model.MODEL_RULE_PAGE_MAX_SIZE
	}

	graph := &dmodel.SearchGraph{}
	graph.And(append(conditions,
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldCreatedAt, dmodel.LessThan, cutoff))...)

	found, err := engineRepo.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: size})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}
