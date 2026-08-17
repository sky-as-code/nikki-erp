package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// The variant's stock figures (INV-PI-012, CR §8.1/§8.2). Product owns no quantity: these are two
// correlated SUM subqueries over the "quants" inverse edge plus a derived difference, so a product
// list can show stock without Product storing any (CR §4.4).
//
// They are pinned against the REAL schema builders — a schema edit that changed the query shape,
// the correlation or the opt-in projection breaks these first.

func TestVariantStock_ResolveAsAggregatePlans(t *testing.T) {
	finalizedVariantSchema(t)

	schemaPlan := computed.PlanFor(ProductVariantSchemaName)
	require.NotNil(t, schemaPlan)

	for _, name := range []string{"on_hand_quantity", "reserved_quantity"} {
		fieldPlan := schemaPlan.Fields[name]
		require.NotNil(t, fieldPlan, name)

		assert.Equal(t, computed.ComputeAggregate, fieldPlan.Def.Kind, name)
		assert.Equal(t, computed.TypeDecimal, fieldPlan.Type, name)
		require.NotNil(t, fieldPlan.SqlSource, name)
		assert.Equal(t, "quants", fieldPlan.SqlSource.Edge, name)
		assert.Equal(t, StockQuantSchemaName, fieldPlan.SqlSource.SourceSchemaName, name)
		// one:many, correlated by the quant's FK — not a junction table.
		assert.False(t, fieldPlan.SqlSource.Many, name)
	}
}

// available_quantity is an expression over the two aggregates rather than a third subquery, so the
// identity available = on_hand - reserved holds by construction and cannot drift from its operands.
func TestVariantStock_AvailableIsDerivedFromTheAggregates(t *testing.T) {
	finalizedVariantSchema(t)

	fieldPlan := computed.PlanFor(ProductVariantSchemaName).Fields["available_quantity"]
	require.NotNil(t, fieldPlan)

	assert.Equal(t, computed.ComputeExpression, fieldPlan.Def.Kind)
	assert.Equal(t, computed.TypeDecimal, fieldPlan.Type)
	assert.Nil(t, fieldPlan.SqlSource, "it must not cost a subquery of its own")

	depends := map[string]bool{}
	for _, dep := range fieldPlan.Dependencies {
		depends[dep.Field] = true
	}
	assert.True(t, depends["on_hand_quantity"], "available depends on on-hand")
	assert.True(t, depends["reserved_quantity"], "available depends on reserved")
}

func TestVariantStock_EmittedSubqueryShape(t *testing.T) {
	finalizedVariantSchema(t)
	registry := dmodel.GetSchemaRegistry()
	variant := registry.Get(ProductVariantSchemaName)
	fieldPlan := computed.PlanFor(ProductVariantSchemaName).Fields["on_hand_quantity"]
	require.NotNil(t, fieldPlan)

	builder := orm.NewPgQueryBuilder().(*orm.PgQueryBuilder)
	sql, cErrs, err := builder.ComputedSubqueryExpr(registry, variant, "t0", fieldPlan, nil)
	require.NoError(t, err)
	require.Empty(t, cErrs)

	assert.Contains(t, sql, `SUM("on_hand_quantity")`)
	assert.Contains(t, sql, `"inventory_stock_quants"`)
	assert.Contains(t, sql, `"product_variant_id" = t0."id"`,
		"the subquery must correlate on the quant's variant FK")
	// SUM over zero rows is NULL; the declared default is what makes a variant with no stock
	// answer 0 rather than nothing (AC-PROD-INT-004).
	assert.Contains(t, sql, "COALESCE")
	assert.NotContains(t, sql, "GROUP BY")

	// The quant is not archivable — a balance is zeroed or deleted, never archived — so unlike
	// variant_count there is no is_archived scope to apply, and every quant row counts.
	assert.NotContains(t, sql, "is_archived")
}

