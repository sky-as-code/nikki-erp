package v1

import (
	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The custom action handlers of every Sales resource, collected in one file.
//
// Each is one line: composable.ServeAction converts the echo context, calls the application
// method, and answers 400 for a business refusal, 404 for a missing record, or 200 with the
// shaped data. MutateResponse shapes a write, Identity a read that returns a payload. No
// authorization here -- the application method asserts it, and SmokeAuthz fails the request if
// no layer did.

func (this *SalesFulfillmentMethodRest) Archive(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "archive fulfillment method", payload,
		this.methodSvc.Archive, composable.MutateResponse)
}

func (this *SalesFulfillmentMethodRest) Unarchive(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "unarchive fulfillment method", payload,
		this.methodSvc.Unarchive, composable.MutateResponse)
}

func (this *SalesChannelRest) Suspend(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "suspend sales channel", payload,
		this.channelSvc.Suspend, composable.MutateResponse)
}

func (this *SalesChannelRest) Activate(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "activate sales channel", payload,
		this.channelSvc.Activate, composable.MutateResponse)
}

func (this *SalesChannelRest) Archive(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "archive sales channel", payload,
		this.channelSvc.Archive, composable.MutateResponse)
}

func (this *SalesChannelRest) Resolve(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "resolve sales channel", payload,
		this.channelSvc.Resolve, composable.Identity[any])
}

func (this *SalesPointRest) Suspend(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "suspend sales point", payload,
		this.pointSvc.Suspend, composable.MutateResponse)
}

func (this *SalesPointRest) Activate(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "activate sales point", payload,
		this.pointSvc.Activate, composable.MutateResponse)
}

func (this *SalesPointRest) Archive(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "archive sales point", payload,
		this.pointSvc.Archive, composable.MutateResponse)
}

func (this *SalesPointRest) Unarchive(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "unarchive sales point", payload,
		this.pointSvc.Unarchive, composable.MutateResponse)
}

func (this *SalesPricelistRest) SetDefault(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "set default pricelist", payload,
		this.pricelistSvc.SetDefault, composable.MutateResponse)
}

// The sales order's twelve actions. create_order is collection-level; the rest address a record.

func (this *SalesOrderRest) CreateOrder(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create sales order", payload,
		this.orderSvc.CreateOrder, composable.Identity[any])
}

func (this *SalesOrderRest) Reprice(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reprice sales order", payload,
		this.orderSvc.Reprice, composable.Identity[any])
}

func (this *SalesOrderRest) Confirm(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "confirm sales order", payload,
		this.orderSvc.Confirm, composable.Identity[any])
}

func (this *SalesOrderRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "cancel sales order", payload,
		this.orderSvc.Cancel, composable.Identity[any])
}

func (this *SalesOrderRest) ApplyVoucher(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "apply voucher", payload,
		this.orderSvc.ApplyVoucher, composable.Identity[any])
}

func (this *SalesOrderRest) ExplainPrice(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "explain order price", payload,
		this.orderSvc.ExplainPrice, composable.Identity[any])
}

func (this *SalesOrderRest) GrantManualDiscount(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "grant manual discount", payload,
		this.orderSvc.GrantManualDiscount, composable.Identity[any])
}

func (this *SalesOrderRest) RevokeManualDiscount(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "revoke manual discount", payload,
		this.orderSvc.RevokeManualDiscount, composable.Identity[any])
}

func (this *SalesOrderRest) AssignParties(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "assign order parties", payload,
		this.orderSvc.AssignParties, composable.Identity[any])
}

func (this *SalesOrderRest) AssignSoldToParty(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "assign sold-to party", payload,
		this.orderSvc.AssignSoldToParty, composable.Identity[any])
}

func (this *SalesOrderRest) AssignBillToParty(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "assign bill-to party", payload,
		this.orderSvc.AssignBillToParty, composable.Identity[any])
}

func (this *SalesOrderRest) AssignPayerParty(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "assign payer party", payload,
		this.orderSvc.AssignPayerParty, composable.Identity[any])
}

// The bill's five actions. merge is collection-level; the rest address a record.

func (this *SalesBillRest) Split(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "split bill", payload,
		this.billSvc.Split, composable.Identity[any])
}

func (this *SalesBillRest) Merge(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "merge bills", payload,
		this.billSvc.Merge, composable.Identity[any])
}

