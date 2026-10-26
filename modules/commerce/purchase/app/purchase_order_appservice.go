package app

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/services"
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The purchase order's authorized surface: the built-in CRUD plus the lifecycle actions. Each
// action carries its own permission rather than reusing "update", because confirming and approving
// are materially different powers from editing a description. The permission codes match the
// action codes seeded in the purchase IAM migration.

type PurchaseOrderApplicationServiceImpl struct {
	composable.CrudApplicationService
	orders *services.PurchaseOrderDomainServiceImpl
}

// NewPurchaseOrderApplicationService needs the derived order service, which alone carries the
// lifecycle operations. A base wired with any other domain service is a wiring bug, so it fails the
// boot rather than the first request.
func NewPurchaseOrderApplicationService(base composable.CrudApplicationService) itOrder.PurchaseOrderCrudApplicationService {
	orders, ok := base.DomainService().(*services.PurchaseOrderDomainServiceImpl)
	if !ok {
		panic("the purchase order onion is not running the derived order domain service")
	}
	return &PurchaseOrderApplicationServiceImpl{CrudApplicationService: base, orders: orders}
}

// recordAction asserts the permission and the record's org, then runs the operation.
func (this *PurchaseOrderApplicationServiceImpl) recordAction(
	ctx corectx.Context, permission string, cmd itOrder.ActionCommand, run func() (*actionResult, error),
) (*actionResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, permission, cmd); cErrs != nil || err != nil {
		return actionFailure(cErrs, err)
	}
	return run()
}

func (this *PurchaseOrderApplicationServiceImpl) ConfirmAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionConfirm, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Confirm(
			ctx, readStringParam(cmd, paramRecordId), readStringParam(cmd, paramAlternativeChoice)))
	})
}

func (this *PurchaseOrderApplicationServiceImpl) ApproveAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionApprove, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Approve(ctx, readStringParam(cmd, paramRecordId)))
	})
}

func (this *PurchaseOrderApplicationServiceImpl) CancelAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionCancel, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Cancel(
			ctx, readStringParam(cmd, paramRecordId), readStringParam(cmd, paramReason)))
	})
}

func (this *PurchaseOrderApplicationServiceImpl) SendAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionSend, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Send(ctx, readStringParam(cmd, paramRecordId)))
	})
}

func (this *PurchaseOrderApplicationServiceImpl) LockAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionLock, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Lock(ctx, readStringParam(cmd, paramRecordId)))
	})
}

// UnlockAction validates the reason in the service: an action whose params mix resource fields
// with action-specific input is not described by any single schema.
func (this *PurchaseOrderApplicationServiceImpl) UnlockAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionUnlock, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Unlock(
			ctx, readStringParam(cmd, paramRecordId), readStringParam(cmd, paramReason)))
	})
}

func (this *PurchaseOrderApplicationServiceImpl) AcknowledgeAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionAcknowledge, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Acknowledge(ctx, readStringParam(cmd, paramRecordId)))
	})
}

// RepriceAction takes only the order id: which prices apply is the resolver's answer, and letting
// a caller pass one in would make this an override with extra steps. It answers which lines moved.
func (this *PurchaseOrderApplicationServiceImpl) RepriceAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionReprice, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Reprice(ctx, readStringParam(cmd, paramRecordId)))
	})
}

// MergeAction acts on several orders at once, so it takes their ids in the body and asserts no
// single record.
func (this *PurchaseOrderApplicationServiceImpl) MergeAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	if _, cErrs := this.AssertAction(ctx, constants.ActionMerge, cmd); cErrs != nil {
		return actionFailure(cErrs, nil)
	}
	return toActionResult(this.orders.MergeOrders(ctx, readStringList(cmd, paramMergeOrderIds)))
}

// CreateAlternativeAction raises an order, so it carries the create permission.
func (this *PurchaseOrderApplicationServiceImpl) CreateAlternativeAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionCreate, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.CreateAlternative(
			ctx, readStringParam(cmd, paramRecordId), readStringParam(cmd, paramAlternativeVendor)))
	})
}

// CompareAlternativesAction changes nothing, so it needs only read.
func (this *PurchaseOrderApplicationServiceImpl) CompareAlternativesAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionRead, cmd, func() (*actionResult, error) {
		comparison, err := services.CompareAlternatives(ctx, readStringParam(cmd, paramRecordId))
		if err != nil {
			return nil, err
		}
		return &actionResult{Data: comparison, HasData: true}, nil
	})
}

// DuplicateAction answers the new order rather than an affected count, so the caller has its id.
func (this *PurchaseOrderApplicationServiceImpl) DuplicateAction(
	ctx corectx.Context, cmd itOrder.ActionCommand,
) (*itOrder.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionCreate, cmd, func() (*actionResult, error) {
		return toActionResult(this.orders.Duplicate(ctx, readStringParam(cmd, paramRecordId)))
	})
}
