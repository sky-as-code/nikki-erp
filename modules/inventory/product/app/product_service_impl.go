package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unit"
)

func NewProductServiceImpl(
	repo itProduct.ProductRepository,
	unitSvc itUnit.UnitService,
	cqrsBus cqrs.CqrsBus,
) itProduct.ProductService {
	return &ProductServiceImpl{
		repo:    repo,
		unitSvc: unitSvc,
		cqrsBus: cqrsBus,
	}
}

type ProductServiceImpl struct {
	repo       itProduct.ProductRepository
	unitSvc    itUnit.UnitService
	variantSvc itVariant.VariantService
	cqrsBus    cqrs.CqrsBus
}

func (s *ProductServiceImpl) SetVariantService(variantSvc itVariant.VariantService) {
	s.variantSvc = variantSvc
}

func (s *ProductServiceImpl) CreateProduct(ctx corectx.Context, cmd itProduct.CreateProductCommand) (*itProduct.CreateProductResult, error) {
	result, err := corecrud.Create(ctx, corecrud.CreateParam[domain.Product, *domain.Product]{
		Action:         "create product",
		BaseRepoGetter: s.repo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, product *domain.Product, vErrs *ft.ClientErrors) error {
			unitId := product.GetUnitId()
			if unitId == nil {
				return nil
			}
			unitIdStr := string(*unitId)
			unitResult, err := s.unitSvc.GetUnit(ctx, itUnit.GetUnitQuery{Id: &unitIdStr})
			if err != nil {
				return err
			}
			if !unitResult.HasData {
				vErrs.Append(*ft.NewBusinessViolation("unit_id", "unit.not_found", "unit does not exist"))
			}
			return nil
		},
	})
	if err != nil || result == nil || result.ClientErrors != nil {
		return result, err
	}

	productId := result.Data.GetId()
	if productId != nil && s.variantSvc != nil {
		variantCmd := itVariant.CreateVariantCommand{Variant: *domain.NewVariant()}
		variantCmd.SetProductId(productId)
		variantCmd.SetOrgId(result.Data.GetOrgId())
		variantCmd.SetName(result.Data.GetName())
		variantCmd.SetBarcode(&cmd.BarCode)
		variantCmd.SetSku(&cmd.Sku)
		variantCmd.SetProposedPrice(&cmd.ProposedPrice)
		s.variantSvc.CreateVariant(ctx, variantCmd)
	}

	return result, nil
}

func (s *ProductServiceImpl) UpdateProduct(ctx corectx.Context, cmd itProduct.UpdateProductCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Product, *domain.Product]{
		Action:       "update product",
		DbRepoGetter: s.repo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, product *domain.Product, foundProduct *domain.Product, vErrs *ft.ClientErrors) error {
			unitId := product.GetUnitId()
			if unitId != nil {
				unitIdStr := string(*unitId)
				unitResult, err := s.unitSvc.GetUnit(ctx, itUnit.GetUnitQuery{Id: &unitIdStr})
				if err != nil {
					return err
				}
				if !unitResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation("unit_id", "unit.not_found", "unit does not exist"))
				}
			}

			defaultVariantId := product.GetDefaultVariantId()
			productId := product.GetId()
			if defaultVariantId != nil && productId != nil && s.variantSvc != nil {
				defaultVariantIdStr := string(*defaultVariantId)
				variantResult, err := s.variantSvc.GetVariant(ctx, itVariant.GetVariantQuery{
					Id: &defaultVariantIdStr,
				})
				if err != nil {
					return err
				}
				if !variantResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation("default_variant_id", "variant.not_found", "default variant does not exist"))
				}
			}

			return nil
		},
	})
}

func (s *ProductServiceImpl) DeleteProduct(ctx corectx.Context, cmd itProduct.DeleteProductCommand) (*itProduct.DeleteProductResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete product",
		DbRepoGetter: s.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (s *ProductServiceImpl) GetProduct(ctx corectx.Context, query itProduct.GetProductQuery) (*itProduct.GetProductResult, error) {
	var q dyn.GetOneQuery
	if query.Id != nil {
		q.Id = *query.Id
	}
	q.Columns = query.Columns
	return corecrud.GetOne[domain.Product](ctx, corecrud.GetOneParam{
		Action:       "get product",
		DbRepoGetter: s.repo,
		Query:        q,
	})
}

func (s *ProductServiceImpl) ProductExists(ctx corectx.Context, query itProduct.ProductExistsQuery) (*itProduct.ProductExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "product exists",
		DbRepoGetter: s.repo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (s *ProductServiceImpl) SearchProducts(ctx corectx.Context, query itProduct.SearchProductsQuery) (*itProduct.SearchProductsResult, error) {
	return corecrud.Search[domain.Product](ctx, corecrud.SearchParam{
		Action:       "search products",
		DbRepoGetter: s.repo,
		Query:        dyn.SearchQuery(query),
	})
}

func (s *ProductServiceImpl) SetProductIsArchived(ctx corectx.Context, cmd itProduct.SetProductIsArchivedCommand) (*itProduct.SetProductIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, s.repo, dyn.SetIsArchivedCommand(cmd))
}
