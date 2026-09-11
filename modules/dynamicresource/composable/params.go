package composable

import (
	"encoding/json"
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The generic helpers of modules/core/dynamicmodel/crud take typed commands, while the onion
// speaks DynamicFields end to end. These functions convert one into the other. They only
// reshape data; every value is validated afterwards by the crud helper itself.

// ParamId reads the record id out of a command or query. Empty when absent.
func ParamId(params dmodel.DynamicFields) model.Id {
	return readId(params, basemodel.FieldId)
}

// ParamOrgId reads the org the caller named, or nil when the params carry none.
func ParamOrgId(params dmodel.DynamicFields) *model.Id {
	raw := readString(params, basemodel.FieldOrgId)
	if raw == "" {
		return nil
	}
	orgId := model.Id(raw)
	return &orgId
}

// ParamString reads a string-typed param. model.Id and model.Etag are string types, so the
// string case covers them too. Anything else answers empty.
func ParamString(params dmodel.DynamicFields, field string) string {
	return readString(params, field)
}

// ParamBool reads a bool or *bool param. The second value reports whether one was present.
func ParamBool(params dmodel.DynamicFields, field string) (bool, bool) {
	return readBool(params, field)
}

// decodeParams reshapes a field map into the given typed command through a JSON round trip,
// which reuses the json tags already declared on the core command structs.
func decodeParams(params dmodel.DynamicFields, target any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "decodeParams.Marshal")
	}
	return errors.Wrap(json.Unmarshal(raw, target), "decodeParams.Unmarshal")
}

// clientErrorsForDecodeFailure turns a decode failure caused by a wrong-typed query parameter
// into the invalid-format error the caller should see.
//
// A value of the wrong type is the caller's mistake, so answering 500 would report our own
// failure for their bad request. Only json.UnmarshalTypeError carries the offending field;
// anything else is a genuine fault and is left to bubble up.
func clientErrorsForDecodeFailure(err error) (ft.ClientErrors, bool) {
	var typeErr *json.UnmarshalTypeError
	if !stdErr.As(err, &typeErr) || typeErr.Field == "" {
		return nil, false
	}
	cErrs := ft.ClientErrors{}
	cErrs.Append(*dmodel.NewInvalidDataTypeErr(typeErr.Field, typeErr.Type.String()))
	return cErrs, true
}

func paramsToDeleteCommand(params dmodel.DynamicFields) dyn.DeleteOneCommand {
	return dyn.DeleteOneCommand{
		Id: readId(params, basemodel.FieldId),
	}
}

func paramsToSetArchivedCommand(params dmodel.DynamicFields) dyn.SetIsArchivedCommand {
	cmd := dyn.SetIsArchivedCommand{
		Id:   readId(params, basemodel.FieldId),
		Etag: model.Etag(readString(params, basemodel.FieldEtag)),
	}
	if archived, ok := readBool(params, basemodel.FieldIsArchived); ok {
		cmd.IsArchived = &archived
	}
	return cmd
}

func paramsToExistsQuery(params dmodel.DynamicFields) (dyn.ExistsQuery, error) {
	query := dyn.ExistsQuery{}
	err := decodeParams(params, &query)
	return query, err
}

func paramsToGetOneQuery(params dmodel.DynamicFields) (dyn.GetOneQuery, error) {
	query := dyn.GetOneQuery{}
	err := decodeParams(params, &query)
	return query, err
}

func paramsToSearchQuery(params dmodel.DynamicFields) (dyn.SearchQuery, error) {
	query := dyn.SearchQuery{}
	err := decodeParams(params, &query)
	return query, err
}

func readId(params dmodel.DynamicFields, field string) model.Id {
	return model.Id(readString(params, field))
}

func readString(params dmodel.DynamicFields, field string) string {
	val, ok := params[field]
	if !ok || val == nil {
		return ""
	}
	// model.Id and model.Etag are string types, so the string cases cover them too.
	switch typed := val.(type) {
	case string:
		return typed
	case *string:
		if typed == nil {
			return ""
		}
		return *typed
	}
	return ""
}

func readBool(params dmodel.DynamicFields, field string) (bool, bool) {
	val, ok := params[field]
	if !ok || val == nil {
		return false, false
	}
	switch typed := val.(type) {
	case bool:
		return typed, true
	case *bool:
		if typed == nil {
			return false, false
		}
		return *typed, true
	default:
		return false, false
	}
}