func (this *SalesBillRest) Pay(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "pay bill", payload,
		this.billSvc.Pay, composable.Identity[any])
}

func (this *SalesBillRest) Settle(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "settle bill", payload,
		this.billSvc.Settle, composable.Identity[any])
}

func (this *SalesBillRest) StartGatewayPayment(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "start gateway payment", payload,
		this.billSvc.StartGatewayPayment, composable.Identity[any])
}

// The channel's payment-method routes. They were served by the legacy engine loop until the bill
// migration made ChannelPaymentAppService injectable.

func (this *SalesChannelRest) PaymentMethods(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "list channel payment methods", payload,
		this.channelSvc.PaymentMethods, composable.Identity[any])
}

func (this *SalesChannelRest) EnablePaymentMethod(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "enable channel payment method", payload,
		this.channelSvc.EnablePaymentMethod, composable.MutateResponse)
}

func (this *SalesChannelRest) DisablePaymentMethod(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "disable channel payment method", payload,
		this.channelSvc.DisablePaymentMethod, composable.MutateResponse)
}

// The dispense loop, and the order's two fulfillment views.

func (this *SalesOrderFulfillmentRest) CreateAttempt(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create fulfillment attempt", payload,
		this.fulfillmentSvc.CreateAttempt, composable.Identity[any])
}

func (this *SalesOrderFulfillmentRest) ApplyAttemptResult(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "apply attempt result", payload,
		this.fulfillmentSvc.ApplyAttemptResult, composable.Identity[any])
}

func (this *SalesOrderFulfillmentRest) ReassignTarget(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "reassign fulfillment target", payload,
		this.fulfillmentSvc.ReassignTarget, composable.Identity[any])
}

func (this *SalesOrderRest) ListFulfillments(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "list order fulfillments", payload,
		this.orderSvc.ListFulfillments, composable.Identity[any])
}

func (this *SalesOrderRest) SearchFulfillmentTargets(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "search fulfillment targets", payload,
		this.orderSvc.SearchFulfillmentTargets, composable.Identity[any])
}

// The return lifecycle, and the order's two refund views.

func (this *SalesReturnRest) CreateReturn(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create return", payload,
		this.returnSvc.CreateReturn, composable.Identity[any])
}

func (this *SalesReturnRest) Process(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "process return", payload,
		this.returnSvc.Process, composable.Identity[any])
}

func (this *SalesReturnRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "cancel return", payload,
		this.returnSvc.Cancel, composable.Identity[any])
}

func (this *SalesOrderRest) CreateRefunds(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create order refunds", payload,
		this.orderSvc.CreateRefunds, composable.Identity[any])
}

func (this *SalesOrderRest) ViewRefunds(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "view order refunds", payload,
		this.orderSvc.ViewRefunds, composable.Identity[any])
}

func (this *SalesOrderRest) ListBills(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "list order bills", payload,
		this.orderSvc.ListBills, composable.Identity[any])
}

// The quotation's three transitions.

func (this *SalesQuotationRest) Convert(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "convert quotation", payload,
		this.quotationSvc.Convert, composable.Identity[any])
}

func (this *SalesQuotationRest) Send(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "send quotation", payload,
		this.quotationSvc.Send, composable.Identity[any])
}

func (this *SalesQuotationRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "cancel quotation", payload,
		this.quotationSvc.Cancel, composable.Identity[any])
}

// The fiscal request and the billing instruction lifecycle.

func (this *SalesFiscalRequestRest) RequestInvoice(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "request invoice", payload,
		this.fiscalSvc.RequestInvoice, composable.Identity[any])
}

func (this *SalesBillingInstructionRest) CreateInstruction(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "create billing instruction", payload,
		this.billingSvc.CreateInstruction, composable.Identity[any])
}

func (this *SalesBillingInstructionRest) UpdateInstruction(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "update billing instruction", payload,
		this.billingSvc.UpdateInstruction, composable.Identity[any])
}

func (this *SalesBillingInstructionRest) MarkReady(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "mark billing instruction ready", payload,
		this.billingSvc.MarkReady, composable.Identity[any])
}

func (this *SalesBillingInstructionRest) RevertToDraft(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "revert billing instruction to draft", payload,
		this.billingSvc.RevertToDraft, composable.Identity[any])
}

func (this *SalesBillingInstructionRest) Cancel(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "cancel billing instruction", payload,
		this.billingSvc.Cancel, composable.Identity[any])
}
