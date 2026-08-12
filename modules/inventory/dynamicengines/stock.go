package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// Stock Location and Stock Operation Type are plain CRUD master data: everything they need is
// already expressed by their schema. Stock Quant is not — it is current state rather than a
// document, so its engine takes its write actions away.

func stockLocationEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StockLocationSchemaName,
		DefaultFields: []string{
			models.StockLocationFieldCode,
			models.StockLocationFieldName,
			models.StockLocationFieldLocationType,
			models.StockLocationFieldParentLocationId,
		},
	}
}

func stockOperationTypeEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StockOperationTypeSchemaName,
		DefaultFields: []string{
			models.StockOperationTypeFieldCode,
			models.StockOperationTypeFieldName,
			models.StockOperationTypeFieldOperationCode,
			models.StockOperationTypeFieldReservationMethod,
			models.StockOperationTypeFieldBackorderPolicy,
		},
	}
}

func stockQuantEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.StockQuantSchemaName,
		DefaultFields: []string{
			models.StockQuantFieldProductVariantId,
			models.StockQuantFieldLocationId,
			models.StockQuantFieldLotRef,
			models.StockQuantFieldOnHandQuantity,
			models.StockQuantFieldReservedQuantity,
			models.StockQuantFieldAvailableQuantity,
			models.StockQuantFieldBaseUomId,
		},
		DefineActions: defineStockQuantActions,
	}
}

// defineStockQuantActions closes the quant's write surface.
//
// A balance is not something a client sets; it is the running total of the movements that have
// completed against it. Leaving create, update and delete open would allow an on-hand quantity
// with no movement behind it, which no report could explain and no audit could trace. Corrections
// go through an inventory adjustment, a transfer or a scrap. See BR §3.3, §4.2.2.6, AC-STOCK-002.
//
// The actions are refused rather than removed so that a caller gets a 400 naming the reason,
// instead of a 404 that reads as "wrong URL".
func defineStockQuantActions(engine drif.DynamicResourceEngine) error {
	for _, action := range []string{drif.ActionCreate, drif.ActionUpdate, drif.ActionDelete} {
		err := engine.ModifyAction(drif.DynamicActionDelta{
			ActionName:    action,
			ValidateExtra: rejectQuantWrite,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to attach the stock quant '%s' guard", action)
		}
	}
	return nil
}

func rejectQuantWrite(
	_ corectx.Context, _ dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	services.AssertQuantNotClientWritable(vErrs)
	return nil
}