func TestVariantStock_ProjectsOnlyWhenRequested(t *testing.T) {
	finalizedVariantSchema(t)
	registry := dmodel.GetSchemaRegistry()
	variant := registry.Get(ProductVariantSchemaName)

	builder := orm.NewPgQueryBuilder()
	withFields, cErrs, err := builder.SqlSelectGraph(variant, registry, nil, orm.SqlSelectGraphOpts{
		Columns: orm.ToSelectColumns([]string{"id", "on_hand_quantity", "available_quantity"}),
	})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Contains(t, *withFields, `AS "on_hand_quantity"`)

	// A product list that does not ask for stock must not pay for the subqueries. This is what
	// makes the columns optional in the CR's sense (§8.1) rather than a cost on every read.
	wildcard, cErrs, err := builder.SqlSelectGraph(variant, registry, nil, orm.SqlSelectGraphOpts{})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.NotContains(t, *wildcard, `SUM("on_hand_quantity")`,
		"the default projection must not pay for the stock subqueries")
}

// available_quantity is an expression, so it is filled in Go after the read rather than projected
// as SQL — it legitimately does not appear in the SELECT. What must happen is that asking for it
// alone still drags both aggregate operands into the projection, or it would evaluate against
// absent operands and quietly read as 0 - 0 for a variant that does hold stock.
func TestVariantStock_AvailableAlonePullsInItsOperands(t *testing.T) {
	finalizedVariantSchema(t)

	plan, cErrs := computed.BuildEvalPlan(ProductVariantSchemaName, []string{"available_quantity"})
	require.Empty(t, cErrs)
	require.NotNil(t, plan)

	assert.Contains(t, plan.Wanted, "on_hand_quantity", "the on-hand aggregate must be evaluated")
	assert.Contains(t, plan.Wanted, "reserved_quantity", "the reserved aggregate must be evaluated")
	assert.Contains(t, plan.ExtraFields, "on_hand_quantity",
		"the on-hand subquery must be added to the projection")
	assert.Contains(t, plan.ExtraFields, "reserved_quantity",
		"the reserved subquery must be added to the projection")
}

// The point of INV-PI-012: the stock figures reach the product list as ordinary selectable
// columns, offered by the same picker that offers the template_* fields. They are deliberately NOT
// in the engine's DefaultFields — a user opts each one in, and only then does the request pay for
// the subquery (CR §8.1, "optional" columns).
func TestVariantStock_AreOfferedAsSelectableColumns(t *testing.T) {
	fields := simplizedFields(t, finalizedVariantSchema(t))
	selectable := selectableFieldNames(fields)

	for _, name := range []string{"on_hand_quantity", "reserved_quantity", "available_quantity"} {
		require.Contains(t, fields, name, "field %q must be served by meta/schema", name)
		assert.True(t, selectable[name],
			"stock field %q must be offered as a selectable column", name)
	}
}

// The fields are readable and writable-never. A client that echoes a stock figure back in a PUT is
// told so, rather than having it silently dropped — and no column exists for it to reach anyway.
func TestVariantStock_AreVirtualAndNotWritable(t *testing.T) {
	schema := finalizedVariantSchema(t)

	// Columns() is the PHYSICAL list — note that Column(name) singular is just an alias for
	// Field(name) and answers for virtual fields too, so it cannot decide this.
	physical := map[string]bool{}
	for _, col := range schema.Columns() {
		physical[col.Name()] = true
	}
	readable := map[string]bool{}
	for _, field := range schema.ReadableFields() {
		readable[field.Name()] = true
	}

	for _, name := range []string{"on_hand_quantity", "reserved_quantity", "available_quantity"} {
		field, ok := schema.Field(name)
		require.True(t, ok, name)

		assert.True(t, field.IsComputed(), name)
		assert.False(t, field.IsPersisted(), name)
		assert.True(t, field.IsVirtual(), name)
		assert.False(t, field.IsEdgeModel(), "%s: a client may render it as a column", name)

		assert.False(t, physical[name], "%s must not occupy a database column", name)
		// Virtual but still selectable: this is what lets the frontend offer it as a list column
		// without any viewkit change.
		assert.True(t, readable[name], "%s must stay selectable by a client", name)
	}
}
