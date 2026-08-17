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

// The rules behind a product's inventory unit.
//
// Product does not own the unit and does not own conversion: the UoM master, its categories, its
// factors and its rounding all live in Essential, and Inventory holds a reference to one of them
// (CR §11.1, §11.3, PROD-INT-INV-009, PROD-INT-INV-010). What lives here is only the question
// Stock is entitled to answer — may this product's unit still be changed?

// AssertInventoryUomNotInUse refuses a unit change on a product whose stock has been used.
//
// The change is not merely inconvenient at this point: every quantity ever recorded for the
// product was written in the old unit, and reinterpreting them is a data migration rather than an
// edit. The requirement routes that through an explicit administrative operation, out of scope
// here (CR §12.2, §12.3, TS-PROD-09).
func AssertInventoryUomNotInUse(vErrs *ft.ClientErrors) {
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockProductConfigSchemaName,
		"stock_product_config.inventory_uom_in_use",
		"the inventory unit cannot be changed once the product has stock or stock history; "+
			"an administrative migration is required to reinterpret existing quantities",
	))
}

// AssertInventoryUomNotArchived refuses an archived unit for new configuration.
//
// An archived UoM stays resolvable so historical records keep displaying it; what it may not do is
// be adopted afresh (CR §11.7, §11.8, AC-PROD-INT-023).
func AssertInventoryUomNotArchived(vErrs *ft.ClientErrors) {
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockProductConfigSchemaName,
		"stock_product_config.inventory_uom_archived",
		"an archived unit of measure cannot be chosen as an inventory unit",
	))
}

// IsTemplateStockInUse reports whether any of a template's variants has stock or stock history.
//
// History counts, unlike the archive guard where it deliberately does not. The two ask different
// questions: archiving asks "would anything be stranded", which completed movement does not
// affect, while this asks "would anything already recorded change meaning", which is precisely
// what completed movement is (CR §12.1 versus §14.2).
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

	// A quant is enough on its own: one exists only because stock was put there, and it holds a
	// quantity in the unit being changed even when that quantity is now zero.
	hasQuants, err := anyMatching(ctx, models.StockQuantSchemaName,
		models.StockQuantFieldProductVariantId, variantIds)
	if err != nil {
		return false, err
	}
	if hasQuants {
		return true, nil
	}

	// Moves are checked too, and in every state. A cancelled or draft move still names a quantity
	// in the current unit, and a product whose only trace is one of those is not a product that
	// has never been used.
	return anyMatching(ctx, models.StockMoveSchemaName,
		models.StockMoveFieldProductVariantId, variantIds)
}

// IsUomUsable reports whether a unit of measure may be chosen for new configuration.
//
// Inventory does not own the UoM and must not decide this from its own tables. The answer comes
// from Essential's own record: a unit that has been archived is still resolvable for history but
// no longer selectable (CR §11.7, §11.8).
//
// A unit that cannot be found reads as unusable. Treating an unknown id as acceptable would let a
// typo configure a product with a unit that does not exist, and every balance afterwards would be
// expressed in nothing.
func IsUomUsable(ctx corectx.Context, uomId string) (bool, error) {
	if uomId == "" {
		return false, nil
	}

	engine, err := engineFor(uomSchemaName)
	if err != nil {
		// Essential not being present is a deployment without the UoM module rather than a bad
		// request. Nothing can be validated, so nothing is claimed: the unit passes, and Stock
		// still refuses to move goods it cannot express a quantity for.
		return true, nil
	}

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: uomByIdGraph(uomId),
		Page:  0,
		Size:  1,
		// nil rather than false: the archived rows are exactly the ones this needs to see, since
		// finding one is how it reports the unit as unusable rather than as missing.
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

// uomSchemaName is Essential's unit-of-measure resource.
//
// Spelled out rather than imported: Inventory reaching into another module's model package would
// couple the two at compile time, which is exactly what the interfaces/ ports exist to avoid. The
// name must match essential/domain/models/uom.go's UomSchemaName verbatim.
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
