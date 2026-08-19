package dynamicengines

import (
	stdErr "errors"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// The order actions that move money.
//
// They are actions on the order resource rather than hand-written REST, so they get the engine's
// permission check, its parameter plumbing and its error shaping for free — and so the permission
// codes are the same ones the IAM seed grants. Each is a Generic action: taking a payment is
// neither a create of the order (the order already exists by then) nor an update of it.
//
// Each has its own permission rather than reusing "update". Being allowed to correct an order's
// description is not the same authority as being allowed to hand money back, and one permission
// covering both would make the smaller grant imply the larger.

// Parameter names of the three actions. They are snake_case to match the resource fields, and
// amount is deliberately the same name it has on the order.
const (
	paramOrderId         = "order_id"
	paramAmount          = "amount"
	paramSource          = "source"
	paramPaymentMethodId = "payment_method_id"

	// paramPaymentProfileId names the merchant account to collect into. Optional: a request
	// without it is collected with the credentials in this deployment's configuration, which is
	// what every caller did before profiles existed.
	paramPaymentProfileId = "payment_profile_id"
	paramContent          = "content"
	paramReturnUrl        = "return_url"
	paramMetadata         = "metadata"
	paramPosId            = "pos_id"
)

// defineOrderActions adds create_payment, refund and remove_pos_orders.
func defineOrderActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  constants.ActionCreatePayment,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    constants.ActionCreatePayment,
			Permission:  constants.ActionCreatePayment,
			MainProcess: processCreatePayment,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  constants.ActionRefund,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    constants.ActionRefund,
			Permission:  constants.ActionRefund,
			MainProcess: processRefund,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  constants.ActionRemovePosOrders,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    constants.ActionRemovePosOrders + "/:pos_id",
			Permission:  constants.ActionRemovePosOrders,
			MainProcess: processRemovePosOrders,
		}),
	)
}

// processRemovePosOrders clears the payment prompts queued on one card terminal.
func processRemovePosOrders(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := requireOrderService()
	if err != nil {
		return nil, err
	}

	result, cErrs, err := service.RemovePosOrders(ctx, services.RemovePosOrdersCommand{
		PosId: readString(input.Params, paramPosId),
	})
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *cErrs}, nil
	}

	return &drif.ActionResult{
		HasData: true,
		Data:    map[string]any{"affected_count": result.AffectedCount},
	}, nil
}

// orderServiceOf reaches the order domain service the module built during Init.
//
// It is held in a package variable rather than reached through the engine, because it is not a
// derived resource service: the two operations it carries are not CRUD on an order, so there is
// nothing to derive from and nothing to install with SetResourceService.
var orderService *services.OrderDomainService

// SetOrderService installs the service the order actions delegate to. Init calls it before any
// request is served.
func SetOrderService(service *services.OrderDomainService) {
	orderService = service
}

func requireOrderService() (*services.OrderDomainService, error) {
	if orderService == nil {
		return nil, errors.New(
			"the order domain service was not installed; PaymentInvoiceModule.Init must call " +
				"dynamicengines.SetOrderService")
	}
	return orderService, nil
}

// processCreatePayment records an order and asks its gateway to start collecting.
func processCreatePayment(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := requireOrderService()
	if err != nil {
		return nil, err
	}

	cmd, vErrs := buildCreatePaymentCommand(input.Params)
	if vErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, cErrs, err := service.CreatePayment(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *cErrs}, nil
	}

	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"order_id": result.OrderId,
			// order_code is what the gateway knows the order by and what its callback arrives
			// under, so a caller reconciling against the gateway's own records needs it.
			"order_code":  result.OrderCode,
			"qr_code_url": result.QrCodeUrl,
			"pay_url":     result.PayUrl,
		},
	}, nil
}

