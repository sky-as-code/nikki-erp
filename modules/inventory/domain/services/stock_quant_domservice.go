package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// NewStockQuantDomainService derives the quant service from the engine's default one, which it
// embeds so built-in actions keep running unchanged.
//
// available_quantity is a computed field declared in stock_quant.json, evaluated by the engine on
// every read.
func NewStockQuantDomainService(
	base composable.CrudDomainService, usageDispatcher *usagecheck.Dispatcher,
) *StockQuantDomainServiceImpl {
	return &StockQuantDomainServiceImpl{
		CrudDomainService: base,
		usageDispatcher:   usageDispatcher,
	}
}

// StockQuantDomainServiceImpl carries the quant's domain behaviors: counting, adjustment,
// reservation and location-usage reads.
type StockQuantDomainServiceImpl struct {
	composable.CrudDomainService

	usageDispatcher *usagecheck.Dispatcher
}

var _ composable.CrudDomainService = (*StockQuantDomainServiceImpl)(nil)

// Delete refuses a balance another module still names.
//
// A quant used to be reachable only through its location, so nothing outside Inventory could hold
// one and deleting it broke nothing. That changed when a vending slot began storing the id of the
// balance it holds: the row IS that slot's contents, and deleting it leaves the machine pointing
// at nothing — the slot would read as empty, restock would have nowhere to land, and no error
// would say why.
//
// Asked rather than enforced by a foreign key alone, for the reason doc 08 gives: the check names
// the module and the record that is still using it, while a constraint violation names a column.
func (this *StockQuantDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	vErrs := ft.NewClientErrors()
	if err := assertQuantDeletable(ctx, this.usageDispatcher, cmd, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, cmd, options...)
}

// assertQuantDeletable asks every module holding a stock quant id whether this balance is still in
// use. A nil dispatcher or a missing id is not an excuse to skip the question; it is only the
// absence of anything to ask about.
func assertQuantDeletable(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher,
	params dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	quantId := derefString(models.NewStockQuantFrom(params).GetId())
	if quantId == "" {
		return nil
	}

	return usagecheck.AssertDeletable(ctx, dispatcher, modconstants.InventoryModuleName,
		models.StockQuantFieldId, usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceStockQuant,
			Identifier: usagecheck.NewIdentifier(
				quantId, readStringParam(params, basemodel.FieldOrgId)),
		}, vErrs)
}
