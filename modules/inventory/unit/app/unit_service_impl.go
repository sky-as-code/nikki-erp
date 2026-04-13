package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unit"
)

func NewUnitServiceImpl(unitRepo itUnit.UnitRepository) itUnit.UnitService {
	return &UnitServiceImpl{
		unitRepo: unitRepo,
	}
}

type UnitServiceImpl struct {
	unitRepo itUnit.UnitRepository
}

func (this *UnitServiceImpl) CreateUnit(ctx corectx.Context, cmd itUnit.CreateUnitCommand) (*itUnit.CreateUnitResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Unit, *domain.Unit]{
		Action:         "create unit",
		BaseRepoGetter: this.unitRepo,
		Data:           cmd,
		ValidateExtra:  this.validateBaseUnitCreate,
	})
}

func (this *UnitServiceImpl) UpdateUnit(ctx corectx.Context, cmd itUnit.UpdateUnitCommand) (*itUnit.UpdateUnitResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Unit, *domain.Unit]{
		Action:        "update unit",
		DbRepoGetter:  this.unitRepo,
		Data:          cmd,
		ValidateExtra: this.validateBaseUnitUpdate,
	})
}

func (this *UnitServiceImpl) DeleteUnit(ctx corectx.Context, cmd itUnit.DeleteUnitCommand) (*itUnit.DeleteUnitResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete unit",
		DbRepoGetter: this.unitRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *UnitServiceImpl) GetUnit(ctx corectx.Context, query itUnit.GetUnitQuery) (*itUnit.GetUnitResult, error) {
	var q dyn.GetOneQuery
	if query.Id != nil {
		q.Id = *query.Id
	}
	q.Columns = query.Columns
	return corecrud.GetOne[domain.Unit](ctx, corecrud.GetOneParam{
		Action:       "get unit",
		DbRepoGetter: this.unitRepo,
		Query:        q,
	})
}

func (this *UnitServiceImpl) SearchUnits(ctx corectx.Context, query itUnit.SearchUnitsQuery) (*itUnit.SearchUnitsResult, error) {
	return corecrud.Search[domain.Unit](ctx, corecrud.SearchParam{
		Action:       "search units",
		DbRepoGetter: this.unitRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *UnitServiceImpl) UnitExists(ctx corectx.Context, query itUnit.UnitExistsQuery) (*itUnit.UnitExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if unit exists",
		DbRepoGetter: this.unitRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

// Business rule validation ----------------------------------------------------------------

func (this *UnitServiceImpl) validateBaseUnitCreate(ctx corectx.Context, unit *domain.Unit, vErrs *ft.ClientErrors) error {
	baseUnitId := unit.GetBaseUnit()
	multiplier := unit.GetMultiplier()

	if baseUnitId != nil {
		if multiplier == nil || *multiplier <= 0 {
			vErrs.Append(*ft.NewBusinessViolation("multiplier", "unit.multiplier_required_with_base_unit",
				"multiplier must be provided and greater than 0 when base unit is provided"))
			return nil
		}
		baseUnit, err := this.assertBaseUnitValid(ctx, baseUnitId, vErrs)
		if err != nil || baseUnit == nil {
			return err
		}
		if baseUnit.GetCategoryId() != unit.GetCategoryId() {
			vErrs.Append(*ft.NewBusinessViolation("category_id", "unit.category_must_match_base_unit",
				"unit category must be same as base unit category"))
		}
	} else {
		if multiplier != nil {
			vErrs.Append(*ft.NewBusinessViolation("multiplier", "unit.multiplier_not_allowed_without_base_unit",
				"multiplier must not be provided when base unit is not provided"))
		}
	}
	return nil
}

func (this *UnitServiceImpl) validateBaseUnitUpdate(ctx corectx.Context, unit *domain.Unit, foundUnit *domain.Unit, vErrs *ft.ClientErrors) error {
	fields := unit.GetFieldData()
	_, hasBaseUnit := fields[domain.UnitFieldBaseUnit]
	_, hasMultiplier := fields[domain.UnitFieldMultiplier]

	if hasBaseUnit || hasMultiplier {
		if foundUnit.GetStatus() != nil && *foundUnit.GetStatus() == domain.UnitStatusActive {
			vErrs.Append(*ft.NewBusinessViolation("status", "unit.active_unit_immutable_fields",
				"cannot update base_unit or multiplier of an active unit"))
		}
	}

	return this.validateBaseUnitCreate(ctx, unit, vErrs)
}

func (this *UnitServiceImpl) assertBaseUnitValid(ctx corectx.Context, baseUnitId *string, vErrs *ft.ClientErrors) (*domain.Unit, error) {
	result, err := this.unitRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{basemodel.FieldId: *baseUnitId},
	})
	if err != nil {
		return nil, err
	}
	if !result.HasData {
		vErrs.Append(*ft.NewBusinessViolation("base_unit", "unit.base_unit_not_found",
			"base unit does not exist"))
		return nil, nil
	}
	baseUnit := result.Data
	if baseUnit.GetBaseUnit() != nil {
		vErrs.Append(*ft.NewBusinessViolation("base_unit", "unit.base_unit_cannot_be_derived",
			"base unit cannot itself be a derived unit"))
		return nil, nil
	}
	return &baseUnit, nil
}
