package engine

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// The crud helpers are generic over a domain type, so the engine instantiates them with
// DynamicEntity and then unwraps the entity back into the plain field map it speaks.

func unwrapEntityResult(result *dyn.OpResult[it.DynamicEntity]) *dyn.OpResult[dmodel.DynamicFields] {
	out := &dyn.OpResult[dmodel.DynamicFields]{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	if result.HasData {
		out.Data = result.Data.GetFieldData()
	}
	return out
}

func unwrapSingleResult(
	result *dyn.OpResult[dyn.SingleResultData[it.DynamicEntity]],
) *dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]] {
	out := &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	if result.HasData {
		out.Data = dyn.SingleResultData[dmodel.DynamicFields]{
			Item: result.Data.Item.GetFieldData(),
			Meta: result.Data.Meta,
		}
	}
	return out
}

func unwrapPagedResult(
	result *dyn.OpResult[dyn.PagedResultData[it.DynamicEntity]],
) *dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]] {
	out := &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	paged := result.Data
	out.Data = dyn.PagedResultData[dmodel.DynamicFields]{
		Items: array.Map(paged.Items, func(item it.DynamicEntity) dmodel.DynamicFields {
			return item.GetFieldData()
		}),
		Total:         paged.Total,
		Page:          paged.Page,
		Size:          paged.Size,
		DesiredFields: paged.DesiredFields,
		MaskedFields:  paged.MaskedFields,
		SchemaEtag:    paged.SchemaEtag,
	}
	return out
}

// findByKeys fetches the record identified by keys, as the entity the validator hooks speak.
//
// It returns (nil, nil) with vErrs populated when the record does not exist, so a caller can
// tell "no such record" from a failed read: the first is the caller's mistake, the second the
// server's.
func (this *DynamicResourceServiceImpl) findByKeys(
	ctx corectx.Context, keys dmodel.DynamicFields, vErrs *ft.ClientErrors,
) (*it.DynamicEntity, error) {
	found, err := this.repository.FindByKeys(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "findByKeys")
	}
	if found.ClientErrors.Count() > 0 {
		vErrs.Concat(found.ClientErrors)
		return nil, nil
	}
	if !found.HasData {
		vErrs.Append(*ft.NewAnonymousNotFoundError())
		return nil, nil
	}
	return it.NewDynamicEntityFrom(found.Data), nil
}

// assertActionSupported reports the action as unsupported when the engine was created
// without it. A client error rather than a hard error: asking for an action this resource
// does not have is the caller's mistake, not a server fault.
//
// It guards the direct-call path. The REST surface needs no guard, because an undefined
// action has no route.
func (this *DynamicResourceServiceImpl) assertActionSupported(action it.CrudAction) *ft.ClientErrors {
	if len(this.crudActions) == 0 || this.crudActions[action] {
		return nil
	}
	cErrs := ft.NewClientErrors()
	cErrs.Append(*ft.NewAnonymousBusinessViolation(
		ft.ErrorKey("err_action_not_supported"),
		"This resource does not support the requested action",
		map[string]any{"action": string(action), "resource": this.schema.Name()},
	))
	return cErrs
}

// toActionResult lifts a typed service result into the shape every action returns,
// so that one signature can carry records, paged results and existence checks alike.
func toActionResult[TData any](result *dyn.OpResult[TData], err error) (*it.ActionResult, error) {
	if err != nil {
		return nil, err
	}
	out := &it.ActionResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	if result.HasData {
		out.Data = result.Data
	}
	return out, nil
}

// readDesiredFields returns the fields the caller asked for, or the service default.
func (this *DynamicResourceServiceImpl) readDesiredFields(params dmodel.DynamicFields) ([]string, error) {
	raw, ok := params[fieldNameFields]
	if !ok || raw == nil {
		return this.defaultFields, nil
	}
	fields, ok := raw.([]string)
	if ok {
		if len(fields) == 0 {
			return this.defaultFields, nil
		}
		return fields, nil
	}

	anyFields, ok := raw.([]any)
	if !ok {
		return nil, errors.New("'fields' must be an array of field names")
	}
	if len(anyFields) == 0 {
		return this.defaultFields, nil
	}
	return array.Map(anyFields, func(item any) string {
		name, _ := item.(string)
		return name
	}), nil
}

const fieldNameFields = "fields"

// uniqueKeysToGraph builds an equality search graph out of every param that names a
// column of the schema. It rejects a params map that carries no usable key.
func (this *DynamicResourceServiceImpl) uniqueKeysToGraph(
	params dmodel.DynamicFields,
) (*dmodel.SearchGraph, ft.ClientErrors) {
	cErrs := ft.ClientErrors{}
	keyNode := dmodel.NewSearchNode()
	hasKey := false

	for _, field := range this.schema.Columns() {
		name := field.Name()
		if name == fieldNameFields {
			continue
		}
		val, ok := params[name]
		if !ok || val == nil {
			continue
		}
		keyNode.NewCondition(name, dmodel.Equals, val)
		hasKey = true
	}

	if !hasKey {
		cErrs.Append(*ft.NewAnonymousBusinessViolation(
			"common.err_missing_unique_key",
			"at least one unique key field is required to identify the record",
		))
		return nil, cErrs
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*keyNode)
	return graph, cErrs
}
