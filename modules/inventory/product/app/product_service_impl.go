package app

import (
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttr "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itAttrGrp "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
	itAttrVal "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
	ext "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/external"
	itMedia "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/media"
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
	MediaSvc    itMedia.InventoryMediaService
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
		mediaSvc:      param.MediaSvc,
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
	mediaSvc      itMedia.InventoryMediaService
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
	result, err := corecrud.GetOne[domain.Product](ctx, corecrud.GetOneParam{
		Action:       "get product",
		DbRepoGetter: this.productRepo,
		Query:        dyn.GetOneQuery(query),
	})
	if err != nil || result == nil || !result.HasData {
		return result, err
	}

	if err := this.populateProductThumbnailMedia(ctx, &result.Data); err != nil {
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

	if err := this.populateProductsThumbnailMedia(ctx, result.Data.Items); err != nil {
		return nil, err
	}

	return result, nil
}

func (this *ProductServiceImpl) SetProductIsArchived(ctx corectx.Context, cmd itProduct.SetProductIsArchivedCommand) (*itProduct.SetProductIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.productRepo, dyn.SetIsArchivedCommand(cmd))
}

func (this *ProductServiceImpl) UploadProductThumnail(ctx corectx.Context, cmd itProduct.UploadProductThumbnailCommand) (*itProduct.UploadProductThumbnailResult, error) {
	res := &itProduct.UploadProductThumbnailResult{}
	getProdRes, err := this.GetProduct(ctx, itProduct.GetProductQuery{
		Id: cmd.ProductId,
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
	oldMediaRef := product.GetThumbnailMediaRef()

	if oldMediaRef != nil && *oldMediaRef != "" {
		delRes, delErr := this.mediaSvc.Delete(ctx, itMedia.DeleteMediaCommand{Id: model.Id(*oldMediaRef)})
		if delErr != nil {
			return nil, delErr
		}
		if delRes.ClientErrors.Count() > 0 {
			return delRes, nil
		}
	}

	uploadRes, err := this.mediaSvc.Upload(ctx, itMedia.UploadMediaCommand{
		Source:     "inventory/product",
		File:       cmd.File,
		FileHeader: cmd.FileHeader,
	})
	if err != nil {
		return nil, err
	}

	if uploadRes.ClientErrors.Count() > 0 {
		res.ClientErrors = uploadRes.ClientErrors
		return res, nil
	}

	if !uploadRes.HasData {
		return res, nil
	}

	media := uploadRes.Data

	product.SetThumbnailMediaRef(media.GetId())

	return this.UpdateProduct(ctx, itProduct.UpdateProductCommand{
		Product: product,
	})
}

func (this *ProductServiceImpl) populateProductThumbnailMedia(ctx corectx.Context, product *domain.Product) error {
	mediaRef := product.GetThumbnailMediaRef()
	if mediaRef == nil || *mediaRef == "" {
		return nil
	}

	mediaRes, err := this.mediaSvc.GetMedia(ctx, itMedia.GetMediaQuery{
		Id:     model.Id(*mediaRef),
		Fields: []string{domain.InventoryMediaFieldStorageKey, domain.InventoryMediaFieldMediaType},
	})
	if err != nil {
		return err
	}
	if !mediaRes.HasData {
		return nil
	}

	product.SetThumbnail(mediaRes.Data.GetFieldData())
	return nil
}

func (this *ProductServiceImpl) populateProductsThumbnailMedia(ctx corectx.Context, products []domain.Product) error {
	mediaIds := collectThumbnailMediaIds(products)
	if len(mediaIds) == 0 {
		return nil
	}

	graph := dmodel.NewSearchGraph().NewCondition(basemodel.FieldId, dmodel.In, anySlice(mediaIds)...)
	searchRes, err := this.mediaSvc.SearchMedia(ctx, itMedia.SearchMediaQuery{
		Fields: []string{domain.InventoryMediaFieldStorageKey, domain.InventoryMediaFieldMediaType},
		Graph:  graph,
		Page:   0,
		Size:   len(mediaIds),
	})
	if err != nil {
		return err
	}
	if !searchRes.HasData {
		return nil
	}

	mediaById := make(map[model.Id]dmodel.DynamicFields, len(searchRes.Data.Items))
	for _, item := range searchRes.Data.Items {
		if id := item.GetId(); id != nil {
			mediaById[*id] = item.GetFieldData()
		}
	}

	for i := range products {
		mediaRef := products[i].GetThumbnailMediaRef()
		if mediaRef == nil || *mediaRef == "" {
			continue
		}
		if media, ok := mediaById[model.Id(*mediaRef)]; ok {
			products[i].SetThumbnail(media)
		}
	}

	return nil
}

func collectThumbnailMediaIds(products []domain.Product) []model.Id {
	seen := make(map[model.Id]struct{})
	ids := make([]model.Id, 0)

	for _, product := range products {
		mediaRef := product.GetThumbnailMediaRef()
		if mediaRef == nil || *mediaRef == "" {
			continue
		}
		id := model.Id(*mediaRef)
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	return ids
}
