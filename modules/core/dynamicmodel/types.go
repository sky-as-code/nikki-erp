package dynamicmodel

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type BaseRepository interface {
	// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
	CheckUniqueCollisions(ctx corectx.Context, data dmodel.DynamicFields) (*crud.OpResult[[][]string], error)
	// DeleteOne deletes a single record by primary key then returns the number of affected rows.
	// If affected rows is 0, the record is not found.
	DeleteOne(ctx corectx.Context, keys dmodel.DynamicFields) (*crud.OpResult[int], error)
	Insert(ctx corectx.Context, data dmodel.DynamicFields) (*crud.OpResult[dmodel.DynamicFields], error)
	GetOne(ctx corectx.Context, param RepoGetOneParam) (*crud.OpResult[dmodel.DynamicFields], error)
	Search(ctx corectx.Context, param RepoSearchParam) (*crud.OpResult[crud.PagedResultData[dmodel.DynamicFields]], error)
	Update(ctx corectx.Context, data dmodel.DynamicFields) (*crud.OpResult[dmodel.DynamicFields], error)
	GetSchema() *dmodel.ModelSchema
}

type RepoGetOneParam struct {
	Filter  dmodel.DynamicFields
	Columns []string
}

type RepoSearchParam struct {
	Graph   *dmodel.SearchGraph
	Columns []string
	Filter  []dmodel.DynamicFields
	Page    int
	Size    int
}

type BaseRepoGetter interface {
	GetBaseRepo() BaseRepository
}

type DynamicModelPtr[TDomain any] interface {
	*TDomain
	dmodel.DynamicModel
}
