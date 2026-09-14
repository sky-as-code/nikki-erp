package advanced

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// F4: computed fields reached through a to-one edge resolve against the edge's own alias.

func TestF4_EdgeRelatedFieldDelegatesToSecondJoin(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("template.uom_name", dmodel.Equals, "kg").OrderBy("template.uom_name", dmodel.Asc)
	sql := advSelect(t, fx, schemaVariant, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "template.uom_name")})

	assert.Contains(t, sql, `LEFT JOIN "adv_templates" AS t1 ON t0."template_id" = t1."id" LEFT JOIN "adv_uoms" AS t2 ON t1."uom_id" = t2."id"`)
	assert.Contains(t, sql, `t2."name" AS "template.uom_name"`)
	assert.Contains(t, sql, `t1."id" AS "template.id"`)
	assert.Contains(t, sql, `WHERE t2."name" = E'kg'`)
	assert.Contains(t, sql, `ORDER BY t2."name" ASC`)
	assert.NotContains(t, sql, "LATERAL")
}

func TestF4_EdgeAggregateLateralOwnedByEdgeAlias(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("template.on_hand_quantity", dmodel.GreaterThan, 0)
	sql := advSelect(t, fx, schemaVariant, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "sku", "template.on_hand_quantity")})

	assert.Equal(t, 1, strings.Count(sql, "SUM("), "filter and selection share the edge-owned lateral")
	assert.Contains(t, sql, `LEFT JOIN "adv_templates" AS t1 ON t0."template_id" = t1."id" LEFT JOIN LATERAL (SELECT COALESCE((SELECT SUM("quantity") FROM "adv_quants" WHERE "template_id" = t1."id" AND "tenant_id" = t1."tenant_id" AND "is_archived" = FALSE), 0) AS v) AS c0 ON TRUE`)
	assert.Contains(t, sql, `c0.v AS "template.on_hand_quantity"`)
	assert.Contains(t, sql, "WHERE c0.v > 0")
	assert.NotContains(t, sql, "DISTINCT")
}

func TestF4_EdgeExpressionCompilesAgainstEdgeAlias(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("template.stock_value", dmodel.GreaterEqual, 100)
	sql := advSelect(t, fx, schemaVariant, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "template.price_band")})
	assert.Contains(t, sql, `WHERE (c0.v * t1."base_price") >= 100`)
	assert.Contains(t, sql, `(CASE WHEN (t1."base_price" > 100) THEN 'high' ELSE 'low' END) AS "template.price_band"`)
}

func TestF4_BareToOneEdgeProjectsSqlRenderableKinds(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "template")})

	assert.Contains(t, sql, `t1."code" AS "template.code"`)
	assert.Contains(t, sql, `c0.v AS "template.on_hand_quantity"`)
	assert.Contains(t, sql, `AS "template.has_stock"`)
	assert.Contains(t, sql, `t2."name" AS "template.uom_name"`, "related field follows its own join")
	assert.Contains(t, sql, `AS "template.stock_value"`)
	assert.NotContains(t, sql, "template.external_score", "Go function fields are never projected in SQL")
	assert.NotContains(t, sql, "template.age_days", "datetime expression has no SQL rendering")
	assert.NotContains(t, sql, "template.located_quantity", "context-bound aggregate skipped when unbound")
	assert.NotContains(t, sql, `"template.tenant_id"`)
}

func TestF4_BareEdgeIncludesContextAggregateWhenBound(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{
		Columns: cols("id", "template"), ComputedContext: map[string]any{"location_id": "WH-1"},
	})
	assert.Contains(t, sql, `AS "template.located_quantity"`)
	assert.Contains(t, sql, "WH-1")
}

func TestF4_ExplicitContextAggregateOnEdgeStillReportsMissingKey(t *testing.T) {
	fx := newFixture(t)
	cErrs := advSelectErr(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "template.located_quantity")})
	assert.Contains(t, (*cErrs)[0].Key, "err_computed_context_missing")
}

func TestF4_NestedSqlKindsProjectedInsideJsonb(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "tags")})
	assert.Contains(t, sql, `'template_count', COALESCE((SELECT COUNT(*) FROM "adv_templates" WHERE "id" IN (SELECT "template_id" FROM "adv_template_tag_rels" WHERE "tag_id" = d."id" AND "tenant_id" = d."tenant_id") AND "is_archived" = FALSE), 0)`)
	assert.Contains(t, sql, `'label', d."label"`)
}

func TestF4_ComputedThroughToManyEdgeRejectedInSelectAndFilter(t *testing.T) {
	fx := newFixture(t)
	// A to-many leaf that is Go-computed is simply not projected; an SQL-kind one is fine inside
	// the jsonb element; but filtering on either through the fan-out edge is refused.
	graph := dmodel.NewSearchGraph().NewCondition("tags.template_count", dmodel.GreaterThan, 1)
	cErrs := advSelectErr(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, (*cErrs)[0].Key, "err_computed_field_not_filterable")
}

func TestF4_PlanNestedProjectionTypesComputedLeavesByDeclaredField(t *testing.T) {
	fx := newFixture(t)
	qb := NewAdvancedPgQueryBuilder().(orm.NestedProjectionCapable)
	plan, cErrs, err := qb.PlanNestedProjection(fx.schema(t, schemaVariant), fx.registry,
		cols("template.uom_name", "template.on_hand_quantity"))
	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, plan)
	assert.Equal(t, "uom_name", plan.ToOne["template"].Leaves["template.uom_name"])
	assert.Equal(t, "on_hand_quantity", plan.ToOne["template"].Leaves["template.on_hand_quantity"])
	_, hasUomHop := plan.ToOne["template.uom"]
	assert.False(t, hasUomHop, "the join behind a related field is an implementation detail, not a projected hop")
}
