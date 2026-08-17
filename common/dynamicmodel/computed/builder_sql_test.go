package computed_test

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The SQL kinds have two authoring forms that must land on the identical Definition. Filters are
// compared through their marshaled wire form: the chained API stores a typed dmodel.Operator
// where JSON decoding stores a plain string, and Condition.Operator() normalizes both, so wire
// equality is the honest equivalence.

func assertSameFilter(t *testing.T, want *dmodel.SearchNode, got *dmodel.SearchNode) {
	t.Helper()
	if want == nil || got == nil {
		assert.Equal(t, want, got)
		return
	}
	wantJson, err := json.Marshal(want)
	require.NoError(t, err)
	gotJson, err := json.Marshal(got)
	require.NoError(t, err)
	assert.JSONEq(t, string(wantJson), string(gotJson))
}

func TestParseDefinitionJson_AggregateCountMatchesChainedForm(t *testing.T) {
	raw := []byte(`{
		"kind": "aggregate", "is_stored": false,
		"source": "orders", "function": "count",
		"filter": {"if": ["status", "=", "completed"]},
		"context": ["company_id"],
		"default": 0
	}`)
	parsed, isStored, err := computed.ParseDefinitionJson(raw, "completed_order_count")
	require.NoError(t, err)
	assert.False(t, isStored)

	chained := computed.Aggregate("orders", computed.AggCount,
		computed.AggFilter(dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "completed")),
		computed.AggContext("company_id"),
		computed.AggDefault(int64(0)),
	)
	parsedNode := parsed.(computed.AggregateExpr)
	chainedNode := chained.(computed.AggregateExpr)
	assertSameFilter(t, chainedNode.Filter, parsedNode.Filter)
	parsedNode.Filter, chainedNode.Filter = nil, nil
	assert.Equal(t, chainedNode, parsedNode)
}

func TestParseDefinitionJson_AggregateSumWithInnerExpression(t *testing.T) {
	raw := []byte(`{
		"kind": "aggregate", "is_stored": false,
		"source": "lines", "function": "sum",
		"expression": {"op": "multiply", "args": [{"field": "quantity"}, {"field": "unit_price"}]}
	}`)
	parsed, _, err := computed.ParseDefinitionJson(raw, "total_amount")
	require.NoError(t, err)

	chained := computed.Aggregate("lines", computed.AggSum,
		computed.AggExpr(computed.Mul(computed.F("quantity"), computed.F("unit_price"))))
	assert.Equal(t, chained, parsed)
}

func TestParseDefinitionJson_ExistsMatchesChainedForm(t *testing.T) {
	raw := []byte(`{
		"kind": "exists", "is_stored": false,
		"source": "invoices",
		"filter": {"and": [
			{"if": ["status", "=", "posted"]},
			{"if": ["company_id", "=", "${ctx.company_id}"]}
		]},
		"context": ["company_id"]
	}`)
	parsed, _, err := computed.ParseDefinitionJson(raw, "has_posted_invoice")
	require.NoError(t, err)

	chained := computed.Exists("invoices",
		dmodel.NewSearchNode().And(
			*dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "posted"),
			*dmodel.NewSearchNode().NewCondition("company_id", dmodel.Equals, computed.Ctx("company_id")),
		),
		"company_id")
	parsedNode := parsed.(computed.ExistsExpr)
	chainedNode := chained.(computed.ExistsExpr)
	assertSameFilter(t, chainedNode.Filter, parsedNode.Filter)
	parsedNode.Filter, chainedNode.Filter = nil, nil
	assert.Equal(t, chainedNode, parsedNode)
}

func TestParseDefinitionJson_LookupMatchesChainedForm(t *testing.T) {
	raw := []byte(`{
		"kind": "lookup", "is_stored": false,
		"source": "purchase_lines", "field": "unit_price",
		"order_by": [
			{"field": "order_date", "direction": "desc"},
			{"field": "id"}
		],
		"filter": {"if": ["status", "=", "done"]},
		"default": 0.0
	}`)
	parsed, _, err := computed.ParseDefinitionJson(raw, "last_purchase_price")
	require.NoError(t, err)

	chained := computed.Lookup("purchase_lines", "unit_price",
		computed.Desc("order_date"), computed.Asc("id"),
		computed.LookupFilter(dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "done")),
		computed.LookupDefault(decimal.RequireFromString("0.0")),
	)
	parsedNode := parsed.(computed.LookupExpr)
	chainedNode := chained.(computed.LookupExpr)
	assertSameFilter(t, chainedNode.Filter, parsedNode.Filter)
	parsedNode.Filter, chainedNode.Filter = nil, nil
	assert.Equal(t, chainedNode, parsedNode)
}

