package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
)

func EntToUnitCategory(entUnitCategory *ent.UnitCategory) *UnitCategory {
	unitCategory := &UnitCategory{}
	model.MustCopy(entUnitCategory, unitCategory)

	return unitCategory
}

func EntToUnitCategories(entUnitCategories []*ent.UnitCategory) []UnitCategory {
	if entUnitCategories == nil {
		return nil
	}
	return array.Map(entUnitCategories, func(entUnitCategory *ent.UnitCategory) UnitCategory {
		return *EntToUnitCategory(entUnitCategory)
	})
}

func (cmd CreateUnitCategoryCommand) ToDomainModel() *UnitCategory {
	unitCategory := &UnitCategory{}
	model.MustCopy(cmd, unitCategory)
	return unitCategory
}

func (cmd UpdateUnitCategoryCommand) ToDomainModel() *UnitCategory {
	unitCategory := &UnitCategory{}
	model.MustCopy(cmd, unitCategory)
	return unitCategory
}

func (this DeleteUnitCategoryCommand) ToDomainModel() *UnitCategory {
	unitCategory := &UnitCategory{}
	unitCategory.Id = &this.Id
	return unitCategory
}
