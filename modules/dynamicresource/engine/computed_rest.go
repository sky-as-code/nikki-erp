package engine

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// On-demand evaluation of a single "function"-kind computed field.
//
// A read computes these fields from rows that exist. A form needs the same answer for a model the
// user is still editing and has not saved — otherwise a value derived from a field they just
// changed stays stale until after a save. This action takes that unsaved model in the body and
// runs the same registered function over it.

const (
	// paramComputeField is the path param naming the field to compute.
	paramComputeField = "field"
	// paramComputeModel is the request body's unsaved model.
	paramComputeModel = "model"
	// paramComputeArgs is the request body's caller-supplied extras.
	paramComputeArgs = "args"
)

// defineComputeFieldAction registers POST {resource}/meta/compute/{field}.
//
// It closes over the engine because the function registry lives there and ProcessInput carries
// only the service and repository — widening that struct for one action would put an engine
// handle in front of every action that has no use for it.
func defineComputeFieldAction(resourceEngine it.DynamicResourceEngine) error {
	return resourceEngine.DefineAction(it.DynamicActionDefinition{
		ActionName: it.ActionComputeField,
		// Generic, not Read: the unsaved model travels in the body, which a GET has no place for.
		ActionType: it.ActionTypeGeneric,
		RestPath:   "meta/compute/:field",
		// Computing a derived value discloses no more than reading the record would.
		Permission: it.PermissionRead,
		MainProcess: func(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			return processComputeField(ctx, resourceEngine, input)
		},
	})
}

func processComputeField(
	ctx corectx.Context, resourceEngine it.DynamicResourceEngine, input it.ProcessInput,
) (*it.ActionResult, error) {
	fieldName, _ := input.Params[paramComputeField].(string)
	schema := resourceEngine.Schema()

	field, fieldPlan, errs := resolveComputeTarget(schema, fieldName)
	if errs.Count() > 0 {
		return &it.ActionResult{ClientErrors: errs}, nil
	}

	fn, ok := resourceEngine.ComputedFieldFunction(fieldPlan.FunctionName)
	if !ok {
		return nil, errors.Errorf(
			"computed-field function '%s' of '%s.%s' is not registered",
			fieldPlan.FunctionName, schema.Name(), fieldName)
	}

	values, err := fn(ctx, it.ComputeFnRequest{
		SchemaName: schema.Name(),
		FieldName:  fieldName,
		Models:     []dmodel.DynamicFields{computeModelParam(input.Params)},
		Args:       computeArgsParam(input.Params),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "computing '%s.%s'", schema.Name(), fieldName)
	}
	if len(values) != 1 {
		return nil, errors.Errorf(
			"computed function '%s' returned %d values for a single model of '%s.%s'",
			fieldPlan.FunctionName, len(values), schema.Name(), fieldName)
	}

	dataType := field.DataType()
	return &it.ActionResult{
		HasData: true,
		Data: dmodel.DynamicFields{
			// The base type name and its array-ness travel separately: an array is a modifier on
			// a base type in the data-type model, not a distinct name the client could parse out.
			"data_type": dataType.String(),
			"is_array":  dataType.IsArray(),
			"value":     values[0],
		},
	}, nil
}

// resolveComputeTarget checks that the named field exists and is computed by a function. Both
// failures are the caller's to fix, so they answer 400 rather than 500.
func resolveComputeTarget(
	schema *dmodel.ModelSchema, fieldName string,
) (*dmodel.ModelField, *computed.FieldPlan, ft.ClientErrors) {
	var errs ft.ClientErrors
	if fieldName == "" {
		errs.Append(*ft.NewValidationError(paramComputeField,
			ft.ErrorKey("err_field_required"), "the field to compute is required"))
		return nil, nil, errs
	}

	field, ok := schema.Field(fieldName)
	if !ok {
		errs.Append(*ft.NewValidationError(paramComputeField,
			ft.ErrorKey("err_unknown_field"), "no such field on this resource"))
		return nil, nil, errs
	}

	schemaPlan := computed.PlanFor(schema.Name())
	if schemaPlan == nil {
		errs.Append(*ft.NewValidationError(paramComputeField,
			ft.ErrorKey("err_field_not_computed_by_function"),
			"this field is not computed by a function"))
		return nil, nil, errs
	}
	fieldPlan, ok := schemaPlan.Fields[fieldName]
	if !ok || fieldPlan.Def.Kind != computed.ComputeFunction {
		errs.Append(*ft.NewValidationError(paramComputeField,
			ft.ErrorKey("err_field_not_computed_by_function"),
			"this field is not computed by a function"))
		return nil, nil, errs
	}
	return field, fieldPlan, errs
}

// computeModelParam reads the unsaved model. An absent model is legal: a function may derive its
// value from context alone, and an empty map says "nothing filled in yet" rather than being wrong.
func computeModelParam(params dmodel.DynamicFields) dmodel.DynamicFields {
	switch typed := params[paramComputeModel].(type) {
	case dmodel.DynamicFields:
		return typed
	case map[string]any:
		return dmodel.DynamicFields(typed)
	}
	return dmodel.DynamicFields{}
}

func computeArgsParam(params dmodel.DynamicFields) map[string]any {
	switch typed := params[paramComputeArgs].(type) {
	case dmodel.DynamicFields:
		return typed
	case map[string]any:
		return typed
	}
	return nil
}
