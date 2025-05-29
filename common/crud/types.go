package crud

type PagingOptions struct {
	Page int
	Size int
}

type PagedResult[T any] struct {
	Items []T
	Total int
}
