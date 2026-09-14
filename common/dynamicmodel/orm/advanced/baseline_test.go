package advanced

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// Baseline parity: for everything PgQueryBuilder already does, the advanced builder must render
// the same statement, or one that differs only by qualifying root columns with the root alias.

type sqlPair struct {
	base, adv string
}

func (this *fixture) selectBoth(t *testing.T, schemaName string, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts) sqlPair {
	t.Helper()
	schema := this.schema(t, schemaName)
	base := renderSql(t, orm.NewPgQueryBuilder(), schema, this.registry, graph, opts)
	adv := renderSql(t, NewAdvancedPgQueryBuilder(), schema, this.registry, graph, opts)
	return sqlPair{base: base, adv: adv}
}

func renderSql(
	t *testing.T, qb orm.QueryBuilder, schema *dmodel.ModelSchema, reg *dmodel.SchemaRegistry,
	graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts,
) string {
	t.Helper()
	sql, cErrs, err := qb.SqlSelectGraph(schema, reg, graph, opts)
	require.NoError(t, err)
	require.Nil(t, cErrs, "unexpected client errors: %v", cErrs)
	require.NotNil(t, sql)
	return *sql
}

func advSelect(t *testing.T, fx *fixture, schemaName string, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts) string {
	t.Helper()
	return renderSql(t, NewAdvancedPgQueryBuilder(), fx.schema(t, schemaName), fx.registry, graph, opts)
}

func advSelectErr(t *testing.T, fx *fixture, schemaName string, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts) *ft.ClientErrors {
	t.Helper()
	sql, cErrs, err := NewAdvancedPgQueryBuilder().SqlSelectGraph(fx.schema(t, schemaName), fx.registry, graph, opts)
	require.NoError(t, err)
	require.Nil(t, sql)
	require.NotNil(t, cErrs)
	return cErrs
}

func cols(names ...string) []orm.SelectColumn { return orm.ToSelectColumns(names) }

func TestBaseline_PlainSelectMatchesPgQueryBuilder(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("code", dmodel.Equals, "ABC").OrderBy("code", dmodel.Asc)
	pair := fx.selectBoth(t, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "code"), Page: 2, Size: 10})
	assert.Equal(t, pair.base, pair.adv)
	assert.Equal(t, `SELECT "id", "code" FROM "adv_templates" WHERE "code" = E'ABC' ORDER BY "code" ASC LIMIT 10 OFFSET 20`, pair.adv)
}

func TestBaseline_WildcardWithoutColumns(t *testing.T) {
	fx := newFixture(t)
	pair := fx.selectBoth(t, schemaTemplate, nil, orm.SqlSelectGraphOpts{})
	assert.Equal(t, pair.base, pair.adv)
	assert.Equal(t, `SELECT * FROM "adv_templates"`, pair.adv)
}

func TestBaseline_ManyToOneFilterJoinParity(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("template.code", dmodel.Equals, "ABC")
	pair := fx.selectBoth(t, schemaVariant, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "sku", "template.code")})
	// Same join, same predicate; the advanced builder additionally projects the edge's primary
	// key so the repository can tell an absent relation from an empty one.
	assert.Equal(t, strings.ReplaceAll(pair.adv, `t1."id" AS "template.id", `, ""), pair.base)
	assert.Contains(t, pair.adv, `LEFT JOIN "adv_templates" AS t1 ON t0."template_id" = t1."id"`)
	assert.Contains(t, pair.adv, `t1."code" AS "template.code"`)
	assert.NotContains(t, pair.adv, "DISTINCT")
}

func TestBaseline_OneToManyFilterForcesDistinctParity(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("quants.note", dmodel.Contains, "x").OrderBy("code", dmodel.Desc)
	pair := fx.selectBoth(t, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "code")})
	assert.Equal(t, pair.base, pair.adv)
	assert.Equal(t, 1, strings.Count(pair.adv, "SELECT DISTINCT"))
	assert.Contains(t, pair.adv, `t1."template_id" = t0."id"`)
}

func TestBaseline_DistinctTokenParity(t *testing.T) {
	fx := newFixture(t)
	pair := fx.selectBoth(t, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: []orm.SelectColumn{orm.SelectColumn("category").AsDistinct()}})
	assert.Equal(t, pair.base, pair.adv)
	assert.Equal(t, `SELECT DISTINCT "category" FROM "adv_templates"`, pair.adv)
}

func TestBaseline_ManyToManyFilterParity(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("tags.label", dmodel.Equals, "new")
	pair := fx.selectBoth(t, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Equal(t, pair.base, pair.adv)
	assert.Contains(t, pair.adv, `"adv_template_tag_rels" AS t1`)
	assert.Contains(t, pair.adv, `"adv_tags" AS t2`)
	assert.Contains(t, pair.adv, "SELECT DISTINCT")
}

func TestBaseline_LinkedNotLinkedParity(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().NewCondition("quants", dmodel.Linked, "01Q"),
		*dmodel.NewSearchNode().NewCondition("tags", dmodel.NotLinked, "01T"),
	)
	pair := fx.selectBoth(t, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "code")})
	assert.Contains(t, pair.adv, `t0."id" IN (SELECT "template_id" FROM "adv_quants" WHERE "id" = '01Q' AND "tenant_id" = t0."tenant_id")`)
	assert.Contains(t, pair.adv, `t0."id" NOT IN (SELECT th."template_id" FROM "adv_template_tag_rels" AS th WHERE th."tag_id" = '01T' AND th."tenant_id" = t0."tenant_id")`)
	assert.Contains(t, pair.adv, `FROM "adv_templates" AS t0`)
	// The only difference to the base builder is that root columns are alias-qualified once the
	// root carries an alias.
	assert.Equal(t, strings.ReplaceAll(pair.adv, `t0."id", t0."code"`, `"id", "code"`), pair.base)
}

