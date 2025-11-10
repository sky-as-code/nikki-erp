package crud

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

type DomainModelProducer[TDomain any] interface {
	ToDomainModel() TDomain
}

type DomainModelBulkProducer[TDomain any] interface {
	ToDomainModels() []TDomain
}

type DeleteCommander[TDomain any] interface {
	DomainModelProducer[TDomain]
	Validatable
}

type Searchable interface {
	Validatable
	GetGraph() *string
}

type Etagger interface {
	GetEtag() *model.Etag
	SetEtag(etag model.Etag)
}

type Validatable interface {
	Validate() ft.ValidationErrors
}

type ValidatableForEdit interface {
	Validate(forEdit bool) ft.ValidationErrors
}

type RepoFindByIdParam interface {
	GetId() model.Id
	SetId(id model.Id)
}

type PagingOptions = db.PagingOptions
type PagedResult[T any] = db.PagedResult[T]

type OpResult[TData any] struct {
	Data TData `json:"data"`

	// Indicates whether "Data" has zero value (only supports empty struct, nil pointer and empty array)
	//
	// If ClientError is nil but HasData is false,
	// it means the query is successfull but doesn't return any data.
	HasData     bool            `json:"hasData"`
	ClientError *ft.ClientError `json:"error,omitempty"`
}

func (this OpResult[TData]) GetClientError() *ft.ClientError {
	return this.ClientError
}

func (this OpResult[TData]) GetHasData() bool {
	return this.HasData
}

type DeletionResultData struct {
	Id           model.Id  `json:"id"`
	DeletedAt    time.Time `json:"deletedAt"`
	DeletedCount *int      `json:"deletedCount,omitempty"`
}

type DeletionResult = OpResult[*DeletionResultData]

func NewSuccessDeletionResult(id model.Id, deletedCount ...*int) *DeletionResult {
	var del *int
	if len(deletedCount) > 0 {
		del = deletedCount[0]
	}
	return &DeletionResult{
		Data: &DeletionResultData{
			Id:           id,
			DeletedAt:    time.Now(),
			DeletedCount: del,
		},
		HasData: true,
	}
}

type ExistsResult = OpResult[bool]

func NewSuccessExistsResult(isExisting bool) *ExistsResult {
	return &ExistsResult{
		Data:    isExisting,
		HasData: true,
	}
}

type SearchQuery struct {
	Page  *int    `json:"page" query:"page"`
	Size  *int    `json:"size" query:"size"`
	Graph *string `json:"graph" query:"graph"`
}

func (this *SearchQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchQuery) GetGraph() *string {
	return this.Graph
}

func (this *SearchQuery) ValidationRules() []*val.FieldRules {
	return []*val.FieldRules{
		val.Field(&this.Page,
			val.Min(model.MODEL_RULE_PAGE_INDEX_START),
			val.Max(model.MODEL_RULE_PAGE_INDEX_END),
		),
		val.Field(&this.Size,
			val.Min(model.MODEL_RULE_PAGE_MIN_SIZE),
			val.Max(model.MODEL_RULE_PAGE_MAX_SIZE),
		),
		// PageIndexValidateRule(&this.Page),
		// PageSizeValidateRule(&this.Size),
	}
}

func PageIndexValidateRule(field **int) *val.FieldRules {
	return val.Field(field,
		val.Min(model.MODEL_RULE_PAGE_INDEX_START),
		val.Max(model.MODEL_RULE_PAGE_INDEX_END),
	)
}

func PageSizeValidateRule(field **int) *val.FieldRules {
	return val.Field(field, val.When(*field != nil,
		val.NotEmpty,
		val.Min(model.MODEL_RULE_PAGE_MIN_SIZE),
		val.Max(model.MODEL_RULE_PAGE_MAX_SIZE),
	))
}
