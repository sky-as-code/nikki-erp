package baserepo

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm/advanced"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// The fold is what makes a single-statement edge projection indistinguishable from per-row
// hydration to a caller. These tests run it over hand-built rows, no database involved.

const foldSchemaPrefix = "test_baserepo_fold_"

func foldRegistry(t *testing.T) *dmodel.SchemaRegistry {
	t.Helper()
	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(dmodel.DefineModel(foldSchemaPrefix+"uom").ShouldBuildDb().TableName("fold_uoms").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("tenant_id").DataType(dmodel.FieldDataTypeUlid()).TenantKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 50))).
		Build()))
	require.NoError(t, reg.Register(dmodel.DefineModel(foldSchemaPrefix+"template").ShouldBuildDb().TableName("fold_templates").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("tenant_id").DataType(dmodel.FieldDataTypeUlid()).TenantKey()).
		Field(dmodel.DefineField().Name("code").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("uom_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("uom").ManyToOne(foldSchemaPrefix+"uom", dmodel.DynamicFields{"uom_id": "id"})).
		EdgeFrom(dmodel.Edge("quants").Existing(foldSchemaPrefix+"quant", "template")).
		Build()))
	require.NoError(t, reg.Register(dmodel.DefineModel(foldSchemaPrefix+"quant").ShouldBuildDb().TableName("fold_quants").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("tenant_id").DataType(dmodel.FieldDataTypeUlid()).TenantKey()).
		Field(dmodel.DefineField().Name("template_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(dmodel.DefineField().Name("quantity").DataType(dmodel.FieldDataTypeDecimal("0", "1000000", 3))).
		Field(dmodel.DefineField().Name("count").DataType(dmodel.FieldDataTypeInt64(0, 1000000))).
		Field(dmodel.DefineField().Name("active").DataType(dmodel.FieldDataTypeBoolean())).
		Field(dmodel.DefineField().Name("note").DataType(dmodel.FieldDataTypeString(0, 50))).
		EdgeTo(dmodel.Edge("template").ManyToOne(foldSchemaPrefix+"template", dmodel.DynamicFields{"template_id": "id"})).
		Build()))
	require.NoError(t, reg.FinalizeRelations())
	return reg
}

func toOnePlan() *orm.NestedProjection {
	return &orm.NestedProjection{
		ToOne: map[string]orm.ToOneEdgeProjection{
			"template": {
				DestSchemaName: foldSchemaPrefix + "template",
				PkAliases:      []string{"template.id"},
				Leaves:         map[string]string{"template.id": "id", "template.code": "code", "template.tenant_id": "tenant_id"},
			},
			"template.uom": {
				DestSchemaName: foldSchemaPrefix + "uom",
				PkAliases:      []string{"template.uom.id"},
				Leaves:         map[string]string{"template.uom.id": "id", "template.uom.name": "name"},
			},
		},
		ToMany: map[string]orm.ToManyEdgeProjection{},
	}
}

func TestNestedFold_ToOneFlatKeysBecomeMap(t *testing.T) {
	reg := foldRegistry(t)
	rows := []dmodel.DynamicFields{{
		"id": "01V", "sku": "SKU-1",
		"template.id": "01T", "template.code": "TPL", "template.tenant_id": "01TEN",
	}}
	require.NoError(t, foldNestedRows(reg, rows, toOnePlan()))

	row := rows[0]
	assert.Equal(t, "SKU-1", row["sku"])
	assert.Equal(t, dmodel.DynamicFields{"id": "01T", "code": "TPL", "uom": nil}, row["template"], "tenant key stripped")
	_, flatLeft := row["template.code"]
	assert.False(t, flatLeft, "flat aliases are removed")
	assert.Nil(t, row["template"].(dmodel.DynamicFields)["uom"], "missing second hop is nil under its parent, key present")
}

func TestNestedFold_NullPkYieldsNilEdge(t *testing.T) {
	reg := foldRegistry(t)
	rows := []dmodel.DynamicFields{{"id": "01V", "sku": "SKU-1"}}
	require.NoError(t, foldNestedRows(reg, rows, toOnePlan()))

	value, present := rows[0]["template"]
	assert.True(t, present, "the edge key is present even when the relation is absent")
	assert.Nil(t, value)
}

func TestNestedFold_DeepPathNests(t *testing.T) {
	reg := foldRegistry(t)
	rows := []dmodel.DynamicFields{{
		"id": "01V",
		"template.id": "01T", "template.code": "TPL",
		"template.uom.id": "01U", "template.uom.name": "Kilogram",
	}}
	require.NoError(t, foldNestedRows(reg, rows, toOnePlan()))

	template := rows[0]["template"].(dmodel.DynamicFields)
	assert.Equal(t, dmodel.DynamicFields{"id": "01U", "name": "Kilogram"}, template["uom"])
	assert.Equal(t, "TPL", template["code"])
}

func TestNestedFold_ToManyJsonbDecodedAndTyped(t *testing.T) {
	reg := foldRegistry(t)
	plan := &orm.NestedProjection{
		ToOne: map[string]orm.ToOneEdgeProjection{},
		ToMany: map[string]orm.ToManyEdgeProjection{
			"quants": {DestSchemaName: foldSchemaPrefix + "quant", ColumnAlias: "quants", Leaves: []string{"id", "quantity", "count", "active", "note"}},
		},
	}
	rows := []dmodel.DynamicFields{
		{"id": "01T", "quants": []byte(`[{"id":"01Q","quantity":12.345,"count":7,"active":true,"note":null,"tenant_id":"01TEN"},{"id":"01R","quantity":0,"count":0,"active":false,"note":"back"}]`)},
		{"id": "02T", "quants": "[]"},
		{"id": "03T"},
	}
	require.NoError(t, foldNestedRows(reg, rows, plan))

	quants := rows[0]["quants"].([]dmodel.DynamicFields)
	require.Len(t, quants, 2)
	assert.Equal(t, "01Q", quants[0]["id"])
	assert.True(t, decimal.RequireFromString("12.345").Equal(quants[0]["quantity"].(decimal.Decimal)), "exact decimal, never float64")
	assert.Equal(t, int64(7), quants[0]["count"])
	assert.Equal(t, true, quants[0]["active"])
	_, hasNote := quants[0]["note"]
	assert.False(t, hasNote, "a JSON null is an absent field, as a NULL column is")
	_, hasTenant := quants[0]["tenant_id"]
	assert.False(t, hasTenant)
	assert.Equal(t, "back", quants[1]["note"])
	assert.Equal(t, []dmodel.DynamicFields{}, rows[1]["quants"], "an empty array is an empty slice")
	assert.Equal(t, []dmodel.DynamicFields{}, rows[2]["quants"], "a missing column is an empty slice too")
}

func TestNestedFold_ToManyUnderAbsentParentIsDropped(t *testing.T) {
	reg := foldRegistry(t)
	plan := toOnePlan()
	plan.ToMany["template.quants"] = orm.ToManyEdgeProjection{
		DestSchemaName: foldSchemaPrefix + "quant", ColumnAlias: "template.quants", Leaves: []string{"id"},
	}
	rows := []dmodel.DynamicFields{
		{"id": "01V", "template.quants": `[]`},
		{"id": "02V", "template.id": "01T", "template.quants": `[{"id":"01Q"}]`},
	}
	require.NoError(t, foldNestedRows(reg, rows, plan))

	assert.Nil(t, rows[0]["template"])
	_, leaked := rows[0]["template.quants"]
	assert.False(t, leaked)
	template := rows[1]["template"].(dmodel.DynamicFields)
	assert.Equal(t, []dmodel.DynamicFields{{"id": "01Q"}}, template["quants"])
}

func TestNestedFold_ProjectedScanFieldsMapAliasesToDestinationFields(t *testing.T) {
	reg := foldRegistry(t)
	root := reg.Get(foldSchemaPrefix + "quant")
	fields := projectedScanFields(reg, root, []string{"id", "note", "template.code", "template"}, toOnePlan())

	assert.Equal(t, "note", fields["note"].Name())
	assert.Equal(t, foldSchemaPrefix+"template", schemaOf(t, reg, "code", fields["template.code"]))
	assert.Equal(t, "name", fields["template.uom.name"].Name())
	_, edgeMapped := fields["template"]
	assert.False(t, edgeMapped, "an edge field is never a scanned column")
}

func schemaOf(t *testing.T, reg *dmodel.SchemaRegistry, fieldName string, field *dmodel.ModelField) string {
	t.Helper()
	require.NotNil(t, field)
	assert.Equal(t, fieldName, field.Name())
	if tpl := reg.Get(foldSchemaPrefix + "template"); tpl != nil {
		if f, ok := tpl.Field(fieldName); ok && f == field {
			return tpl.Name()
		}
	}
	return ""
}

// ---- repository wiring ------------------------------------------------------------------

type stubCapableBuilder struct {
	orm.QueryBuilder
	plan  *orm.NestedProjection
	cErrs *ft.ClientErrors
	asked [][]orm.SelectColumn
}

func (this *stubCapableBuilder) PlanNestedProjection(
	_ *dmodel.ModelSchema, _ *dmodel.SchemaRegistry, columns []orm.SelectColumn,
) (*orm.NestedProjection, *ft.ClientErrors, error) {
	this.asked = append(this.asked, columns)
	return this.plan, this.cErrs, nil
}

func TestSearch_CapableBuilderIsAskedForThePlan(t *testing.T) {
	repo := virtualRepo(t)
	stub := &stubCapableBuilder{plan: &orm.NestedProjection{}}
	repo.queryBuilder = stub

	plan, cErrs, err := repo.nestedProjectionPlan([]string{"id", "peer.name"})
	require.NoError(t, err)
	assert.Empty(t, cErrs)
	assert.Same(t, stub.plan, plan)
	require.Len(t, stub.asked, 1)
	assert.Equal(t, orm.ToSelectColumns([]string{"id", "peer.name"}), stub.asked[0])
}

func TestSearch_NonCapableBuilderKeepsHydrationPath(t *testing.T) {
	repo := virtualRepo(t)
	repo.queryBuilder = orm.NewPgQueryBuilder()

	plan, cErrs, err := repo.nestedProjectionPlan([]string{"id", "peer.name"})
	require.NoError(t, err)
	assert.Empty(t, cErrs)
	assert.Nil(t, plan)
	assert.True(t, repo.hasNestedOrEdgeColumns([]string{"id", "peer.name"}))
}

func TestSearch_CapableBuilderRelaxesSelectDepthValidation(t *testing.T) {
	repo := virtualRepo(t)
	repo.queryBuilder = orm.NewPgQueryBuilder()
	require.NotNil(t, repo.validateSelectColumns([]string{"peer.owner.name"}), "the hydration path caps selection at one dot")

	repo.queryBuilder = advanced.NewAdvancedPgQueryBuilder()
	assert.Nil(t, repo.validateSelectColumns([]string{"peer.owner.name"}), "a projecting builder validates dotted paths itself")
	assert.NotNil(t, repo.validateSelectColumns([]string{"nosuch"}), "root fields are still validated here")
}

func TestSearch_AdvancedBuilderPlansAgainstTheGlobalRegistry(t *testing.T) {
	repo := virtualRepo(t)
	repo.queryBuilder = advanced.NewAdvancedPgQueryBuilder()

	plan, cErrs, err := repo.nestedProjectionPlan([]string{"id", "peer.name"})
	require.NoError(t, err)
	require.Empty(t, cErrs)
	require.NotNil(t, plan)
	assert.Equal(t, map[string]string{"peer.id": "id", "peer.name": "name"}, plan.ToOne["peer"].Leaves)

	plan, cErrs, err = repo.nestedProjectionPlan([]string{"id", "sku"})
	require.NoError(t, err)
	require.Empty(t, cErrs)
	assert.Nil(t, plan, "no edge in the projection means the plain path")

	_, cErrs, err = repo.nestedProjectionPlan([]string{"peer.nosuch"})
	require.NoError(t, err)
	require.NotEmpty(t, cErrs)
	assert.Contains(t, cErrs[0].Key, "err_unknown_schema_field")
}
