package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/external"
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
		ft.ErrorKey("stock_product_config.inventory_uom_in_use", "inventory"),
		"the inventory unit cannot be changed once the product has stock or stock history; "+
			"an administrative migration is required to reinterpret existing quantities",
	))
}

// AssertInventoryUomNotArchived refuses an archived unit for new configuration. An archived UoM
// stays resolvable so historical records keep displaying it, but may not be adopted afresh.
func AssertInventoryUomNotArchived(vErrs *ft.ClientErrors) {
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockProductConfigSchemaName,
		ft.ErrorKey("stock_product_config.inventory_uom_archived", "inventory"),
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

	variantEngine, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return false, err
	}

	rows, err := models.FindTemplateVariants(
		ctx, variantEngine, templateId, MaxCascadeVariants)
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
func IsUomUsable(ctx corectx.Context, uoms itExt.UomLookupExtService, uomId string) (bool, error) {
	if uomId == "" {
		return false, nil
	}
	if uoms == nil {
		// No UoM module bound means nothing can be validated, so the unit passes; Stock still refuses
		// to move goods it cannot express a quantity for.
		return true, nil
	}
	found, err := uoms.GetUom(ctx, itExt.GetUomQuery{Id: model.Id(uomId)})
	if err != nil {
		return false, errors.Wrap(err, "IsUomUsable")
	}
	if found == nil || !found.HasData {
		return false, nil
	}
	return !found.Data.IsArchived, nil
}

// anyMatching reports whether any row of a schema matches one of the given ids.
func anyMatching(
	ctx corectx.Context, schemaName string, field string, ids []string,
) (bool, error) {
	engine, err := repoFor(schemaName)
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