func TestNewDefinition_DerivesSqlKinds(t *testing.T) {
	filter := dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "posted")

	aggregate, err := computed.NewDefinition(false, computed.Aggregate("variants", computed.AggCount))
	require.NoError(t, err)
	assert.Equal(t, computed.ComputeAggregate, aggregate.Kind)
	require.NotNil(t, aggregate.Aggregate)

	exists, err := computed.NewDefinition(false, computed.Exists("invoices", filter))
	require.NoError(t, err)
	assert.Equal(t, computed.ComputeExists, exists.Kind)
	require.NotNil(t, exists.Exists)

	lookup, err := computed.NewDefinition(false, computed.Lookup("lines", "price", computed.Desc("id")))
	require.NoError(t, err)
	assert.Equal(t, computed.ComputeLookup, lookup.Kind)
	require.NotNil(t, lookup.Lookup)
}

func TestNewDefinition_SqlKindStructuralErrors(t *testing.T) {
	filter := dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "posted")
	cases := map[string]computed.Expr{
		"count with operand":           computed.Aggregate("orders", computed.AggCount, computed.AggField("id")),
		"count_distinct without field": computed.Aggregate("orders", computed.AggCountDistinct),
		"sum without operand":          computed.Aggregate("lines", computed.AggSum),
		"sum with both operands":       computed.Aggregate("lines", computed.AggSum, computed.AggField("qty"), computed.AggExpr(computed.F("qty"))),
		"unknown aggregate function":   computed.Aggregate("lines", computed.AggregateFunction("median"), computed.AggField("qty")),
		"aggregate without source":     computed.Aggregate("", computed.AggCount),
		"exists without filter":        computed.Exists("invoices", nil),
		"lookup without order_by":      computed.Lookup("lines", "price"),
		"lookup without field":         computed.Lookup("lines", "", computed.Desc("id")),
		"nested aggregate in inner expression": computed.Aggregate("lines", computed.AggSum,
			computed.AggExpr(computed.Aggregate("lines", computed.AggSum, computed.AggField("qty")))),
		"exists nested in expression": computed.Fn("coalesce", computed.Exists("invoices", filter), computed.Lit(false)),
	}
	for name, root := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := computed.NewDefinition(false, root)
			require.Error(t, err)
		})
	}
}

func TestParseDefinitionJson_SqlKindDecodeErrors(t *testing.T) {
	cases := map[string]string{
		"bad order direction": `{"kind": "lookup", "is_stored": false, "source": "lines",
			"field": "price", "order_by": [{"field": "id", "direction": "descending"}]}`,
		"object default": `{"kind": "aggregate", "is_stored": false, "source": "lines",
			"function": "count", "default": {"$raw": "1)) OR 1=1 --"}}`,
		"malformed filter": `{"kind": "exists", "is_stored": false, "source": "invoices",
			"filter": {"if": ["a", "=", "b"], "and": [{"if": ["c", "=", "d"]}]}}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := computed.ParseDefinitionJson([]byte(raw), "evil")
			require.Error(t, err)
		})
	}
}

func TestCtx_RendersWholeStringPlaceholder(t *testing.T) {
	assert.Equal(t, "${ctx.company_id}", computed.Ctx("company_id"))
}

func sqlKindModelJson(computedBlock string) string {
	return `{
		"name": "cf_sql_kind_schema",
		"fields": [
			{"name": "id", "data_type": "ulid", "primary_key": true},
			{"name": "subject", "data_type": {"type": "int64", "min": 0, "max": 1000000}, "computed": ` + computedBlock + `}
		]
	}`
}

func TestModelJsonSchema_AcceptsSqlKindBlocks(t *testing.T) {
	valid := sqlKindModelJson(`{
		"kind": "aggregate", "is_stored": false, "source": "orders", "function": "count",
		"filter": {"if": ["status", "=", "completed"]}, "default": 0
	}`)
	_, errs := dmodel.ParseModelJsonSafe(valid)
	assert.Equal(t, 0, errs.Count(), "a well-formed aggregate block must pass the JSON Schema: %v", errs)
}

func TestModelJsonSchema_RejectsCrossKindProperties(t *testing.T) {
	cases := map[string]string{
		"related with source": `{"kind": "related", "is_stored": false,
			"field": "template.name", "source": "orders"}`,
		"aggregate with order_by": `{"kind": "aggregate", "is_stored": false,
			"source": "orders", "function": "count", "order_by": [{"field": "id"}]}`,
		"exists with default": `{"kind": "exists", "is_stored": false,
			"source": "orders", "filter": {"if": ["a", "=", "b"]}, "default": false}`,
		"lookup missing order_by": `{"kind": "lookup", "is_stored": false,
			"source": "lines", "field": "price"}`,
		"filter with unknown property": `{"kind": "exists", "is_stored": false,
			"source": "orders", "filter": {"sql": "1=1"}}`,
	}
	for name, block := range cases {
		t.Run(name, func(t *testing.T) {
			_, errs := dmodel.ParseModelJsonSafe(sqlKindModelJson(block))
			assert.Greater(t, errs.Count(), 0, "the JSON Schema must reject this block")
		})
	}
}
