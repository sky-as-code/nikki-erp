package app

import (
	"time"

	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/objectkey"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/upload"
	"github.com/sky-as-code/nikki-erp/modules/inventory/legacy/domain"
	itAttr "github.com/sky-as-code/nikki-erp/modules/inventory/legacy/interfaces/attribute"
	itAttrGrp "github.com/sky-as-code/nikki-erp/modules/inventory/legacy/interfaces/attributegroup"
	itAttrVal "github.com/sky-as-code/nikki-erp/modules/inventory/legacy/interfaces/attributevalue"
	ext "github.com/sky-as-code/nikki-erp/modules/inventory/legacy/interfaces/external"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/legacy/interfaces/product"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/legacy/interfaces/variant"
)

type ProductServiceParam struct {
	dig.In

	CqrsBus     cqrs.CqrsBus
	AttrRepo    itAttr.AttributeRepository
	AttrGrpRepo itAttrGrp.AttributeGroupRepository
	AttrValRepo itAttrVal.AttributeValueRepository
	ProductRepo itProduct.ProductRepository
	VariantRepo itVariant.VariantRepository
	Storage     filestorage.FileStorageAdapter
	UomSvc      ext.UomExtService
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
		uomSvc:        param.UomSvc,
		storage:       param.Storage,
	}
}

type ProductServiceImpl struct {
	cqrsBus       cqrs.CqrsBus
	attrRepo      itAttr.AttributeRepository
	attrGrpRepo   itAttrGrp.AttributeGroupRepository
	attrValueRepo itAttrVal.AttributeValueRepository
	productRepo   itProduct.ProductRepository
	variantRepo   itVariant.VariantRepository
	uomSvc        ext.UomExtService
	storage       filestorage.FileStorageAdapter
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
			uomResult, err := this.uomSvc.GetUom(ctx, ext.GetUomQuery{Id: *unitId})
			if err != nil {
				return err
			}
			if !uomResult.HasData {
				vErrs.Append(*ft.NewBusinessViolation("unit_id", "uom.not_found", "unit of measure does not exist"))
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
				uomResult, err := this.uomSvc.GetUom(ctx, ext.GetUomQuery{Id: *unitId})
				if err != nil {
					return err
				}
				if !uomResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation("unit_id", "uom.not_found", "unit of measure does not exist"))
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
	result, err := corecrud.GetOne[domain.Product](ctx, corecrud.GetOneParam{
		Action:       "get product",
		DbRepoGetter: this.productRepo,
		Query:        dyn.GetOneQuery(query),
	})
	if err != nil || result == nil || !result.HasData {
		return result, err
	}

	if err := this.populateProductThumbnailUrl(ctx, &result.Data); err != nil {
		return nil, err
	}

	return result, nil
}

func (this *ProductServiceImpl) ProductExists(ctx corectx.Context, query itProduct.ProductExistsQuery) (*itProduct.ProductExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "product exists",
		DbRepoGetter: this.productRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *ProductServiceImpl) SearchProducts(ctx corectx.Context, query itProduct.SearchProductsQuery) (*itProduct.SearchProductsResult, error) {
	result, err := corecrud.Search[domain.Product](ctx, corecrud.SearchParam{
		Action:       "search products",
		DbRepoGetter: this.productRepo,
		Query:        dyn.SearchQuery(query),
	})
	if err != nil || result == nil || !result.HasData {
		return result, err
	}

	if err := this.populateProductsThumbnailUrl(ctx, result.Data.Items); err != nil {
		return nil, err
	}

	return result, nil
}

func (this *ProductServiceImpl) SetProductIsArchived(ctx corectx.Context, cmd itProduct.SetProductIsArchivedCommand) (*itProduct.SetProductIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.productRepo, dyn.SetIsArchivedCommand(cmd))
}

func (this *ProductServiceImpl) UploadProductThumnail(ctx corectx.Context, cmd itProduct.UploadProductThumbnailCommand) (*itProduct.UploadProductThumbnailResult, error) {
	res := &itProduct.UploadProductThumbnailResult{}
	getProdRes, err := corecrud.GetOne[domain.Product](ctx, corecrud.GetOneParam{
		Action:       "get product",
		DbRepoGetter: this.productRepo,
		Query:        dyn.GetOneQuery{Id: cmd.ProductId},
	})
	if err != nil {
		return nil, err
	}

	if getProdRes.ClientErrors.Count() > 0 {
		res.ClientErrors = getProdRes.ClientErrors
		return res, nil
	}

	if !getProdRes.HasData {
		res.ClientErrors = ft.ClientErrors{*ft.NewNotFoundError("id")}
		return res, nil
	}

	product := getProdRes.Data
	if key := product.GetThumbnailKey(); key != nil && *key != "" {
		if err := this.storage.Remove(ctx.InnerContext(), *key); err != nil {
			return nil, err
		}
	}

	if cmd.File == nil {
		return res, nil
	}

	storageKey, err := objectkey.BuildFromFileHeader("inventory/product", cmd.FileHeader)
	if err != nil {
		return nil, err
	}

	mediaType, uploadReader, err := upload.SniffContentTypeAndRewind(cmd.File, cmd.FileHeader)
	if err != nil {
		return nil, err
	}

	size := int64(0)
	if cmd.FileHeader != nil {
		size = cmd.FileHeader.Size
	}
	if err = this.storage.Put(
		ctx.InnerContext(),
		storageKey,
		uploadReader,
		filestorage.NewPutOptions(mediaType, size),
	); err != nil {
		return nil, err
	}

	product.SetThumbnailKey(&storageKey)

	return this.UpdateProduct(ctx, itProduct.UpdateProductCommand{
		Product: product,
	})
}

func (this *ProductServiceImpl) populateProductThumbnailUrl(ctx corectx.Context, product *domain.Product) error {
	key := product.GetThumbnailKey()
	if key == nil || *key == "" {
		return nil
	}

	url, err := this.storage.GeneratePresignedUrl(ctx.InnerContext(), *key, time.Hour)
	if err != nil {
		return err
	}

	product.SetThumbnailUrl(&url)
	return nil
}

func (this *ProductServiceImpl) populateProductsThumbnailUrl(ctx corectx.Context, products []domain.Product) error {
	seen := make(map[string]struct{})
	uniqueKeys := make([]string, 0)

	for i := range products {
		key := products[i].GetThumbnailKey()
		if key == nil || *key == "" {
			continue
		}
		if _, exists := seen[*key]; exists {
			continue
		}
		seen[*key] = struct{}{}
		uniqueKeys = append(uniqueKeys, *key)
	}

	if len(uniqueKeys) == 0 {
		return nil
	}

	urls, err := filestorage.GeneratePresignedBulk(ctx.InnerContext(), this.storage, uniqueKeys, time.Hour)
	if err != nil {
		return err
	}

	for i := range products {
		key := products[i].GetThumbnailKey()
		if key == nil || *key == "" {
			continue
		}
		if url, ok := urls[*key]; ok {
			products[i].SetThumbnailUrl(&url)
		}
	}

	return nil
}
