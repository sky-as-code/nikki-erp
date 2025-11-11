package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/variant/interfaces"
)

func NewProductServiceImpl(
	productRepo itProduct.ProductRepository,
	unitService itUnit.UnitService,
	variantService itVariant.VariantService,
) itProduct.ProductService {
	return &ProductServiceImpl{
		productRepo:    productRepo,
		unitService:    unitService,
		variantService: variantService,
	}
}

type ProductServiceImpl struct {
	productRepo    itProduct.ProductRepository
	unitService    itUnit.UnitService
	variantService itVariant.VariantService
}

// Create

func (this *ProductServiceImpl) CreateProduct(ctx crud.Context, cmd itProduct.CreateProductCommand) (*itProduct.CreateProductResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*itProduct.Product, itProduct.CreateProductCommand, itProduct.CreateProductResult]{
		Action:              "create product",
		Command:             cmd,
		RepoCreate:          this.productRepo.Create,
		AssertBusinessRules: this.assertCreateProduct,
		Sanitize:            this.sanitizeProduct,
		SetDefault:          this.setProductDefaults,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProduct.CreateProductResult {
			return &itProduct.CreateProductResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *itProduct.Product) *itProduct.CreateProductResult {
			return &itProduct.CreateProductResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (this *ProductServiceImpl) UpdateProduct(ctx crud.Context, cmd itProduct.UpdateProductCommand) (*itProduct.UpdateProductResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*itProduct.Product, itProduct.UpdateProductCommand, itProduct.UpdateProductResult]{
		Action:              "update product",
		Command:             cmd,
		AssertExists:        this.assertProductIdExists,
		AssertBusinessRules: this.assertUpdateProduct,
		RepoUpdate:          this.productRepo.Update,
		Sanitize:            this.sanitizeProduct,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProduct.UpdateProductResult {
			return &itProduct.UpdateProductResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *itProduct.Product) *itProduct.UpdateProductResult {
			return &itProduct.UpdateProductResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (this *ProductServiceImpl) DeleteProduct(ctx crud.Context, cmd itProduct.DeleteProductCommand) (*itProduct.DeleteProductResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*itProduct.Product, itProduct.DeleteProductCommand, itProduct.DeleteProductResult]{
		Action:       "delete product",
		Command:      cmd,
		AssertExists: this.assertProductIdExists,
		RepoDelete: func(ctx crud.Context, model *itProduct.Product) (int, error) {
			return this.productRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProduct.DeleteProductResult {
			return &itProduct.DeleteProductResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *itProduct.Product, deletedCount int) *itProduct.DeleteProductResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (this *ProductServiceImpl) GetProductById(ctx crud.Context, query itProduct.GetProductByIdQuery) (*itProduct.GetProductByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*itProduct.Product, itProduct.GetProductByIdQuery, itProduct.GetProductByIdResult]{
		Action: "get product by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itProduct.GetProductByIdQuery, vErrs *ft.ValidationErrors) (*itProduct.Product, error) {
			dbProduct, err := this.productRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbProduct == nil {
				vErrs.AppendNotFound("id", "product id")
			}
			return dbProduct, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProduct.GetProductByIdResult {
			return &itProduct.GetProductByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *itProduct.Product) *itProduct.GetProductByIdResult {
			return &itProduct.GetProductByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *ProductServiceImpl) SearchProducts(ctx crud.Context, query itProduct.SearchProductsQuery) (*itProduct.SearchProductsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[itProduct.Product, itProduct.SearchProductsQuery, itProduct.SearchProductsResult]{
		Action: "search products",
		Query:  query,
		SetQueryDefaults: func(q *itProduct.SearchProductsQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			// Expect repository to provide ParseSearchGraph like Party repo
			return this.productRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query itProduct.SearchProductsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[itProduct.Product], error) {
			return this.productRepo.Search(ctx, itProduct.SearchParam{
				Predicate:    predicate,
				Order:        order,
				Page:         *query.Page,
				Size:         *query.Size,
				WithVariants: query.WithVariants,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProduct.SearchProductsResult {
			return &itProduct.SearchProductsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[itProduct.Product]) *itProduct.SearchProductsResult {
			return &itProduct.SearchProductsResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// assert methods
// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (this *ProductServiceImpl) assertCreateProduct(ctx crud.Context, product *itProduct.Product, vErrs *ft.ValidationErrors) error {
	// dbProduct, err := this.productRepo.FindById(ctx, itProduct.FindByIdParam{
	// 	Id: *product.Id,
	// })
	// if err != nil {
	// 	return err
	// }

	// if dbProduct != nil {
	// 	vErrs.Append("id", "product with the given id already exists")
	// 	return nil
	// }

	unit, err := this.unitService.GetUnitById(ctx, itUnit.GetUnitByIdQuery{
		Id: *product.Unit,
	})
	if err != nil {
		return err
	}

	if unit.Data == nil {
		vErrs.Append("unit", "unit does not exist")
		return nil
	}

	// if product.DefaultsVariantId == nil {
	// 	variant, err := this.variantService.CreateVariant(ctx, itVariant.CreateVariantCommand{
	// 		ProductId: *product.Id,
	// 		Sku:       "DEF-" + *product.Id,
	// 		// Price:     0,
	// 	})

	// 	if err != nil {
	// 		return err
	// 	}

	// 	if variant.Data == nil {
	// 		vErrs.Append("defaultsVariantId", "failed to create default variant")
	// 		return nil
	// 	}

	// 	product.DefaultsVariantId = variant.Data.Id
	// }

	return nil
}

func (this *ProductServiceImpl) assertUpdateProduct(ctx crud.Context, product *itProduct.Product, _ *itProduct.Product, vErrs *ft.ValidationErrors) error {

	return nil
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (this *ProductServiceImpl) sanitizeProduct(product *itProduct.Product) {

}

func (this *ProductServiceImpl) assertProductIdExists(ctx crud.Context, product *itProduct.Product, vErrs *ft.ValidationErrors) (*itProduct.Product, error) {
	dbProduct, err := this.productRepo.FindById(ctx, itProduct.FindByIdParam{
		Id:           *product.Id,
		WithVariants: false,
	})
	if err != nil {
		return nil, err
	}

	if dbProduct == nil {
		vErrs.Append("id", "product not found")
		return nil, nil
	}

	return dbProduct, nil
}

func (this *ProductServiceImpl) setProductDefaults(product *itProduct.Product) {
	product.SetDefaults()
}
