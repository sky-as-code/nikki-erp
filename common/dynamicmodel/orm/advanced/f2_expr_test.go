package advanced

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// F2: related and expression computed fields are filterable and sortable in SQL.

func TestF2_RelatedFilterRewritesToJoinLeaf(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("uom_name", dmodel.Equals, "Kilogram").OrderBy("uom_name", dmodel.Desc)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "uom_name")})

	assert.Contains(t, sql, `LEFT JOIN "adv_uoms" AS t1 ON t0."uom_id" = t1."id"`)
	assert.Contains(t, sql, `WHERE t1."name" = E'Kilogram'`)
	assert.Contains(t, sql, `ORDER BY t1."name" DESC`)
	// Selected on the root it stays Go-filled, as with PgQueryBuilder.
	assert.True(t, strings.HasPrefix(sql, `SELECT t0."id" FROM`), sql)
	assert.NotContains(t, sql, "LATERAL")
}

func TestF2_RelatedFilterSharesJoinWithExplicitEdgePath(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().NewCondition("uom_name", dmodel.Contains, "kilo"),
		*dmodel.NewSearchNode().NewCondition("uom.symbol", dmodel.Equals, "kg"),
	)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Equal(t, 1, strings.Count(sql, "LEFT JOIN"))
	assert.Contains(t, sql, `t1."name" ILIKE E'%kilo%'`)
	assert.Contains(t, sql, `t1."symbol" = E'kg'`)
}

func TestF2_ExpressionArithmeticCompiles(t *testing.T) {
	fx := newFixture(t)
	// stock_value = on_hand_quantity * base_price: the aggregate operand becomes the shared lateral.
	graph := dmodel.NewSearchGraph().NewCondition("stock_value", dmodel.GreaterThan, 1000).OrderBy("stock_value", dmodel.Desc)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id", "on_hand_quantity")})

	assert.Equal(t, 1, strings.Count(sql, "SUM("), "aggregate operand shares the projected lateral")
	assert.Contains(t, sql, `WHERE (c0.v * t0."base_price") > 1000`)
	assert.Contains(t, sql, `ORDER BY (c0.v * t0."base_price") DESC`)
}

func TestF2_ExpressionCaseCompiles(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("price_band", dmodel.Equals, "high")
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, sql, `WHERE (CASE WHEN ("base_price" > 100) THEN 'high' ELSE 'low' END) = E'high'`)
	assert.NotContains(t, sql, " AS t0", "no join or lateral was needed, so the root stays unaliased")
}

func TestF2_DivisionCastsNumericAndCoalesceCompiles(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("unit_value", dmodel.LessThan, 5)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, sql, `WHERE ((t0."base_price")::numeric / COALESCE(c0.v, 1)) < 5`)
}

func TestF2_ConcatIsNullPropagating(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("display_code", dmodel.StartsWith, "AB")
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, sql, `WHERE ("code" || '-' || "category") ILIKE E'AB%'`)
	assert.NotContains(t, sql, "CONCAT(")
}

func TestF2_DatetimeFunctionRejected(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("age_days", dmodel.GreaterThan, 30)
	cErrs := advSelectErr(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, (*cErrs)[0].Key, "err_computed_field_not_filterable")
	assert.Contains(t, (*cErrs)[0].Message, "date_diff")
}

func TestF2_FunctionKindRejectedWithDistinctMessage(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("external_score", dmodel.Equals, 3)
	cErrs := advSelectErr(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, (*cErrs)[0].Key, "err_computed_field_not_filterable")
	assert.Contains(t, (*cErrs)[0].Message, "evaluated in Go")
}

func TestF2_ValueConversionUsesComputedType(t *testing.T) {
	fx := newFixture(t)
	// has_stock is boolean: a string value must be refused by the shared converter, not rendered.
	graph := dmodel.NewSearchGraph().NewCondition("has_stock", dmodel.Equals, "yes")
	cErrs := advSelectErr(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.NotEmpty(t, *cErrs)

	// on_hand_quantity is decimal: the IN operator converts each value.
	graph = dmodel.NewSearchGraph().NewCondition("on_hand_quantity", dmodel.In, 1, 2.5)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, sql, "WHERE c0.v IN (1, 2.5)")
}

func TestF2_InjectionInFilterValueEscaped(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().NewCondition("uom_name", dmodel.Equals, "x'; DROP TABLE adv_uoms; --")
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.Contains(t, sql, `t1."name" = E'x\'; DROP TABLE adv_uoms; --'`)
}

func TestF2_RelatedLangJsonLeafKeepsLocalizedSemantics(t *testing.T) {
	fx := newFixture(t)
	lang := model.LanguageCode("vi-VN")
	// variant.template_code is related to a plain string; use the variant's edge to a lang-json
	// column directly to prove the leaf field's type drives the rendering.
	graph := dmodel.NewSearchGraph().NewCondition("template.name", dmodel.Contains, "áo").OrderBy("template_code", dmodel.Asc)
	sql := advSelect(t, fx, schemaVariant, graph, orm.SqlSelectGraphOpts{Columns: cols("id"), Language: &lang})
	assert.Contains(t, sql, `(t1."name" ->> 'vi-VN') ILIKE`)
	assert.Contains(t, sql, `ORDER BY t1."code" ASC`)
	assert.Equal(t, 1, strings.Count(sql, "LEFT JOIN"))
}

func TestF2_CountWithExpressionFilterUsesLateral(t *testing.T) {
	fx := newFixture(t)
	schema := fx.schema(t, schemaTemplate)
	graph := dmodel.NewSearchGraph().NewCondition("stock_value", dmodel.GreaterThan, 0)
	sql, cErrs, err := NewAdvancedPgQueryBuilder().SqlCountGraph(schema, fx.registry, graph, orm.SqlSelectGraphOpts{Columns: cols("id")})
	assert.NoError(t, err)
	assert.Nil(t, cErrs)
	assert.Contains(t, *sql, "LEFT JOIN LATERAL")
	assert.Contains(t, *sql, `WHERE (c0.v * t0."base_price") > 0`)
}
