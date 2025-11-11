package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unitcategory/interfaces"
)

func NewUnitCategoryServiceImpl(
	unitCategoryRepo it.UnitCategoryRepository,
) it.UnitCategoryService {
	return &UnitCategoryServiceImpl{
		unitCategoryRepo: unitCategoryRepo,
	}
}

type UnitCategoryServiceImpl struct {
	unitCategoryRepo it.UnitCategoryRepository
}

// Create

func (s *UnitCategoryServiceImpl) CreateUnitCategory(ctx crud.Context, cmd it.CreateUnitCategoryCommand) (*it.CreateUnitCategoryResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*it.UnitCategory, it.CreateUnitCategoryCommand, it.CreateUnitCategoryResult]{
		Action:     "create unit category",
		Command:    cmd,
		RepoCreate: s.unitCategoryRepo.Create,
		Sanitize:   s.sanitizeUnitCategory,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateUnitCategoryResult {
			return &it.CreateUnitCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.UnitCategory) *it.CreateUnitCategoryResult {
			return &it.CreateUnitCategoryResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (s *UnitCategoryServiceImpl) UpdateUnitCategory(ctx crud.Context, cmd it.UpdateUnitCategoryCommand) (*it.UpdateUnitCategoryResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*it.UnitCategory, it.UpdateUnitCategoryCommand, it.UpdateUnitCategoryResult]{
		Action:       "update unit category",
		Command:      cmd,
		AssertExists: s.assertUnitCategoryIdExists,
		RepoUpdate:   s.unitCategoryRepo.Update,
		Sanitize:     s.sanitizeUnitCategory,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateUnitCategoryResult {
			return &it.UpdateUnitCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.UnitCategory) *it.UpdateUnitCategoryResult {
			return &it.UpdateUnitCategoryResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *UnitCategoryServiceImpl) DeleteUnitCategory(ctx crud.Context, cmd it.DeleteUnitCategoryCommand) (*it.DeleteUnitCategoryResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*it.UnitCategory, it.DeleteUnitCategoryCommand, it.DeleteUnitCategoryResult]{
		Action:       "delete unit category",
		Command:      cmd,
		AssertExists: s.assertUnitCategoryIdExists,
		RepoDelete: func(ctx crud.Context, model *it.UnitCategory) (int, error) {
			return s.unitCategoryRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteUnitCategoryResult {
			return &it.DeleteUnitCategoryResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *it.UnitCategory, deletedCount int) *it.DeleteUnitCategoryResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *UnitCategoryServiceImpl) GetUnitCategoryById(ctx crud.Context, query it.GetUnitCategoryByIdQuery) (*it.GetUnitCategoryByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*it.UnitCategory, it.GetUnitCategoryByIdQuery, it.GetUnitCategoryByIdResult]{
		Action: "get unit category by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetUnitCategoryByIdQuery, vErrs *ft.ValidationErrors) (*it.UnitCategory, error) {
			dbUnitCategory, err := s.unitCategoryRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbUnitCategory == nil {
				vErrs.AppendNotFound("id", "unit category id")
			}
			return dbUnitCategory, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetUnitCategoryByIdResult {
			return &it.GetUnitCategoryByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.UnitCategory) *it.GetUnitCategoryByIdResult {
			return &it.GetUnitCategoryByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (s *UnitCategoryServiceImpl) SearchUnitCategories(ctx crud.Context, query it.SearchUnitCategoriesQuery) (*it.SearchUnitCategoriesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[it.UnitCategory, it.SearchUnitCategoriesQuery, it.SearchUnitCategoriesResult]{
		Action: "search unit categories",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchUnitCategoriesQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return s.unitCategoryRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query it.SearchUnitCategoriesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[it.UnitCategory], error) {
			return s.unitCategoryRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchUnitCategoriesResult {
			return &it.SearchUnitCategoriesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[it.UnitCategory]) *it.SearchUnitCategoriesResult {
			return &it.SearchUnitCategoriesResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *UnitCategoryServiceImpl) sanitizeUnitCategory(_ *it.UnitCategory) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *UnitCategoryServiceImpl) assertUnitCategoryIdExists(ctx crud.Context, unitCategory *it.UnitCategory, vErrs *ft.ValidationErrors) (*it.UnitCategory, error) {
	dbUnitCategory, err := s.unitCategoryRepo.FindById(ctx, it.FindByIdParam{
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
