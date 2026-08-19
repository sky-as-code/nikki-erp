package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// These tests pin the JSON-backed builders to the Go builders they replaced. The legacy chains
// below are verbatim copies of what domain/party_entity.go, comm_channel_entity.go and
// relationship_entity.go contained before the conversion, and exist only as the reference to diff
// against. They are what makes deleting those files safe.
//
// Two differences are deliberate and are asserted separately rather than hidden in the copies:
//
//   - tax_id and website were GLOBALLY unique and now carry no unique constraint at all, only an
//     org-scoped index. That was a multi-tenancy bug — two organizations legitimately record the
//     same supplier — so the legacy chains keep the old Unique() and the equality tests below
//     exclude the unique sets, which are then checked explicitly with the new shape. See
//     TestPartyHasNoUniqueConstraints for why "unique per organization" was not the answer.
//   - The parties table gained an org_id search index it never had, despite every query being
//     org-scoped.
//
// Everything else — field order, types, defaults, nullability, edges, cascade behaviour — must be
// identical, and the tests fail if it is not.

func TestPartySchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEquivalent(t, partySchemaBuilderLegacy().Build(), PartySchemaBuilder().Build())
}

func TestCommChannelSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, commChannelSchemaBuilderLegacy().Build(), CommChannelSchemaBuilder().Build())
}

func TestRelationshipSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, relationshipSchemaBuilderLegacy().Build(), RelationshipSchemaBuilder().Build())
}

// The deliberate half of the party change: tax_id and website are no longer unique at all.
//
// They used to carry a GLOBAL unique constraint, which was a multi-tenancy bug — the first
// organization to record a supplier locked every other organization in the deployment out of
// recording its own. That much is not in question.
//
// What replaced it is an index without a unique constraint, and that needs explaining, because
// "unique per organization" was the obvious fix and is not what is here. Neither construct the
// framework offers can express it:
//
//   - CompositeUnique requires every field to be requiredForCreate. Both columns are optional, so
//     Build() panics with "use PartialUnique() instead".
//   - PartialUnique emits a PAIR of indexes, and the second is "UNIQUE (org_id) WHERE col IS NULL".
//     That permits exactly ONE party per organization with no tax id — and most contacts, every
//     individual and every walk-in customer, have none. Verified against Postgres: the second
//     untaxed party in an organization is rejected with a duplicate-key error naming org_id.
//
// So the constraint is dropped and the index kept. The index is what the org-scoped queries and any
// duplicate check need; enforcing "no two suppliers in this org share a tax id" belongs in the
// application layer, where it can warn a user rather than refuse a write — which is the better
// behaviour anyway, since a shared tax id across branches of one group is legitimate.
func TestPartyHasNoUniqueConstraints(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := PartySchemaBuilder().Build()

	assert.Empty(t, schema.PartialUniques(),
		"a partial unique would also forbid a second party with no tax_id in the same org")
	assert.Empty(t, schema.CompositeUniques(),
		"tax_id and website are optional, so they cannot be composite-unique")

	taxId, ok := schema.Field(PartyFieldTaxId)
	require.True(t, ok)
	assert.False(t, taxId.IsUnique(), "tax_id must no longer be globally unique")

	website, ok := schema.Field(PartyFieldWebsite)
	require.True(t, ok)
	assert.False(t, website.IsUnique(), "website must no longer be globally unique")
}

// The lookups that replaced the dropped constraints. Both are org-scoped, so org_id leads.
func TestPartyIndexesSupportOrgScopedLookup(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := PartySchemaBuilder().Build()

	assertHasSearchIndexOn(t, schema, PartyFieldTaxId)
	assertHasSearchIndexOn(t, schema, PartyFieldWebsite)
}

// Every party query is scoped by organization, so the column needs an index. It had none.
func TestPartyHasOrgIdSearchIndex(t *testing.T) {
	requireBaseSchemasRegistered(t)

	assertHasSearchIndexOn(t, PartySchemaBuilder().Build(), PartyFieldOrgId)
}

// A channel is almost always fetched by its party — that is the query the removed nested route
// used to serve, and it is now a filter instead.
func TestCommChannelHasPartyIdSearchIndex(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := CommChannelSchemaBuilder().Build()

	assertHasSearchIndexOn(t, schema, CommChannelFieldPartyId)
	assertHasSearchIndexOn(t, schema, CommChannelFieldOrgId)
}

func TestRelationshipHasBothPartySearchIndexes(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := RelationshipSchemaBuilder().Build()

	assertHasSearchIndexOn(t, schema, RelationshipFieldPartyId)
	assertHasSearchIndexOn(t, schema, RelationshipFieldTargetPartyId)
}

