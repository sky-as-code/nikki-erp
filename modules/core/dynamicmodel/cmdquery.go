package dynamicmodel

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type SetIsArchivedCommand struct {
	Id         model.Id   `json:"id" param:"id"`
	Etag       model.Etag `json:"etag" param:"etag"`
	IsArchived *bool      `json:"is_archived" param:"is_archived"` // Pointer to trigger missing field error
}

type ToDomainModelFunc[TDomain any] func(data dmodel.DynamicFields) TDomain
type BeforeValidationFunc[TDomain any] func(ctx corectx.Context, model TDomain, vErrs *ft.ClientErrors) (TDomain, error)
type AfterValidationFunc[TDomain any] func(ctx corectx.Context, model TDomain) (TDomain, error)
type ValidateExtraFunc[TDomain any] func(ctx corectx.Context, model TDomain, vErrs *ft.ClientErrors) error

type CreateParam[
	TDomain any,
	TDomainPtr DynamicModelPtr[TDomain],
] struct {
	// Action name for logging and error messages
	Action         string
	BaseRepoGetter DynamicModelRepository

	// Data to create
	Data dmodel.DynamicModelGetter

	// Optional function to do some processing on the domain model before validation.
	BeforeValidation BeforeValidationFunc[TDomainPtr]

	// Optional function to do some processing on the domain model after validation.
	AfterValidation AfterValidationFunc[TDomainPtr]

	// Optional function for advanced validation (business rules) in addition to built-in schema validation.
	ValidateExtra ValidateExtraFunc[TDomainPtr]
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

type ExistsQuery struct {
	Ids []model.Id `json:"ids" param:"ids"`
}

type ExistsResultData struct {
	Existing    []model.Id `json:"existing"`
	NotExisting []model.Id `json:"not_existing"`
}

func (this ExistsResultData) Exists(id model.Id) bool {
	return array.Contains(this.Existing, id)
}

type GetOneQuery struct {
	Id      model.Id `json:"id" param:"id"`
	Columns []string `json:"columns" query:"columns"`
}

type SearchQuery struct {
	Columns []string            `json:"columns" query:"columns"`
	Graph   *dmodel.SearchGraph `json:"graph" query:"graph"`
	Page    int                 `json:"page" query:"page"`
	Size    int                 `json:"size" query:"size"`
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
