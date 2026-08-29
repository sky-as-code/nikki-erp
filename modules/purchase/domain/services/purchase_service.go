// Package services holds the rules of the Purchase module: the totals recomputation, the order and
// agreement lifecycles, and the audit trail that records them. Everything here is reached from a
// dynamicengines action callback, which adapts and validates the request; the writes belong here.
package services

import (
	"time"

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

// engineFor resolves another resource's engine from the registry. It is a variable so a test can
// substitute its own engines; the registry is a package singleton populated during Init.
var engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for '%s'", schemaName)
	}
	return engine, nil
}

// EngineFor exposes the registry lookup to dynamicengines, whose action callbacks receive only
// their own engine. It delegates to engineFor so a test's substitution is honoured here too.
func EngineFor(schemaName string) (drif.DynamicResourceEngine, error) {
	return engineFor(schemaName)
}

// withOrderTransaction runs body inside one transaction on a cloned context. The transaction must
// go on the clone, never on ctx itself, or a committed transaction stays visible to whatever runs
// next; the clone carries the caller's identity, which the audit columns and actor need. There is
// no join-existing-transaction branch: BeginTx returns ErrTxNested, and nesting these entry points
// is a bug rather than a case to handle.
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

// orderNotFoundResult reports a missing order as a client error, since the id came from the caller.
func orderNotFoundResult(orderId string) *dyn.OpResult[dyn.MutateResultData] {
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *orderNotFoundErrors(orderId)}
}

func orderViolationResult(key, message string) *dyn.OpResult[dyn.MutateResultData] {
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *orderViolationErrors(key, message)}
}

// The builders above wrap these, which return the errors alone. Split because not every order
// operation returns a MutateResultData: reprice needs the same refusals in a different envelope.
func orderNotFoundErrors(orderId string) *ft.ClientErrors {
	return orderViolationErrors(
		"purchase_order.not_found",
		"no purchase order with id '"+orderId+"'",
	)
}

func orderViolationErrors(key, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.PurchaseOrderSchemaName, key, message))
	return vErrs
}

// mutateOk is the success envelope of a lifecycle operation. AffectedCount is 1 for the order
// itself, never a count of the lines and audit events written along the way.
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

// decimalOf reads a decimal field, treating absent and null alike as zero. It avoids
// DynamicFields.GetDecimal, which type-asserts unchecked and panics on a value the repository
// returned in another shape.
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

func newOrderErrors() *ft.ClientErrors {
	return ft.NewClientErrors()
}

// timeNow is the clock the pricing path reads, as a variable so a test can pin it against fixed
// validity windows. It is the only clock seam: vendorpricing itself has none.
var timeNow = time.Now
