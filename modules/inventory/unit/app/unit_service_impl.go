package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
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

// Create

func (this *UnitServiceImpl) CreateUnit(ctx crud.Context, cmd itUnit.CreateUnitCommand) (*itUnit.CreateUnitResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.Unit, itUnit.CreateUnitCommand, itUnit.CreateUnitResult]{
		Action:              "create unit",
		Command:             cmd,
		AssertBusinessRules: this.assertCreateUnitRules,
		SetDefault:          this.setDefaultsUnit,
		RepoCreate:          this.unitRepo.Create,
		Sanitize:            this.sanitizeUnit,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnit.CreateUnitResult {
			return &itUnit.CreateUnitResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Unit) *itUnit.CreateUnitResult {
			return &itUnit.CreateUnitResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (this *UnitServiceImpl) UpdateUnit(ctx crud.Context, cmd itUnit.UpdateUnitCommand) (*itUnit.UpdateUnitResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Unit, itUnit.UpdateUnitCommand, itUnit.UpdateUnitResult]{
		Action:              "update unit",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateUnitRules,
		AssertExists:        this.assertUnitIdExists,
		RepoUpdate:          this.unitRepo.Update,
		Sanitize:            this.sanitizeUnit,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnit.UpdateUnitResult {
			return &itUnit.UpdateUnitResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Unit) *itUnit.UpdateUnitResult {
			return &itUnit.UpdateUnitResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (this *UnitServiceImpl) DeleteUnit(ctx crud.Context, cmd itUnit.DeleteUnitCommand) (*itUnit.DeleteUnitResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Unit, itUnit.DeleteUnitCommand, itUnit.DeleteUnitResult]{
		Action:       "delete unit",
		Command:      cmd,
		AssertExists: this.assertUnitIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.Unit) (int, error) {
			return this.unitRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnit.DeleteUnitResult {
			return &itUnit.DeleteUnitResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Unit, deletedCount int) *itUnit.DeleteUnitResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (this *UnitServiceImpl) GetUnitById(ctx crud.Context, query itUnit.GetUnitByIdQuery) (*itUnit.GetUnitByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Unit, itUnit.GetUnitByIdQuery, itUnit.GetUnitByIdResult]{
		Action: "get unit by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itUnit.GetUnitByIdQuery, vErrs *ft.ValidationErrors) (*domain.Unit, error) {
			dbUnit, err := this.unitRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbUnit == nil {
				vErrs.AppendNotFound("id", "unit id")
				return nil, nil
			}
			return dbUnit, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnit.GetUnitByIdResult {
			return &itUnit.GetUnitByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Unit) *itUnit.GetUnitByIdResult {
			return &itUnit.GetUnitByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *UnitServiceImpl) SearchUnits(ctx crud.Context, query itUnit.SearchUnitsQuery) (*itUnit.SearchUnitsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Unit, itUnit.SearchUnitsQuery, itUnit.SearchUnitsResult]{
		Action: "search units",
		Query:  query,
		SetQueryDefaults: func(q *itUnit.SearchUnitsQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.unitRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itUnit.SearchUnitsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Unit], error) {
			return this.unitRepo.Search(ctx, itUnit.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itUnit.SearchUnitsResult {
			return &itUnit.SearchUnitsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[domain.Unit]) *itUnit.SearchUnitsResult {
			return &itUnit.SearchUnitsResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (this *UnitServiceImpl) assertCreateUnitRules(ctx crud.Context, unit *domain.Unit, vErrs *ft.ValidationErrors) error {
	if unit.BaseUnit != nil {
		baseUnit, err := this.assertDerivedUnit(ctx, unit, vErrs) // check base unit can be a derived unit
		if err != nil {
			return err
		}

		if baseUnit.CategoryId != unit.CategoryId { // unit and base unit category must be same
			vErrs.Append("categoryId", "unit category must be same as base unit category")
			return nil
		}
	}

	return nil
}

func (this *UnitServiceImpl) assertUpdateUnitRules(ctx crud.Context, unit *domain.Unit, dbUnit *domain.Unit, vErrs *ft.ValidationErrors) error {
	if dbUnit.Status != nil && *dbUnit.Status == "active" { // only update BaseUnit and Multiplier when status is not active
		if unit.BaseUnit != nil || unit.Multiplier != nil {
			vErrs.Append("status", "cannot update BaseUnit or Multiplier of an active unit")
			return nil
		}
	}

	if unit.BaseUnit != nil {
		baseUnit, err := this.assertDerivedUnit(ctx, unit, vErrs) // check base unit can be a derived unit
		if err != nil {
			return err
		}

		if baseUnit.CategoryId != unit.CategoryId { // unit and base unit category must be same
			vErrs.Append("categoryId", "unit category must be same as base unit category")
			return nil
		}
	}

	return nil
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (this *UnitServiceImpl) sanitizeUnit(_ *domain.Unit) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (this *UnitServiceImpl) assertUnitIdExists(ctx crud.Context, unit *domain.Unit, vErrs *ft.ValidationErrors) (*domain.Unit, error) {
	dbUnit, err := this.unitRepo.FindById(ctx, itUnit.FindByIdParam{
		Id: *unit.Id,
	})
	if err != nil {
		return nil, err
	}

	if dbUnit == nil {
		vErrs.Append("id", "unit not found")
		return nil, nil
	}

	return dbUnit, nil
}

func (this *UnitServiceImpl) assertDerivedUnit(ctx crud.Context, unit *domain.Unit, vErrs *ft.ValidationErrors) (*domain.Unit, error) {
	baseUnit, err := this.GetUnitById(ctx, itUnit.GetUnitByIdQuery{
		Id: *unit.BaseUnit,
	})
	if err != nil {
		return nil, err
	}

	if baseUnit.Data == nil {
		vErrs.Append("id", "base unit does not exist")
		return nil, nil
	}

	if baseUnit.Data.BaseUnit != nil {
		vErrs.Append("id", "base unit cannot be a Derived Unit")
		return nil, nil
	}

	return baseUnit.Data, nil
}

func (this *UnitServiceImpl) setDefaultsUnit(unit *domain.Unit) {
	unit.SetDefaults()
}
