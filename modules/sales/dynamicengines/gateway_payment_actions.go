package dynamicengines

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Starting a gateway collection is its own action rather than a flag on `pay`.
//
// The two have different outcomes and different shapes: `pay` records money already taken and can
// settle the bill on the spot, while this one opens a collection and answers with something for the
// customer to pay with, having settled nothing. Folding them together would give one endpoint two
// return shapes and two meanings for success.
//
// It shares the `pay` permission: both are the power to take money against a bill, and splitting
// that in two would let a role start collections it could not record.

const (
	ActionStartGatewayPayment = "start_gateway_payment"
)

// paymentOrders is the port onto paymentinvoice's gateway, set through a setter rather than
// imported: this package may not import infra/, where the binding lives.
var paymentOrders itExt.PaymentOrderExtService

// SetPaymentOrderPort must be called by Init before any request is served.
func SetPaymentOrderPort(port itExt.PaymentOrderExtService) {
	paymentOrders = port
}

func defineGatewayPaymentActions(engine drif.DynamicResourceEngine) error {
	return engine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  ActionStartGatewayPayment,
		ActionType:  drif.ActionTypeGeneric,
		RestPath:    ":id/start_gateway_payment",
		Permission:  PermissionPayBill,
		MainProcess: processStartGatewayPayment,
	})
}

func processStartGatewayPayment(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	result, vErrs, err := services.StartGatewayPayment(ctx, services.StartGatewayPaymentParams{
		SalesBillId:     readStringParam(input.Params, paramId),
		PaymentMethodId: readStringParam(input.Params, paramPaymentMethodId),
		Amount:          readDecimalParam(input.Params, "amount"),
		CurrencyCode:    readStringParam(input.Params, "currency_code"),
		Content:         readStringParam(input.Params, "content"),
	}, paymentMethods, paymentOrders, channelPayments, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	// No bill status here, unlike pay: nothing has been settled. The customer has been handed
	// something to pay with, and the bill moves when the provider says the money arrived.
	return &drif.ActionResult{HasData: true, Data: map[string]any{
		"sales_payment_id": result.SalesPaymentId,
		"sales_bill_id":    result.SalesBillId,
		"payment_order_id": result.PaymentOrderId,
		"order_code":       result.OrderCode,
		"qr_code_url":      result.QrCodeUrl,
		"pay_url":          result.PayUrl,
	}}, nil
}
