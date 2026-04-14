package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unitcategory"
)

func NewUnitCategoryServiceImpl(
	repo itUnitCategory.UnitCategoryRepository,
) itUnitCategory.UnitCategoryService {
	return &UnitCategoryServiceImpl{repo: repo}
}

type UnitCategoryServiceImpl struct {
	repo itUnitCategory.UnitCategoryRepository
}

func (s *UnitCategoryServiceImpl) CreateUnitCategory(ctx corectx.Context, cmd itUnitCategory.CreateUnitCategoryCommand) (*itUnitCategory.CreateUnitCategoryResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.UnitCategory, *domain.UnitCategory]{
		Action:         "create unit category",
		BaseRepoGetter: s.repo,
		Data:           cmd,
	})
}

func (s *UnitCategoryServiceImpl) UpdateUnitCategory(ctx corectx.Context, cmd itUnitCategory.UpdateUnitCategoryCommand) (*itUnitCategory.UpdateUnitCategoryResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.UnitCategory, *domain.UnitCategory]{
		Action:       "update unit category",
		DbRepoGetter: s.repo,
		Data:         cmd,
	})
}

func (s *UnitCategoryServiceImpl) DeleteUnitCategory(ctx corectx.Context, cmd itUnitCategory.DeleteUnitCategoryCommand) (*itUnitCategory.DeleteUnitCategoryResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete unit category",
		DbRepoGetter: s.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (s *UnitCategoryServiceImpl) GetUnitCategory(ctx corectx.Context, query itUnitCategory.GetUnitCategoryQuery) (*itUnitCategory.GetUnitCategoryResult, error) {
	var q dyn.GetOneQuery
	if query.Id != nil {
		q.Id = *query.Id
	}
	q.Columns = query.Columns
	return corecrud.GetOne[domain.UnitCategory](ctx, corecrud.GetOneParam{
		Action:       "get unit category",
		DbRepoGetter: s.repo,
		Query:        q,
	})
}

func (s *UnitCategoryServiceImpl) SearchUnitCategories(ctx corectx.Context, query itUnitCategory.SearchUnitCategoriesQuery) (*itUnitCategory.SearchUnitCategoriesResult, error) {
	return corecrud.Search[domain.UnitCategory](ctx, corecrud.SearchParam{
		Action:       "search unit categories",
		DbRepoGetter: s.repo,
		Query:        dyn.SearchQuery(query),
	})
}

func (s *UnitCategoryServiceImpl) UnitCategoryExists(ctx corectx.Context, query itUnitCategory.UnitCategoryExistsQuery) (*itUnitCategory.UnitCategoryExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if unit category exists",
		DbRepoGetter: s.repo,
		Query:        dyn.ExistsQuery(query),
	})
}
