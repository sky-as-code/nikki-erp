package app

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The custom actions receive the request as the field map the REST engine bound. These helpers
// read what each action needs out of it; a missing or mistyped value reads as absent, and the
// domain service decides whether absent is acceptable.
//
// They were moved here verbatim from dynamicengines/lifecycle_actions.go and
// dynamicengines/voucher_actions.go, which owned them while the actions lived in the engine
// definitions. Parsing belongs beside authorization: both are things the application layer does
// to a request before a rule ever sees it.

// paramRecordId is the path parameter every id-addressed action carries.
const paramRecordId = "id"

const (
	paramCode            = "code"
	paramSoldToPartyId   = "sold_to_party_id"
	paramBillToPartyId   = "bill_to_party_id"
	paramPayerPartyId    = "payer_party_id"
	paramPaymentMethodId = "payment_method_id"

	// Read by the fulfillment and availability actions.
	paramItems            = "items"
	paramProductVariantId = "product_variant_id"
	paramQuantity         = "quantity"
	paramOrgId            = "org_id"

	reasonItemsMalformed    = "sales_order_fulfillment.items_malformed"
	reasonQuantityMalformed = "sales_order_fulfillment.quantity_malformed"
	paramVoucherCode        = "code"
	paramNowUnix            = "now_unix"
)

// readBoolParam reads a flag.
//
// A checkbox that did not arrive is an unticked one, and defaulting to false is the conservative
// reading for every flag that grants rather than restricts.
func readBoolParam(params dmodel.DynamicFields, field string) bool {
	value, ok := params[field]
	if !ok || value == nil {
		return false
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	if typed, ok := value.(*bool); ok && typed != nil {
		return *typed
	}
	return false
}

func readStringParam(params dmodel.DynamicFields, field string) string {
	value, ok := params[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	if typed, ok := value.(*string); ok && typed != nil {
		return *typed
	}
	return ""
}

func readStringsParam(params map[string]any, field string) []string {
	raw, ok := params[field].([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		if typed, ok := item.(string); ok {
			values = append(values, typed)
		}
	}
	return values
}

// readOptionalDecimalParam keeps the difference between "absent" and "zero", which
// readDecimalParam cannot: a reconciliation field has to keep that difference, because a missing
// estimate is not a claim that the price was zero.
func readOptionalDecimalParam(params map[string]any, field string) *decimal.Decimal {
	value, ok := params[field]
	if !ok || value == nil {
		return nil
	}
	if text, isText := value.(string); isText && text == "" {
		return nil
	}
	parsed := readDecimalParam(params, field)
	return &parsed
}

// readDecimalParam accepts whatever shape JSON delivered. A decimal crosses as a string so it does
// not lose precision; a float64 is accepted so a bare number does not silently become zero.
func readDecimalParam(params map[string]any, field string) decimal.Decimal {
	value, ok := params[field]
	if !ok || value == nil {
		return decimal.Zero
	}
	switch typed := value.(type) {
	case string:
		if parsed, err := decimal.NewFromString(typed); err == nil {
			return parsed
		}
	case float64:
		return decimal.NewFromFloat(typed)
	case int:
		return decimal.NewFromInt(int64(typed))
	case int64:
		return decimal.NewFromInt(typed)
	}
	return decimal.Zero
}

// assertRecordAction is the authorization every id-addressed custom action performs: the org
// scope and permission, then that the named record belongs to that org.
//
// Collection-level actions call AssertAction directly instead: there is no record yet to place in
// an org.
func assertRecordAction(
	appSvc composable.CrudApplicationService, ctx corectx.Context, permission string, params dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	orgId, cErrs := appSvc.AssertAction(ctx, permission, params)
	if cErrs != nil {
		return cErrs, nil
	}
	if orgId == nil {
		return nil, nil
	}
	return appSvc.AssertRecordInOrg(ctx, params, *orgId)
}

// The failure shapers keep the module's error contract: a business refusal travels as
// ClientErrors with a nil Go error and answers 400, while a genuine fault travels as err and
// answers 500. Returning the wrong one turns a fixable refusal into an outage or hides a bug.

func mutateFailure(cErrs *ft.ClientErrors, err error) (*dyn.OpResult[dyn.MutateResultData], error) {
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *cErrs}, nil
}

func anyFailure(cErrs *ft.ClientErrors, err error) (*dyn.OpResult[any], error) {
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[any]{ClientErrors: *cErrs}, nil
}

// anyResult lifts a typed read into the untyped result a custom route answers with.
func anyResult(data any) *dyn.OpResult[any] {
	return &dyn.OpResult[any]{Data: data, HasData: true}
}

// Shared by the fulfillment and availability readers.

// readDecimalValue accepts every numeric shape a JSON body can arrive in: a whole number comes back
// from the decoder as a float64, and a client sending an exact decimal sends a string.
func readDecimalValue(value any) (decimal.Decimal, bool) {
	switch typed := value.(type) {
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

func itemsViolation(key, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesOrderFulfillmentSchemaName, key, message))
	return vErrs
}
