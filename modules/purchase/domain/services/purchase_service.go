// Package services holds the rules of the Purchase module: the totals recomputation, the order and
// agreement lifecycles, and the audit trail that records them.
//
// Everything here is reachable from a dynamicengines action callback, which adapts and validates a
// request; the writes belong in this package. See docs/wiki/07. ERP backend module.md §6.7.
package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// engineFor resolves another resource's engine from the registry.
//
// It is a variable rather than a plain function so that a test can supply its own engines: the
// registry is a package singleton populated during Init, which a unit test has no way to build.
var engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for '%s'", schemaName)
	}
	return engine, nil
}

// EngineFor exposes the registry lookup to the dynamicengines package, whose action callbacks
// receive only their own engine but need another resource's repository to apply a cross-schema
// rule. It delegates to engineFor so a test's substitution is honoured here too.
func EngineFor(schemaName string) (drif.DynamicResourceEngine, error) {
	return engineFor(schemaName)
}

// withOrderTransaction runs body inside one transaction on a scoped copy of the context.
//
// The transaction goes on a clone, never on ctx itself: setting it on the caller's context would
// leave a committed transaction visible to whatever runs next. CloneRequestContext carries the
// caller's identity across, which the audit columns and the audit event's actor both need.
//
// There is no "join an existing transaction" branch, because pgTxClient.BeginTx returns
// ErrTxNested: the lifecycle actions are entry points, and nesting one inside another is a bug
// rather than a case to handle.
func withOrderTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	engine, err := engineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return err
	}

	tranx, err := engine.ResourceRepository().BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "withOrderTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "withOrderTransaction")
}

// orderNotFoundResult is the answer when an operation names an order that does not exist. It is a
// client error rather than a server one: the id came from the caller.
func orderNotFoundResult(orderId string) *dyn.OpResult[dyn.MutateResultData] {
	return orderViolationResult(
		"purchase_order.not_found",
		"no purchase order with id '"+orderId+"'",
	)
}

func orderViolationResult(key, message string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.PurchaseOrderSchemaName, key, message))
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

// mutateOk is the success envelope of a lifecycle operation.
//
// AffectedCount is 1 for the order itself, not a count of the lines and audit events the operation
// touched: the caller asked to act on one order, and reporting the internal write count would make
// the number mean something different for each operation.
func mutateOk() *dyn.OpResult[dyn.MutateResultData] {
	return &dyn.OpResult[dyn.MutateResultData]{
		Data:    dyn.MutateResultData{AffectedCount: 1},
		HasData: true,
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefBool(value *bool) bool {
	return value != nil && *value
}

// decimalOf reads a decimal field, treating absent and null alike as zero.
//
// It does not use DynamicFields.GetDecimal, which type-asserts without checking and so panics on a
// value the repository handed back in another shape. A total is summed from whatever the database
// returns, and a panic in that path would take down the request rather than reporting a bad row.
func decimalOf(fields dmodel.DynamicFields, key string) decimal.Decimal {
	value, ok := fields[key]
	if !ok || value == nil {
		return decimal.Zero
	}
	switch typed := value.(type) {
	case decimal.Decimal:
		return typed
	case *decimal.Decimal:
		if typed == nil {
			return decimal.Zero
		}
		return *typed
	case string:
		parsed, err := decimal.NewFromString(typed)
		if err != nil {
			return decimal.Zero
		}
		return parsed
	case float64:
		return decimal.NewFromFloat(typed)
	case int64:
		return decimal.NewFromInt(typed)
	default:
		return decimal.Zero
	}
}

func stringOf(fields dmodel.DynamicFields, key string) string {
	value, ok := fields[key]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	// model.Id and the enum types are string kinds, so a fmt-free fallback keeps this total.
	if typed, ok := value.(interface{ String() string }); ok {
		return typed.String()
	}
	return ""
}

func boolOf(fields dmodel.DynamicFields, key string) bool {
	value, ok := fields[key]
	if !ok || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case *bool:
		return typed != nil && *typed
	default:
		return false
	}
}

// newOrderErrors is the client-error collector the lifecycle operations report refusals through.
func newOrderErrors() *ft.ClientErrors {
	return ft.NewClientErrors()
}
