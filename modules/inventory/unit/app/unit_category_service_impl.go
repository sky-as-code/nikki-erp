package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unitcategory"
)

func NewUnitCategoryServiceImpl(
	unitCategoryRepo itUnitCategory.UnitCategoryRepository,
) itUnitCategory.UnitCategoryService {
	return &UnitCategoryServiceImpl{
		unitCategoryRepo: unitCategoryRepo,
	}
}

type UnitCategoryServiceImpl struct {
	unitCategoryRepo itUnitCategory.UnitCategoryRepository
}

// Create

func (s *UnitCategoryServiceImpl) CreateUnitCategory(ctx crud.Context, cmd itUnitCategory.CreateUnitCategoryCommand) (*itUnitCategory.CreateUnitCategoryResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.UnitCategory, itUnitCategory.CreateUnitCategoryCommand, itUnitCategory.CreateUnitCategoryResult]{
		Action:     "create unit category",
		Command:    cmd,
		RepoCreate: s.unitCategoryRepo.Create,
		Sanitize:   s.sanitizeUnitCategory,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnitCategory.CreateUnitCategoryResult {
			return &itUnitCategory.CreateUnitCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.UnitCategory) *itUnitCategory.CreateUnitCategoryResult {
			return &itUnitCategory.CreateUnitCategoryResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (s *UnitCategoryServiceImpl) UpdateUnitCategory(ctx crud.Context, cmd itUnitCategory.UpdateUnitCategoryCommand) (*itUnitCategory.UpdateUnitCategoryResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.UnitCategory, itUnitCategory.UpdateUnitCategoryCommand, itUnitCategory.UpdateUnitCategoryResult]{
		Action:       "update unit category",
		Command:      cmd,
		AssertExists: s.assertUnitCategoryIdExists,
		RepoUpdate:   s.unitCategoryRepo.Update,
		Sanitize:     s.sanitizeUnitCategory,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnitCategory.UpdateUnitCategoryResult {
			return &itUnitCategory.UpdateUnitCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.UnitCategory) *itUnitCategory.UpdateUnitCategoryResult {
			return &itUnitCategory.UpdateUnitCategoryResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *UnitCategoryServiceImpl) DeleteUnitCategory(ctx crud.Context, cmd itUnitCategory.DeleteUnitCategoryCommand) (*itUnitCategory.DeleteUnitCategoryResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.UnitCategory, itUnitCategory.DeleteUnitCategoryCommand, itUnitCategory.DeleteUnitCategoryResult]{
		Action:       "delete unit category",
		Command:      cmd,
		AssertExists: s.assertUnitCategoryIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.UnitCategory) (int, error) {
			return s.unitCategoryRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnitCategory.DeleteUnitCategoryResult {
			return &itUnitCategory.DeleteUnitCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *domain.UnitCategory, deletedCount int) *itUnitCategory.DeleteUnitCategoryResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *UnitCategoryServiceImpl) GetUnitCategoryById(ctx crud.Context, query itUnitCategory.GetUnitCategoryByIdQuery) (*itUnitCategory.GetUnitCategoryByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.UnitCategory, itUnitCategory.GetUnitCategoryByIdQuery, itUnitCategory.GetUnitCategoryByIdResult]{
		Action: "get unit category by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itUnitCategory.GetUnitCategoryByIdQuery, vErrs *ft.ValidationErrors) (*domain.UnitCategory, error) {
			dbUnitCategory, err := s.unitCategoryRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbUnitCategory == nil {
				vErrs.AppendNotFound("id", "unit category id")
			}
			return dbUnitCategory, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnitCategory.GetUnitCategoryByIdResult {
			return &itUnitCategory.GetUnitCategoryByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.UnitCategory) *itUnitCategory.GetUnitCategoryByIdResult {
			return &itUnitCategory.GetUnitCategoryByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (s *UnitCategoryServiceImpl) SearchUnitCategories(ctx crud.Context, query itUnitCategory.SearchUnitCategoriesQuery) (*itUnitCategory.SearchUnitCategoriesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.UnitCategory, itUnitCategory.SearchUnitCategoriesQuery, itUnitCategory.SearchUnitCategoriesResult]{
		Action: "search unit categories",
		Query:  query,
		SetQueryDefaults: func(q *itUnitCategory.SearchUnitCategoriesQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return s.unitCategoryRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query itUnitCategory.SearchUnitCategoriesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.UnitCategory], error) {
			return s.unitCategoryRepo.Search(ctx, itUnitCategory.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnitCategory.SearchUnitCategoriesResult {
			return &itUnitCategory.SearchUnitCategoriesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[domain.UnitCategory]) *itUnitCategory.SearchUnitCategoriesResult {
			return &itUnitCategory.SearchUnitCategoriesResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *UnitCategoryServiceImpl) sanitizeUnitCategory(_ *domain.UnitCategory) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *UnitCategoryServiceImpl) assertUnitCategoryIdExists(ctx crud.Context, unitCategory *domain.UnitCategory, vErrs *ft.ValidationErrors) (*domain.UnitCategory, error) {
	dbUnitCategory, err := s.unitCategoryRepo.FindById(ctx, itUnitCategory.FindByIdParam{
		Id: *unitCategory.Id,
	})
	if err != nil {
		return nil, err
	}

	if dbUnitCategory == nil {
		vErrs.Append("id", "unit category not found")
		return nil, nil
	}

	return dbUnitCategory, nil
}
