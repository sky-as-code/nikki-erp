package unit

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

type UnitRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.Unit) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.Unit) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, unit domain.Unit) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Unit], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Unit]], error)
	Update(ctx corectx.Context, unit domain.Unit) (*dyn.OpResult[dyn.MutateResultData], error)
}

type UnitService interface {
	CreateUnit(ctx corectx.Context, cmd CreateUnitCommand) (*CreateUnitResult, error)
	UpdateUnit(ctx corectx.Context, cmd UpdateUnitCommand) (*UpdateUnitResult, error)
	DeleteUnit(ctx corectx.Context, cmd DeleteUnitCommand) (*DeleteUnitResult, error)
	GetUnit(ctx corectx.Context, query GetUnitQuery) (*GetUnitResult, error)
	SearchUnits(ctx corectx.Context, query SearchUnitsQuery) (*SearchUnitsResult, error)
	UnitExists(ctx corectx.Context, query UnitExistsQuery) (*UnitExistsResult, error)
}
