package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itProductCategory "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

func NewProductCategoryServiceImpl(
	productCategoryRepo itProductCategory.ProductCategoryRepository,
) itProductCategory.ProductCategoryService {
	return &ProductCategoryServiceImpl{
		productCategoryRepo: productCategoryRepo,
	}
}

type ProductCategoryServiceImpl struct {
	productCategoryRepo itProductCategory.ProductCategoryRepository
}

// Create

func (s *ProductCategoryServiceImpl) CreateProductCategory(ctx crud.Context, cmd itProductCategory.CreateProductCategoryCommand) (*itProductCategory.CreateProductCategoryResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.ProductCategory, itProductCategory.CreateProductCategoryCommand, itProductCategory.CreateProductCategoryResult]{
		Action:     "create product category",
		Command:    cmd,
		RepoCreate: s.productCategoryRepo.Create,
		SetDefault: s.SetDefaults,
		Sanitize:   s.sanitizeProductCategory,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProductCategory.CreateProductCategoryResult {
			return &itProductCategory.CreateProductCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.ProductCategory) *itProductCategory.CreateProductCategoryResult {
			return &itProductCategory.CreateProductCategoryResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (s *ProductCategoryServiceImpl) UpdateProductCategory(ctx crud.Context, cmd itProductCategory.UpdateProductCategoryCommand) (*itProductCategory.UpdateProductCategoryResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.ProductCategory, itProductCategory.UpdateProductCategoryCommand, itProductCategory.UpdateProductCategoryResult]{
		Action:       "update product category",
		Command:      cmd,
		AssertExists: s.assertProductCategoryId,
		RepoUpdate:   s.productCategoryRepo.Update,
		Sanitize:     s.sanitizeProductCategory,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProductCategory.UpdateProductCategoryResult {
			return &itProductCategory.UpdateProductCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.ProductCategory) *itProductCategory.UpdateProductCategoryResult {
			return &itProductCategory.UpdateProductCategoryResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *ProductCategoryServiceImpl) DeleteProductCategory(ctx crud.Context, cmd itProductCategory.DeleteProductCategoryCommand) (*itProductCategory.DeleteProductCategoryResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.ProductCategory, itProductCategory.DeleteProductCategoryCommand, itProductCategory.DeleteProductCategoryResult]{
		Action:       "delete product category",
		Command:      cmd,
		AssertExists: s.assertProductCategoryId,
		RepoDelete: func(ctx crud.Context, model *domain.ProductCategory) (int, error) {
			return s.productCategoryRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProductCategory.DeleteProductCategoryResult {
			return &itProductCategory.DeleteProductCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *domain.ProductCategory, deletedCount int) *itProductCategory.DeleteProductCategoryResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *ProductCategoryServiceImpl) GetProductCategoryById(ctx crud.Context, query itProductCategory.GetProductCategoryByIdQuery) (*itProductCategory.GetProductCategoryByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.ProductCategory, itProductCategory.GetProductCategoryByIdQuery, itProductCategory.GetProductCategoryByIdResult]{
		Action: "get product category by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itProductCategory.GetProductCategoryByIdQuery, vErrs *ft.ValidationErrors) (*domain.ProductCategory, error) {
			dbProductCategory, err := s.productCategoryRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbProductCategory == nil {
				vErrs.AppendNotFound("id", "product category id")
			}
			return dbProductCategory, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProductCategory.GetProductCategoryByIdResult {
			return &itProductCategory.GetProductCategoryByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.ProductCategory) *itProductCategory.GetProductCategoryByIdResult {
			return &itProductCategory.GetProductCategoryByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (s *ProductCategoryServiceImpl) SearchProductCategories(ctx crud.Context, query itProductCategory.SearchProductCategoriesQuery) (*itProductCategory.SearchProductCategoriesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.ProductCategory, itProductCategory.SearchProductCategoriesQuery, itProductCategory.SearchProductCategoriesResult]{
		Action: "search product categories",
		Query:  query,
		SetQueryDefaults: func(q *itProductCategory.SearchProductCategoriesQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return s.productCategoryRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query itProductCategory.SearchProductCategoriesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.ProductCategory], error) {
			return s.productCategoryRepo.Search(ctx, itProductCategory.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itProductCategory.SearchProductCategoriesResult {
			return &itProductCategory.SearchProductCategoriesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[domain.ProductCategory]) *itProductCategory.SearchProductCategoriesResult {
			return &itProductCategory.SearchProductCategoriesResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *ProductCategoryServiceImpl) SetDefaults(productCategory *domain.ProductCategory) {
	productCategory.SetDefaults()
}

func (s *ProductCategoryServiceImpl) sanitizeProductCategory(_ *domain.ProductCategory) {
}

func (s *ProductCategoryServiceImpl) assertProductCategoryId(ctx crud.Context, productCategory *domain.ProductCategory, vErrs *ft.ValidationErrors) (*domain.ProductCategory, error) {
	dbProductCategory, err := s.productCategoryRepo.FindById(ctx, itProductCategory.FindByIdParam{
		Id: *productCategory.Id,
	})
	if err != nil {
		return nil, err
	}

	if dbProductCategory == nil {
		vErrs.Append("id", "product category not found")
		return nil, nil
	}

	return dbProductCategory, nil
}
