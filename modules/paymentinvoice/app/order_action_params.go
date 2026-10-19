package app

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// The request parsing of the order actions. The actions declare no param schema — the params are a
// mix of order fields and method-specific input that no single schema describes — so every shape
// problem is checked here and reported as a client error naming the field.

// Parameter names of the three actions. They are snake_case to match the resource fields, and
// amount is deliberately the same name it has on the order.
const (
	paramOrderId   = "order_id"
	paramOrderCode = "order_code"
	paramAmount    = "amount"
	paramSource    = "source"
	paramOrgId     = "org_id"

	// A payment method may be named by id or by code. The code is what the standalone service was
	// called with, so a caller migrating off it keeps working without learning our ids.
	paramPaymentMethodId   = "payment_method_id"
	paramPaymentMethodCode = "payment_method_code"

	// paramPaymentProfileId names the merchant account to collect into. Optional: a request
	// without it is collected with the credentials in this deployment's configuration, which is
	// what every caller did before profiles existed.
	paramPaymentProfileId = "payment_profile_id"
	paramContent          = "content"
	paramReturnUrl        = "return_url"
	paramMetadata         = "metadata"
	paramPosId            = "pos_id"
)

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
		OrgId:             readString(params, paramOrgId),
		Source:            readString(params, paramSource),
		PaymentMethodId:   readString(params, paramPaymentMethodId),
		PaymentMethodCode: readString(params, paramPaymentMethodCode),
		PaymentProfileId:  readString(params, paramPaymentProfileId),
		Content:           readOptionalString(params, paramContent),
		ReturnUrl:         readOptionalString(params, paramReturnUrl),
		Metadata:          buildCreateMetadata(params),
	}

	if cmd.PaymentMethodId == "" && cmd.PaymentMethodCode == "" {
		vErrs.Append(*ft.NewBusinessViolation(paramPaymentMethodId,
			ft.ErrorKey("paymentinvoice.payment_method_required", "paymentinvoice"),
			"a payment method must be identified, by id or by code"))
	}

	// Every record this module writes carries an organization, and it cannot be derived from the
	// request: a caller may belong to several. Refusing here names the missing field, where the
	// schema would otherwise reject the composed record as a server fault.
	if cmd.OrgId == "" {
		vErrs.Append(*ft.NewBusinessViolation(paramOrgId,
			ft.ErrorKey("paymentinvoice.org_required", "paymentinvoice"),
			"the organization the order belongs to must be identified"))
	}

	amount, ok := readDecimal(params, paramAmount)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(paramAmount,
			ft.ErrorKey("paymentinvoice.amount_malformed", "paymentinvoice"),
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
		OrderId:   readString(params, paramOrderId),
		OrderCode: readString(params, paramOrderCode),
		Content:   readOptionalString(params, paramContent),
	}

	if cmd.OrderId == "" && cmd.OrderCode == "" {
		vErrs.Append(*ft.NewBusinessViolation(paramOrderId,
			ft.ErrorKey("paymentinvoice.order_id_required", "paymentinvoice"),
			"the order to refund must be identified, by order_id or by order_code"))
	}

	amount, ok := readDecimal(params, paramAmount)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(paramAmount,
			ft.ErrorKey("paymentinvoice.amount_malformed", "paymentinvoice"),
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
