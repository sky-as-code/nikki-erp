package v1

import (
	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The custom action handlers. Each hands the bound payload to an application method, which asserts
// the action's own permission before delegating, and answers the action's data as the body, which
// is what the legacy engine actions returned.

func (this *PurchaseOrderRest) Confirm(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "confirm purchase order", payload,
		this.orderSvc.ConfirmAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Approve(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "approve purchase order", payload,
		this.orderSvc.ApproveAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "cancel purchase order", payload,
		this.orderSvc.CancelAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Send(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "send purchase order", payload,
		this.orderSvc.SendAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Lock(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "lock purchase order", payload,
		this.orderSvc.LockAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Unlock(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "unlock purchase order", payload,
		this.orderSvc.UnlockAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Acknowledge(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "acknowledge purchase order", payload,
		this.orderSvc.AcknowledgeAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Merge(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "merge purchase orders", payload,
		this.orderSvc.MergeAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) CreateAlternative(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "create alternative purchase order", payload,
		this.orderSvc.CreateAlternativeAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) CompareAlternatives(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "compare alternative purchase orders", payload,
		this.orderSvc.CompareAlternativesAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Reprice(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "reprice purchase order", payload,
		this.orderSvc.RepriceAction, dynamicresource.Identity[any])
}

func (this *PurchaseOrderRest) Duplicate(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "duplicate purchase order", payload,
		this.orderSvc.DuplicateAction, dynamicresource.Identity[any])
}

func (this *AgreementRest) Confirm(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "confirm purchase agreement", payload,
		this.agreementSvc.ConfirmAction, dynamicresource.Identity[any])
}

func (this *AgreementRest) Close(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "close purchase agreement", payload,
		this.agreementSvc.CloseAction, dynamicresource.Identity[any])
}

func (this *AgreementRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "cancel purchase agreement", payload,
		this.agreementSvc.CancelAction, dynamicresource.Identity[any])
}

func (this *AgreementRest) CreateRfq(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "create rfq from purchase agreement", payload,
		this.agreementSvc.CreateRfqAction, dynamicresource.Identity[any])
}
