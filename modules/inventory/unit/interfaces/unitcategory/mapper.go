package unitcategory

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

func EntToUnitCategory(entUnitCategory *ent.UnitCategory) *domain.UnitCategory {
	unitCategory := &domain.UnitCategory{}
	model.MustCopy(entUnitCategory, unitCategory)

	return unitCategory
}

func EntToUnitCategories(entUnitCategories []*ent.UnitCategory) []domain.UnitCategory {
	if entUnitCategories == nil {
		return nil
	}
	return array.Map(entUnitCategories, func(entUnitCategory *ent.UnitCategory) domain.UnitCategory {
		return *EntToUnitCategory(entUnitCategory)
	})
}

func (cmd CreateUnitCategoryCommand) ToDomainModel() *domain.UnitCategory {
	unitCategory := &domain.UnitCategory{}
	model.MustCopy(cmd, unitCategory)
	return unitCategory
}

func (cmd UpdateUnitCategoryCommand) ToDomainModel() *domain.UnitCategory {
	unitCategory := &domain.UnitCategory{}
	model.MustCopy(cmd, unitCategory)
	return unitCategory
}

func (this DeleteUnitCategoryCommand) ToDomainModel() *domain.UnitCategory {
	unitCategory := &domain.UnitCategory{}
	unitCategory.Id = &this.Id
	return unitCategory
}