func TestBaseline_LangJsonOrderAndFilterParity(t *testing.T) {
	fx := newFixture(t)
	lang := model.LanguageCode("en-US")
	graph := dmodel.NewSearchGraph().NewCondition("name", dmodel.Contains, "shirt").OrderBy("name", dmodel.Asc)
	pair := fx.selectBoth(t, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "name"), Language: &lang})
	assert.Equal(t, pair.base, pair.adv)
	assert.Contains(t, pair.adv, `ORDER BY ("name" ->> 'en-US') ASC`)
	assert.Contains(t, pair.adv, `("name" ->> 'en-US') ILIKE`)
}

func TestBaseline_AggregateSelectTokens(t *testing.T) {
	fx := newFixture(t)
	pair := fx.selectBoth(t, schemaQuant, nil, orm.SqlSelectGraphOpts{Columns: cols("SUM(quantity)", "COUNT(DISTINCT::location_id)")})
	assert.Equal(t, pair.base, pair.adv)
	assert.Equal(t, `SELECT SUM("quantity"), COUNT(DISTINCT "location_id") FROM "adv_quants"`, pair.adv)
}

func TestBaseline_GoComputedFieldSelectedIsSkipped(t *testing.T) {
	fx := newFixture(t)
	pair := fx.selectBoth(t, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "external_score", "display_code")})
	assert.Equal(t, pair.base, pair.adv)
	assert.Equal(t, `SELECT "id" FROM "adv_templates"`, pair.adv)
}

func TestBaseline_OnlyGoComputedFieldsAnchorOnPrimaryKey(t *testing.T) {
	fx := newFixture(t)
	pair := fx.selectBoth(t, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("external_score")})
	assert.Equal(t, pair.base, pair.adv)
	assert.Equal(t, `SELECT "id" FROM "adv_templates"`, pair.adv)
}

func TestBaseline_UnknownFieldIsClientError(t *testing.T) {
	fx := newFixture(t)
	cErrs := advSelectErr(t, fx, schemaTemplate, dmodel.NewSearchGraph().NewCondition("nope", dmodel.Equals, 1), orm.SqlSelectGraphOpts{})
	assert.Equal(t, "common:err_unknown_schema_field", (*cErrs)[0].Key)
}

func TestBaseline_FilterDepthCapIsFiveDots(t *testing.T) {
	fx := newFixture(t)
	cErrs := advSelectErr(t, fx, schemaVariant,
		dmodel.NewSearchGraph().NewCondition("template.uom.a.b.c.d.e", dmodel.Equals, 1), orm.SqlSelectGraphOpts{})
	assert.Equal(t, "common:err_graph_field_path_too_deep", (*cErrs)[0].Key)
}

func TestBaseline_CountAndExistsParity(t *testing.T) {
	fx := newFixture(t)
	schema := fx.schema(t, schemaTemplate)
	graph := dmodel.NewSearchGraph().NewCondition("quants.quantity", dmodel.GreaterThan, 0)
	opts := orm.SqlSelectGraphOpts{Columns: cols("id", "code")}

	baseCount, _, err := orm.NewPgQueryBuilder().SqlCountGraph(schema, fx.registry, graph, opts)
	require.NoError(t, err)
	advCount, cErrs, err := NewAdvancedPgQueryBuilder().SqlCountGraph(schema, fx.registry, graph, opts)
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Equal(t, *baseCount, *advCount)
	assert.Contains(t, *advCount, `SELECT COUNT(*) FROM (SELECT DISTINCT t0."id", t0."code"`)

	plain := dmodel.NewSearchGraph().NewCondition("code", dmodel.Equals, "A")
	baseCount, _, err = orm.NewPgQueryBuilder().SqlCountGraph(schema, fx.registry, plain, opts)
	require.NoError(t, err)
	advCount, _, err = NewAdvancedPgQueryBuilder().SqlCountGraph(schema, fx.registry, plain, opts)
	require.NoError(t, err)
	assert.Equal(t, *baseCount, *advCount)
	assert.Equal(t, `SELECT COUNT(*) FROM "adv_templates" WHERE "code" = E'A'`, *advCount)

	baseExists, _, err := orm.NewPgQueryBuilder().SqlExistsGraph(schema, fx.registry, graph)
	require.NoError(t, err)
	advExists, _, err := NewAdvancedPgQueryBuilder().SqlExistsGraph(schema, fx.registry, graph)
	require.NoError(t, err)
	assert.Equal(t, *baseExists, *advExists)
	assert.True(t, strings.HasPrefix(*advExists, "SELECT EXISTS (SELECT 1 FROM"))
}

func TestBaseline_DmlAndDdlInheritedFromPgQueryBuilder(t *testing.T) {
	fx := newFixture(t)
	schema := fx.schema(t, schemaUom)
	data := dmodel.DynamicFields{"id": "01U", "name": "Kilogram", "symbol": "kg", "tenant_id": "01T"}
	base, _, err := orm.NewPgQueryBuilder().SqlInsert(schema, data, false)
	require.NoError(t, err)
	adv, _, err := NewAdvancedPgQueryBuilder().SqlInsert(schema, data, false)
	require.NoError(t, err)
	assert.Equal(t, *base, *adv)
}
