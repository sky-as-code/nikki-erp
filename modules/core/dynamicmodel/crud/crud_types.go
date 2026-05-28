package crud

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type BeforeValidationFn[T any] func(ctx corectx.Context, model T, vErrs *ft.ClientErrors) (T, error)
type AfterValidationSuccessFn[T any] func(ctx corectx.Context, model T) (T, error)
type CreateValidateExtraFn[T any] func(ctx corectx.Context, inputModel T, vErrs *ft.ClientErrors) error
type UpdateValidateExtraFn[T any] func(ctx corectx.Context, inputModel T, foundModel T, vErrs *ft.ClientErrors) error
type DeleteValidateExtraFn func(ctx corectx.Context, keyFields dmodel.DynamicFields, vErrs *ft.ClientErrors) error

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
