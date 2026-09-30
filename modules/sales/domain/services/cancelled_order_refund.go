package services

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

type OrderPaymentHooks struct {
	Lock       lock.DistributedLock
	Policy     func(ctx corectx.Context) SalesPolicy
	RefundDeps RefundProcessingDeps
}

var orderPaymentHooks OrderPaymentHooks

func SetOrderPaymentHooks(hooks OrderPaymentHooks) {
	orderPaymentHooks = hooks
}

var errOrderLockUnavailable = errors.New("the order is being changed by another request")

func acquireOrderLock(ctx corectx.Context, dLock lock.DistributedLock, orderId string) (func(), error) {
	if dLock == nil || orderId == "" {
		return func() {}, nil
	}
	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, errors.Wrapf(err, "acquiring the lock of order '%s'", orderId)
	}
	if !acquired {
		return nil, errors.Wrapf(errOrderLockUnavailable, "order '%s'", orderId)
	}
	released := false
	return func() {
		if released {
			return
		}
		released = true
		_ = dLock.Release(ctx, key)
	}, nil
}

type CancelledOrderRefund struct {
	SalesReturnId string
	Status        string
}

func refundCancelledOrderUnderLock(ctx corectx.Context, orderId string) (*CancelledOrderRefund, error) {
	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || order == nil {
		return nil, err
	}
	if stringOf(order, models.SalesOrderFieldStatus) != string(models.SalesOrderStatusCancelled) || !isPaid(order) {
		return nil, nil
	}

	lines, err := searchBy(ctx, models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}
	claimed, err := refundOnlyClaimsOf(ctx, orderId)
	if err != nil {
		return nil, err
	}
	requests := make([]CreateReturnLine, 0, len(lines))
	for _, line := range lines {
		lineId := stringOf(line, models.SalesOrderLineFieldId)
		owed := decimalOf(line, models.SalesOrderLineFieldOrderedQuantity).
			Sub(decimalOf(line, models.SalesOrderLineFieldFulfilledQuantity)).
			Sub(claimed[lineId])
		if owed.IsPositive() {
			requests = append(requests, CreateReturnLine{SalesOrderLineId: lineId, RequestedQty: owed})
		}
	}
	if len(requests) == 0 {
		return nil, nil
	}

	policy := SalesPolicy{}
	if orderPaymentHooks.Policy != nil {
		policy = orderPaymentHooks.Policy(ctx)
	}
	created, vErrs, err := createReturnUnderLock(ctx, CreateReturnParams{
		SalesOrderId:        orderId,
		Reason:              "Paid after cancellation",
		RefundReason:        models.SalesRefundReasonCustomerRequested,
		ReturnType:          models.SalesReturnTypeRefundOnly,
		Lines:               requests,
		AllowCancelledOrder: true,
	}, order, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil && vErrs.Count() > 0 {
		return nil, errors.New("the refund of cancelled order '" + orderId + "' was refused: " +
			describeFirstViolation(vErrs))
	}

	refund := &CancelledOrderRefund{SalesReturnId: created.SalesReturnId, Status: created.Status}
	if !boolOf(order, models.SalesOrderFieldAutoConfirmRefund) {
		return refund, nil
	}
	confirmed, cErrs, err := confirmRefundRequestUnderLock(ctx, ConfirmRefundRequestParams{
		SalesReturnId: created.SalesReturnId, Automatic: true,
	}, orderPaymentHooks.RefundDeps)
	if err != nil {
		return nil, err
	}
	if cErrs == nil && confirmed != nil {
		refund.Status = confirmed.Status
	}
	return refund, nil
}
