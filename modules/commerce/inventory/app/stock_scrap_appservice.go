package app

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockScrapApplicationService is handed the composable default by the scrap onion.
func NewStockScrapApplicationService(base composable.CrudApplicationService) itStock.StockScrapApplicationService {
	scrapSvc, ok := base.DomainService().(*services.StockScrapDomainServiceImpl)
	if !ok {
		panic(errors.New("the stock scrap onion must be built with NewStockScrapDomainService"))
	}
	return &StockScrapApplicationServiceImpl{CrudApplicationService: base, scrapSvc: scrapSvc}
}

type StockScrapApplicationServiceImpl struct {
	composable.CrudApplicationService
	scrapSvc *services.StockScrapDomainServiceImpl
}

// DoScrap carries its own permission so a role that may fix a typo in a scrap note cannot
// destroy stock.
func (this *StockScrapApplicationServiceImpl) DoScrap(
	ctx corectx.Context, cmd itStock.DoScrapCommand,
) (*itStock.DoScrapResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionDoScrap, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.scrapSvc.DoScrap(ctx, readStringField(cmd, paramRecordId))
}
