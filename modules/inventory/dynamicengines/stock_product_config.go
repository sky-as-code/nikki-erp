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

// Stock's settings for a product line: today, the unit its stock is counted in.
//
// The guards here are what make changing that unit safe. Changing it after stock has moved would
// silently reinterpret every quantity ever recorded — a balance of "100" meaning 100 units becomes
// 100 dozen — so once a product has been used, an ordinary update must refuse (CR §12).

func stockProductConfigEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StockProductConfigSchemaName,
		DefaultFields: []string{
			models.StockProductConfigFieldProductTemplateId,
			models.StockProductConfigFieldInventoryUomId,
		},
		DefineActions: defineStockProductConfigActions,
	}
}

// defineStockProductConfigActions attaches the inventory-unit guards to create and update.
//
// Both are guarded, for different reasons. On create the unit must be one that may still be
// adopted, so an archived UoM cannot be chosen for new configuration. On update that check still
// applies, and the in-use check is added on top: a product whose stock has already moved may not
// have the meaning of its recorded quantities changed underneath it.
//
// The update declares KeysToFetch so the engine hands the stored row to the guard. Without it the
// guard would have to read the record itself, and would be comparing against whatever it happened
// to fetch rather than against what the engine is about to overwrite.
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

// stockProductConfigKeysToFetch names the record the update guard compares against.
//
// It is required, not an optimisation: without it the engine passes a nil foundModel, the guard
// finds nothing to compare the incoming unit with, and every change would be waved through — the
// in-use rule would exist in the code and never once fire.
func stockProductConfigKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.StockProductConfigFieldId: params[models.StockProductConfigFieldId],
	}
}

// validateInventoryUomSelectable refuses a UoM that may no longer be adopted.
//
// Only archived-ness is checked. Whether the unit exists and what it converts to are Essential's
// to answer, and this asks rather than reimplementing any of it (CR §11.7, §11.9,
// AC-PROD-INT-023, AC-PROD-INT-026).
func validateInventoryUomSelectable(
	ctx corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	uomId := readStringField(params, models.StockProductConfigFieldInventoryUomId)
	if uomId == "" {
		return nil
	}
	return checkUomUsable(ctx, uomId, vErrs)
}

// validateInventoryUomChange refuses a change of unit once the product's stock has been used.
//
// A product that has never moved stock may be reconfigured freely: nothing was recorded in the old
// unit, so nothing changes meaning (CR §12.1, TS-PROD-08). Once it has, the change is refused and
// has to go through an explicit administrative migration, which is deliberately not this action
// (CR §12.2, §12.3, TS-PROD-09).
//
// An update that leaves the unit alone passes untouched, so editing anything else on the row stays
// possible however much stock the product has moved.
func validateInventoryUomChange(
	ctx corectx.Context,
	params dmodel.DynamicFields,
	foundModel *dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) error {
	newUomId := readStringField(params, models.StockProductConfigFieldInventoryUomId)
	if newUomId == "" {
		return nil
	}
	// No stored row means the record is gone, which the engine's own update reports far better
	// than a guard inventing a message for it.
	if foundModel == nil {
		return nil
	}

	stored := models.NewStockProductConfigFrom(*foundModel)
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

