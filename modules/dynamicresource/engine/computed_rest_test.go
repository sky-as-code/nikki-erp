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

// computeFieldEngine builds an engine whose schema carries one array-valued function field, the
// shape that motivated the kind: a list chosen by a scalar on the same row.
func computeFieldEngine(t *testing.T, schemaName string) it.DynamicResourceEngine {
	t.Helper()
	schema := dmodel.DefineModel(schemaName).
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("mode").DataType(dmodel.FieldDataTypeString(0, 20))).
		Field(dmodel.DefineField().Name("tag_ids").DataType(dmodel.FieldDataTypeUlid().ArrayType()).
			Computed(false, computed.GoFunction("tags").DependsOn("mode").Build())).
		Build()
	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(schema))
	require.NoError(t, reg.FinalizeRelations())

	resourceEngine := NewDynamicResourceEngine(NewEngineParam{Schema: schema})
	require.NoError(t, DefineBuiltinActions(resourceEngine))
	return resourceEngine
}

// The endpoint exists so a form can see a derived value for a model that was never saved, so the
// posted model — not any stored row — must be what the function receives.
func TestComputeFieldUsesThePostedModel(t *testing.T) {
	resourceEngine := computeFieldEngine(t, "cfr_posted")
	var seen dmodel.DynamicFields
	require.NoError(t, resourceEngine.DefineComputedFieldFunction("tags",
		func(_ corectx.Context, req it.ComputeFnRequest) ([]any, error) {
			seen = req.Models[0]
			return []any{[]string{"a", "b"}}, nil
		}))

	result, err := resourceEngine.ExecuteAction(ownerContext(), it.ActionComputeField,
		dmodel.DynamicFields{
			"field": "tag_ids",
			"model": map[string]any{"mode": "override"},
			"args":  map[string]any{"as_of": "2026-08-26"},
		})

	require.NoError(t, err)
	require.True(t, result.HasData)
	assert.Equal(t, "override", seen["mode"])

	data := result.Data.(dmodel.DynamicFields)
	assert.Equal(t, "ulid", data["data_type"])
	assert.Equal(t, true, data["is_array"], "array-ness travels separately from the base type")
	assert.Equal(t, []string{"a", "b"}, data["value"])
}

func TestComputeFieldPassesArgs(t *testing.T) {
	resourceEngine := computeFieldEngine(t, "cfr_args")
	var seen map[string]any
	require.NoError(t, resourceEngine.DefineComputedFieldFunction("tags",
		func(_ corectx.Context, req it.ComputeFnRequest) ([]any, error) {
			seen = req.Args
			return []any{nil}, nil
		}))

	_, err := resourceEngine.ExecuteAction(ownerContext(), it.ActionComputeField,
		dmodel.DynamicFields{
			"field": "tag_ids",
			"args":  map[string]any{"as_of": "2026-08-26"},
		})

	require.NoError(t, err)
	assert.Equal(t, "2026-08-26", seen["as_of"])
}

// A bad field name is the caller's mistake, so it must answer 400 through ClientErrors rather
// than a 500 from a returned error.
func TestComputeFieldRejectsUnknownField(t *testing.T) {
	resourceEngine := computeFieldEngine(t, "cfr_unknown")

	result, err := resourceEngine.ExecuteAction(ownerContext(), it.ActionComputeField,
		dmodel.DynamicFields{"field": "no_such_field"})

	require.NoError(t, err)
	assert.NotEqual(t, 0, result.ClientErrors.Count())
}

func TestComputeFieldRejectsNonFunctionField(t *testing.T) {
	resourceEngine := computeFieldEngine(t, "cfr_plain")

	result, err := resourceEngine.ExecuteAction(ownerContext(), it.ActionComputeField,
		dmodel.DynamicFields{"field": "mode"})

	require.NoError(t, err)
	assert.NotEqual(t, 0, result.ClientErrors.Count())
}
