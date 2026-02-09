package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unit"
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

func (this *ProductServiceImpl) CreateProduct(ctx crud.Context, cmd itProduct.CreateProductCommand) (result *itProduct.CreateProductResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to add or remove users"); e != nil {
			err = e
		}
	}()

	var dbProduct *domain.Product
	product := cmd.ToDomainModel()
	product.SetDefaults()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			err = this.assertCreateProduct(ctx, cmd, vErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbProduct, err = this.productRepo.Create(ctx, product)
			ft.PanicOnErr(err)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			err = this.assertCreateVariant(ctx, cmd, *dbProduct.Id, vErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()

	if vErrs.Count() > 0 {
		return &itProduct.CreateProductResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &itProduct.CreateProductResult{
		HasData: true,
		Data:    dbProduct,
	}, nil
}

// Update

func (this *ProductServiceImpl) UpdateProduct(ctx crud.Context, cmd itProduct.UpdateProductCommand) (*itProduct.UpdateProductResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Product, itProduct.UpdateProductCommand, itProduct.UpdateProductResult]{
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
		ToSuccessResult: func(model *domain.Product) *itProduct.UpdateProductResult {
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
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Product, itProduct.DeleteProductCommand, itProduct.DeleteProductResult]{
		Action:       "delete product",
		Command:      cmd,
		AssertExists: this.assertProductIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.Product) (int, error) {
			return this.productRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProduct.DeleteProductResult {
			return &itProduct.DeleteProductResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *domain.Product, deletedCount int) *itProduct.DeleteProductResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (this *ProductServiceImpl) GetProductById(ctx crud.Context, query itProduct.GetProductByIdQuery) (*itProduct.GetProductByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Product, itProduct.GetProductByIdQuery, itProduct.GetProductByIdResult]{
		Action: "get product by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itProduct.GetProductByIdQuery, vErrs *ft.ValidationErrors) (*domain.Product, error) {
			dbProduct, err := this.productRepo.FindById(ctx, q)
			ft.PanicOnErr(err)
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
		ToSuccessResult: func(model *domain.Product) *itProduct.GetProductByIdResult {
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
	result, err := crud.Search(ctx, crud.SearchParam[domain.Product, itProduct.SearchProductsQuery, itProduct.SearchProductsResult]{
		Action: "search products",
		Query:  query,
		SetQueryDefaults: func(q *itProduct.SearchProductsQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			// Expect repository to provide ParseSearchGraph like Party repo
			return this.productRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query itProduct.SearchProductsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Product], error) {
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
		ToSuccessResult: func(paged *crud.PagedResult[domain.Product]) *itProduct.SearchProductsResult {
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
func (this *ProductServiceImpl) assertCreateProduct(ctx crud.Context, cmd itProduct.CreateProductCommand, vErrs *ft.ValidationErrors) error {
	if cmd.UnitId != nil {
		unit, err := this.unitService.GetUnitById(ctx, itUnit.GetUnitByIdQuery{
			Id: *cmd.UnitId,
		})
		if err != nil {
			return err
		}

		if unit.Data == nil {
			vErrs.Append("unit", "unit does not exist")
			return nil
		}
	}

	return nil
}

func (this *ProductServiceImpl) assertCreateVariant(ctx crud.Context, cmd itProduct.CreateProductCommand, productId string, vErrs *ft.ValidationErrors) error {
	defaultVariant, err := this.variantService.CreateVariant(ctx, itVariant.CreateVariantCommand{
		ProductId:     productId,
		Sku:           cmd.Sku,
		Barcode:       cmd.BarCode,
		ProposedPrice: cmd.ProposedPrice,
	})
	ft.PanicOnErr(err)

	if defaultVariant.Data == nil {
		vErrs.Append("defaultVariant", "failed to create default variant")
		return nil
	}

	return nil
}

func (this *ProductServiceImpl) assertUpdateProduct(ctx crud.Context, product *domain.Product, _ *domain.Product, vErrs *ft.ValidationErrors) error {

	return nil
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (this *ProductServiceImpl) sanitizeProduct(product *domain.Product) {
}

func (this *ProductServiceImpl) assertProductIdExists(ctx crud.Context, product *domain.Product, vErrs *ft.ValidationErrors) (*domain.Product, error) {
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
