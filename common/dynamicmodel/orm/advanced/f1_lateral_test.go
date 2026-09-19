package advanced

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// F1: an SQL-computed field is materialised once as a lateral, whatever mentions it.

func TestF1_FilterAndSelectShareOneLateral(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("on_hand_quantity", dmodel.GreaterThan, 0)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "on_hand_quantity")})

	assert.Equal(t, 1, strings.Count(sql, "SUM("), "the SUM subquery must appear exactly once")
	assert.Contains(t, sql, `LEFT JOIN LATERAL (SELECT COALESCE((SELECT SUM("quantity") FROM "adv_quants" WHERE "template_id" = t0."id" AND "tenant_id" = t0."tenant_id" AND "is_archived" = FALSE), 0) AS v) AS c0 ON TRUE`)
	assert.Contains(t, sql, `c0.v AS "on_hand_quantity"`)
	assert.Contains(t, sql, `WHERE c0.v > 0`)
	assert.Contains(t, sql, `FROM "adv_templates" AS t0`)
	assert.NotContains(t, sql, "GROUP BY")
	assert.NotContains(t, sql, "DISTINCT")
}

func TestF1_OrderBySharesLateral(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("on_hand_quantity", dmodel.GreaterEqual, 10).OrderBy("on_hand_quantity", dmodel.Desc)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "on_hand_quantity"), Size: 20})

	assert.Equal(t, 1, strings.Count(sql, "SUM("))
	assert.Contains(t, sql, "ORDER BY c0.v DESC LIMIT 20")
	assert.Contains(t, sql, "WHERE c0.v >= 10")
}

func TestF1_FilterOnlyStillProjectsNothingExtra(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("on_hand_quantity", dmodel.GreaterThan, 0)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "code")})

	assert.True(t, strings.HasPrefix(sql, `SELECT t0."id", t0."code" FROM`), sql)
	assert.Contains(t, sql, "WHERE c0.v > 0")
}

func TestF1_TwoComputedFieldsTwoLaterals(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().NewCondition("on_hand_quantity", dmodel.GreaterThan, 0),
		*dmodel.NewSearchNode().NewCondition("has_stock", dmodel.Equals, true),
	)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "has_stock", "on_hand_quantity")})

	assert.Equal(t, 2, strings.Count(sql, "LEFT JOIN LATERAL"))
	// Laterals are numbered in first-mention order: the select list is resolved first.
	assert.Contains(t, sql, `c0.v AS "has_stock"`)
	assert.Contains(t, sql, `c1.v AS "on_hand_quantity"`)
	assert.Contains(t, sql, `AS c1 ON TRUE`)
	assert.Contains(t, sql, "WHERE (c1.v > 0 AND c0.v = TRUE)")
}

func TestF1_ExistsKindInFilter(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("has_stock", dmodel.Equals, true)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})

	assert.Contains(t, sql, `LEFT JOIN LATERAL (SELECT EXISTS (SELECT 1 FROM "adv_quants" WHERE "template_id" = t0."id" AND "tenant_id" = t0."tenant_id" AND "is_archived" = FALSE AND "quantity" > 0) AS v) AS c0 ON TRUE WHERE c0.v = TRUE`)
}

func TestF1_LookupKindInFilterAndOrder(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("largest_quant_note", dmodel.Contains, "damaged").OrderBy("largest_quant_note", dmodel.Asc)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})

	assert.Contains(t, sql, `LATERAL (SELECT (SELECT "note" FROM "adv_quants" WHERE "template_id" = t0."id" AND "tenant_id" = t0."tenant_id" AND "is_archived" = FALSE ORDER BY "quantity" DESC LIMIT 1) AS v) AS c0 ON TRUE`)
	assert.Contains(t, sql, "WHERE c0.v ILIKE E'%damaged%'")
	assert.Contains(t, sql, "ORDER BY c0.v ASC")
}

func TestF1_IsSetOnComputedWithoutDefaultSeesNull(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("largest_quant_note", dmodel.IsNotSet)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, sql, "WHERE c0.v IS NULL")
}

func TestF1_ContextPlaceholderBoundOnce(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("located_quantity", dmodel.GreaterThan, 0)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{
		Columns:         cols("id", "located_quantity"),
		ComputedContext: map[string]any{"location_id": "WH-1"},
	})
	assert.Equal(t, 1, strings.Count(sql, "WH-1"), "one lateral, one bound placeholder")
	assert.Contains(t, sql, `"location_id" = E'WH-1'`)
}

