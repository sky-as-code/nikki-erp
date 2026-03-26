package dynamicmodel

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type BaseRepository interface {
	Insert(ctx corectx.Context, data dmodel.DynamicFields) (*crud.OpResult[dmodel.DynamicFields], error)
	Update(ctx corectx.Context, data dmodel.DynamicFields, prevEtag string) (*crud.OpResult[dmodel.DynamicFields], error)
	GetOne(ctx corectx.Context, param GetOneParam) (*crud.OpResult[dmodel.DynamicFields], error)
	Search(ctx corectx.Context, param SearchParam) (*crud.OpResult[crud.PagedResult[dmodel.DynamicFields]], error)
	Archive(ctx corectx.Context, keys dmodel.DynamicFields) (*crud.OpResult[dmodel.DynamicFields], error)
	Delete(ctx corectx.Context, keys dmodel.DynamicFields) (*crud.OpResult[int64], error)
	// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
	CheckUniqueCollisions(ctx corectx.Context, data dmodel.DynamicFields) (*crud.OpResult[[][]string], error)
	GetSchema() *dmodel.ModelSchema
}

type GetOneParam struct {
	Filter          dmodel.DynamicFields
	Columns         []string
	IncludeArchived bool
}

type SearchParam struct {
	Graph           *dmodel.SearchGraph
	Columns         []string
	Filter          []dmodel.DynamicFields
	IncludeArchived bool
	Page            int
	Size            int
}

type BaseRepoGetter interface {
	GetBaseRepo() BaseRepository
}

type DynamicModelPtr[TDomain any] interface {
	*TDomain
	dmodel.DynamicModel
}
