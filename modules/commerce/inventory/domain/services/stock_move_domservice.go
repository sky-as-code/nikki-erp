package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveDomainService is handed the composable default by the stock_move onion.
func NewStockMoveDomainService(base composable.CrudDomainService) itStock.StockMoveDomainService {
	return &StockMoveDomainServiceImpl{CrudDomainService: base}
}

type StockMoveDomainServiceImpl struct {
	composable.CrudDomainService
}

// Create refuses a new move for a variant that cannot start new business. Only creation is
// checked: a move already on a document keeps progressing after its variant is archived, so the
// document can still be completed or cancelled.
func (this *StockMoveDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, validateMoveVariantTransactable)
	return this.CrudDomainService.Create(ctx, params, opts)
}

func validateMoveVariantTransactable(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	variantId := readStringParam(inputModel.GetFieldData(), models.StockMoveFieldProductVariantId)
	return assertVariantTransactable(ctx, variantId, models.StockMoveFieldProductVariantId, vErrs)
}
