package computed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// invokerRecorder captures how the eval stage calls a registered function, so the batching claim
// — one invocation per function per page, not per row — is actually pinned.
type invokerRecorder struct {
	calls  int
	rows   int
	fields []string
	result func(rows []dmodel.DynamicFields) []any
	err    error
}

func (this *invokerRecorder) fn() computed.FunctionInvokerFn {
	return func(functionName string, fieldName string, rows []dmodel.DynamicFields) ([]any, error) {
		this.calls++
		this.rows = len(rows)
		this.fields = append(this.fields, fieldName)
		if this.err != nil {
			return nil, this.err
		}
		return this.result(rows), nil
	}
}

func functionEvalSchema(t *testing.T, name string) {
	t.Helper()
	schema := dmodel.DefineModel(name).
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("mode").DataType(dmodel.FieldDataTypeString(0, 20))).
		Field(dmodel.DefineField().Name("tag_ids").DataType(dmodel.FieldDataTypeUlid().ArrayType()).
			Computed(false, computed.GoFunction("tags").DependsOn("mode").Build())).
		Build()
	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(schema))
	require.NoError(t, reg.FinalizeRelations())
}

func TestApply_FunctionRunsOncePerPage(t *testing.T) {
	functionEvalSchema(t, "cf_ev_batch")
	plan, errs := computed.BuildEvalPlan("cf_ev_batch", []string{"id", "tag_ids"})
	require.Equal(t, 0, errs.Count())
	require.NotNil(t, plan)

	rows := []dmodel.DynamicFields{
		{"id": "r1", "mode": "override"},
		{"id": "r2", "mode": "inherit"},
		{"id": "r3", "mode": "override"},
	}
	recorder := &invokerRecorder{result: func(rows []dmodel.DynamicFields) []any {
		values := make([]any, len(rows))
		for i, row := range rows {
			values[i] = []string{row["mode"].(string)}
		}
		return values
	}}

	require.NoError(t, plan.Apply(rows, computed.EvalDeps{Invoke: recorder.fn()}))

	assert.Equal(t, 1, recorder.calls, "one invocation must serve the whole page")
	assert.Equal(t, 3, recorder.rows)
	assert.Equal(t, []any{[]string{"override"}}, []any{rows[0]["tag_ids"]})
	assert.Equal(t, []string{"inherit"}, rows[1]["tag_ids"])
}

// A short result would shift values onto the wrong rows, which is worse than failing.
func TestApply_FunctionResultLengthMismatchIsAnError(t *testing.T) {
	functionEvalSchema(t, "cf_ev_short")
	plan, errs := computed.BuildEvalPlan("cf_ev_short", []string{"id", "tag_ids"})
	require.Equal(t, 0, errs.Count())

	rows := []dmodel.DynamicFields{{"id": "r1"}, {"id": "r2"}}
	recorder := &invokerRecorder{result: func([]dmodel.DynamicFields) []any {
		return []any{"only-one"}
	}}

	err := plan.Apply(rows, computed.EvalDeps{Invoke: recorder.fn()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "returned 1 values for 2 rows")
}

func TestApply_FunctionErrorIsWrappedWithFieldIdentity(t *testing.T) {
	functionEvalSchema(t, "cf_ev_err")
	plan, errs := computed.BuildEvalPlan("cf_ev_err", []string{"id", "tag_ids"})
	require.Equal(t, 0, errs.Count())

	recorder := &invokerRecorder{err: assert.AnError}
	err := plan.Apply([]dmodel.DynamicFields{{"id": "r1"}}, computed.EvalDeps{Invoke: recorder.fn()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tag_ids")
	assert.Contains(t, err.Error(), "tags")
}

// A plan wanting a function field with no invoker supplied must say so, rather than silently
// leaving the field absent as if it had been computed to nothing.
func TestApply_MissingInvokerIsAnError(t *testing.T) {
	functionEvalSchema(t, "cf_ev_noinvoker")
	plan, errs := computed.BuildEvalPlan("cf_ev_noinvoker", []string{"id", "tag_ids"})
	require.Equal(t, 0, errs.Count())

	err := plan.Apply([]dmodel.DynamicFields{{"id": "r1"}}, computed.EvalDeps{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no function invoker")
}
