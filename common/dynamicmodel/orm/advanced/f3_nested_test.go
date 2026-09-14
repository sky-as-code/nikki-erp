package advanced

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// F3: nested edge fields are projected inside the root statement — joins for to-one edges,
// jsonb aggregation for to-many ones — so a page is always one query.

// ---- to-one -----------------------------------------------------------------------------

func TestF3_ToOneLeafAliasedEdgeDotLeaf(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "template.code")})
	assert.Equal(t, `SELECT t0."id", t1."id" AS "template.id", t1."code" AS "template.code" FROM "adv_variants" AS t0 LEFT JOIN "adv_templates" AS t1 ON t0."template_id" = t1."id"`, sql)
}

func TestF3_ToOnePkAlwaysProjectedOnce(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "template.code", "template.id", "template.category")})
	assert.Equal(t, 1, strings.Count(sql, `AS "template.id"`))
	assert.Contains(t, sql, `t1."category" AS "template.category"`)
}

func TestF3_DeepToOnePathThreeHops(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "template.uom.name")})
	assert.Contains(t, sql, `LEFT JOIN "adv_templates" AS t1 ON t0."template_id" = t1."id" LEFT JOIN "adv_uoms" AS t2 ON t1."uom_id" = t2."id"`)
	assert.Contains(t, sql, `t1."id" AS "template.id"`)
	assert.Contains(t, sql, `t2."id" AS "template.uom.id"`)
	assert.Contains(t, sql, `t2."name" AS "template.uom.name"`)
	assert.NotContains(t, sql, "DISTINCT")
}

func TestF3_DepthCapFourHopsRejected(t *testing.T) {
	fx := newFixture(t)
	cErrs := advSelectErr(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("template.uom.a.b.c")})
	assert.Equal(t, "common:err_graph_field_path_too_deep", (*cErrs)[0].Key)
}

func TestF3_BareToOneEdgeExpandsReadableColumns(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "uom")})
	assert.Contains(t, sql, `t1."id" AS "uom.id"`)
	assert.Contains(t, sql, `t1."name" AS "uom.name"`)
	assert.Contains(t, sql, `t1."symbol" AS "uom.symbol"`)
	assert.NotContains(t, sql, `"uom.tenant_id"`, "the destination tenant key is never returned")
}

func TestF3_NestedLangJsonLeafRawColumn(t *testing.T) {
	fx := newFixture(t)
	lang := model.LanguageCode("en-US")
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "template.name"), Language: &lang})
	// The projection carries the whole document; localisation is a read-side concern.
	assert.Contains(t, sql, `t1."name" AS "template.name"`)
	assert.NotContains(t, sql, "->>")
}

func TestF3_ToOneSelectSharesJoinWithFilter(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("template.category", dmodel.Equals, "tools")
	sql := advSelect(t, fx, schemaVariant, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "template.code")})
	assert.Equal(t, 1, strings.Count(sql, "LEFT JOIN"))
	assert.Contains(t, sql, `WHERE t1."category" = E'tools'`)
}

// ---- to-many ----------------------------------------------------------------------------

func TestF3_OneToManyJsonbAggLateral(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "quants.note", "quants.quantity")})
	assert.Equal(t, `SELECT t0."id", e0.v AS "quants" FROM "adv_templates" AS t0 LEFT JOIN LATERAL (SELECT COALESCE(jsonb_agg(jsonb_build_object('id', d."id", 'note', d."note", 'quantity', d."quantity") ORDER BY d."id"), '[]'::jsonb) AS v FROM "adv_quants" AS d WHERE d."template_id" = t0."id" AND d."tenant_id" = t0."tenant_id") AS e0 ON TRUE`, sql)
}

func TestF3_ToManyLeavesMergeIntoOneLateral(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("quants.note", "id", "quants.location_id")})
	assert.Equal(t, 1, strings.Count(sql, "LEFT JOIN LATERAL"))
	assert.Equal(t, 1, strings.Count(sql, "jsonb_agg"))
	assert.Contains(t, sql, `'location_id', d."location_id"`)
}

func TestF3_BareToManyEdgeExpandsReadableColumns(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "quants")})
	assert.Contains(t, sql, `'note', d."note"`)
	assert.Contains(t, sql, `'quantity', d."quantity"`)
	assert.Contains(t, sql, `'is_archived', d."is_archived"`)
	assert.NotContains(t, sql, `'tenant_id'`)
	assert.Contains(t, sql, `e0.v AS "quants"`)
}

func TestF3_ManyToManyJunctionInsideLateral(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "tags.label")})
	assert.Contains(t, sql, `FROM "adv_tags" AS d JOIN "adv_template_tag_rels" AS j ON j."tag_id" = d."id" AND j."tenant_id" = d."tenant_id" WHERE j."template_id" = t0."id" AND j."tenant_id" = t0."tenant_id"`)
	assert.Contains(t, sql, `jsonb_build_object('id', d."id", 'label', d."label")`)
	assert.NotContains(t, sql, "DISTINCT")
}

