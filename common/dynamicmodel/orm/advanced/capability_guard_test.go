package advanced

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// Guard: the schema's is_filterable / is_sortable flags and the builder must agree on every
// computed field. A field advertised as filterable that the builder refuses (or the reverse)
// would show a filter in the UI that fails, or hide one that works.

const schemaScored = "adv_scored"

// scoredSchema adds the one shape the shared fixture lacks: an expression over a Go-only field.
func scoredSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel(schemaScored).ShouldBuildDb().TableName("adv_scoreds").
		Field(idField()).
		Field(tenantField()).
		Field(stringField("code")).
		Field(dmodel.DefineField().Name("score").DataType(dmodel.FieldDataTypeInt32(0, 100)).Computed(false,
			computed.GoFunction("adv.scored").DependsOn("code"))).
		Field(dmodel.DefineField().Name("score_plus_one").DataType(dmodel.FieldDataTypeInt32(0, 101)).Computed(false,
			computed.Add(computed.F("score"), computed.Lit(1)))).
		Field(stringField("code_upper").Computed(false, computed.Fn("upper", computed.F("code")))).
		Build()
}

func newGuardFixture(t *testing.T) *fixture {
	t.Helper()
	reg := dmodel.NewSchemaRegistry()
	for _, schema := range []*dmodel.ModelSchema{
		uomSchema(), templateSchema(), variantSchema(), quantSchema(), tagSchema(), tagRelSchema(), scoredSchema(),
	} {
		require.NoError(t, reg.Register(schema))
	}
	require.NoError(t, reg.FinalizeRelations())
	return &fixture{registry: reg}
}

// guardContext satisfies every ${ctx.*} placeholder the fixture declares.
var guardContext = map[string]any{"location_id": "LOC1"}

func TestGuard_SqlRenderableMatchesBuilderForEveryComputedField(t *testing.T) {
	fx := newGuardFixture(t)
	checked := 0
	for _, schemaName := range []string{schemaTemplate, schemaVariant, schemaTag, schemaScored} {
		schema := fx.schema(t, schemaName)
		for name, field := range schema.Fields() {
			if !field.IsVirtual() || field.IsEdgeModel() {
				continue
			}
			checked++
			renderable := computed.IsSqlRenderable(schemaName, name)
			assertBuilderAgrees(t, fx, schema, name, renderable)
		}
	}
	assert.GreaterOrEqual(t, checked, 15, "the guard must actually cover the fixture's computed fields")
}

func assertBuilderAgrees(t *testing.T, fx *fixture, schema *dmodel.ModelSchema, name string, renderable bool) {
	t.Helper()
	qb := NewAdvancedPgQueryBuilder()
	opts := orm.SqlSelectGraphOpts{Columns: cols("id"), ComputedContext: guardContext}
	filter := dmodel.NewSearchGraph().NewCondition(name, dmodel.IsSet)
	_, filterErrs, err := qb.SqlSelectGraph(schema, fx.registry, filter, opts)
	require.NoError(t, err)
	order := dmodel.NewSearchGraph().OrderBy(name, dmodel.Asc)
	_, orderErrs, err := qb.SqlSelectGraph(schema, fx.registry, order, opts)
	require.NoError(t, err)

	label := schema.Name() + "." + name
	if renderable {
		assert.Nil(t, filterErrs, "%s is advertised filterable but the builder refused it: %v", label, filterErrs)
		assert.Nil(t, orderErrs, "%s is advertised sortable but the builder refused it: %v", label, orderErrs)
		return
	}
	if assert.NotNil(t, filterErrs, "%s is advertised not filterable but the builder accepted it", label) {
		assert.Contains(t, (*filterErrs)[0].Key, "err_computed_field_not_filterable", label)
	}
	assert.NotNil(t, orderErrs, "%s is advertised not sortable but the builder accepted it", label)
}

func TestGuard_KnownCapabilities(t *testing.T) {
	newGuardFixture(t)
	expected := map[string]bool{
		schemaTemplate + ".on_hand_quantity":   true,
		schemaTemplate + ".has_stock":          true,
		schemaTemplate + ".largest_quant_note": true,
		schemaTemplate + ".uom_name":           true,
		schemaTemplate + ".stock_value":        true,
		schemaTemplate + ".display_code":       true,
		schemaTemplate + ".age_days":           false, // date_diff has no SQL rendering
		schemaTemplate + ".external_score":     false, // function kind
		schemaVariant + ".template_uom_name":   true,
		schemaScored + ".score_plus_one":       false, // reads a function-kind field
		schemaScored + ".code_upper":           true,
	}
	for path, want := range expected {
		schemaName, fieldName := splitGuardPath(path)
		assert.Equal(t, want, computed.IsSqlRenderable(schemaName, fieldName), path)
	}
}

func TestGuard_SchemaPayloadCarriesTheFlags(t *testing.T) {
	fx := newGuardFixture(t)
	payload, err := json.Marshal(fx.schema(t, schemaTemplate).ToSimplized())
	require.NoError(t, err)
	var decoded struct {
		Fields map[string]struct {
			IsFilterable bool `json:"is_filterable"`
			IsSortable   bool `json:"is_sortable"`
		} `json:"fields"`
	}
	require.NoError(t, json.Unmarshal(payload, &decoded))

	cases := map[string]bool{
		"code":             true,  // physical column
		"on_hand_quantity": true,  // aggregate
		"uom_name":         true,  // related
		"age_days":         false, // date/time builtin
		"external_score":   false, // function kind
		"uom":              false, // edge
		"quants":           false, // edge
	}
	for name, want := range cases {
		field, ok := decoded.Fields[name]
		if !assert.True(t, ok, "field %s must be in the payload", name) {
			continue
		}
		assert.Equal(t, want, field.IsFilterable, "%s is_filterable", name)
		assert.Equal(t, want, field.IsSortable, "%s is_sortable", name)
	}
}

func splitGuardPath(path string) (string, string) {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[:i], path[i+1:]
		}
	}
	return "", path
}
