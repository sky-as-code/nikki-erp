package crud

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

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

type DeletionResultData struct {
	Id        model.Id  `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type DeletionResult = OpResult[*DeletionResultData]

func NewSuccessDeletionResult(id model.Id) *DeletionResult {
	return &DeletionResult{
		Data: &DeletionResultData{
			Id:        id,
			DeletedAt: time.Now(),
		},
		HasData: true,
	}
}
