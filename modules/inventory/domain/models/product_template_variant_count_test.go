package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// variant_count is the module's adoption of the SQL-computed aggregate kind: one correlated
// COUNT subquery over the template's "variants" inverse edge, opt-in by explicit projection.
// These tests pin the resolved plan and the emitted SQL against the REAL schema builders, so a
// schema edit that would silently change the query shape fails here first.

func TestVariantCount_ResolvesAsAggregatePlan(t *testing.T) {
	finalizedVariantSchema(t)

	schemaPlan := computed.PlanFor(ProductTemplateSchemaName)
	require.NotNil(t, schemaPlan)
	fieldPlan := schemaPlan.Fields["variant_count"]
	require.NotNil(t, fieldPlan)

	assert.Equal(t, computed.ComputeAggregate, fieldPlan.Def.Kind)
	assert.Equal(t, computed.TypeInt64, fieldPlan.Type)
	require.NotNil(t, fieldPlan.SqlSource)
	assert.Equal(t, "variants", fieldPlan.SqlSource.Edge)
	assert.Equal(t, ProductVariantSchemaName, fieldPlan.SqlSource.SourceSchemaName)
	assert.False(t, fieldPlan.SqlSource.Many)
}

func TestVariantCount_EmittedSubqueryShape(t *testing.T) {
	finalizedVariantSchema(t)
	registry := dmodel.GetSchemaRegistry()
	template := registry.Get(ProductTemplateSchemaName)
	fieldPlan := computed.PlanFor(ProductTemplateSchemaName).Fields["variant_count"]
	require.NotNil(t, fieldPlan)

	builder := orm.NewPgQueryBuilder().(*orm.PgQueryBuilder)
	sql, cErrs, err := builder.ComputedSubqueryExpr(registry, template, "t0", fieldPlan, nil)
	require.NoError(t, err)
	require.Empty(t, cErrs)

	assert.Contains(t, sql, "SELECT COUNT(*) FROM")
	assert.Contains(t, sql, `"inventory_product_variants"`)
	assert.Contains(t, sql, `"product_template_id" = t0."id"`,
		"the subquery must correlate on the variant's template FK")
	assert.Contains(t, sql, `"is_archived" = FALSE`, "archived variants never count")
	assert.NotContains(t, sql, "JOIN")
	assert.NotContains(t, sql, "GROUP BY")
}

func TestVariantCount_ProjectsOnlyWhenRequested(t *testing.T) {
	finalizedVariantSchema(t)
	registry := dmodel.GetSchemaRegistry()
	template := registry.Get(ProductTemplateSchemaName)

	builder := orm.NewPgQueryBuilder()
	withField, cErrs, err := builder.SqlSelectGraph(template, registry, nil, orm.SqlSelectGraphOpts{
		Columns: orm.ToSelectColumns([]string{"id", "variant_count"}),
	})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Contains(t, *withField, `AS "variant_count"`)

	wildcard, cErrs, err := builder.SqlSelectGraph(template, registry, nil, orm.SqlSelectGraphOpts{})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.NotContains(t, *wildcard, "COUNT(*)",
		"the default projection must not pay for the subquery")
}
