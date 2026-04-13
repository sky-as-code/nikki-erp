package unit

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

func (this CreateUnitCommand) ToDomainModel() *domain.Unit {
	unit := &domain.Unit{}
	model.MustCopy(this, unit)
	return unit
}

func (this UpdateUnitCommand) ToDomainModel() *domain.Unit {
	unit := &domain.Unit{}
	model.MustCopy(this, unit)
	return unit
}