func TestF3_ToManyNoDistinctNoFanOut(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("code", dmodel.Equals, "A").OrderBy("code", dmodel.Asc)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "quants.note", "tags.label"), Size: 50})
	assert.NotContains(t, sql, "DISTINCT")
	assert.NotContains(t, sql, `LEFT JOIN "adv_quants"`)
	assert.Equal(t, 2, strings.Count(sql, "LEFT JOIN LATERAL"))
	assert.Contains(t, sql, `AS e0 ON TRUE`)
	assert.Contains(t, sql, `AS e1 ON TRUE`)
	assert.Contains(t, sql, `ORDER BY t0."code" ASC LIMIT 50`)
}

func TestF3_ToManyNotArchiveFiltered(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "quants.note")})
	// Parity with per-row hydration, which scopes by tenant only.
	assert.NotContains(t, sql, "is_archived")
}

func TestF3_JsonbNotJson(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "quants.note")})
	assert.NotContains(t, sql, "json_agg(")
	assert.NotContains(t, sql, "json_build_object(")
	assert.Contains(t, sql, "'[]'::jsonb")
}

func TestF3_FilterFanOutPlusEdgeLateralOneDistinct(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("quants.location_id", dmodel.Equals, "WH-1")
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "quants.note")})
	assert.Equal(t, 1, strings.Count(sql, "SELECT DISTINCT"))
	assert.Contains(t, sql, `LEFT JOIN "adv_quants" AS t1 ON t1."template_id" = t0."id"`)
	assert.Contains(t, sql, `e0.v AS "quants"`)
}

func TestF3_ToManyThroughToOnePrefix(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "template.quants.quantity")})
	assert.Contains(t, sql, `t1."id" AS "template.id"`)
	assert.Contains(t, sql, `e0.v AS "template.quants"`)
	assert.Contains(t, sql, `WHERE d."template_id" = t1."id" AND d."tenant_id" = t1."tenant_id"`)
}

func TestF3_NestedBeyondToManyRejected(t *testing.T) {
	fx := newFixture(t)
	cErrs := advSelectErr(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("quants.template.code")})
	assert.Contains(t, (*cErrs)[0].Key, "err_graph_nested_beyond_to_many")
}

func TestF3_TooManyEdgeLateralsRejected(t *testing.T) {
	fx := newFixture(t)
	// The fixture has three to-many edges reachable from variant→template: quants, variants, tags,
	// plus two more via the second hop path; six distinct groups exceed the cap of five.
	columns := cols("template.quants", "template.variants", "template.tags")
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: columns})
	assert.Equal(t, 3, strings.Count(sql, "jsonb_agg"))

	prev := MaxNestedEdgeLateralsForTest(2)
	defer MaxNestedEdgeLateralsForTest(prev)
	cErrs := advSelectErr(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: columns})
	assert.Contains(t, (*cErrs)[0].Key, "err_too_many_nested_edges")
}

func TestF3_CountIgnoresProjectedEdges(t *testing.T) {
	fx := newFixture(t)
	sql := advCount(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "quants.note", "uom.name")})
	assert.Equal(t, `SELECT COUNT(*) FROM "adv_templates"`, sql)
}

// ---- projection plan --------------------------------------------------------------------

func TestF3_PlanNestedProjectionMatchesSql(t *testing.T) {
	fx := newFixture(t)
	qb := NewAdvancedPgQueryBuilder().(orm.NestedProjectionCapable)
	plan, cErrs, err := qb.PlanNestedProjection(fx.schema(t, schemaVariant), fx.registry,
		cols("id", "template.code", "template.uom.name", "template.quants.note", "sku"))
	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, plan)

	tpl := plan.ToOne["template"]
	assert.Equal(t, schemaTemplate, tpl.DestSchemaName)
	assert.Equal(t, []string{"template.id"}, tpl.PkAliases)
	assert.Equal(t, map[string]string{"template.id": "id", "template.code": "code"}, tpl.Leaves)

	uom := plan.ToOne["template.uom"]
	assert.Equal(t, schemaUom, uom.DestSchemaName)
	assert.Equal(t, map[string]string{"template.uom.id": "id", "template.uom.name": "name"}, uom.Leaves)

	quants := plan.ToMany["template.quants"]
	assert.Equal(t, schemaQuant, quants.DestSchemaName)
	assert.Equal(t, "template.quants", quants.ColumnAlias)
	assert.Equal(t, []string{"id", "note"}, quants.Leaves)
}

func TestF3_PlanNestedProjectionNilWithoutEdges(t *testing.T) {
	fx := newFixture(t)
	qb := NewAdvancedPgQueryBuilder().(orm.NestedProjectionCapable)
	plan, cErrs, err := qb.PlanNestedProjection(fx.schema(t, schemaVariant), fx.registry, cols("id", "sku", "template_code"))
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Nil(t, plan)
}

func TestF3_PlanNestedProjectionReportsClientError(t *testing.T) {
	fx := newFixture(t)
	qb := NewAdvancedPgQueryBuilder().(orm.NestedProjectionCapable)
	plan, cErrs, err := qb.PlanNestedProjection(fx.schema(t, schemaVariant), fx.registry, cols("template.nope"))
	require.NoError(t, err)
	assert.Nil(t, plan)
	require.NotNil(t, cErrs)
	assert.Equal(t, "common:err_unknown_schema_field", (*cErrs)[0].Key)
}
