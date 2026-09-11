package app

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The custom actions receive the request as the field map the REST engine bound. These helpers
// read what each action needs out of it; a missing or mistyped value reads as absent, and the
// domain service decides whether absent is acceptable.

// paramRecordId is the path parameter every id-addressed action carries.
const paramRecordId = "id"

func readOptionalBool(params dmodel.DynamicFields, field string) *bool {
	value, ok := composable.ParamBool(params, field)
	if !ok {
		return nil
	}
	return &value
}

// readStringSliceField reads a list of ids. A decoded JSON array arrives as []any of strings;
// []string is accepted too for params built in Go. Anything else reads as absent rather than
// erroring: an unparseable list means there is nothing to act on.
func readStringSliceField(params dmodel.DynamicFields, field string) []string {
	value, ok := params[field]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok && text != "" {
				result = append(result, text)
			}
		}
		return result
	}
	return nil
}

// readDecimalField reads a quantity. Quantities travel as strings so a float's rounding never
// reaches a balance, but a JSON client may still send a bare number; both are parsed through
// decimal, the only representation the rest of the module works in.
func readDecimalField(params dmodel.DynamicFields, schemaName string, field string) (decimal.Decimal, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	value, ok := params[field]
	if !ok || value == nil {
		vErrs.Append(*ft.NewBusinessViolation(schemaName, schemaName+"."+field+"_required", "'"+field+"' is required"))
		return decimal.Zero, vErrs
	}
	parsed, ok := toDecimal(value)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(schemaName, schemaName+"."+field+"_malformed", "'"+field+"' must be a decimal number"))
		return decimal.Zero, vErrs
	}
	return parsed, vErrs
}

func toDecimal(value any) (decimal.Decimal, bool) {
	switch typed := value.(type) {
	case string:
		parsed, err := decimal.NewFromString(typed)
		return parsed, err == nil
	case float64:
		return decimal.NewFromFloat(typed), true
	case int:
		return decimal.NewFromInt(int64(typed)), true
	case int64:
		return decimal.NewFromInt(typed), true
	case decimal.Decimal:
		return typed, true
	}
	return decimal.Zero, false
}

func derefId(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

// assertRecordAction is the authorization every id-addressed custom action performs: the org
// scope and permission, then that the named record belongs to that org.
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
