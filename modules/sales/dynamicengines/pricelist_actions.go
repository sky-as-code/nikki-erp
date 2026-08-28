package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// The pricelist's one operation that is not CRUD.
//
// Making a list the organization's default is not an update to that list, even though it writes a
// field on it: it also demotes whichever list held the place before. Exposing it as a plain PATCH
// of is_default would let a client set two lists to true, or none, and no single-record update
// could prevent it — the rule spans rows, and only an operation that owns both writes can hold it.

const (
	// PermissionSetDefaultPricelist is `update`, not a permission of its own.
	//
	// Choosing which list prices an order that named none is the same class of power as editing
	// what a list charges, and whoever may do one may sensibly do the other. That is unlike confirm
	// or cancel on an order, which commit the business to something no edit can, and so earn
	// permissions of their own.
	PermissionSetDefaultPricelist = "update"

	ActionSetDefaultPricelist = "set_default"
)

// defineSalesPricelistActions adds set_default to the pricelist engine.
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

// pricelistServiceOf reaches the derived service the module installed during Init.
//
// A failed assertion means Init did not install it — a wiring bug rather than a request problem, so
// it answers with a Go error and a 500 rather than a business violation the caller cannot act on.
func pricelistServiceOf(input drif.ProcessInput) (*services.SalesPricelistDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.SalesPricelistDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the sales pricelist engine is not running the derived pricelist service; " +
				"SalesModule.Init must install it with SetResourceService")
	}
	return service, nil
}
