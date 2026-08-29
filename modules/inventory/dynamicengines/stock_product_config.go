package dynamicengines

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// Stock's settings for a product line: today, the unit its stock is counted in. Changing that unit
// after stock has moved would silently reinterpret every quantity ever recorded — a balance of
// "100" units becomes 100 dozen — so once a product has been used, an ordinary update must refuse.

func stockProductConfigEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.StockProductConfigSchemaName,
		DefineActions: defineStockProductConfigActions,
	}
}

// defineStockProductConfigActions attaches the inventory-unit guards to create and update. Create
// checks the UoM may still be adopted; update adds the in-use check on top, so a product whose
// stock has moved cannot have the meaning of its recorded quantities changed underneath it.
//
// The update declares KeysToFetch so the engine hands the stored row to the guard; otherwise the
// guard would compare against whatever it fetched itself rather than what the engine is about to
// overwrite.
func defineStockProductConfigActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.ModifyAction(drif.DynamicActionDelta{
			ActionName:    drif.ActionCreate,
			ValidateExtra: validateInventoryUomSelectable,
		}),
		engine.ModifyAction(drif.DynamicActionDelta{
			ActionName:    drif.ActionUpdate,
			KeysToFetch:   stockProductConfigKeysToFetch,
			ValidateExtra: validateInventoryUomChange,
		}),
	)
}

// stockProductConfigKeysToFetch names the record the update guard compares against. Required, not
// an optimisation: without it the engine passes a nil foundModel and every change is waved through,
// so the in-use rule would never fire.
func stockProductConfigKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.StockProductConfigFieldId: params[models.StockProductConfigFieldId],
	}
}

// validateInventoryUomSelectable refuses a UoM that may no longer be adopted. Only archived-ness
// is checked: whether the unit exists and what it converts to are Essential's to answer, and this
// asks rather than reimplementing any of it.
func validateInventoryUomSelectable(
	ctx corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	uomId := readStringField(params, models.StockProductConfigFieldInventoryUomId)
	if uomId == "" {
		return nil
	}
	return checkUomUsable(ctx, uomId, vErrs)
}

// validateInventoryUomChange refuses a change of unit once the product's stock has been used. A
// product that has never moved stock may be reconfigured freely — nothing was recorded in the old
// unit — but once it has, the change must go through an explicit administrative migration, which
// is deliberately not this action. An update leaving the unit alone passes untouched, so editing
// anything else on the row stays possible however much stock has moved.
func validateInventoryUomChange(
	ctx corectx.Context,
	inputModel *drif.DynamicEntity,
	foundModel *drif.DynamicEntity,
	vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	newUomId := readStringField(params, models.StockProductConfigFieldInventoryUomId)
	if newUomId == "" {
		return nil
	}
	// No stored row means the record is gone, which the engine's own update reports better than a
	// guard inventing a message for it.
	if foundModel == nil {
		return nil
	}

	stored := models.NewStockProductConfigFrom(foundModel.GetFieldData())
	if derefId(stored.GetInventoryUomId()) == newUomId {
		return nil
	}

	if err := checkUomUsable(ctx, newUomId, vErrs); err != nil || vErrs.Count() > 0 {
		return err
	}

	inUse, err := services.IsTemplateStockInUse(ctx, derefId(stored.GetProductTemplateId()))
	if err != nil {
		return err
	}
	if inUse {
		services.AssertInventoryUomNotInUse(vErrs)
	}
	return nil
}

// checkUomUsable asks Essential whether a unit may still be chosen.
func checkUomUsable(ctx corectx.Context, uomId string, vErrs *ft.ClientErrors) error {
	usable, err := services.IsUomUsable(ctx, uomId)
	if err != nil {
		return err
	}
	if !usable {
		services.AssertInventoryUomNotArchived(vErrs)
	}
	return nil
}