// processRefund gives money back for an order already paid.
func processRefund(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := requireOrderService()
	if err != nil {
		return nil, err
	}

	cmd, vErrs := buildRefundCommand(input.Params)
	if vErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, cErrs, err := service.Refund(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *cErrs}, nil
	}

	// rested_amount keeps the old service's spelling: the ordering system reads this key.
	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"order_id":      result.OrderId,
			"refund_amount": result.RefundAmount.String(),
			"rested_amount": result.RestedAmount.String(),
		},
	}, nil
}

// buildCreatePaymentCommand turns an untyped request body into a typed command.
//
// The action declares no ParamSchema — the params are a mix of order fields and method-specific
// input that no single schema describes — so every shape problem is checked here and reported as
// a client error naming the field.
func buildCreatePaymentCommand(
	params dmodel.DynamicFields,
) (services.CreatePaymentCommand, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	cmd := services.CreatePaymentCommand{
		Source:           readString(params, paramSource),
		PaymentMethodId:  readString(params, paramPaymentMethodId),
		PaymentProfileId: readString(params, paramPaymentProfileId),
		Content:          readOptionalString(params, paramContent),
		ReturnUrl:        readOptionalString(params, paramReturnUrl),
		Metadata:         buildCreateMetadata(params),
	}

	if cmd.PaymentMethodId == "" {
		vErrs.Append(*ft.NewBusinessViolation(paramPaymentMethodId,
			"paymentinvoice.payment_method_required",
			"a payment method must be identified"))
	}

	amount, ok := readDecimal(params, paramAmount)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(paramAmount,
			"paymentinvoice.amount_malformed",
			"the amount must be a number"))
	}
	cmd.Amount = amount

	return cmd, vErrs
}

// buildCreateMetadata collects the method-specific input.
//
// pos_id is accepted at the top level as well as inside metadata, because that is where the old
// service's callers put it and the vending machines still do. It is folded into the map either
// way, so nothing downstream has to know about the two spellings.
func buildCreateMetadata(params dmodel.DynamicFields) map[string]any {
	metadata := map[string]any{}
	if raw, ok := params[paramMetadata].(map[string]any); ok {
		for key, value := range raw {
			metadata[key] = value
		}
	}
	if posId := readString(params, paramPosId); posId != "" {
		metadata[models.OrderMetaPosId] = posId
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
}

func buildRefundCommand(params dmodel.DynamicFields) (services.RefundCommand, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	cmd := services.RefundCommand{
		OrderId: readString(params, paramOrderId),
		Content: readOptionalString(params, paramContent),
	}

	if cmd.OrderId == "" {
		vErrs.Append(*ft.NewBusinessViolation(paramOrderId,
			"paymentinvoice.order_id_required",
			"the order to refund must be identified"))
	}

	amount, ok := readDecimal(params, paramAmount)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(paramAmount,
			"paymentinvoice.amount_malformed",
			"the amount must be a number"))
	}
	cmd.Amount = amount

	return cmd, vErrs
}

func readString(params dmodel.DynamicFields, field string) string {
	value, ok := params[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

func readOptionalString(params dmodel.DynamicFields, field string) *string {
	if value := readString(params, field); value != "" {
		return &value
	}
	return nil
}

// readDecimal accepts the several shapes JSON and the query builder produce for one number.
//
// A string is accepted because that is how an exact amount survives JSON — a float64 cannot hold
// every decimal, and rounding one here would collect something other than what was asked for. It
// reports false rather than defaulting to zero, so a malformed amount is refused instead of
// silently becoming a free order.
func readDecimal(params dmodel.DynamicFields, field string) (decimal.Decimal, bool) {
	value, ok := params[field]
	if !ok || value == nil {
		return decimal.Zero, false
	}

	switch typed := value.(type) {
	case decimal.Decimal:
		return typed, true
	case string:
		parsed, err := decimal.NewFromString(typed)
		if err != nil {
			return decimal.Zero, false
		}
		return parsed, true
	case float64:
		return decimal.NewFromFloat(typed), true
	case int:
		return decimal.NewFromInt(int64(typed)), true
	case int64:
		return decimal.NewFromInt(typed), true
	}
	return decimal.Zero, false
}
