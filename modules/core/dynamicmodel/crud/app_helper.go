package crud

import (
	"go.bryk.io/pkg/errors"

	// dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type FieldsResolver interface {
	GetListFields(ctx corectx.Context, uiName string, userId model.Id) (*dyn.OpResult[[]string], error)
}

type UiSearchParam[TDomain any, TDomainPtr dyn.DynamicModelPtr[TDomain]] struct {
	Action        string
	FieldResolver FieldsResolver
	Schema        *dmodel.ModelSchema
	// Name of the saved search view.
	ExpectedSearchName string
	// Default name to use when ExpectedSearchName is not specified.
	DefaultSearchName string
	SearchFn          SearchFn[TDomain]
}

type SearchFn[TDomain any] func(
	fn AfterValidationSuccessFn[dyn.SearchQuery],
) (*dyn.OpResult[dyn.PagedResultData[TDomain]], error)

func UiSearch[TDomain any, TDomainPtr dyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, param UiSearchParam[TDomain, TDomainPtr],
) (_ *dyn.OpResult[dyn.PagedResultData[TDomain]], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	desiredFields := []string{}
	maskedFields := []string{}

	result, err := param.SearchFn(func(ctx corectx.Context, query dyn.SearchQuery) (dyn.SearchQuery, error) {
		isClientSpecifiedFields := len(query.Fields) > 0

		if !isClientSpecifiedFields {
			uiFields, err := getListFields(ctx, param.FieldResolver, param.ExpectedSearchName)
			if err != nil {
				return query, err
			}
			query.Fields = uiFields
		}
		desiredFields = query.Fields

		// TODO: Determine masked fields with field-level permission
		// maskedFields = []string{}

		return query, nil
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.PagedResultData[TDomain]]{ClientErrors: result.ClientErrors}, nil
	}

	for _, item := range result.Data.Items {
		data := TDomainPtr(&item).GetFieldData()
		for _, field := range maskedFields {
			// User doesn't have permission to read this field,
			// but can be aware of its existence because it is included in the result.
			data[field] = nil
		}
	}
	result.Data.MaskedFields = maskedFields
	result.Data.DesiredFields = desiredFields
	result.Data.SchemaEtag = param.Schema.Etag()

	return result, nil
}

func getListFields(ctx corectx.Context, fieldResolver FieldsResolver, uiName string) ([]string, error) {
	uiFields, err := fieldResolver.GetListFields(ctx, uiName, ctx.GetPermissions().UserId)
	if err != nil {
		return nil, err
	}
	if uiFields.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(uiFields.ClientErrors.ToError(), "UiSearch")
	}
	if !uiFields.HasData {
		// TODO: Set default fields
	}
	return uiFields.Data, nil
}

type UiGetOneParam[TDomain any, TDomainPtr dyn.DynamicModelPtr[TDomain]] struct {
	Action   string
	Schema   *dmodel.ModelSchema
	GetOneFn GetOneFn[TDomain]
}

type GetOneFn[TDomain any] func() (*dyn.OpResult[TDomain], error)

func UiGetOne[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context, param UiGetOneParam[TDomain, TDomainPtr],
) (_ *dyn.OpResult[dyn.SingleResultData[TDomain]], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	result, err := param.GetOneFn()
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.SingleResultData[TDomain]]{ClientErrors: result.ClientErrors}, nil
	}

	maskedFields := []string{}
	data := TDomainPtr(&result.Data).GetFieldData()
	for _, field := range maskedFields {
		// User doesn't have permission to read this field,
		// but can be aware of its existence because it is included in the result.
		data[field] = nil
	}

	return &dyn.OpResult[dyn.SingleResultData[TDomain]]{Data: dyn.SingleResultData[TDomain]{
		Item: result.Data,
		Meta: dyn.SingleMetaData{
			DesiredFields: param.Schema.FieldNames(),
			SchemaEtag:    param.Schema.Etag(),
		},
	}, HasData: true}, nil
}
