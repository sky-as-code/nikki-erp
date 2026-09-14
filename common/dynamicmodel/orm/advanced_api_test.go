package orm

import (
	"testing"

	"github.com/huandu/go-sqlbuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The resolver hook is what lets orm/advanced answer field paths PgQueryBuilder itself rejects.
// These tests pin its contract: consulted first, ok=false falls back to the planner, and the
// linked/not_linked path keeps reading the planner's root alias regardless of the resolver.

type stubRefResolver struct {
	answers map[string]string
	field   *dmodel.ModelField
}

func (this *stubRefResolver) ResolveFilterRef(name string) (*dmodel.ModelField, string, bool, error) {
	if ref, ok := this.answers[name]; ok {
		return this.field, ref, true, nil
	}
	return nil, "", false, nil
}

func (this *stubRefResolver) ResolveOrderRef(name string) (*dmodel.ModelField, string, bool, error) {
	return this.ResolveFilterRef(name)
}

func advancedApiFixture(t *testing.T) (*PgQueryBuilder, *dmodel.SchemaRegistry, *dmodel.ModelSchema) {
	t.Helper()
	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(computedOrderLineSchema()))
	order := computedOrderSchema(
		dmodel.DefineField().Name("code").DataType(dmodel.FieldDataTypeString(0, 20)))
	require.NoError(t, reg.Register(order))
	require.NoError(t, reg.FinalizeRelations())
	return NewPgQueryBuilder().(*PgQueryBuilder), reg, order
}

func TestAdvancedApi_CompilePredicateHonoursResolver(t *testing.T) {
	qb, reg, order := advancedApiFixture(t)
	plan := qb.NewJoinPlan(reg, order)
	plan.EnsureRootAliased()
	synthetic := dmodel.DefineField().Name("line_total").DataType(dmodel.FieldDataTypeInt64(0, 1000)).Build()
	resolver := &stubRefResolver{answers: map[string]string{"line_total": "c0.v"}, field: synthetic}

	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().NewCondition("line_total", dmodel.GreaterThan, 5),
		*dmodel.NewSearchNode().NewCondition("code", dmodel.Equals, "A"),
	)
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	predicate, cErrs, err := qb.CompileGraphPredicate(plan, resolver, nil, sb, graph)
	require.NoError(t, err)
	require.Empty(t, cErrs)
	raw, args := sb.Select("1").From("x").Where(predicate).Build()
	sql, err := Interpolate(raw, args)
	require.NoError(t, err)
	assert.Contains(t, sql, "c0.v > 5", "resolver-provided ref must be used verbatim")
	assert.Contains(t, sql, `"code" = E'A'`, "ok=false must fall back to the planner's own resolution")

	orderExprs, err := qb.CompileOrderExprs(plan, resolver, nil,
		dmodel.SearchOrder{{"line_total", "desc"}, {"code", "asc"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"c0.v DESC", `"code" ASC`}, orderExprs)
}

func TestAdvancedApi_LinkedUsesPlannerRootAliasWithResolver(t *testing.T) {
	qb, reg, order := advancedApiFixture(t)
	plan := qb.NewJoinPlan(reg, order)
	plan.EnsureRootAliased()
	resolver := &stubRefResolver{answers: map[string]string{}}
	graph := dmodel.NewSearchGraph().NewCondition("lines", dmodel.Linked, "01LINE")
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	predicate, cErrs, err := qb.CompileGraphPredicate(plan, resolver, nil, sb, graph)
	require.NoError(t, err)
	require.Empty(t, cErrs)
	assert.Contains(t, predicate, `t0."id" IN (SELECT "order_id" FROM "cf_ormc_lines"`)
}

func TestAdvancedApi_ResolverUnknownFieldBecomesClientError(t *testing.T) {
	qb, reg, order := advancedApiFixture(t)
	plan := qb.NewJoinPlan(reg, order)
	graph := dmodel.NewSearchGraph().NewCondition("nope", dmodel.Equals, 1)
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	_, _, err := qb.CompileGraphPredicate(plan, &stubRefResolver{}, nil, sb, graph)
	require.Error(t, err)
	_, cErrs, err := SqlGraphOutcome("", nil, err)
	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.Equal(t, "common:err_unknown_schema_field", (*cErrs)[0].Key)
}
