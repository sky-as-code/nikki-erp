package dynamicentity

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

type OpResult[TData any] struct {
	// The result data when success. It is only meaningful if IsEmpty is false and ClientErrors is nil.
	// Otherwise, it could be nil or an empty struct.
	Data TData `json:"data"`

	// Contains validation errors, business errors...,
	// or nil if there is no violation.
	ClientErrors ft.ClientErrors `json:"errors,omitempty"`

	// Indicates whether "Data" has zero value (ie: empty struct, empty array)
	//
	// If ClientErrors is nil but IsEmpty is true,
	// it means the query is successfull but no data is found.
	IsEmpty bool `json:"isEmpty"`
}

type PagingOptions struct {
	Page int `json:"page" query:"page"`
	Size int `json:"size" query:"size"`
}

type PagedResult[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}

type BaseRepository interface {
	Insert(ctx Context, data schema.DynamicFields) (*OpResult[schema.DynamicFields], error)
	Update(ctx Context, data schema.DynamicFields, prevEtag string) (*OpResult[schema.DynamicFields], error)
	GetOne(ctx Context, param GetOneParam) (*OpResult[schema.DynamicFields], error)
	Search(ctx Context, param SearchParam) (*OpResult[PagedResult[schema.DynamicFields]], error)
	Archive(ctx Context, keys schema.DynamicFields) (*OpResult[schema.DynamicFields], error)
	Delete(ctx Context, keys schema.DynamicFields) (*OpResult[int64], error)
	// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
	CheckUniqueCollisions(ctx Context, data schema.DynamicFields) (*OpResult[[][]string], error)
	GetSchema() *schema.EntitySchema
}

type GetOneParam struct {
	Filter          schema.DynamicFields
	Columns         []string
	IncludeArchived bool
}

type SearchParam struct {
	Graph           schema.SearchGraph
	Columns         []string
	Filter          []schema.DynamicFields
	IncludeArchived bool
	Page            int
	Size            int
}

type BaseRepoGetter interface {
	GetBaseRepo() BaseRepository
}

type DynamicModelPtr[TDomain any] interface {
	*TDomain
	schema.DynamicModel
}
