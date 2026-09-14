package advanced

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// One synthetic product-like graph covering every shape the advanced builder must handle:
//
//	adv_variant  --template-->  adv_template  --uom-->  adv_uom
//	                            adv_template  <--quants--  adv_quant   (one:many, tenant + is_archived)
//	                            adv_template  <--variants-- adv_variant
//	                            adv_template  <==tags==>  adv_tag      (many:many via adv_template_tag_rel)
//
// adv_template carries a computed field of every kind. The plan cache in the computed package
// is global and rebuilt by each FinalizeRelations, so a test always builds the fixture right
// before rendering SQL and never runs in parallel.

const (
	schemaUom      = "adv_uom"
	schemaTemplate = "adv_template"
	schemaVariant  = "adv_variant"
	schemaQuant    = "adv_quant"
	schemaTag      = "adv_tag"
	schemaTagRel   = "adv_template_tag_rel"
)

func idField() *dmodel.FieldBuilder {
	return dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()
}

func tenantField() *dmodel.FieldBuilder {
	return dmodel.DefineField().Name("tenant_id").DataType(dmodel.FieldDataTypeUlid()).TenantKey()
}

func stringField(name string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(name).DataType(dmodel.FieldDataTypeString(0, 100))
}

func decimalField(name string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(name).DataType(dmodel.FieldDataTypeDecimal("-1000000000", "1000000000", 6))
}

func uomSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel(schemaUom).ShouldBuildDb().TableName("adv_uoms").
		Field(idField()).
		Field(tenantField()).
		Field(stringField("name")).
		Field(stringField("symbol")).
		Build()
}

