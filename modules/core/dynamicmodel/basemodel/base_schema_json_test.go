package basemodel

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Each JSON twin must produce exactly the schema its Go builder produces, so that swapping
// one for the other is invisible to the rest of the system.
func TestJsonBaseSchemasMatchGoBuilders(t *testing.T) {
	testCases := []struct {
		name     string
		fromGo   func() *dmodel.ModelSchemaBuilder
		fromJson func() *dmodel.ModelSchemaBuilder
	}{
		{"base_model", BaseModelSchemaBuilder, BaseModelSchemaBuilderJson},
		{"org_base_model", OrgIdModelSchemaBuilder, OrgIdModelSchemaBuilderJson},
		{"archivable_model", ArchivableModelSchemaBuilder, ArchivableModelSchemaBuilderJson},
		{"auditable_model", AuditableModelSchemaBuilder, AuditableModelSchemaBuilderJson},
		{
			"auditable_readonly_model",
			AuditableReadonlyModelSchemaBuilder,
			AuditableReadonlyModelSchemaBuilderJson,
		},
		{"traceable_model", TraceableModelSchemaBuilder, TraceableModelSchemaBuilderJson},
		{
			"traceable_readonly_model",
			TraceableReadonlyModelSchemaBuilder,
			TraceableReadonlyModelSchemaBuilderJson,
		},
		{"versioned_model", VersionedModelSchemaBuilder, VersionedModelSchemaBuilderJson},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertSchemasEqual(t, testCase.fromGo().Build(), testCase.fromJson().Build())
		})
	}
}

// Registration is intended to run exactly once per process; registerModelInOrder is what
// guarantees that. Calling it again is a real error, not something to swallow here.
func TestRegisterJsonBaseSchemasRejectsDuplicate(t *testing.T) {
	require.NoError(t, registerBaseSchemasOnceForTest(t))
	assert.Error(t, RegisterJsonBaseSchemas(), "re-registering must surface as an error")
}

func TestRegisterJsonBaseSchemas(t *testing.T) {
	require.NoError(t, registerBaseSchemasOnceForTest(t))

	names := []string{
		BaseModelSchemaName, OrgBaseModelSchemaName, ArchivableModelSchemaName,
		AuditableModelSchemaName, AuditableReadonlyModelSchemaName, TraceableModelSchemaName,
		TraceableReadonlyModelSchemaName, VersionedModelSchemaName,
	}
	for _, name := range names {
		assert.NotNil(t, dmodel.GetSchemaBuilder(name), "builder '%s' must be registered", name)
	}
}

// The registry is process-wide while tests share one binary, so registration can only
// happen on the first call. Later callers reuse that result.
var (
	registerForTestOnce sync.Once
	registerForTestErr  error
)

func registerBaseSchemasOnceForTest(t *testing.T) error {
	t.Helper()
	registerForTestOnce.Do(func() {
		registerForTestErr = RegisterJsonBaseSchemas()
	})

	return registerForTestErr
}

func TestTraceableJsonKeepsServiceInjection(t *testing.T) {
	schema := TraceableModelSchemaBuilderJson().Build()

	for _, fieldName := range []string{FieldCreatedBy, FieldUpdatedBy} {
		field, ok := schema.Field(fieldName)
		require.True(t, ok)
		assert.True(t, field.IsServiceInjected(), "%s must stay service-injected", fieldName)
		assert.NotNil(t, field.InjectFn())
	}
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
	assert.Equal(t, expected.PartialUniques(), actual.PartialUniques())
	assert.Equal(t, expected.SearchIndexGroups(), actual.SearchIndexGroups())

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
	assert.Equal(t, expectedField.Default(), actualField.Default(), "%s default", fieldName)
}

func assertRelationsEqual(t *testing.T, expected []dmodel.ModelRelation, actual []dmodel.ModelRelation) {
	t.Helper()

	require.Equal(t, len(expected), len(actual), "relation count")
	for i := range expected {
		// ModelRelation is comparable as a whole except for M2mThroughModel, which is a
		// pointer resolved later by FinalizeRelations.
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
