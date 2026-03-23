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

type BaseRepository interface {
	Insert(ctx Context, data schema.DynamicFields) (schema.DynamicFields, error)
	Update(ctx Context, data schema.DynamicFields) (schema.DynamicFields, error)
	FindByPk(ctx Context, keys schema.DynamicFields) (schema.DynamicFields, error)
	Search(ctx Context, graph schema.SearchGraph, columns []string, filter ...schema.DynamicFields) ([]schema.DynamicFields, error)
	Archive(ctx Context, keys schema.DynamicFields) (schema.DynamicFields, error)
	Delete(ctx Context, keys schema.DynamicFields) (int64, error)
	// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
	CheckUniqueCollisions(ctx Context, data schema.DynamicFields) ([][]string, error)
	GetSchema() *schema.EntitySchema
}

type BaseRepoGetter interface {
	GetBaseRepo() BaseRepository
}

type DynamicModelPtr[TDomain any] interface {
	*TDomain
	schema.DynamicModel
}
