package computed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Multi-hop related paths: the chain variant -> template -> uom, declared either as the full
// path or as a related field pointing at the template's own related field.

func relatedChainRegistry(t *testing.T, variantFields ...*dmodel.FieldBuilder) *dmodel.SchemaRegistry {
	t.Helper()
	uom := dmodel.DefineModel("cf_rel_uom").ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Field(dmodel.DefineField().Name("owner_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("owner").ManyToOne("cf_rel_owner", dmodel.DynamicFields{"owner_id": "id"})).
		Build()
	owner := dmodel.DefineModel("cf_rel_owner").ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
	template := dmodel.DefineModel("cf_rel_template").ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("uom_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(dmodel.DefineField().Name("uom_name").DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("uom.name"))).
		Field(dmodel.DefineField().Name("uom_upper").DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Fn("upper", computed.F("uom_name")))).
		EdgeTo(dmodel.Edge("uom").ManyToOne("cf_rel_uom", dmodel.DynamicFields{"uom_id": "id"})).
		Build()
	variant := dmodel.DefineModel("cf_rel_variant").ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("template_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("template").ManyToOne("cf_rel_template", dmodel.DynamicFields{"template_id": "id"}))
	for _, field := range variantFields {
		variant.Field(field)
	}
	reg := dmodel.NewSchemaRegistry()
	for _, schema := range []*dmodel.ModelSchema{uom, owner, template, variant.Build()} {
		require.NoError(t, reg.Register(schema))
	}
	return reg
}

func relatedField(name, path string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(name).DataType(dmodel.FieldDataTypeString(0, 200)).
		Computed(false, computed.Related(path))
}

func TestRelated_TwoHopPathResolvesToNestedLeaf(t *testing.T) {
	reg := relatedChainRegistry(t, relatedField("template_uom_name", "template.uom.name"))
	require.NoError(t, reg.FinalizeRelations())

	plan := computed.PlanFor("cf_rel_variant").Fields["template_uom_name"]
	require.NotNil(t, plan)
	assert.Equal(t, "template", plan.RelatedEdge)
	assert.Equal(t, "uom.name", plan.RelatedLeaf, "everything past the first hop is a nested path on the batched read")
	assert.Equal(t, "cf_rel_template", plan.RelatedSchemaName)
	assert.Equal(t, "template_id", plan.RelatedFkColumn)
	assert.Equal(t, "id", plan.RelatedRefColumn)
	assert.Equal(t, computed.TypeString, plan.Type)
	assert.Equal(t, []string{"template_id"}, plan.PhysicalOperands)
	assert.Contains(t, plan.Dependencies, computed.FieldRef{Schema: "cf_rel_uom", Field: "name"})
	assert.Contains(t, plan.Dependencies, computed.FieldRef{Schema: "cf_rel_template", Field: "uom_id"})
}

func TestRelated_LeafThatIsRelatedIsFlattened(t *testing.T) {
	reg := relatedChainRegistry(t, relatedField("template_uom_name", "template.uom_name"))
	require.NoError(t, reg.FinalizeRelations())

	plan := computed.PlanFor("cf_rel_variant").Fields["template_uom_name"]
	require.NotNil(t, plan)
	assert.Equal(t, "template", plan.RelatedEdge)
	assert.Equal(t, "uom.name", plan.RelatedLeaf)
	assert.Equal(t, computed.TypeString, plan.Type)
}

func TestRelated_LeafThatIsExpressionRejected(t *testing.T) {
	reg := relatedChainRegistry(t, relatedField("bad", "template.uom_upper"))
	err := reg.FinalizeRelations()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "physical column or at another related field")
}

func TestRelated_DepthLimitCountsFlattenedHops(t *testing.T) {
	reg := relatedChainRegistry(t, relatedField("bad", "template.uom.owner.name"))
	err := reg.FinalizeRelations()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolves to 3 edges")
	assert.Contains(t, err.Error(), "the limit is 2")
}

func TestRelated_EvalPlanReadsNestedLeafFromFirstHop(t *testing.T) {
	reg := relatedChainRegistry(t, relatedField("template_uom_name", "template.uom.name"))
	require.NoError(t, reg.FinalizeRelations())

	plan, cErrs := computed.BuildEvalPlan("cf_rel_variant", []string{"id", "template_uom_name"})
	require.Empty(t, cErrs)
	require.NotNil(t, plan)
	require.Len(t, plan.RelatedReads, 1)
	read := plan.RelatedReads[0]
	assert.Equal(t, "cf_rel_template", read.SchemaName)
	assert.Equal(t, map[string]string{"template_uom_name": "uom.name"}, read.Leaves)

	var askedFields []string
	rows := []dmodel.DynamicFields{
		{"id": "01V", "template_id": "01T"},
		{"id": "02V", "template_id": "02T"},
		{"id": "03V", "template_id": "03T"},
	}
	search := func(schemaName, keyColumn string, keys []any, fields []string) ([]dmodel.DynamicFields, error) {
		askedFields = fields
		return []dmodel.DynamicFields{
			{"id": "01T", "uom": dmodel.DynamicFields{"id": "01U", "name": "Kilogram"}},
			{"id": "02T", "uom": nil},
		}, nil
	}
	require.NoError(t, plan.Apply(rows, computed.EvalDeps{Search: search}))

	assert.ElementsMatch(t, []string{"id", "uom.name"}, askedFields, "the nested path is projected on the source read")
	assert.Equal(t, "Kilogram", rows[0]["template_uom_name"])
	_, hasSecond := rows[1]["template_uom_name"]
	assert.False(t, hasSecond, "a template without a unit leaves the field absent")
	_, hasThird := rows[2]["template_uom_name"]
	assert.False(t, hasThird, "a missing template leaves the field absent")
}
