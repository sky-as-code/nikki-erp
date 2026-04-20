package app

import (
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttr "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itAttrGrp "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
	itAttrVal "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
	ext "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/external"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

type ProductServiceParam struct {
	dig.In

	CqrsBus     cqrs.CqrsBus
	AttrRepo    itAttr.AttributeRepository
	AttrGrpRepo itAttrGrp.AttributeGroupRepository
	AttrValRepo itAttrVal.AttributeValueRepository
	ProductRepo itProduct.ProductRepository
	VariantRepo itVariant.VariantRepository
	UnitSvc     ext.UnitExtService
}

func NewProductService(param ProductServiceParam) itProduct.ProductService {
	return newProductServiceImpl(param)
}

func newProductServiceImpl(param ProductServiceParam) *ProductServiceImpl {
	return &ProductServiceImpl{
		cqrsBus:       param.CqrsBus,
		attrRepo:      param.AttrRepo,
		attrGrpRepo:   param.AttrGrpRepo,
		attrValueRepo: param.AttrValRepo,
		productRepo:   param.ProductRepo,
		variantRepo:   param.VariantRepo,
		unitSvc:       param.UnitSvc,
	}
}

type ProductServiceImpl struct {
	cqrsBus       cqrs.CqrsBus
	attrRepo      itAttr.AttributeRepository
	attrGrpRepo   itAttrGrp.AttributeGroupRepository
	attrValueRepo itAttrVal.AttributeValueRepository
	productRepo   itProduct.ProductRepository
	variantRepo   itVariant.VariantRepository
	unitSvc       ext.UnitExtService
}

func (this *ProductServiceImpl) CreateProduct(ctx corectx.Context, cmd itProduct.CreateProductCommand) (*itProduct.CreateProductResult, error) {
	result, err := corecrud.Create(ctx, corecrud.CreateParam[domain.Product, *domain.Product]{
		Action:         "create product",
		BaseRepoGetter: this.productRepo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, product *domain.Product, vErrs *ft.ClientErrors) error {
			unitId := product.GetUnitId()
			if unitId == nil {
				return nil
			}
			unitResult, err := this.unitSvc.GetUnit(ctx, ext.GetUnitQuery{Id: *unitId})
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
	if productId != nil {
		variantCmd := itVariant.CreateVariantCommand{Variant: *domain.NewVariant()}
		variantCmd.SetProductId(productId)
		variantCmd.SetOrgId(cmd.GetOrgId())
		variantCmd.SetName(cmd.GetName())
		variantCmd.SetBarcode(&cmd.BarCode)
		variantCmd.SetSku(&cmd.Sku)
		variantCmd.SetProposedPrice(&cmd.ProposedPrice)
		this.CreateVariant(ctx, variantCmd)
	}

	return result, nil
}

func (this *ProductServiceImpl) UpdateProduct(ctx corectx.Context, cmd itProduct.UpdateProductCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Product, *domain.Product]{
		Action:       "update product",
		DbRepoGetter: this.productRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, product *domain.Product, foundProduct *domain.Product, vErrs *ft.ClientErrors) error {
			unitId := product.GetUnitId()
			if unitId != nil {
				unitResult, err := this.unitSvc.GetUnit(ctx, ext.GetUnitQuery{Id: *unitId})
				if err != nil {
					return err
				}
				if !unitResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation("unit_id", "unit.not_found", "unit does not exist"))
				}
			}

			defaultVariantId := product.GetDefaultVariantId()
			productId := product.GetId()
			if defaultVariantId != nil && productId != nil {
				variantResult, err := this.GetVariant(ctx, itVariant.GetVariantQuery{
					Id: *defaultVariantId,
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

func (this *ProductServiceImpl) DeleteProduct(ctx corectx.Context, cmd itProduct.DeleteProductCommand) (*itProduct.DeleteProductResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete product",
		DbRepoGetter: this.productRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ProductServiceImpl) GetProduct(ctx corectx.Context, query itProduct.GetProductQuery) (*itProduct.GetProductResult, error) {
	return corecrud.GetOne[domain.Product](ctx, corecrud.GetOneParam{
		Action:       "get product",
		DbRepoGetter: this.productRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ProductServiceImpl) ProductExists(ctx corectx.Context, query itProduct.ProductExistsQuery) (*itProduct.ProductExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "product exists",
		DbRepoGetter: this.productRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *ProductServiceImpl) SearchProducts(ctx corectx.Context, query itProduct.SearchProductsQuery) (*itProduct.SearchProductsResult, error) {
	return corecrud.Search[domain.Product](ctx, corecrud.SearchParam{
		Action:       "search products",
		DbRepoGetter: this.productRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *ProductServiceImpl) SetProductIsArchived(ctx corectx.Context, cmd itProduct.SetProductIsArchivedCommand) (*itProduct.SetProductIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.productRepo, dyn.SetIsArchivedCommand(cmd))
}
