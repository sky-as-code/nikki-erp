package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockMoveDomainService is handed the composable default by the stock_move onion.
func NewStockMoveDomainService(base composable.CrudDomainService) itStock.StockMoveDomainService {
	return &StockMoveDomainServiceImpl{CrudDomainService: base}
}

type StockMoveDomainServiceImpl struct {
	composable.CrudDomainService
}

// Create refuses a new move for a variant that cannot start new business, or naming a group location
// as its source or destination. Only creation checks the variant: a move already on a document keeps progressing after its variant is archived, so the
// document can still be completed or cancelled.
func (this *StockMoveDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(
		opts.ValidateExtra, validateMoveVariantTransactable, validateMovementLocationsHoldStock)
	return this.CrudDomainService.Create(ctx, params, opts)
}

// Update refuses to re-point a move at a group location, which holds no stock.
func (this *StockMoveDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.UpdateOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, validateMovementLocationsHoldStock)
	return this.CrudDomainService.Update(ctx, params, opts)
}

func validateMoveVariantTransactable(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	variantId := readStringParam(inputModel.GetFieldData(), models.StockMoveFieldProductVariantId)
	return assertVariantTransactable(ctx, variantId, models.StockMoveFieldProductVariantId, vErrs)
}
