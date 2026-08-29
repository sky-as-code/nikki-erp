package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The rules behind a product's inventory unit. Inventory holds only a reference: the UoM master,
// its categories, conversion factors and rounding all live in Essential. What lives here is the one
// question Stock can answer — may this product's unit still be changed?

// AssertInventoryUomNotInUse refuses a unit change on a product whose stock has been used: every
// quantity ever recorded was written in the old unit, so reinterpreting them is a data migration
// rather than an edit.
func AssertInventoryUomNotInUse(vErrs *ft.ClientErrors) {
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockProductConfigSchemaName,
		"stock_product_config.inventory_uom_in_use",
		"the inventory unit cannot be changed once the product has stock or stock history; "+
			"an administrative migration is required to reinterpret existing quantities",
	))
}

// AssertInventoryUomNotArchived refuses an archived unit for new configuration. An archived UoM
// stays resolvable so historical records keep displaying it, but may not be adopted afresh.
func AssertInventoryUomNotArchived(vErrs *ft.ClientErrors) {
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockProductConfigSchemaName,
		"stock_product_config.inventory_uom_archived",
		"an archived unit of measure cannot be chosen as an inventory unit",
	))
}

// IsTemplateStockInUse reports whether any of a template's variants has stock or stock history.
// History counts here, unlike in the archive guard: archiving asks whether anything would be
// stranded, while this asks whether anything already recorded would change meaning.
func IsTemplateStockInUse(ctx corectx.Context, templateId string) (bool, error) {
	if templateId == "" {
		return false, nil
	}

	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return false, err
	}

	rows, err := models.FindTemplateVariants(
		ctx, variantEngine.ResourceRepository(), templateId, MaxCascadeVariants)
	if err != nil {
		return false, errors.Wrap(err, "IsTemplateStockInUse")
	}

	variantIds := make([]string, 0, len(rows))
	for _, row := range rows {
		variant := models.NewProductVariantFrom(row)
		if id := derefId(variant.GetId()); id != "" {
			variantIds = append(variantIds, id)
		}
	}
	if len(variantIds) == 0 {
		return false, nil
	}

	// A quant is enough on its own: it exists only because stock was put there, and it holds a
	// quantity in the unit being changed even when that quantity is now zero.
	hasQuants, err := anyMatching(ctx, models.StockQuantSchemaName,
		models.StockQuantFieldProductVariantId, variantIds)
	if err != nil {
		return false, err
	}
	if hasQuants {
		return true, nil
	}

	// Moves are checked in every state: a cancelled or draft move still names a quantity in the
	// current unit.
	return anyMatching(ctx, models.StockMoveSchemaName,
		models.StockMoveFieldProductVariantId, variantIds)
}

// IsUomUsable reports whether a unit of measure may be chosen for new configuration. The answer
// comes from Essential's record, not Inventory's tables: an archived unit stays resolvable for
// history but is no longer selectable.
//
// A unit that cannot be found reads as unusable, or a typo would configure a product with a unit
// that does not exist and every balance afterwards would be expressed in nothing.
func IsUomUsable(ctx corectx.Context, uomId string) (bool, error) {
	if uomId == "" {
		return false, nil
	}

	engine, err := engineFor(uomSchemaName)
	if err != nil {
		// Essential missing means a deployment without the UoM module, not a bad request. Nothing can
		// be validated, so the unit passes; Stock still refuses to move goods it cannot express a
		// quantity for.
		return true, nil
	}

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: uomByIdGraph(uomId),
		Page:  0,
		Size:  1,
		// nil rather than false: the archived rows are the ones this needs to see, since finding one
		// is how it reports the unit as unusable rather than missing.
		IncludeArchived: nil,
	})
	if err != nil {
		return false, errors.Wrap(err, "IsUomUsable")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return false, nil
	}

	row := found.Data.Items[0]
	archived, _ := row[basemodel.FieldIsArchived].(bool)
	return !archived, nil
}

// uomSchemaName is Essential's unit-of-measure resource. Spelled out rather than imported, which
// would couple the modules at compile time; it must match essential/domain/models/uom.go's
// UomSchemaName verbatim.
const uomSchemaName = "essential_uom"

func uomByIdGraph(uomId string) *dmodel.SearchGraph {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldId, dmodel.Equals, uomId),
	)
	return graph
}

// anyMatching reports whether any row of a schema matches one of the given ids.
func anyMatching(
	ctx corectx.Context, schemaName string, field string, ids []string,
) (bool, error) {
	engine, err := engineFor(schemaName)
	if err != nil {
		return false, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(field, dmodel.In, toAnySlice(ids)...),
	)

	total, err := countMatching(ctx, engine, graph)
	if err != nil {
		return false, errors.Wrapf(err, "anyMatching(%s)", schemaName)
	}
	return total > 0, nil
}
