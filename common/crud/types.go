package crud

type PagingOptions struct {
	Page int `json:"page" query:"page"`
	Size int `json:"size" query:"size"`
}

type PagedResult[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}
