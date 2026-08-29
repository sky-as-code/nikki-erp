package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// Setting the default pricelist is not a plain PATCH of is_default: it also demotes the previous
// default, and only an operation owning both writes can keep exactly one default per org.

const (
	// Deliberately `update` rather than a permission of its own: choosing the default list is the
	// same class of power as editing what a list charges.
	PermissionSetDefaultPricelist = "update"

	ActionSetDefaultPricelist = "set_default"
)

func defineSalesPricelistActions(engine drif.DynamicResourceEngine) error {
	return engine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  ActionSetDefaultPricelist,
		ActionType:  drif.ActionTypeGeneric,
		RestPath:    ":id/set_default",
		Permission:  PermissionSetDefaultPricelist,
		MainProcess: processSetDefaultPricelist,
	})
}

func processSetDefaultPricelist(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := pricelistServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.SetDefault(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

// pricelistServiceOf reaches the derived service installed during Init. A failed assertion is a
// wiring bug, so it answers a Go error (500) rather than a business violation.
func pricelistServiceOf(input drif.ProcessInput) (*services.SalesPricelistDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.SalesPricelistDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the sales pricelist engine is not running the derived pricelist service; " +
				"SalesModule.Init must install it with SetResourceService")
	}
	return service, nil
}
