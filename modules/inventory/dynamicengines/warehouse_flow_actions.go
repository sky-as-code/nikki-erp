package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// The flow reconfigurations reach the application service rather than a domain one, because each
// writes the warehouse and provisions its locations together.
//
// Neither creates a stock move. A flow is policy for transactions made from now on: a receipt
// already under way keeps the shape it was created with, and existing quants are untouched.

// warehouseAppService resolves the orchestration layer from the container.
//
// The action callbacks receive only their own engine, so unlike the domain services there is no
// handle to type-assert; it is looked up instead. A failure here means Init did not register it,
// which is a wiring bug rather than a request problem.
func warehouseAppService() (itWarehouse.WarehouseAppService, error) {
	var service itWarehouse.WarehouseAppService
	err := deps.Invoke(func(resolved itWarehouse.WarehouseAppService) { service = resolved })
	if err != nil {
		return nil, errors.Wrap(err,
			"the warehouse application service is not registered; "+
				"InventoryModule.Init must publish it")
	}
	return service, nil
}

func processConfigureIncomingFlow(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := warehouseAppService()
	if err != nil {
		return nil, err
	}
	result, err := service.ConfigureIncomingFlow(ctx, itWarehouse.ConfigureFlowCommand{
		WarehouseId: readStringField(input.Params, paramWarehouseId),
		Flow:        readStringField(input.Params, paramFlow),
	})
	return toFlowActionResult(result, err)
}

func processConfigureOutgoingFlow(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := warehouseAppService()
	if err != nil {
		return nil, err
	}
	result, err := service.ConfigureOutgoingFlow(ctx, itWarehouse.ConfigureFlowCommand{
		WarehouseId: readStringField(input.Params, paramWarehouseId),
		Flow:        readStringField(input.Params, paramFlow),
	})
	return toFlowActionResult(result, err)
}

func toFlowActionResult(
	result *itWarehouse.ConfigureFlowResult, err error,
) (*drif.ActionResult, error) {
	if err != nil {
		return nil, err
	}
	out := &drif.ActionResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	if result.HasData {
		out.Data = dyn.MutateResultData{AffectedCount: result.Data.AffectedCount}
	}
	return out, nil
}
