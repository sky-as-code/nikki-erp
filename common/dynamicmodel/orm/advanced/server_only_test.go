package advanced

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// A server_only field (a sequence definition's FF1 key) is readable on its own root, but never
// through an edge: a bare edge leaves it out, and naming it as an edge leaf is an unknown field.
//
//	adv_secret_ref --def--> adv_secret_def <--children-- adv_secret_child

const (
	schemaSecretDef   = "adv_secret_def"
	schemaSecretRef   = "adv_secret_ref"
	schemaSecretChild = "adv_secret_child"
)

func serverOnlyField(name string) *dmodel.FieldBuilder {
	return stringField(name).Metadata(map[string]any{dmodel.MetadataServerOnly: true})
}

func newServerOnlyFixture(t *testing.T) *fixture {
	t.Helper()
	def := dmodel.DefineModel(schemaSecretDef).ShouldBuildDb().TableName("adv_secret_defs").
		Field(idField()).Field(tenantField()).Field(stringField("code")).Field(serverOnlyField("params")).
		EdgeFrom(dmodel.Edge("children").Existing(schemaSecretChild, "def")).
		Build()
	ref := dmodel.DefineModel(schemaSecretRef).ShouldBuildDb().TableName("adv_secret_refs").
		Field(idField()).Field(tenantField()).
		Field(dmodel.DefineField().Name("def_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("def").ManyToOne(schemaSecretDef, dmodel.DynamicFields{"def_id": "id"})).
		Build()
	child := dmodel.DefineModel(schemaSecretChild).ShouldBuildDb().TableName("adv_secret_children").
		Field(idField()).Field(tenantField()).Field(serverOnlyField("tweak")).Field(stringField("note")).
		Field(dmodel.DefineField().Name("def_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("def").ManyToOne(schemaSecretDef, dmodel.DynamicFields{"def_id": "id"})).
		Build()
	reg := dmodel.NewSchemaRegistry()
	for _, schema := range []*dmodel.ModelSchema{def, ref, child} {
		require.NoError(t, reg.Register(schema))
	}
	require.NoError(t, reg.FinalizeRelations())
	return &fixture{registry: reg}
}

func TestServerOnly_RootColumnStillSelectable(t *testing.T) {
	fx := newServerOnlyFixture(t)
	sql := advSelect(t, fx, schemaSecretDef, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "params")})
	assert.Contains(t, sql, `"params"`)
}

func TestServerOnly_BareToOneEdgeOmitsIt(t *testing.T) {
	fx := newServerOnlyFixture(t)
	sql := advSelect(t, fx, schemaSecretRef, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "def")})
	assert.Contains(t, sql, `AS "def.code"`)
	assert.NotContains(t, sql, "params")
}

func TestServerOnly_BareToManyEdgeOmitsIt(t *testing.T) {
	fx := newServerOnlyFixture(t)
	sql := advSelect(t, fx, schemaSecretRef, nil, orm.SqlSelectGraphOpts{Columns: cols("id", "def.children")})
	assert.Contains(t, sql, `'note'`)
	assert.NotContains(t, sql, "tweak")
}

func TestServerOnly_EdgeLeafRefused(t *testing.T) {
	cases := map[string]struct {
		schema string
		graph  *dmodel.SearchGraph
		fields []string
	}{
		"select to-one":  {schemaSecretRef, nil, []string{"id", "def.params"}},
		"select to-many": {schemaSecretDef, nil, []string{"id", "children.tweak"}},
		"filter":         {schemaSecretRef, dmodel.NewSearchGraph().NewCondition("def.params", dmodel.Equals, "k"), nil},
		"sort":           {schemaSecretRef, dmodel.NewSearchGraph().OrderBy("def.params", dmodel.Asc), nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fx := newServerOnlyFixture(t)
			fields := tc.fields
			if fields == nil {
				fields = []string{"id"}
			}
			cErrs := advSelectErr(t, fx, tc.schema, tc.graph, orm.SqlSelectGraphOpts{Columns: cols(fields...)})
			assert.Equal(t, "common:err_unknown_schema_field", (*cErrs)[0].Key)
		})
	}
}
