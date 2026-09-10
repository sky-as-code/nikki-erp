package composable

import (
	"sort"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ComputedFieldFn produces the value of a "function"-kind computed field.
//
// It receives the whole page at once rather than one row at a time: a search returns up to a
// page of rows, and a per-row signature would turn any lookup the function performs into an N+1.
// It must return exactly one value per row in Models, in the same order; a length mismatch is an
// error, never a partial fill.
//
// The function may resolve services from the dependency container. Resolve them once, when the
// module registers the function, and close over them.
type ComputedFieldFn func(ctx corectx.Context, req ComputeFnRequest) ([]any, error)

// ComputeFnRequest is what a computed-field function is given.
type ComputeFnRequest struct {
	// SchemaName and FieldName identify what is being computed, so one function can serve several
	// fields or several schemas.
	SchemaName string
	FieldName  string

	// Models are the rows to compute over: a page of rows on a read, exactly one on a
	// meta/compute call, where it is the unsaved model the client posted.
	Models []dmodel.DynamicFields

	// Args carries the caller-supplied extras of a meta/compute call. Nil on a read.
	Args map[string]any
}

// assertComputedFunctionsDefined matches the schema's declarations against what was registered.
//
// It reports every missing function at once rather than the first, so a module adding several
// fields fixes them in one pass instead of rediscovering them one boot at a time.
func assertComputedFunctionsDefined(
	schemaName string, lookup func(name string) (ComputedFieldFn, bool),
) error {
	schemaPlan := computed.PlanFor(schemaName)
	if schemaPlan == nil {
		return nil
	}

	var missing []string
	for fieldName, fieldPlan := range schemaPlan.Fields {
		if fieldPlan.Def.Kind != computed.ComputeFunction {
			continue
		}
		if _, ok := lookup(fieldPlan.FunctionName); !ok {
			missing = append(missing, fieldName+" -> "+fieldPlan.FunctionName)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	// Map iteration order is random; sorting keeps the boot failure reproducible.
	sort.Strings(missing)
	return errors.Errorf(
		"resource '%s' declares computed fields whose functions were never registered "+
			"in DynamicResourceEngineOnionImpl.ComputedFunctions: %s",
		schemaName, strings.Join(missing, ", "),
	)
}

// On-demand evaluation of a single "function"-kind computed field.
//
// A read computes these fields from rows that exist. A form needs the same answer for a model the
// user is still editing and has not saved, otherwise a value derived from a field they just
// changed stays stale until after a save. ComputeField takes that unsaved model and runs the same
// registered function over it.

const (
	// ParamComputeField names the field to compute; the REST engine fills it from the path.
	ParamComputeField = "field"
	// ParamComputeModel is the request body's unsaved model.
	ParamComputeModel = "model"
	// ParamComputeArgs is the request body's caller-supplied extras.
	ParamComputeArgs = "args"
)

func (this *DefaultDomainServiceImpl) ComputeField(
	ctx corectx.Context, cmd ComputeFieldCommand,
) (*ComputeFieldResult, error) {
	fieldName, _ := cmd[ParamComputeField].(string)

	field, fieldPlan, errs := resolveComputeTarget(this.schema, fieldName)
	if errs.Count() > 0 {
		return &ComputeFieldResult{ClientErrors: errs}, nil
	}

	fn, ok := this.lookupComputedFunction(fieldPlan.FunctionName)
	if !ok {
		return nil, errors.Errorf(
			"computed-field function '%s' of '%s.%s' is not registered",
			fieldPlan.FunctionName, this.schema.Name(), fieldName)
	}

	values, err := fn(ctx, ComputeFnRequest{
		SchemaName: this.schema.Name(),
		FieldName:  fieldName,
		Models:     []dmodel.DynamicFields{computeModelParam(cmd)},
		Args:       computeArgsParam(cmd),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "computing '%s.%s'", this.schema.Name(), fieldName)
	}
	if len(values) != 1 {
		return nil, errors.Errorf(
			"computed function '%s' returned %d values for a single model of '%s.%s'",
			fieldPlan.FunctionName, len(values), this.schema.Name(), fieldName)
	}

	dataType := field.DataType()
	return &ComputeFieldResult{
		HasData: true,
		Data: ComputeFieldResultData{
			// The base type name and its array-ness travel separately: an array is a modifier on
			// a base type in the data-type model, not a distinct name the client could parse out.
			DataType: dataType.String(),
			IsArray:  dataType.IsArray(),
			Value:    values[0],
		},
	}, nil
}

func (this *DefaultDomainServiceImpl) lookupComputedFunction(name string) (ComputedFieldFn, bool) {
	if this.computedFunction == nil {
		return nil, false
	}
	return this.computedFunction(name)
}

// resolveComputeTarget checks that the named field exists and is computed by a function. Both
// failures are the caller's to fix, so they answer 400 rather than 500.
func resolveComputeTarget(
	schema *dmodel.ModelSchema, fieldName string,
) (*dmodel.ModelField, *computed.FieldPlan, ft.ClientErrors) {
	var errs ft.ClientErrors
	if fieldName == "" {
		errs.Append(*ft.NewValidationError(ParamComputeField,
			ft.ErrorKey("err_field_required"), "the field to compute is required"))
		return nil, nil, errs
	}

	field, ok := schema.Field(fieldName)
	if !ok {
		errs.Append(*ft.NewValidationError(ParamComputeField,
			ft.ErrorKey("err_unknown_field"), "no such field on this resource"))
		return nil, nil, errs
	}

	schemaPlan := computed.PlanFor(schema.Name())
	if schemaPlan == nil {
		errs.Append(*ft.NewValidationError(ParamComputeField,
			ft.ErrorKey("err_field_not_computed_by_function"),
			"this field is not computed by a function"))
		return nil, nil, errs
	}
	fieldPlan, ok := schemaPlan.Fields[fieldName]
	if !ok || fieldPlan.Def.Kind != computed.ComputeFunction {
		errs.Append(*ft.NewValidationError(ParamComputeField,
			ft.ErrorKey("err_field_not_computed_by_function"),
			"this field is not computed by a function"))
		return nil, nil, errs
	}
	return field, fieldPlan, errs
}

// computeModelParam reads the unsaved model. An absent model is legal: a function may derive its
// value from context alone, and an empty map says "nothing filled in yet" rather than being wrong.
func computeModelParam(params dmodel.DynamicFields) dmodel.DynamicFields {
	switch typed := params[ParamComputeModel].(type) {
	case dmodel.DynamicFields:
		return typed
	case map[string]any:
		return dmodel.DynamicFields(typed)
	}
	return dmodel.DynamicFields{}
}

func computeArgsParam(params dmodel.DynamicFields) map[string]any {
	switch typed := params[ParamComputeArgs].(type) {
	case dmodel.DynamicFields:
		return typed
	case map[string]any:
		return typed
	}
	return nil
}
