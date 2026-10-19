package v1

import (
	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The custom action handlers. Each hands the bound payload to an application method, which asserts
// the action's own permission before delegating, and answers the action's data as the body, which
// is what the legacy engine actions returned.

func (this *PurchaseOrderRest) Confirm(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "confirm purchase order", payload,
		this.orderSvc.ConfirmAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Approve(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "approve purchase order", payload,
		this.orderSvc.ApproveAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "cancel purchase order", payload,
		this.orderSvc.CancelAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Send(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "send purchase order", payload,
		this.orderSvc.SendAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Lock(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "lock purchase order", payload,
		this.orderSvc.LockAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Unlock(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "unlock purchase order", payload,
		this.orderSvc.UnlockAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Acknowledge(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "acknowledge purchase order", payload,
		this.orderSvc.AcknowledgeAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Merge(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "merge purchase orders", payload,
		this.orderSvc.MergeAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) CreateAlternative(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create alternative purchase order", payload,
		this.orderSvc.CreateAlternativeAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) CompareAlternatives(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "compare alternative purchase orders", payload,
		this.orderSvc.CompareAlternativesAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Reprice(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reprice purchase order", payload,
		this.orderSvc.RepriceAction, composable.Identity[any])
}

func (this *PurchaseOrderRest) Duplicate(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "duplicate purchase order", payload,
		this.orderSvc.DuplicateAction, composable.Identity[any])
}

func (this *AgreementRest) Confirm(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "confirm purchase agreement", payload,
		this.agreementSvc.ConfirmAction, composable.Identity[any])
}

func (this *AgreementRest) Close(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "close purchase agreement", payload,
		this.agreementSvc.CloseAction, composable.Identity[any])
}

func (this *AgreementRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "cancel purchase agreement", payload,
		this.agreementSvc.CancelAction, composable.Identity[any])
}

func (this *AgreementRest) CreateRfq(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create rfq from purchase agreement", payload,
		this.agreementSvc.CreateRfqAction, composable.Identity[any])
}
