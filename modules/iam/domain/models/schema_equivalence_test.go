package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)


// These tests pin the JSON-backed builders to the Go builders they replaced. The legacy
// chains below are verbatim copies of what role.go / group.go contained before the
// conversion, and exist only as the reference to diff against.

func TestRoleSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, roleSchemaBuilderLegacy().Build(), RoleSchemaBuilder().Build())
}

func TestGroupSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, groupSchemaBuilderLegacy().Build(), GroupSchemaBuilder().Build())
}

func TestGroupUserRelSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, groupUserRelSchemaBuilderLegacy().Build(), GroupUserRelSchemaBuilder().Build())
}

func requireBaseSchemasRegistered(t *testing.T) {
	t.Helper()
	// Normally done by CoreModule.RegisterModels during app start-up.
	_ = basemodel.RegisterJsonBaseSchemas()
}

func roleSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RoleSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", RoleSchemaName)).
		TableName("iam_roles").
		RecordLabelField(RoleFieldName).
		PartialUnique(RoleFieldName, RoleFieldOrgId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(RoleFieldName).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldDescription).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(RoleFieldOwnerGroupId).
				Description(model.LangJson{"en-US": "One of the users in this group can approve grant requests for this role"}),
		).
		Field(
			basemodel.DefineFieldId(RoleFieldOwnerUserId).
				Description(model.LangJson{"en-US": "Only this user can approve grant requests for this role"}),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsPrivate).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				Default(false),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequestable).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequiredAttach).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequiredComment).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false),
		).
		Field(
			basemodel.DefineFieldId(RoleFieldOrgId).
				Description(model.LangJson{"en-US": "If specified, the role only accepts entitlements whose org_unit_id belongs to this organization. " +
					"Otherwise, the role only accepts entitlements with domain scope (org_unit_id is nil)",
				}),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(RoleEdgeRoleRequests).
				Label(model.LangJson{"en-US": "Grant requests"}).
				Existing(RoleRequestSchemaName, RoleReqEdgeRole),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeOwnerGroup).
				Label(model.LangJson{"en-US": "Owner group"}).
				ManyToOne(GroupSchemaName, dmodel.DynamicFields{
					RoleFieldOwnerGroupId: GroupFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeOwnerUser).
				Label(model.LangJson{"en-US": "Owner user"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleFieldOwnerUserId: UserFieldId,
				}),
		).
		EdgeFrom(
			dmodel.Edge(RoleEdgeEntitlements).
				Label(model.LangJson{"en-US": "Entitlements"}).
				Existing(EntitlementSchemaName, EntitlementEdgeRole),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeAssignedGroups).
				Label(model.LangJson{"en-US": "Assigned groups"}).
				ManyToMany(GroupSchemaName, RoleGroupAssignmentSchemaName, "role"),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeAssignedUsers).
				Label(model.LangJson{"en-US": "Assigned users"}).
				ManyToMany(UserSchemaName, RoleUserAssignmentSchemaName, "role"),
		)
}

func groupUserRelSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(GrpUsrRelSchemaName).
		TableName("iam_group_user_rel").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		CompositeUnique(GrpUsrRelFieldGroupId, GrpUsrRelFieldUserId).
		Field(
			basemodel.DefineFieldId(GrpUsrRelFieldGroupId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(GrpUsrRelFieldUserId).
				RequiredForCreate(),
		)
}

func groupSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(GroupSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", GroupSchemaName)).
		TableName("iam_groups").
		RecordLabelField(GroupFieldName).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(GroupFieldName).
				Label(model.NewLangJsonRefSf("fields.%s", GroupFieldName)).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(GroupFieldDescription).
				Label(model.NewLangJsonRefSf("fields.%s", GroupFieldDescription)).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(GroupFieldOwnerId).
				RequiredForCreate().
				Description(model.LangJson{model.LanguageCodeEnUs: "User who owns the group, is notified when membership is updated and " +
					"is responsible for reviewing the membership periodically.",
				}),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(GroupEdgeOwner).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					GroupFieldOwnerId: UserFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(GroupEdgeUsers).
				ManyToMany(UserSchemaName, GrpUsrRelSchemaName, "group").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(GroupEdgeRoles).
				ManyToMany(RoleSchemaName, RoleGroupAssignmentSchemaName, "receiver_group").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(GroupEdgeOwnRoles).
				Label(model.LangJson{model.LanguageCodeEnUs: "Owned roles"}).
				Existing(RoleSchemaName, RoleEdgeOwnerGroup),
		).
		EdgeFrom(
			dmodel.Edge(GroupEdgeBenefitGrantRequests).
				Label(model.LangJson{model.LanguageCodeEnUs: "Grant requests for this group"}).
				Existing(RoleRequestSchemaName, RoleReqEdgeReceiverGroup),
		)
}

