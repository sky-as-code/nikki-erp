package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces"
	// itUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/unit_category/interfaces"
)

func NewUnitServiceImpl(
	unitRepo it.UnitRepository,
	// unitCategoryService itUnitCategory.UnitCategoryService,
) it.UnitService {
	return &UnitServiceImpl{
		unitRepo: unitRepo,
		// unitCategoryService: unitCategoryService,
	}
}

type UnitServiceImpl struct {
	unitRepo it.UnitRepository
	// unitCategoryService itUnitCategory.UnitCategoryService
}

// Create

func (this *UnitServiceImpl) CreateUnit(ctx crud.Context, cmd it.CreateUnitCommand) (*it.CreateUnitResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*it.Unit, it.CreateUnitCommand, it.CreateUnitResult]{
		Action:              "create unit",
		Command:             cmd,
		AssertBusinessRules: this.assertCreateUnitRules,
		SetDefault:          this.setDefaultsUnit,
		RepoCreate:          this.unitRepo.Create,
		Sanitize:            this.sanitizeUnit,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateUnitResult {
			return &it.CreateUnitResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Unit) *it.CreateUnitResult {
			return &it.CreateUnitResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (this *UnitServiceImpl) UpdateUnit(ctx crud.Context, cmd it.UpdateUnitCommand) (*it.UpdateUnitResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*it.Unit, it.UpdateUnitCommand, it.UpdateUnitResult]{
		Action:              "update unit",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateUnitRules,
		AssertExists:        this.assertUnitIdExists,
		RepoUpdate:          this.unitRepo.Update,
		Sanitize:            this.sanitizeUnit,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateUnitResult {
			return &it.UpdateUnitResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Unit) *it.UpdateUnitResult {
			return &it.UpdateUnitResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (this *UnitServiceImpl) DeleteUnit(ctx crud.Context, cmd it.DeleteUnitCommand) (*it.DeleteUnitResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*it.Unit, it.DeleteUnitCommand, it.DeleteUnitResult]{
		Action:       "delete unit",
		Command:      cmd,
		AssertExists: this.assertUnitIdExists,
		RepoDelete: func(ctx crud.Context, model *it.Unit) (int, error) {
			return this.unitRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteUnitResult {
			return &it.DeleteUnitResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Unit, deletedCount int) *it.DeleteUnitResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (this *UnitServiceImpl) GetUnitById(ctx crud.Context, query it.GetUnitByIdQuery) (*it.GetUnitByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*it.Unit, it.GetUnitByIdQuery, it.GetUnitByIdResult]{
		Action: "get unit by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetUnitByIdQuery, vErrs *ft.ValidationErrors) (*it.Unit, error) {
			dbUnit, err := this.unitRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbUnit == nil {
				vErrs.AppendNotFound("id", "unit id")
			}
			return dbUnit, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetUnitByIdResult {
			return &it.GetUnitByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Unit) *it.GetUnitByIdResult {
			return &it.GetUnitByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *UnitServiceImpl) SearchUnits(ctx crud.Context, query it.SearchUnitsQuery) (*it.SearchUnitsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[it.Unit, it.SearchUnitsQuery, it.SearchUnitsResult]{
		Action: "search units",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchUnitsQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.unitRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query it.SearchUnitsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[it.Unit], error) {
			return this.unitRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchUnitsResult {
			return &it.SearchUnitsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[it.Unit]) *it.SearchUnitsResult {
			return &it.SearchUnitsResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (this *UnitServiceImpl) assertCreateUnitRules(ctx crud.Context, unit *it.Unit, vErrs *ft.ValidationErrors) error {
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

func (this *UnitServiceImpl) assertUpdateUnitRules(ctx crud.Context, unit *it.Unit, dbUnit *it.Unit, vErrs *ft.ValidationErrors) error {
	if *dbUnit.Status == "active" { // only update BaseUnit and Multiplier when status is not active
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
func (this *UnitServiceImpl) sanitizeUnit(_ *it.Unit) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (this *UnitServiceImpl) assertUnitIdExists(ctx crud.Context, unit *it.Unit, vErrs *ft.ValidationErrors) (*it.Unit, error) {
	dbUnit, err := this.unitRepo.FindById(ctx, it.FindByIdParam{
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

// func (this *UnitServiceImpl) assertUnitCategoryExists(ctx crud.Context, unitCategoryId model.Id, vErrs *ft.ValidationErrors) (*itUnitCategory.UnitCategory, error) {
// 	dbUnitCategory, err := this.unitCategoryService.GetUnitCategoryById(ctx, itUnitCategory.GetUnitCategoryByIdQuery{
// 		Id: unitCategoryId,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	if dbUnitCategory.Data == nil {
// 		vErrs.Append("id", "unit category does not exist")
// 		return nil, nil
// 	}

// 	return dbUnitCategory.Data, nil
// }

func (this *UnitServiceImpl) assertDerivedUnit(ctx crud.Context, unit *it.Unit, vErrs *ft.ValidationErrors) (*it.Unit, error) {
	baseUnit, err := this.GetUnitById(ctx, it.GetUnitByIdQuery{
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

func (this *UnitServiceImpl) setDefaultsUnit(unit *it.Unit) {
	unit.SetDefaults()
}

// func (this *UnitServiceImpl) assertUnitName
