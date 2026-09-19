package advanced

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// Count and exists must see exactly the laterals the WHERE clause needs — never the ones only
// the projection or the sort asked for — while the DISTINCT-subquery count keeps the projected
// ones because the list query's grain includes them.

func advCount(t *testing.T, fx *fixture, schemaName string, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts) string {
	t.Helper()
	sql, cErrs, err := NewAdvancedPgQueryBuilder().SqlCountGraph(fx.schema(t, schemaName), fx.registry, graph, opts)
	require.NoError(t, err)
	require.Nil(t, cErrs)
	return *sql
}

func advExists(t *testing.T, fx *fixture, schemaName string, graph *dmodel.SearchGraph) string {
	t.Helper()
	sql, cErrs, err := NewAdvancedPgQueryBuilder().SqlExistsGraph(fx.schema(t, schemaName), fx.registry, graph)
	require.NoError(t, err)
	require.Nil(t, cErrs)
	return *sql
}

func TestCount_FilterLateralIncluded(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("on_hand_quantity", dmodel.GreaterThan, 0)
	sql := advCount(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Equal(t, `SELECT COUNT(*) FROM "adv_templates" AS t0 LEFT JOIN LATERAL (SELECT COALESCE((SELECT SUM("quantity") FROM "adv_quants" WHERE "template_id" = t0."id" AND "tenant_id" = t0."tenant_id" AND "is_archived" = FALSE), 0) AS v) AS c0 ON TRUE WHERE c0.v > 0`, sql)
}

func TestCount_ProjectionAndOrderLateralsExcluded(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("code", dmodel.Equals, "A").OrderBy("has_stock", dmodel.Desc)
	sql := advCount(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "on_hand_quantity", "largest_quant_note")})
	assert.Equal(t, `SELECT COUNT(*) FROM "adv_templates" WHERE "code" = E'A'`, sql)
}

func TestCount_SelectOnlyJoinIsDropped(t *testing.T) {
	fx := newFixture(t)
	// A many:one join that only the projection needs cannot change the count, so it is not planned.
	sql := advCount(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "template.code")})
	assert.Equal(t, `SELECT COUNT(*) FROM "adv_variants"`, sql)
}

func TestCount_DistinctSubqueryProjectsSelectLaterals(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().NewCondition("quants.location_id", dmodel.Equals, "WH-1"),
		*dmodel.NewSearchNode().NewCondition("has_stock", dmodel.Equals, true),
	)
	sql := advCount(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "on_hand_quantity")})
	assert.True(t, strings.HasPrefix(sql, `SELECT COUNT(*) FROM (SELECT DISTINCT t0."id", c1.v AS "on_hand_quantity" FROM "adv_templates" AS t0`), sql)
	assert.Equal(t, 2, strings.Count(sql, "LEFT JOIN LATERAL"))
	assert.Contains(t, sql, `LEFT JOIN "adv_quants" AS t1 ON t1."template_id" = t0."id"`)
	assert.Contains(t, sql, `WHERE (t1."location_id" = E'WH-1' AND c0.v = TRUE)) AS _distinct_count`)
}

func TestCount_ExplicitDistinctTokenWithComputedFilter(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("stock_value", dmodel.GreaterThan, 0)
	sql := advCount(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: []orm.SelectColumn{orm.SelectColumn("category").AsDistinct()}})
	assert.True(t, strings.HasPrefix(sql, `SELECT COUNT(*) FROM (SELECT DISTINCT t0."category" FROM "adv_templates" AS t0 LEFT JOIN LATERAL`), sql)
	assert.Contains(t, sql, `WHERE (c0.v * t0."base_price") > 0) AS _distinct_count`)
}

func TestExists_UsesFilterLateralsOnly(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().NewCondition("uom_name", dmodel.Equals, "kg"),
		*dmodel.NewSearchNode().NewCondition("on_hand_quantity", dmodel.GreaterThan, 0),
	).OrderBy("has_stock", dmodel.Asc)
	sql := advExists(t, fx, schemaTemplate, graph)
	assert.True(t, strings.HasPrefix(sql, `SELECT EXISTS (SELECT 1 FROM "adv_templates" AS t0 LEFT JOIN "adv_uoms" AS t1 ON t0."uom_id" = t1."id" LEFT JOIN LATERAL`), sql)
	assert.Equal(t, 1, strings.Count(sql, "LEFT JOIN LATERAL"))
	assert.NotContains(t, sql, "EXISTS (SELECT 1 FROM \"adv_quants\"", "the has_stock lateral is order-only")
	assert.Contains(t, sql, `WHERE (t1."name" = E'kg' AND c0.v > 0))`)
}

func TestExists_PlainFilterStaysUnaliased(t *testing.T) {
	fx := newFixture(t)
	sql := advExists(t, fx, schemaTemplate, dmodel.NewSearchGraph().NewCondition("code", dmodel.Equals, "A"))
	assert.Equal(t, `SELECT EXISTS (SELECT 1 FROM "adv_templates" WHERE "code" = E'A')`, sql)
}
