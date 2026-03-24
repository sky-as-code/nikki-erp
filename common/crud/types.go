package crud

import (
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