func assertSchemasEqual(t *testing.T, expected *dmodel.ModelSchema, actual *dmodel.ModelSchema) {
	t.Helper()

	assert.Equal(t, expected.Name(), actual.Name())
	assert.Equal(t, expected.TableName(), actual.TableName())
	assert.Equal(t, expected.Label(), actual.Label())
	assert.Equal(t, expected.Description(), actual.Description())
	assert.Equal(t, expected.RecordLabelField(), actual.RecordLabelField())
	assert.Equal(t, expected.RecordSubLabelField(), actual.RecordSubLabelField())
	// Field ORDER, not just membership: it determines column order.
	assert.Equal(t, expected.FieldNames(), actual.FieldNames())
	assert.Equal(t, expected.CompositeUniques(), actual.CompositeUniques())
	assert.Equal(t, expected.PartialUniqueGroups(), actual.PartialUniqueGroups())
	assert.Equal(t, expected.SearchIndexGroups(), actual.SearchIndexGroups())
	assert.Equal(t, expected.PrimaryKeys(), actual.PrimaryKeys())
	assert.Equal(t, expected.AllUniques(), actual.AllUniques())

	for _, fieldName := range expected.FieldNames() {
		assertFieldsEqual(t, expected, actual, fieldName)
	}

	assertRelationsEqual(t, expected.ToRelations(), actual.ToRelations())
	assertRelationsEqual(t, expected.FromRelations(), actual.FromRelations())
}

func assertFieldsEqual(t *testing.T, expected *dmodel.ModelSchema, actual *dmodel.ModelSchema, fieldName string) {
	t.Helper()

	expectedField, ok := expected.Field(fieldName)
	require.True(t, ok)
	actualField, ok := actual.Field(fieldName)
	require.True(t, ok, "field '%s' missing", fieldName)

	assert.Equal(t, expectedField.Label(), actualField.Label(), "%s label", fieldName)
	assert.Equal(t, expectedField.Description(), actualField.Description(), "%s description", fieldName)
	assert.Equal(t, expectedField.DataType().String(), actualField.DataType().String(), "%s type", fieldName)
	assert.Equal(t, expectedField.DataType().IsArray(), actualField.DataType().IsArray(), "%s isArray", fieldName)
	// Compares the concretely-typed option slices, which catches JSON float64 leakage.
	assert.Equal(t, expectedField.DataType().Options(), actualField.DataType().Options(), "%s options", fieldName)

	assert.Equal(t, expectedField.IsRequiredForCreate(), actualField.IsRequiredForCreate(), "%s reqCreate", fieldName)
	assert.Equal(t, expectedField.IsRequiredForUpdate(), actualField.IsRequiredForUpdate(), "%s reqUpdate", fieldName)
	assert.Equal(t, expectedField.IsPrimaryKey(), actualField.IsPrimaryKey(), "%s primaryKey", fieldName)
	assert.Equal(t, expectedField.IsTenantKey(), actualField.IsTenantKey(), "%s tenantKey", fieldName)
	assert.Equal(t, expectedField.IsVersioningKey(), actualField.IsVersioningKey(), "%s versioningKey", fieldName)
	assert.Equal(t, expectedField.IsUnique(), actualField.IsUnique(), "%s unique", fieldName)
	assert.Equal(t, expectedField.IsAutoGenerated(), actualField.IsAutoGenerated(), "%s autoGenerated", fieldName)
	assert.Equal(t, expectedField.IsNoUpdate(), actualField.IsNoUpdate(), "%s noUpdate", fieldName)
	assert.Equal(t, expectedField.IsServiceInjected(), actualField.IsServiceInjected(), "%s injected", fieldName)
	assert.Equal(t, expectedField.IsNullable(), actualField.IsNullable(), "%s nullable", fieldName)
	assert.Equal(t, expectedField.ColumnType(), actualField.ColumnType(), "%s columnType", fieldName)
	assert.Equal(t, expectedField.Default(), actualField.Default(), "%s default", fieldName)
}

func assertRelationsEqual(t *testing.T, expected []dmodel.ModelRelation, actual []dmodel.ModelRelation) {
	t.Helper()

	require.Equal(t, len(expected), len(actual), "relation count")
	for i := range expected {
		assert.Equal(t, expected[i].Edge, actual[i].Edge)
		assert.Equal(t, expected[i].Label(), actual[i].Label())
		assert.Equal(t, expected[i].RelationType, actual[i].RelationType)
		assert.Equal(t, expected[i].DestSchemaName, actual[i].DestSchemaName)
		assert.Equal(t, expected[i].SrcField, actual[i].SrcField)
		assert.Equal(t, expected[i].DestField, actual[i].DestField)
		assert.Equal(t, expected[i].InversePeerSchemaName, actual[i].InversePeerSchemaName)
		assert.Equal(t, expected[i].InversePeerEdgeName, actual[i].InversePeerEdgeName)
		assert.Equal(t, expected[i].OnDelete, actual[i].OnDelete)
		assert.Equal(t, expected[i].OnUpdate, actual[i].OnUpdate)
		assert.Equal(t, expected[i].M2mThroughSchemaName, actual[i].M2mThroughSchemaName)
		assert.Equal(t, expected[i].M2mSrcFieldPrefix, actual[i].M2mSrcFieldPrefix)
		assert.Equal(t, expected[i].UnvalidatedFkMap, actual[i].UnvalidatedFkMap)
		assert.Equal(t, expected[i].ForeignKeys, actual[i].ForeignKeys)
	}
}
