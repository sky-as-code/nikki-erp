package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUnit "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unit"
)

type UnitExtService interface {
	GetUnit(ctx corectx.Context, query GetUnitQuery) (*GetUnitResult, error)
	UnitExists(ctx corectx.Context, query UnitExistsQuery) (*UnitExistsResult, error)
}

type GetUnitQuery = itUnit.GetUnitQuery
type GetUnitResult = itUnit.GetUnitResult
type UnitExistsQuery = itUnit.UnitExistsQuery
type UnitExistsResult = itUnit.UnitExistsResult
