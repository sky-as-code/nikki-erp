package models

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// These tests pin the JSON-backed builders (ESS-01, ESS-02) to the fluent Go builders they
// replaced. The legacy chains below are verbatim copies of the pre-conversion builders and exist
// only as the reference to diff against: field order, types, options (regex included), defaults,
// nullability, uniques and edges must be identical, so the regenerated migration shows no diff.

func TestEssentialSchemaJsonMatchesLegacy(t *testing.T) {
	_ = basemodel.RegisterJsonBaseSchemas()
	assertEssentialSchemasMatchLegacy(t)
}

// assertEssentialSchemasMatchLegacy constructs both builders inside the call, so they pick up
// whatever base model is installed at that moment (see the multi-tenant variant).
func assertEssentialSchemasMatchLegacy(t *testing.T) {
	for name, pair := range map[string][2]*dmodel.ModelSchemaBuilder{
		LanguageSchemaName:       {languageSchemaBuilderLegacy(), LanguageSchemaBuilder()},
		EnumSchemaName:           {enumSchemaBuilderLegacy(), EnumSchemaBuilder()},
		TagSchemaName:            {tagSchemaBuilderLegacy(), TagSchemaBuilder()},
		FieldMetadataSchemaName:  {fieldMetadataSchemaBuilderLegacy(), FieldMetadataSchemaBuilder()},
		ModelMetadataSchemaName:  {modelMetadataSchemaBuilderLegacy(), ModelMetadataSchemaBuilder()},
		ModuleMetadataSchemaName: {moduleMetadataSchemaBuilderLegacy(), ModuleMetadataSchemaBuilder()},
	} {
		t.Run(name, func(t *testing.T) {
			assertSchemasEqual(t, pair[0].Build(), pair[1].Build())
		})
	}
}

// ---------------------------------------------------------------------------
// Legacy builders: verbatim copies of the pre-conversion Go chains.
// ---------------------------------------------------------------------------

func languageSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(LanguageSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", LanguageSchemaName)).
		TableName("essential_languages").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldName).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldIsoCode).
				// Hyphenated BCP47, not the underscore form: every consumer of an iso code speaks
				// hyphen -- the stored `language` setting, the LangJson document
				// keys, ToBCP47LanguageCode, i18next on the frontend. An underscore code could not
				// have been matched by any of them.
				DataType(dmodel.FieldDataTypeString(5, 10, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile(`^[a-z]{2}-[A-Z]{2}$`),
				})).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldDirection).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(LanguageDirectionLtr),
					string(LanguageDirectionRtl),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldDecimalSeparator).
				DataType(dmodel.FieldDataTypeString(1, 5)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldThousandsSeparator).
				DataType(dmodel.FieldDataTypeString(1, 5)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldDateFormat).
				DataType(dmodel.FieldDataTypeString(1, 30)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldTimeFormat).
				DataType(dmodel.FieldDataTypeString(1, 30)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldShortTimeFormat).
				DataType(dmodel.FieldDataTypeString(1, 30)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldFirstDayOfWeek).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(WeekDayMonday),
					string(WeekDayTuesday),
					string(WeekDayWednesday),
					string(WeekDayThursday),
					string(WeekDayFriday),
					string(WeekDaySaturday),
					string(WeekDaySunday),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(LanguageFieldIsEnabled).
				// The set of languages a user may pick. False by default: a language created before its
				// translations ship would render every key as its raw namespace:key to whoever chose it.
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

func enumSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(EnumSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", EnumSchemaName)).
		TableName("essential_enums").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(EnumFieldLabel).
				Label(model.NewLangJsonRefSf("fields.%s", EnumFieldLabel)).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(EnumFieldValue).
				Label(model.NewLangJsonRefSf("fields.%s", EnumFieldValue)).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(EnumFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", EnumFieldType)).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

func tagSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TagSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", TagSchemaName)).
		TableName("essential_tags").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(TagFieldLabel).
				Label(model.NewLangJsonRefSf("fields.%s", TagFieldLabel)).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(TagFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", TagFieldType)).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

func fieldMetadataSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(FieldMetadataSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", FieldMetadataSchemaName)).
		TableName("essential_field_metadata").
		ShouldBuildDb().
		Field(
			basemodel.DefineFieldId(FieldMetadataFieldId).
				Label(model.NewLangJsonRefSf("fields.%s", FieldMetadataFieldId)).
				UseTypeDefault().
				PrimaryKey(),
		).
		Field(
			basemodel.DefineFieldId(FieldMetadataFieldModelMetadataId).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldName).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldDataType).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldIsRequired).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(FieldMetadataFieldDisplayOrder).
				DataType(dmodel.FieldDataTypeInt32(0, 10000)).
				Default(int32(0)).
				RequiredForCreate(),
		).
		CompositeUnique(dmodel.CompositeUniqueParam{
			Fields: []string{FieldMetadataFieldModelMetadataId, FieldMetadataFieldCode},
		}).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(FieldMetadataEdgeModelMetadata).
				ManyToOne(ModelMetadataSchemaName, dmodel.DynamicFields{
					FieldMetadataFieldModelMetadataId: ModelMetadataFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

func modelMetadataSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ModelMetadataSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", ModelMetadataSchemaName)).
		TableName("essential_model_metadata").
		ShouldBuildDb().
		Field(
			basemodel.DefineFieldId(ModelMetadataFieldId).
				Label(model.NewLangJsonRefSf("fields.%s", ModelMetadataFieldId)).
				UseTypeDefault().
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldName).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldCodePrefix).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)).
				Default(""),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldCodeLastSeq).
				DataType(dmodel.FieldDataTypeInt32(0, 1000000000)).
				Default(int32(0)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModelMetadataFieldPadding).
				DataType(dmodel.FieldDataTypeInt32(1, 20)).
				Default(int32(8)).
				RequiredForCreate(),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(ModelMetadataEdgeFieldMetadataList).
				Label(model.LangJson{model.LanguageCodeEnUs: "Field metadata list"}).
				Existing(FieldMetadataSchemaName, FieldMetadataEdgeModelMetadata),
		)
}

func moduleMetadataSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ModuleMetadataSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", ModuleMetadataSchemaName)).
		TableName("essential_modules").
		CompositeUnique(dmodel.CompositeUniqueParam{Fields: []string{ModuleMetadataFieldName}}).
		ShouldBuildDb().
		Field(
			basemodel.DefineFieldId(ModuleMetadataFieldId).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldId)).
				UseTypeDefault().
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldName).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldName)).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldLabel).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldLabel)).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldIsOrphaned).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldIsOrphaned)).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				Default(false),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldIsInternal).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldIsInternal)).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				NoUpdate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldVersion).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldVersion)).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldHasUi).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldHasUi)).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				Default(false),
		).
		Field(
			// Nullable so rows written before this field existed migrate; startup sync fills it.
			dmodel.DefineField().
				Name(ModuleMetadataFieldModelPrefixes).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldModelPrefixes)).
				DataType(dmodel.FieldDataTypeString(1, 50).ArrayType()),
		).
		Field(
			// Nullable for the same reason as model_prefixes.
			dmodel.DefineField().
				Name(ModuleMetadataFieldCategory).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldCategory)).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(modules.ModuleCategoryFoundation),
					string(modules.ModuleCategoryCoremart),
					string(modules.ModuleCategoryCoreparking),
					string(modules.ModuleCategoryItInfra),
				})),
		).
		Field(
			// Always true for now: nothing disables a module yet, so there is no toggle and no sync diff.
			dmodel.DefineField().
				Name(ModuleMetadataFieldIsEnabled).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldIsEnabled)).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				NoUpdate().
				Default(true),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldDeps).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldDeps)).
				DataType(dmodel.FieldDataTypeString(1, 50).ArrayType()),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldDepLevel).
				Label(model.NewLangJsonRefSf("fields.%s", ModuleMetadataFieldDepLevel)).
				DataType(dmodel.FieldDataTypeInt32(0, 999)).
				RequiredForCreate().
				Default("0"),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

// ---------------------------------------------------------------------------
// Comparison helpers, copied from contacts/domain/models/schema_equivalence_test.go.
// ---------------------------------------------------------------------------

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
	// A JSON number decodes as float64 while the Go chain declared int32; the value goes through
	// the field's data type Validate either way, so only the rendered value must match.
	assert.Equal(t, renderDefault(expectedField), renderDefault(actualField), "%s default", fieldName)
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

func renderDefault(field *dmodel.ModelField) string {
	if field.Default() == nil || field.Default().Get() == nil {
		return "<none>"
	}
	return fmt.Sprint(*field.Default().Get())
}
