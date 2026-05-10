package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unitcategory"
)

func NewUnitCategoryDomainServiceImpl(
	repo itUnitCategory.UnitCategoryRepository,
) itUnitCategory.UnitCategoryDomainService {
	return &UnitCategoryDomainServiceImpl{repo: repo}
}

type UnitCategoryDomainServiceImpl struct {
	repo itUnitCategory.UnitCategoryRepository
}

func (s *UnitCategoryDomainServiceImpl) CreateUnitCategory(ctx corectx.Context, cmd itUnitCategory.CreateUnitCategoryCommand) (*itUnitCategory.CreateUnitCategoryResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.UnitCategory, *models.UnitCategory]{
		Action:         "create unit category",
		BaseRepoGetter: s.repo,
		Data:           cmd,
	})
}

func (s *UnitCategoryDomainServiceImpl) UpdateUnitCategory(ctx corectx.Context, cmd itUnitCategory.UpdateUnitCategoryCommand) (*itUnitCategory.UpdateUnitCategoryResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.UnitCategory, *models.UnitCategory]{
		Action:       "update unit category",
		DbRepoGetter: s.repo,
		Data:         cmd,
	})
}

func (s *UnitCategoryDomainServiceImpl) DeleteUnitCategory(ctx corectx.Context, cmd itUnitCategory.DeleteUnitCategoryCommand) (*itUnitCategory.DeleteUnitCategoryResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete unit category",
		DbRepoGetter: s.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (s *UnitCategoryDomainServiceImpl) GetUnitCategory(ctx corectx.Context, query itUnitCategory.GetUnitCategoryQuery) (*itUnitCategory.GetUnitCategoryResult, error) {
	var q dyn.GetOneQuery
	if query.Id != nil {
		q.Id = *query.Id
	}
	q.Fields = query.Columns
	return corecrud.GetOne[models.UnitCategory](ctx, corecrud.GetOneParam{
		Action:       "get unit category",
		DbRepoGetter: s.repo,
		Query:        q,
	})
}

func (s *UnitCategoryDomainServiceImpl) SearchUnitCategories(ctx corectx.Context, query itUnitCategory.SearchUnitCategoriesQuery) (*itUnitCategory.SearchUnitCategoriesResult, error) {
	return corecrud.Search[models.UnitCategory](ctx, corecrud.SearchParam{
		Action:       "search unit categories",
		DbRepoGetter: s.repo,
		Query:        dyn.SearchQuery(query),
	})
}

func (s *UnitCategoryDomainServiceImpl) UnitCategoryExists(ctx corectx.Context, query itUnitCategory.UnitCategoryExistsQuery) (*itUnitCategory.UnitCategoryExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if unit category exists",
		DbRepoGetter: s.repo,
		Query:        dyn.ExistsQuery(query),
	})
}