// The asymmetry that shapes every relationship query, pinned so it cannot be "tidied up" silently.
//
// A relationship joins two parties that may belong to different organizations, so there is no single
// organization to scope it to. Adding org_id here is a decision with a migration behind it.
func TestRelationshipHasNoOrgId(t *testing.T) {
	requireBaseSchemasRegistered(t)

	_, ok := RelationshipSchemaBuilder().Build().Field(basemodel.FieldOrgId)

	assert.False(t, ok, "contacts_relationship must not gain an org_id without a deliberate migration")
}

// Deleting a party takes its channels and its relationships with it. This is the opposite of
// Inventory's NO ACTION, and the API tests assert the resulting behaviour, so the cascade must not
// drift.
func TestPartyChildEdgesCascadeOnDelete(t *testing.T) {
	requireBaseSchemasRegistered(t)

	for _, relation := range CommChannelSchemaBuilder().Build().ToRelations() {
		assert.Equal(t, dmodel.RelationCascadeCascade, relation.OnDelete,
			"comm channel edge %q must cascade", relation.Edge)
	}
	for _, relation := range RelationshipSchemaBuilder().Build().ToRelations() {
		assert.Equal(t, dmodel.RelationCascadeCascade, relation.OnDelete,
			"relationship edge %q must cascade", relation.Edge)
	}
}

func requireBaseSchemasRegistered(t *testing.T) {
	t.Helper()
	// Normally done by CoreModule.RegisterModels during app start-up.
	_ = basemodel.RegisterJsonBaseSchemas()
}

func assertHasSearchIndexOn(t *testing.T, schema *dmodel.ModelSchema, field string) {
	t.Helper()

	for _, group := range schema.SearchIndexGroups() {
		for _, indexed := range group.Fields {
			if indexed == field {
				return
			}
		}
	}
	t.Fatalf("schema %q has no search index covering %q", schema.Name(), field)
}

// ---------------------------------------------------------------------------
// Legacy builders: verbatim copies of the pre-conversion Go chains.
// ---------------------------------------------------------------------------

func partySchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PartySchemaName).
		Label(model.NewLangJsonRefSf("%s.label", PartySchemaName)).
		TableName("contacts_parties").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.OrgIdModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(PartyFieldAvatarUrl).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldAvatarUrl)).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldDisplayName).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldDisplayName)).
				DataType(dmodel.FieldDataTypeString(1, 50)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldLegalName).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldLegalName)).
				DataType(dmodel.FieldDataTypeString(0, 100)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldLegalAddress).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldLegalAddress)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldTaxId).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldTaxId)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_SHORT_NAME_LENGTH)).
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldJobPosition).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldJobPosition)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_SHORT_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldTitle).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldTitle)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldType)).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(PartyTypeIndividual),
					string(PartyTypeCompany),
				})).
				Default(string(PartyTypeIndividual)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldNote).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldNote)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(PartyFieldNationalityId).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldNationalityId)),
		).
		Field(
			basemodel.DefineFieldId(PartyFieldLanguageId).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldLanguageId)),
		).
		Field(
			dmodel.DefineField().
				Name(PartyFieldWebsite).
				Label(model.NewLangJsonRefSf("fields.%s", PartyFieldWebsite)).
				DataType(dmodel.FieldDataTypeUrl()).
				Unique(),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(PartyEdgeCommChannels).
				Label(model.LangJson{model.LanguageCodeEnUs: "Communication Channels"}).
				Existing(CommChannelSchemaName, CommChannelEdgeParty),
		).
		EdgeFrom(
			dmodel.Edge(PartyEdgeRelationshipsAsSource).
				Label(model.LangJson{model.LanguageCodeEnUs: "Relationships (as source)"}).
				Existing(RelationshipSchemaName, RelationshipEdgeSourceParty),
		).
		EdgeFrom(
			dmodel.Edge(PartyEdgeRelationshipsAsTarget).
				Label(model.LangJson{model.LanguageCodeEnUs: "Relationships (as target)"}).
				Existing(RelationshipSchemaName, RelationshipEdgeTargetParty),
		)
}

func commChannelSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(CommChannelSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", CommChannelSchemaName)).
		TableName("contacts_comm_channels").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.OrgIdModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(CommChannelFieldNote).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldNote)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(CommChannelFieldPartyId).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldPartyId)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(CommChannelFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldType)).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(CommChannelTypePhone),
					string(CommChannelTypeZalo),
					string(CommChannelTypeFacebook),
					string(CommChannelTypeEmail),
					string(CommChannelTypePost),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(CommChannelFieldValue).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldValue)).
				DataType(dmodel.FieldDataTypeString(0, 255)),
		).
		Field(
			dmodel.DefineField().
				Name(CommChannelFieldValueJson).
				Label(model.NewLangJsonRefSf("fields.%s", CommChannelFieldValueJson)).
				DataType(dmodel.FieldDataTypeJsonMap()),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(CommChannelEdgeParty).
				Label(model.LangJson{model.LanguageCodeEnUs: "Party"}).
				ManyToOne(PartySchemaName, dmodel.DynamicFields{
					CommChannelFieldPartyId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

func relationshipSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RelationshipSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", RelationshipSchemaName)).
		TableName("contacts_relationships").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(RelationshipFieldPartyId).
				Label(model.NewLangJsonRefSf("fields.%s", RelationshipFieldPartyId)).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(RelationshipFieldTargetPartyId).
				Label(model.NewLangJsonRefSf("fields.%s", RelationshipFieldTargetPartyId)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(RelationshipFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", RelationshipFieldType)).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(RelationshipTypeEmployee),
					string(RelationshipTypeSpouse),
					string(RelationshipTypeParent),
					string(RelationshipTypeSibling),
					string(RelationshipTypeEmergency),
					string(RelationshipTypeSubsidiary),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(RelationshipFieldNote).
				Label(model.NewLangJsonRefSf("fields.%s", RelationshipFieldNote)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(RelationshipEdgeSourceParty).
				Label(model.LangJson{model.LanguageCodeEnUs: "Source Party"}).
				ManyToOne(PartySchemaName, dmodel.DynamicFields{
					RelationshipFieldPartyId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(RelationshipEdgeTargetParty).
				Label(model.LangJson{model.LanguageCodeEnUs: "Target Party"}).
				ManyToOne(PartySchemaName, dmodel.DynamicFields{
					RelationshipFieldTargetPartyId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

// ---------------------------------------------------------------------------
// Comparison helpers, copied from iam/domain/models/schema_equivalence_test.go.
// ---------------------------------------------------------------------------

// assertSchemasEquivalent is assertSchemasEqual minus the unique and search-index sets, for the
// one schema where those were changed on purpose. They are asserted separately above.
func assertSchemasEquivalent(t *testing.T, expected *dmodel.ModelSchema, actual *dmodel.ModelSchema) {
	t.Helper()

	assertSchemaShapeEqual(t, expected, actual)
	assert.Equal(t, expected.PrimaryKeys(), actual.PrimaryKeys())
}

func assertSchemasEqual(t *testing.T, expected *dmodel.ModelSchema, actual *dmodel.ModelSchema) {
	t.Helper()

	assertSchemaShapeEqual(t, expected, actual)
	assert.Equal(t, expected.CompositeUniques(), actual.CompositeUniques())
	assert.Equal(t, expected.PartialUniques(), actual.PartialUniques())
	assert.Equal(t, expected.PrimaryKeys(), actual.PrimaryKeys())
	assert.Equal(t, expected.AllUniques(), actual.AllUniques())
}

func assertSchemaShapeEqual(t *testing.T, expected *dmodel.ModelSchema, actual *dmodel.ModelSchema) {
	t.Helper()

	assert.Equal(t, expected.Name(), actual.Name())
	assert.Equal(t, expected.TableName(), actual.TableName())
	assert.Equal(t, expected.Label(), actual.Label())
	assert.Equal(t, expected.RecordSubLabelField(), actual.RecordSubLabelField())
	// Field ORDER, not just membership: it determines column order.
	assert.Equal(t, expected.FieldNames(), actual.FieldNames())

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

	assert.Equal(t, expectedField.DataType().String(), actualField.DataType().String(), "%s type", fieldName)
	assert.Equal(t, expectedField.DataType().IsArray(), actualField.DataType().IsArray(), "%s isArray", fieldName)
	// Compares the concretely-typed option slices, which catches JSON float64 leakage.
	assert.Equal(t, expectedField.DataType().Options(), actualField.DataType().Options(), "%s options", fieldName)

	assert.Equal(t, expectedField.IsRequiredForCreate(), actualField.IsRequiredForCreate(), "%s reqCreate", fieldName)
	assert.Equal(t, expectedField.IsRequiredForUpdate(), actualField.IsRequiredForUpdate(), "%s reqUpdate", fieldName)
	assert.Equal(t, expectedField.IsPrimaryKey(), actualField.IsPrimaryKey(), "%s primaryKey", fieldName)
	assert.Equal(t, expectedField.IsTenantKey(), actualField.IsTenantKey(), "%s tenantKey", fieldName)
	assert.Equal(t, expectedField.IsVersioningKey(), actualField.IsVersioningKey(), "%s versioningKey", fieldName)
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
