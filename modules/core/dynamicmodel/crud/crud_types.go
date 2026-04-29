package crud

import dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

type ServiceCreateOptions[TModel any] struct {
	AfterValidationSuccess AfterValidationSuccessFn[TModel]
}

type ServiceUpdateOptions[TModel any] ServiceCreateOptions[TModel]

type ServiceDeleteOptions struct {
	AfterValidationSuccess AfterValidationSuccessFn[dyn.DeleteOneCommand]
}

type ServiceSearchOptions struct {
	AfterValidationSuccess AfterValidationSuccessFn[dyn.SearchQuery]
}
