package computed

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Function-kind resolution. Unlike every other kind there is no expression to walk and nothing to
// infer: the value comes from Go code the engine holds, so resolution only has to validate the
// declared dependency and record what the read must project.
//
// Whether the named function is actually registered cannot be answered here — the registry lives
// on the engine, which is built after schema finalize. That check runs at boot; see
// AssertComputedFunctionsDefined on the DynamicResourceEngine.

func (this *resolver) buildFunctionPlan(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, plan *FieldPlan,
) error {
	plan.FunctionName = plan.Def.Function.Name

	dependsOn := plan.Def.Function.DependsOn
	if dependsOn == "" {
		// No declared dependency: the function reads whatever the row already carries, and the
		// frontend has no change to trigger a recompute on.
		plan.Type = Type(field.DataType().String())
		return nil
	}

	// Reusing the expression resolver's field lookup gets existence checking, edge and
	// service-filled rejection, dependency recording for impact analysis, cycle detection through
	// computed-on-computed references, and — most importantly — registration of the dependency as
	// a physical operand, so the read projects the field the function needs to see.
	operands := map[string]bool{}
	resolve := this.fieldTypeResolver(schema, plan, operands)
	if _, err := resolve(dependsOn); err != nil {
		return errors.Wrapf(err, "computed field %s.%s depends_on", schema.Name(), field.Name())
	}
	plan.DependsOn = dependsOn

	// The declared data_type is authoritative: a Go function may return a list, and the inference
	// types are all scalar, so there is nothing to reconcile it against.
	plan.Type = Type(field.DataType().String())
	return nil
}