func templateSchema() *dmodel.ModelSchema {
	quantityGtZero := dmodel.NewSearchNode().NewCondition("quantity", dmodel.GreaterThan, 0)
	inLocation := dmodel.NewSearchNode().NewCondition("location_id", dmodel.Equals, computed.Ctx("location_id"))
	return dmodel.DefineModel(schemaTemplate).ShouldBuildDb().TableName("adv_templates").
		Field(idField()).
		Field(tenantField()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeLangJson(0, 100))).
		Field(stringField("code")).
		Field(stringField("category")).
		Field(dmodel.DefineField().Name("uom_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(decimalField("base_price")).
		Field(dmodel.DefineField().Name("created_at").DataType(dmodel.FieldDataTypeDateTime())).
		Field(dmodel.DefineField().Name("is_archived").DataType(dmodel.FieldDataTypeBoolean())).
		// aggregate
		Field(decimalField("on_hand_quantity").Computed(false,
			computed.Aggregate("quants", computed.AggSum, computed.AggField("quantity"), computed.AggDefault(0)))).
		Field(decimalField("located_quantity").Computed(false,
			computed.Aggregate("quants", computed.AggSum, computed.AggField("quantity"),
				computed.AggFilter(inLocation), computed.AggContext("location_id")))).
		// exists
		Field(dmodel.DefineField().Name("has_stock").DataType(dmodel.FieldDataTypeBoolean()).Computed(false,
			computed.Exists("quants", quantityGtZero))).
		// lookup
		Field(stringField("largest_quant_note").Computed(false,
			computed.Lookup("quants", "note", computed.Desc("quantity")))).
		// related
		Field(stringField("uom_name").Computed(false, computed.Related("uom.name"))).
		// expression over an aggregate and over physical columns
		Field(decimalField("stock_value").Computed(false,
			computed.Mul(computed.F("on_hand_quantity"), computed.F("base_price")))).
		Field(stringField("display_code").Computed(false,
			computed.Fn("concat", computed.F("code"), computed.Lit("-"), computed.F("category")))).
		Field(stringField("price_band").Computed(false,
			computed.Case().
				When(computed.Gt(computed.F("base_price"), computed.Lit(100)), computed.Lit("high")).
				Else(computed.Lit("low")))).
		Field(decimalField("unit_value").Computed(false,
			computed.Div(computed.F("base_price"), computed.Fn("coalesce", computed.F("on_hand_quantity"), computed.Lit(1))))).
		Field(dmodel.DefineField().Name("age_days").DataType(dmodel.FieldDataTypeInt32(0, 100000)).Computed(false,
			computed.Fn("date_diff", computed.F("created_at"), computed.F("created_at")))).
		// function (Go only)
		Field(dmodel.DefineField().Name("external_score").DataType(dmodel.FieldDataTypeInt32(0, 100)).Computed(false,
			computed.GoFunction("adv.score").DependsOn("code"))).
		EdgeTo(dmodel.Edge("uom").ManyToOne(schemaUom, dmodel.DynamicFields{"uom_id": "id"})).
		EdgeTo(dmodel.Edge("tags").ManyToMany(schemaTag, schemaTagRel, "template")).
		EdgeFrom(dmodel.Edge("quants").Existing(schemaQuant, "template")).
		EdgeFrom(dmodel.Edge("variants").Existing(schemaVariant, "template")).
		Build()
}

func variantSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel(schemaVariant).ShouldBuildDb().TableName("adv_variants").
		Field(idField()).
		Field(tenantField()).
		Field(stringField("sku")).
		Field(dmodel.DefineField().Name("template_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(decimalField("price")).
		Field(dmodel.DefineField().Name("is_archived").DataType(dmodel.FieldDataTypeBoolean())).
		Field(stringField("template_code").Computed(false, computed.Related("template.code"))).
		Field(stringField("template_uom_name").Computed(false, computed.Related("template.uom.name"))).
		Field(stringField("template_uom_alias").Computed(false, computed.Related("template.uom_name"))).
		EdgeTo(dmodel.Edge("template").ManyToOne(schemaTemplate, dmodel.DynamicFields{"template_id": "id"})).
		Build()
}

func quantSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel(schemaQuant).ShouldBuildDb().TableName("adv_quants").
		Field(idField()).
		Field(tenantField()).
		Field(dmodel.DefineField().Name("template_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(stringField("location_id")).
		Field(decimalField("quantity")).
		Field(stringField("note")).
		Field(dmodel.DefineField().Name("is_archived").DataType(dmodel.FieldDataTypeBoolean())).
		EdgeTo(dmodel.Edge("template").ManyToOne(schemaTemplate, dmodel.DynamicFields{"template_id": "id"})).
		Build()
}

func tagSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel(schemaTag).ShouldBuildDb().TableName("adv_tags").
		Field(idField()).
		Field(tenantField()).
		Field(stringField("label")).
		Field(dmodel.DefineField().Name("template_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).Computed(false,
			computed.Aggregate("templates", computed.AggCount, computed.AggDefault(0)))).
		EdgeTo(dmodel.Edge("templates").ManyToMany(schemaTemplate, schemaTagRel, "tag")).
		Build()
}

func tagRelSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel(schemaTagRel).ShouldBuildDb().TableName("adv_template_tag_rels").
		Field(dmodel.DefineField().Name("template_id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("tag_id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(tenantField()).
		Build()
}

type fixture struct {
	registry *dmodel.SchemaRegistry
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	reg := dmodel.NewSchemaRegistry()
	for _, schema := range []*dmodel.ModelSchema{
		uomSchema(), templateSchema(), variantSchema(), quantSchema(), tagSchema(), tagRelSchema(),
	} {
		require.NoError(t, reg.Register(schema))
	}
	require.NoError(t, reg.FinalizeRelations())
	return &fixture{registry: reg}
}

func (this *fixture) schema(t *testing.T, name string) *dmodel.ModelSchema {
	t.Helper()
	schema := this.registry.Get(name)
	require.NotNil(t, schema, "schema %s must be registered", name)
	return schema
}

func TestFixture_FinalizesWithAllComputedKinds(t *testing.T) {
	fx := newFixture(t)
	plan := computed.PlanFor(schemaTemplate)
	require.NotNil(t, plan)
	kinds := map[string]computed.ComputeKind{
		"on_hand_quantity":   computed.ComputeAggregate,
		"located_quantity":   computed.ComputeAggregate,
		"has_stock":          computed.ComputeExists,
		"largest_quant_note": computed.ComputeLookup,
		"uom_name":           computed.ComputeRelated,
		"stock_value":        computed.ComputeExpression,
		"display_code":       computed.ComputeExpression,
		"price_band":         computed.ComputeExpression,
		"unit_value":         computed.ComputeExpression,
		"age_days":           computed.ComputeExpression,
		"external_score":     computed.ComputeFunction,
	}
	for name, kind := range kinds {
		fieldPlan := plan.Fields[name]
		require.NotNil(t, fieldPlan, "field %s must have a plan", name)
		require.Equal(t, kind, fieldPlan.Def.Kind, "field %s kind", name)
	}
	variantPlan := computed.PlanFor(schemaVariant)
	require.NotNil(t, variantPlan)
	require.Equal(t, computed.ComputeRelated, variantPlan.Fields["template_code"].Def.Kind)

	template := fx.schema(t, schemaTemplate)
	edges := map[string]dmodel.RelationType{}
	for _, rel := range template.Relations() {
		edges[rel.Edge] = rel.RelationType
	}
	require.Equal(t, dmodel.RelationTypeManyToOne, edges["uom"])
	require.Equal(t, dmodel.RelationTypeOneToMany, edges["quants"])
	require.Equal(t, dmodel.RelationTypeOneToMany, edges["variants"])
	require.Equal(t, dmodel.RelationTypeManyToMany, edges["tags"])
}
