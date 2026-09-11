package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
)

func NewSalesPricelistApplicationService(
	base composable.CrudApplicationService,
) itCatalog.SalesPricelistApplicationService {
	return &SalesPricelistApplicationServiceImpl{
		CrudApplicationService: base,
		pricelistSvc:           base.DomainService().(itCatalog.SalesPricelistDomainService),
	}
}

type SalesPricelistApplicationServiceImpl struct {
	composable.CrudApplicationService
	pricelistSvc itCatalog.SalesPricelistDomainService
}

// SetDefault reuses the update permission: naming the fallback list is an ordinary edit of the
// list's own flag, and a separate grant would have to be discovered and assigned to restore a
// power the caller already has over the record.
func (this *SalesPricelistApplicationServiceImpl) SetDefault(
	ctx corectx.Context, cmd itCatalog.SetDefaultPricelistCommand,
) (*itCatalog.SetDefaultPricelistResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSetDefaultPricelist, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.pricelistSvc.SetDefault(ctx, readStringParam(cmd, paramRecordId))
}

func NewSalesPricelistItemApplicationService(
	base composable.CrudApplicationService,
) itCatalog.SalesPricelistItemApplicationService {
	return &SalesPricelistItemApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPricelistItemApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// The plain-CRUD catalogue resources: no custom action, but each still gets its own application
// service so a rule has a home later and the layer is published by its own type.

func NewSalesComboApplicationService(base composable.CrudApplicationService) itCatalog.SalesComboApplicationService {
	return &SalesComboApplicationServiceImpl{CrudApplicationService: base}
}

type SalesComboApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesComboComponentApplicationService(base composable.CrudApplicationService) itCatalog.SalesComboComponentApplicationService {
	return &SalesComboComponentApplicationServiceImpl{CrudApplicationService: base}
}

type SalesComboComponentApplicationServiceImpl struct {
	composable.CrudApplicationService
}
