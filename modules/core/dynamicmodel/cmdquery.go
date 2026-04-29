package dynamicmodel

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
)

type SetIsArchivedCommand struct {
	Id         model.Id   `json:"id" param:"id"`
	Etag       model.Etag `json:"etag" param:"etag"`
	IsArchived *bool      `json:"is_archived" param:"is_archived"` // Pointer to trigger missing field error
}

type DeleteOneCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (this DeleteOneCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.delete_one_command",
		func() *dmodel.ModelSchemaBuilder {
			return DeleteOneQuerySchemaBuilder()
		},
	)
}

// Helper type for generic type constraints.
type DeleteOneCommandShape interface {
	// MUST change this when `DeleteOneCommand` fields change.
	~struct {
		Id model.Id `json:"id" param:"id"`
	}
}

type ExistsQuery struct {
	Ids []model.Id `json:"ids" query:"ids"`
}

type ExistsResultData struct {
	Existing    []model.Id `json:"existing"`
	NotExisting []model.Id `json:"not_existing"`
}

func (this ExistsResultData) Exists(id model.Id) bool {
	return array.Contains(this.Existing, id)
}

// Helper type for generic type constraints.
type ExistsQueryShape interface {
	// MUST change this when `ExistsQuery` fields change.
	~struct {
		Ids []model.Id `json:"ids" query:"ids"`
	}
}

type GetOneQuery struct {
	Id     model.Id `json:"id" param:"id"`
	Fields []string `json:"fields" query:"fields"`
}

type SearchQuery struct {
	Fields []string `json:"fields" query:"fields"`
	Page   int      `json:"page" query:"page"`
	Size   int      `json:"size" query:"size"`
	// Optional search graph for advanced search
	Graph *dmodel.SearchGraph `json:"graph" query:"graph"`
	// Optional language code to filter fields with LangJson type
	Language *model.LanguageCode `json:"language" query:"language"`

	// Determines the fields to be returned in the response.
	// If not specified, the view "auto" will be used,
	// which will return all fields that user has permission on.
	View string `json:"view" query:"view"`
}

type OpResult[TData any] struct {
	// The result data when success. It is only meaningful if HasData is true and ClientErrors is nil.
	// Otherwise, it could be nil or an empty struct.
	Data TData

	// Contains validation errors, business errors...,
	// or nil if there is no violation.
	ClientErrors ft.ClientErrors

	// Indicates whether "Data" is present (non-zero value: non-empty struct, non-empty array, etc.).
	//
	// If ClientErrors is nil but HasData is false,
	// it means the query is successfull but no data is found.
	HasData bool
}

type MutateResultData struct {
	AffectedCount int                 `json:"affected_count"`
	AffectedAt    model.ModelDateTime `json:"affected_at"`
	Etag          model.Etag          `json:"etag,omitempty"`
}

type PagingOptions struct {
	Page int `json:"page" query:"page"`
	Size int `json:"size" query:"size"`
}

type PagedResultData[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`

	// Determines the fields to be returned in the response.
	// If the request does not specify the fields, the view "auto" will be used,
	// which will return all fields that user has permission to view.
	// View string `json:"view"`

	// List of fields determined by `View` which are requested to read.
	DesiredFields []string `json:"desired_fields"`

	// Subset of `DesiredFields` that user doesn't have permission to read.
	// These fields will be returned as `nil` values. User can be aware of their existence
	// and ask for permission to read them.
	MaskedFields []string `json:"masked_fields"`

	// The etag of the schema used to generate the response.
	// This is used to check if the schema has changed since the last request.
	// If the schema has changed, the client should fetch the new schema and update its cache.
	SchemaEtag string `json:"schema_etag"`
}

type SingleResultData[T any] struct {
	Item T              `json:"item"`
	Meta SingleMetaData `json:"meta"`
}

type SingleMetaData struct {
	// Determines the fields to be returned in the response.
	// If the request does not specify the fields, the view "auto" will be used,
	// which will return all fields that user has permission on.
	// View string `json:"view"`

	// List of fields determined by `View` which are requested to read.
	DesiredFields []string `json:"desired_fields"`

	// Subset of `DesiredFields` that user doesn't have permission to read.
	// These fields will be returned as `nil` values. User can be aware of their existence
	// and ask for permission to read them.
	MaskedFields []string `json:"masked_fields"`

	// The etag of the schema used to generate the response.
	// This is used to check if the schema has changed since the last request.
	// If the schema has changed, the client should fetch the new schema and update its cache.
	SchemaEtag string `json:"schema_etag"`
}
