package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	basemodel "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	m.Run()
}

func probeSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("probe_user_settings").
		Field(
			dmodel.DefineField().
				Name("theme_mode").
				Label(model.NewLangJsonRef("probe.themeMode")).
				DataType(dmodel.FieldDataTypeEnumString([]string{"light", "dark", "auto"})).
				Default("auto").
				Metadata(map[string]any{"allow_override": true}),
		).
		Field(
			dmodel.DefineField().
				Name("language").
				Label(model.NewLangJsonRef("probe.language")).
				DataType(dmodel.FieldDataTypeEnumString([]string{"en-US", "vi-VN"})),
		).
		Build()
}

// The schema column is jsonmap, whose type check accepts an object and rejects a Go struct.
// ToSimplized returns an anonymous struct, so handing it to the repository directly is refused —
// and the refusal arrives as a ClientError rather than a Go error, which is how it went unnoticed
// while registration kept reporting success.
func TestSimplizedDocument_IsAPlainJsonObject(t *testing.T) {
	document, err := simplizedDocument(probeSchema())
	require.NoError(t, err)
	require.NotNil(t, document)

	assert.Equal(t, "probe_user_settings", document["name"])
	_, hasFields := document["fields"]
	assert.True(t, hasFields, "the stored document must carry its fields")
}

// The cross-level uniqueness check reads names out of the stored document. ToSimplized keys fields
// by name instead of emitting a list, so reading it as an array yields an empty set: the check
// would pass vacuously and let two levels of one module declare the same setting name, which
// collides on write because the record unique key carries no level.
func TestFieldNameSetOfDocument_ReadsEveryDeclaredName(t *testing.T) {
	document, err := simplizedDocument(probeSchema())
	require.NoError(t, err)

	names := fieldNameSetOfDocument(document)

	assert.Contains(t, names, "theme_mode")
	assert.Contains(t, names, "language")
	assert.Len(t, names, 2)
}

// The read path rebuilds a schema from the stored document on every settings read. If the stored
// shape and the parser's expected shape disagree, every read of a registered module fails.
func TestSchemaFromDocument_RoundTripsAStoredSchema(t *testing.T) {
	document, err := simplizedDocument(probeSchema())
	require.NoError(t, err)

	rebuilt, err := schemaFromDocument(document)
	require.NoError(t, err, "a document written by simplizedDocument must be readable again")
	require.NotNil(t, rebuilt)

	for _, name := range []string{"theme_mode", "language"} {
		field, ok := rebuilt.Field(name)
		require.True(t, ok, "field %s survives the round trip", name)
		assert.Equal(t, name, field.Name())
	}

	// allow_override drives whether a tenant may lock a setting, so it must survive storage.
	themeMode, ok := rebuilt.Field("theme_mode")
	require.True(t, ok)
	allow, found := themeMode.MetadataValue("allow_override")
	assert.True(t, found, "allow_override must survive the round trip")
	assert.Equal(t, true, allow)
}

// A registered schema must remain byte-identical across boots, or every restart rewrites the row
// and the "identical registration is a no-op" contract is lost.
func TestSameSchemaDocument_MatchesItselfAcrossARoundTrip(t *testing.T) {
	document, err := simplizedDocument(probeSchema())
	require.NoError(t, err)

	again, err := simplizedDocument(probeSchema())
	require.NoError(t, err)

	assert.True(t, sameSchemaDocument(document, again),
		"an unchanged declaration must compare equal, or every boot rewrites the row")
}

// The real declarations, not a hand-made probe. Registration stores a document and every settings
// read parses it back, so a data type that renders but cannot be re-read breaks that module's
// settings pane while registration still reports success — which is exactly how the enum case
// went unnoticed. int32, boolean and bounded string are all newly used by these declarations.
func TestSchemaFromDocument_RoundTripsEveryDeclaredDataType(t *testing.T) {
	cases := map[string]*dmodel.ModelSchema{
		"enum + bounded string": dmodel.DefineModel("probe_mixed").
			Field(dmodel.DefineField().Name("locale").
				DataType(dmodel.FieldDataTypeEnumString([]string{"en-US", "vi-VN"}))).
			Field(dmodel.DefineField().Name("zone").
				DataType(dmodel.FieldDataTypeString(1, 64))).
			Build(),

		"int32 + boolean": dmodel.DefineModel("probe_policy").
			Field(dmodel.DefineField().Name("session_timeout_minutes").
				DataType(dmodel.FieldDataTypeInt32(1, 10080)).
				Metadata(map[string]any{"allow_override": false})).
			Field(dmodel.DefineField().Name("require_mfa").
				DataType(dmodel.FieldDataTypeBoolean()).
				Metadata(map[string]any{"allow_override": false})).
			Build(),
	}

	for name, schema := range cases {
		t.Run(name, func(t *testing.T) {
			document, err := simplizedDocument(schema)
			require.NoError(t, err)

			rebuilt, err := schemaFromDocument(document)
			require.NoError(t, err, "a stored document must be readable again")

			assert.ElementsMatch(t, schema.FieldNames(), rebuilt.FieldNames())

			for _, fieldName := range schema.FieldNames() {
				original, ok := schema.Field(fieldName)
				require.True(t, ok)
				restored, ok := rebuilt.Field(fieldName)
				require.True(t, ok, "field %s survives the round trip", fieldName)

				assert.Equal(t, original.DataType().String(), restored.DataType().String(),
					"field %s keeps its data type", fieldName)

				// allow_override decides whether a tenant may lock a setting, so losing it in
				// storage would silently change who is allowed to edit what.
				want, hadMetadata := original.MetadataValue("allow_override")
				if hadMetadata {
					got, found := restored.MetadataValue("allow_override")
					assert.True(t, found, "field %s keeps allow_override", fieldName)
					assert.Equal(t, want, got, fieldName)
				}
			}
		})
	}
}

// A bounded string must keep its bounds. Losing them would leave the rebuilt field accepting
// values the original rejected, so validation would pass on a read path and fail on a write.
func TestSchemaFromDocument_KeepsStringBounds(t *testing.T) {
	schema := dmodel.DefineModel("probe_bounds").
		Field(dmodel.DefineField().Name("zone").
			DataType(dmodel.FieldDataTypeString(1, 64))).
		Build()

	document, err := simplizedDocument(schema)
	require.NoError(t, err)
	rebuilt, err := schemaFromDocument(document)
	require.NoError(t, err)

	field, ok := rebuilt.Field("zone")
	require.True(t, ok)

	_, vErr := field.Validate("Asia/Ho_Chi_Minh", true)
	assert.Nil(t, vErr, "a valid zone name must pass")

	_, vErr = field.Validate(strings.Repeat("x", 65), true)
	assert.NotNil(t, vErr, "a value past the declared maximum must be refused")
}
