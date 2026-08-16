package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func fkRegistryWith(t *testing.T, schemas ...*dmodel.ModelSchema) *dmodel.SchemaRegistry {
	t.Helper()
	reg := dmodel.NewSchemaRegistry()
	for _, schema := range schemas {
		require.NoError(t, reg.Register(schema))
	}
	require.NoError(t, reg.FinalizeRelations())
	return reg
}

func fkFieldOf(t *testing.T, schema *dmodel.ModelSchema, name string) *dmodel.ModelField {
	t.Helper()
	field, ok := schema.Field(name)
	require.Truef(t, ok, "field %q not found on schema %q", name, schema.Name())
	return field
}

func fkThemeSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("fk_theme").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
}

// The many:one case, mirroring vending_machine's Kiosk -> Theme edge: the key map's key names
// the local column, which is the field that must come out flagged.
func fkKioskSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("fk_kiosk").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("code").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("theme_ref").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("theme").ManyToOne("fk_theme", dmodel.DynamicFields{"theme_ref": "id"})).
		Build()
}

func TestForeignKey_ManyToOneMarksLocalColumn(t *testing.T) {
	kiosk := fkKioskSchema()
	fkRegistryWith(t, kiosk, fkThemeSchema())

	assert.True(t, fkFieldOf(t, kiosk, "theme_ref").IsForeignKey(),
		"the key map's local column is a foreign key")
	assert.False(t, fkFieldOf(t, kiosk, "code").IsForeignKey(),
		"an ordinary field is not a foreign key")
	assert.False(t, fkFieldOf(t, kiosk, "id").IsForeignKey(),
		"a primary key is not a foreign key")
}

// A foreign key is a system field; an ordinary column is not. This is the split that decides
// whether a column reaches the client's field picker.
func TestForeignKey_CountsAsSystemField(t *testing.T) {
	kiosk := fkKioskSchema()
	fkRegistryWith(t, kiosk, fkThemeSchema())

	assert.True(t, fkFieldOf(t, kiosk, "theme_ref").IsSystemField())
	assert.True(t, fkFieldOf(t, kiosk, "id").IsSystemField(), "a primary key is a system field")
	assert.False(t, fkFieldOf(t, kiosk, "code").IsSystemField())
}

// For one:many the key map names a column on the CHILD, so marking it on the declaring schema
// would flag the wrong table.
func TestForeignKey_OneToManyMarksChildColumn(t *testing.T) {
	parent := dmodel.DefineModel("fk_parent").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		EdgeTo(dmodel.Edge("children").OneToMany("fk_child", dmodel.DynamicFields{"parent_id": "id"})).
		Build()
	child := dmodel.DefineModel("fk_child").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("parent_id").DataType(dmodel.FieldDataTypeUlid())).
		Build()

	fkRegistryWith(t, parent, child)

	assert.True(t, fkFieldOf(t, child, "parent_id").IsForeignKey(),
		"the child holds the column, so the child's field is flagged")
	_, hasField := parent.Field("parent_id")
	assert.False(t, hasField, "the parent never had such a column to flag")
}

// Many:many contributes no FK column to either peer; the association columns live on the
// junction and are flagged there.
func TestForeignKey_ManyToManyMarksJunctionColumns(t *testing.T) {
	group := dmodel.DefineModel("fk_group").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		EdgeTo(dmodel.Edge("users").ManyToMany("fk_user", "fk_group_user_rel", "group")).
		Build()
	user := dmodel.DefineModel("fk_user").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		EdgeTo(dmodel.Edge("groups").ManyToMany("fk_group", "fk_group_user_rel", "user")).
		Build()
	through := dmodel.DefineModel("fk_group_user_rel").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("group_id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("user_id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Build()

	fkRegistryWith(t, group, user, through)

	assert.True(t, fkFieldOf(t, through, "group_id").IsForeignKey())
	assert.True(t, fkFieldOf(t, through, "user_id").IsForeignKey())
}

func TestForeignKey_ClonePreservesFlag(t *testing.T) {
	kiosk := fkKioskSchema()
	fkRegistryWith(t, kiosk, fkThemeSchema())

	assert.True(t, fkFieldOf(t, kiosk, "theme_ref").Clone().IsForeignKey())
}

// Clone dropped isVersioningKey before this refactor, which went unnoticed while nothing read
// it. IsSystemField reads it now, so a cloned etag must keep reporting as a system field.
func TestForeignKey_ClonePreservesVersioningKey(t *testing.T) {
	schema := dmodel.DefineModel("fk_versioned").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("etag").DataType(dmodel.FieldDataTypeEtag()).VersioningKey()).
		Build()

	cloned := fkFieldOf(t, schema, "etag").Clone()

	assert.True(t, cloned.IsVersioningKey())
	assert.True(t, cloned.IsSystemField())
}
