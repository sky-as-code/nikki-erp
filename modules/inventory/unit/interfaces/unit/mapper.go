package unit

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

func EntToUnit(entUnit *ent.Unit) *domain.Unit {
	unit := &domain.Unit{}
	model.MustCopy(entUnit, unit)

	return unit
}

func EntToUnits(entUnits []*ent.Unit) []domain.Unit {
	if entUnits == nil {
		return nil
	}
	return array.Map(entUnits, func(entUnit *ent.Unit) domain.Unit {
		return *EntToUnit(entUnit)
	})
}

func (cmd CreateUnitCommand) ToDomainModel() *domain.Unit {
	unit := &domain.Unit{}
	model.MustCopy(cmd, unit)
	return unit
}

func (cmd UpdateUnitCommand) ToDomainModel() *domain.Unit {
	unit := &domain.Unit{}
	model.MustCopy(cmd, unit)
	return unit
}

func (this DeleteUnitCommand) ToDomainModel() *domain.Unit {
	unit := &domain.Unit{}
	unit.Id = &this.Id
	return unit
}