func TestF1_MissingContextIsClientError(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("located_quantity", dmodel.GreaterThan, 0)
	cErrs := advSelectErr(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, (*cErrs)[0].Key, "err_computed_context_missing")
}

func TestF1_CountUsesFilterLateralOnly(t *testing.T) {
	fx := newFixture(t)
	schema := fx.schema(t, schemaTemplate)
	graph := dmodel.NewSearchGraph().NewCondition("on_hand_quantity", dmodel.GreaterThan, 0).OrderBy("largest_quant_note", dmodel.Asc)
	sql, cErrs, err := NewAdvancedPgQueryBuilder().SqlCountGraph(schema, fx.registry, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "has_stock")})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Equal(t, 1, strings.Count(*sql, "LEFT JOIN LATERAL"), *sql)
	assert.Contains(t, *sql, "SUM(")
	assert.NotContains(t, *sql, "EXISTS")
	assert.NotContains(t, *sql, "ORDER BY")
	assert.True(t, strings.HasPrefix(*sql, `SELECT COUNT(*) FROM "adv_templates" AS t0 LEFT JOIN LATERAL`), *sql)
	assert.Contains(t, *sql, "WHERE c0.v > 0")
}

func TestF1_ExistsQueryUsesFilterLateral(t *testing.T) {
	fx := newFixture(t)
	schema := fx.schema(t, schemaTemplate)
	graph := dmodel.NewSearchGraph().NewCondition("has_stock", dmodel.Equals, true)
	sql, cErrs, err := NewAdvancedPgQueryBuilder().SqlExistsGraph(schema, fx.registry, graph)
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.True(t, strings.HasPrefix(*sql, `SELECT EXISTS (SELECT 1 FROM "adv_templates" AS t0 LEFT JOIN LATERAL`), *sql)
	assert.Contains(t, *sql, "WHERE c0.v = TRUE)")
}

func TestF1_LateralCapCountsDistinctFields(t *testing.T) {
	fx := newFixture(t)
	prev := computed.ActiveLimits()
	limited := prev
	limited.MaxSqlComputedFieldsPerRequest = 2
	computed.SetLimits(limited)
	defer computed.SetLimits(prev)

	// The same field three times is one lateral, so it stays under a cap of two.
	graph := dmodel.NewSearchGraph().NewCondition("on_hand_quantity", dmodel.GreaterThan, 0).OrderBy("on_hand_quantity", dmodel.Desc)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "on_hand_quantity")})
	assert.Equal(t, 1, strings.Count(sql, "LEFT JOIN LATERAL"))

	cErrs := advSelectErr(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("on_hand_quantity", "has_stock", "largest_quant_note")})
	assert.Contains(t, (*cErrs)[0].Key, "err_too_many_sql_computed_fields")
}

func TestF1_FanOutFilterPlusComputedKeepsOneDistinct(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().NewCondition("quants.location_id", dmodel.Equals, "WH-1"),
		*dmodel.NewSearchNode().NewCondition("on_hand_quantity", dmodel.GreaterThan, 0),
	).OrderBy("on_hand_quantity", dmodel.Desc)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "code")})

	assert.Equal(t, 1, strings.Count(sql, "SELECT DISTINCT"))
	// Under DISTINCT the sort expression must be projected, and it is the shared lateral column.
	assert.True(t, strings.HasPrefix(sql, `SELECT DISTINCT t0."id", t0."code", c0.v FROM`), sql)
	assert.Contains(t, sql, `LEFT JOIN "adv_quants" AS t1 ON t1."template_id" = t0."id"`)
	assert.Contains(t, sql, "ORDER BY c0.v DESC")
}

func TestF1_ComputedThroughToManyEdgeRejected(t *testing.T) {
	fx := newFixture(t)
	// variants.template_code would be one value per variant row; meaningless on the template.
	graph := dmodel.NewSearchGraph().NewCondition("variants.template_code", dmodel.Equals, "A")
	cErrs := advSelectErr(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, (*cErrs)[0].Key, "err_computed_field_not_filterable")
}

func TestF1_GoFunctionFieldInFilterRejected(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("external_score", dmodel.GreaterThan, 1)
	cErrs := advSelectErr(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, (*cErrs)[0].Key, "err_computed_field_not_filterable")

	cErrs = advSelectErr(t, fx, schemaTemplate, dmodel.NewSearchGraph().OrderBy("external_score", dmodel.Asc), orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, (*cErrs)[0].Key, "err_field_not_sortable")
}
