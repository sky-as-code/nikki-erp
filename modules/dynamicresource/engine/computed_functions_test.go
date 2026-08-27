package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

func stubComputedFn(values ...any) it.ComputedFieldFn {
	return func(_ corectx.Context, req it.ComputeFnRequest) ([]any, error) {
		if len(values) > 0 {
			return values, nil
		}
		out := make([]any, len(req.Models))
		for i := range req.Models {
			out[i] = "computed"
		}
		return out, nil
	}
}

func TestDefineComputedFieldFunction(t *testing.T) {
	engine := newTestEngine()

	require.NoError(t, engine.DefineComputedFieldFunction("tags", stubComputedFn()))

	fn, ok := engine.ComputedFieldFunction("tags")
	assert.True(t, ok)
	assert.NotNil(t, fn)
}

func TestDefineComputedFieldFunctionRejectsDuplicate(t *testing.T) {
	engine := newTestEngine()
	require.NoError(t, engine.DefineComputedFieldFunction("tags", stubComputedFn()))

	err := engine.DefineComputedFieldFunction("tags", stubComputedFn())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already defined")
}

func TestDefineComputedFieldFunctionRejectsEmptyName(t *testing.T) {
	engine := newTestEngine()

	assert.Error(t, engine.DefineComputedFieldFunction("  ", stubComputedFn()))
	assert.Error(t, engine.DefineComputedFieldFunction("tags", nil))
}

// A schema with no computed fields must not fail the assertion, or every existing resource would
// break the boot.
func TestAssertComputedFunctionsDefinedPassesWithoutComputedFields(t *testing.T) {
	engine := newTestEngine()

	assert.NoError(t, engine.AssertComputedFunctionsDefined())
}

// The whole point of the boot check: a declared function nobody registered must be named, with
// both the field and the function, so the fix is obvious from the log line alone.
func TestAssertComputedFunctionsDefinedReportsMissing(t *testing.T) {
	schema := dmodel.DefineModel("cfe_boot_check").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("mode").DataType(dmodel.FieldDataTypeString(0, 20))).
		Field(dmodel.DefineField().Name("tag_ids").DataType(dmodel.FieldDataTypeUlid().ArrayType()).
			Computed(false, computed.GoFunction("never.registered").DependsOn("mode").Build())).
		Build()
	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(schema))
	require.NoError(t, reg.FinalizeRelations())
	engine := NewDynamicResourceEngine(NewEngineParam{Schema: schema})

	err := engine.AssertComputedFunctionsDefined()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tag_ids")
	assert.Contains(t, err.Error(), "never.registered")

	require.NoError(t, engine.DefineComputedFieldFunction("never.registered", stubComputedFn()))
	assert.NoError(t, engine.AssertComputedFunctionsDefined(),
		"registering the function must satisfy the check")
}
