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
	Id      model.Id `json:"id" param:"id"`
	Columns []string `json:"columns" query:"columns"`
}

type SearchQuery struct {
	Columns []string `json:"columns" query:"columns"`
	Page    int      `json:"page" query:"page"`
	Size    int      `json:"size" query:"size"`
	// Optional search graph for advanced search
	Graph *dmodel.SearchGraph `json:"graph" query:"graph"`
	// Optional language code to filter fields with LangJson type
	Language *model.LanguageCode `json:"language" query:"language"`
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
}
