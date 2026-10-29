package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// A group location (a warehouse's root, or any node a user lays out only to organise others) holds
// no stock. A move or transfer naming one as its source or destination would put goods on a node
// that availability, putaway and counts all ignore, so it is refused where moves and transfers are
// written.

// movementLocationFields are the location fields a move and a transfer share.
var movementLocationFields = []string{
	models.StockMoveFieldSourceLocationId,
	models.StockMoveFieldDestinationLocationId,
}

// validateMovementLocationsHoldStock is a ValidateExtra hook for moves and transfers. On update it
// reads only the location fields the input carries.
func validateMovementLocationsHoldStock(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	return assertLocationsHoldStock(ctx, inputModel.GetFieldData(), vErrs)
}

// assertLocationsHoldStock appends inventory_location.group_holds_no_stock for every movement
// location field in params naming a group location.
func assertLocationsHoldStock(ctx corectx.Context, params dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
	for _, field := range movementLocationFields {
		locationId := readStringParam(params, field)
		if locationId == "" {
			continue
		}
		isGroup, err := isGroupLocation(ctx, locationId)
		if err != nil {
			return err
		}
		if isGroup {
			appendLocationViolation(vErrs, field, "inventory_location.group_holds_no_stock",
				"a group location holds no stock; choose one of the locations inside it")
		}
	}
	return nil
}

// isGroupLocation reports whether the location exists and is of type group. A missing location is
// not this guard's to report: the FK and the other validations do.
func isGroupLocation(ctx corectx.Context, locationId string) (bool, error) {
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return false, err
	}
	found, err := engine.FindByKeys(ctx, dmodel.DynamicFields{models.InventoryLocationFieldId: locationId})
	if err != nil {
		return false, errors.Wrap(err, "isGroupLocation")
	}
	if found == nil || !found.HasData {
		return false, nil
	}
	return models.NewInventoryLocationFrom(found.Data).IsGroup(), nil
}
