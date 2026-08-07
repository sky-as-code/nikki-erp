package engine

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The generic helpers of modules/core/dynamicmodel/crud take typed commands, while the
// engine speaks DynamicFields end to end. These functions convert one into the other.
// They only reshape data; every value is validated afterwards by the crud helper itself.

// decodeParams reshapes a field map into the given typed command through a JSON round trip,
// which reuses the json tags already declared on the core command structs.
func decodeParams(params dmodel.DynamicFields, target any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "decodeParams.Marshal")
	}
	return errors.Wrap(json.Unmarshal(raw, target), "decodeParams.Unmarshal")
}

// paramsToDeleteCommand reads the record id out of params.
func paramsToDeleteCommand(params dmodel.DynamicFields) dyn.DeleteOneCommand {
	return dyn.DeleteOneCommand{
		Id: readId(params, basemodel.FieldId),
	}
}

// paramsToSetArchivedCommand reads id, etag and the archived flag out of params.
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

// paramsToExistsQuery reads the id list out of params.
func paramsToExistsQuery(params dmodel.DynamicFields) (dyn.ExistsQuery, error) {
	query := dyn.ExistsQuery{}
	err := decodeParams(params, &query)
	return query, err
}

// paramsToGetOneQuery reads the record id and the desired fields out of params.
func paramsToGetOneQuery(params dmodel.DynamicFields) (dyn.GetOneQuery, error) {
	query := dyn.GetOneQuery{}
	err := decodeParams(params, &query)
	return query, err
}

// paramsToSearchQuery reads paging, fields and the search graph out of params.
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
	// model.Id and model.Etag are string types, so the string case covers them too.
	if typed, ok := val.(string); ok {
		return typed
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
