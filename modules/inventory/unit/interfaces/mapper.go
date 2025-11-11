package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
)

func EntToUnit(entUnit *ent.Unit) *Unit {
	unit := &Unit{}
	model.MustCopy(entUnit, unit)

	return unit
}

func EntToUnits(entUnits []*ent.Unit) []Unit {
	if entUnits == nil {
		return nil
	}
	return array.Map(entUnits, func(entUnit *ent.Unit) Unit {
		return *EntToUnit(entUnit)
	})
}

func (cmd CreateUnitCommand) ToDomainModel() *Unit {
	unit := &Unit{}
	model.MustCopy(cmd, unit)
	return unit
}

func (cmd UpdateUnitCommand) ToDomainModel() *Unit {
	unit := &Unit{}
	model.MustCopy(cmd, unit)
	return unit
}

func (this DeleteUnitCommand) ToDomainModel() *Unit {
	unit := &Unit{}
	unit.Id = &this.Id
	return unit
}
