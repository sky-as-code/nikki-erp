package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unit"
)

func NewUnitApplicationServiceImpl(unitSvc it.UnitDomainService) it.UnitAppService {
	return &UnitApplicationServiceImpl{unitSvc: unitSvc}
}

type UnitApplicationServiceImpl struct {
	unitSvc it.UnitDomainService
}

func (this *UnitApplicationServiceImpl) CreateUnit(ctx corectx.Context, cmd it.CreateUnitCommand) (*it.CreateUnitResult, error) {
	return this.unitSvc.CreateUnit(ctx, cmd)
}

func (this *UnitApplicationServiceImpl) UpdateUnit(ctx corectx.Context, cmd it.UpdateUnitCommand) (*it.UpdateUnitResult, error) {
	return this.unitSvc.UpdateUnit(ctx, cmd)
}

func (this *UnitApplicationServiceImpl) DeleteUnit(ctx corectx.Context, cmd it.DeleteUnitCommand) (*it.DeleteUnitResult, error) {
	return this.unitSvc.DeleteUnit(ctx, cmd)
}

func (this *UnitApplicationServiceImpl) GetUnit(ctx corectx.Context, query it.GetUnitQuery) (*it.GetUnitResult, error) {
	return this.unitSvc.GetUnit(ctx, query)
}

func (this *UnitApplicationServiceImpl) SearchUnits(ctx corectx.Context, query it.SearchUnitsQuery) (*it.SearchUnitsResult, error) {
	return this.unitSvc.SearchUnits(ctx, query)
}

func (this *UnitApplicationServiceImpl) UnitExists(ctx corectx.Context, query it.UnitExistsQuery) (*it.UnitExistsResult, error) {
	return this.unitSvc.UnitExists(ctx, query)
}
