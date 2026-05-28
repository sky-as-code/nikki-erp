package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unitcategory"
)

func NewUnitCategoryApplicationServiceImpl(unitCategorySvc it.UnitCategoryDomainService) it.UnitCategoryAppService {
	return &UnitCategoryApplicationServiceImpl{unitCategorySvc: unitCategorySvc}
}

type UnitCategoryApplicationServiceImpl struct {
	unitCategorySvc it.UnitCategoryDomainService
}

func (this *UnitCategoryApplicationServiceImpl) CreateUnitCategory(ctx corectx.Context, cmd it.CreateUnitCategoryCommand) (*it.CreateUnitCategoryResult, error) {
	return this.unitCategorySvc.CreateUnitCategory(ctx, cmd)
}

func (this *UnitCategoryApplicationServiceImpl) UpdateUnitCategory(ctx corectx.Context, cmd it.UpdateUnitCategoryCommand) (*it.UpdateUnitCategoryResult, error) {
	return this.unitCategorySvc.UpdateUnitCategory(ctx, cmd)
}

func (this *UnitCategoryApplicationServiceImpl) DeleteUnitCategory(ctx corectx.Context, cmd it.DeleteUnitCategoryCommand) (*it.DeleteUnitCategoryResult, error) {
	return this.unitCategorySvc.DeleteUnitCategory(ctx, cmd)
}

func (this *UnitCategoryApplicationServiceImpl) GetUnitCategory(ctx corectx.Context, query it.GetUnitCategoryQuery) (*it.GetUnitCategoryResult, error) {
	return this.unitCategorySvc.GetUnitCategory(ctx, query)
}

func (this *UnitCategoryApplicationServiceImpl) SearchUnitCategories(ctx corectx.Context, query it.SearchUnitCategoriesQuery) (*it.SearchUnitCategoriesResult, error) {
	return this.unitCategorySvc.SearchUnitCategories(ctx, query)
}

func (this *UnitCategoryApplicationServiceImpl) UnitCategoryExists(ctx corectx.Context, query it.UnitCategoryExistsQuery) (*it.UnitCategoryExistsResult, error) {
	return this.unitCategorySvc.UnitCategoryExists(ctx, query)
}
