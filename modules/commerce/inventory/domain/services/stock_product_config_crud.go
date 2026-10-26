package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/external"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Stock's settings for a product line: today, the unit its stock is counted in. Changing that
// unit after stock has moved would silently reinterpret every quantity ever recorded, so once a
// product has been used, an ordinary update must refuse.

// NewStockProductConfigDomainService attaches the inventory-unit guards to create and update.
// Create checks the UoM may still be adopted; update adds the in-use check on top.
func NewStockProductConfigDomainService(
	base composable.CrudDomainService, uoms itExt.UomLookupExtService,
) itStock.StockProductConfigDomainService {
	return &StockProductConfigDomainServiceImpl{CrudDomainService: base, uoms: uoms}
}

type StockProductConfigDomainServiceImpl struct {
	composable.CrudDomainService
	uoms itExt.UomLookupExtService
}

func (this *StockProductConfigDomainServiceImpl) Create(
	ctx corectx.Context, cmd itStock.CreateStockProductConfigCommand, options ...composable.CreateOptions,
) (*itStock.CreateStockProductConfigResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateInventoryUomSelectable)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *StockProductConfigDomainServiceImpl) Update(
	ctx corectx.Context, cmd itStock.UpdateStockProductConfigCommand, options ...composable.UpdateOptions,
) (*itStock.UpdateStockProductConfigResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateInventoryUomChange)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// validateInventoryUomSelectable refuses a UoM that may no longer be adopted. Only archived-ness
// is checked: whether the unit exists and what it converts to are Essential's to answer.
func (this *StockProductConfigDomainServiceImpl) validateInventoryUomSelectable(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	uomId := readStringParam(inputModel.GetFieldData(), models.StockProductConfigFieldInventoryUomId)
	if uomId == "" {
		return nil
	}
	return this.checkUomUsable(ctx, uomId, vErrs)
}

// validateInventoryUomChange refuses a change of unit once the product's stock has been used. A
// product that has never moved stock may be reconfigured freely; once it has, the change must go
// through an explicit administrative migration, which is deliberately not this action. An update
// leaving the unit alone passes untouched.
func (this *StockProductConfigDomainServiceImpl) validateInventoryUomChange(
	ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	newUomId := readStringParam(inputModel.GetFieldData(), models.StockProductConfigFieldInventoryUomId)
	if newUomId == "" || foundModel == nil {
		return nil
	}

	stored := models.NewStockProductConfigFrom(foundModel.GetFieldData())
	if derefString(stored.GetInventoryUomId()) == newUomId {
		return nil
	}
	if err := this.checkUomUsable(ctx, newUomId, vErrs); err != nil || vErrs.Count() > 0 {
		return err
	}

	inUse, err := IsTemplateStockInUse(ctx, derefString(stored.GetProductTemplateId()))
	if err != nil {
		return err
	}
	if inUse {
		AssertInventoryUomNotInUse(vErrs)
	}
	return nil
}

// checkUomUsable asks Essential whether a unit may still be chosen.
func (this *StockProductConfigDomainServiceImpl) checkUomUsable(
	ctx corectx.Context, uomId string, vErrs *ft.ClientErrors,
) error {
	usable, err := IsUomUsable(ctx, this.uoms, uomId)
	if err != nil {
		return err
	}
	if !usable {
		AssertInventoryUomNotArchived(vErrs)
	}
	return nil
}
