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
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// engineFor resolves a resource's engine from the registry.
//
// It is a variable rather than a plain function so that a test can supply its own engines: the
// registry is a package singleton populated during Init, which a unit test has no way to build.
// This mirrors inventory/domain/services/variant_domservice.go.
var engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for '%s'", schemaName)
	}
	return engine, nil
}

// EngineFor exposes the registry lookup to the dynamicengines package, whose action callbacks
// receive only their own engine but need another resource's repository.
func EngineFor(schemaName string) (drif.DynamicResourceEngine, error) {
	return engineFor(schemaName)
}

// appendOrderViolation records a rule the caller broke, against the order resource.
func appendOrderViolation(vErrs *ft.ClientErrors, key string, message string) {
	vErrs.Append(*ft.NewBusinessViolation(models.OrderSchemaName, key, message))
}

// appendFieldViolation records a rule broken by one named input, so the caller is told which of
// their fields to fix rather than only that something was wrong.
func appendFieldViolation(vErrs *ft.ClientErrors, field string, key string, message string) {
	vErrs.Append(*ft.NewBusinessViolation(field, key, message))
}

// withOrderTransaction runs body inside one transaction on a scoped copy of the context.
//
// The transaction goes on a clone rather than on the caller's context, per docs/wiki/02 §5.1:
// mutating the caller's context would leave a committed transaction visible to whatever runs
// next. CloneRequestContext carries the caller's identity across, which the audit columns need.
func withOrderTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	engine, err := engineFor(models.OrderSchemaName)
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

// findOrderByBusinessId looks an order up by the identifier the caller quotes.
//
// Callers hold order_id, never the primary key: the primary key is this module's, while order_id
// is what the ordering system and support were given.
func findOrderByBusinessId(ctx corectx.Context, orderId string) (*models.Order, error) {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.OrderFieldOrderId, dmodel.Equals, orderId),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findOrderByBusinessId")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewOrderFrom(found.Data.Items[0]), nil
}

// orderCodeExists reports whether a generated code is already taken.
func orderCodeExists(ctx corectx.Context, orderCode string) (bool, error) {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return false, err
	}

	found, err := engine.ResourceRepository().Exists(ctx, []dmodel.DynamicFields{
		{models.OrderFieldOrderCode: orderCode},
	})
	if err != nil {
		return false, errors.Wrap(err, "orderCodeExists")
	}
	if found == nil || !found.HasData {
		return false, nil
	}
	return len(found.Data.Existing) > 0, nil
}

// writeOrderFields updates an order through the repository, bypassing the service overrides.
//
// The status columns are declared no_update precisely so a client cannot set them; the payment
// flow, the callbacks and the watchdog write them here instead, having already checked what the
// transition requires.
func writeOrderFields(ctx corectx.Context, orderPk string, fields dmodel.DynamicFields) error {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.OrderFieldId: orderPk}
	for key, value := range fields {
		update[key] = value
	}
	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeOrderFields")
}

// writeTransactionFields updates a transaction through the repository, for the same reason.
func writeTransactionFields(ctx corectx.Context, transactionPk string, fields dmodel.DynamicFields) error {
	engine, err := engineFor(models.TransactionSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.TransactionFieldId: transactionPk}
	for key, value := range fields {
		update[key] = value
	}
	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeTransactionFields")
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefDecimal(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

func ptrOf[T any](value T) *T {
	return &value
}

// mergeMetadata returns a copy of base with overlay applied, so neither input is mutated.
//
// A nil result would erase what is already on the order, so an empty overlay returns base
// unchanged rather than nil.
func mergeMetadata(base map[string]any, overlay map[string]any) map[string]any {
	merged := map[string]any{}
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range overlay {
		merged[key] = value
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}
