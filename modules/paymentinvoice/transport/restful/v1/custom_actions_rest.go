package v1

import (
	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The custom action handlers. Each hands the bound payload to an application method, which asserts
// the action's own permission before delegating, and answers the flat body the legacy actions did.

func (this *OrderRest) CreatePayment(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create payment", payload,
		this.orderSvc.CreatePaymentAction, composable.Identity[map[string]any])
}

func (this *OrderRest) Refund(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "refund order", payload,
		this.orderSvc.RefundAction, composable.Identity[map[string]any])
}

func (this *OrderRest) RemovePosOrders(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "remove pos orders", payload,
		this.orderSvc.RemovePosOrdersAction, composable.Identity[map[string]any])
}

func (this *InvoiceRest) Issue(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "issue invoice", payload,
		this.invoiceSvc.IssueAction, composable.Identity[map[string]any])
}
