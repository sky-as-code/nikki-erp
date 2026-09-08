package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// A fulfillment method has only the archive pair. There is no suspend: unlike a channel, which can
// stop selling while its points stay configured, a method is either offered to new orders or it is
// not — and either way the fulfillments that snapshotted it keep running, which is the whole point
// of the snapshot. Both actions ride on the built-in set_archived permission, because unarchiving is
// the same power in reverse and splitting them would let a role retire a policy it could not restore.
func defineSalesFulfillmentMethodActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionArchive,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/archive",
			Permission:  drif.PermissionSetArchived,
			MainProcess: processFulfillmentMethodArchive,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionUnarchive,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/unarchive",
			Permission:  drif.PermissionSetArchived,
			MainProcess: processFulfillmentMethodUnarchive,
		}),
	)
}

func processFulfillmentMethodArchive(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := fulfillmentMethodServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Archive(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

func processFulfillmentMethodUnarchive(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := fulfillmentMethodServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Unarchive(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

// fulfillmentMethodServiceOf asserts to the derived type, which alone carries Archive and
// Unarchive. A failed assertion is a wiring bug rather than a bad request, so it answers a Go error.
func fulfillmentMethodServiceOf(
	input drif.ProcessInput,
) (*services.SalesFulfillmentMethodDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.SalesFulfillmentMethodDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the sales fulfillment method engine is not running the derived method service; " +
				"SalesModule.Init must install it with SetResourceService")
	}
	return service, nil
}
