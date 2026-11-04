package defaults

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

func TestLoad_ReadsSchemaFile(t *testing.T) {
	fsys := fstest.MapFS{
		"inventory_warehouse.json": {Data: []byte(`[{"$alias":"wh","code":"WH"}]`)},
	}

	rows, err := Load(fsys, "inventory_warehouse", model.LanguageCodeViVn)

	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "WH", rows[0]["code"])
}

func TestLoad_CollapsesStringFieldsToLanguage(t *testing.T) {
	const schemaName = "defaults_test_localized"
	require.NoError(t, dmodel.RegisterSchemaB(dmodel.DefineModel(schemaName).
		Field(dmodel.DefineField().Name("code").DataType(dmodel.FieldDataTypeString(1, 50))).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(1, 50))).
		Field(dmodel.DefineField().Name("label").DataType(dmodel.FieldDataTypeLangJson(1, 50)))))
	fsys := fstest.MapFS{
		schemaName + ".json": {Data: []byte(`[{"code":"Stock","name":{"en-US":"Stock","vi-VN":"Kho"},` +
			`"label":{"en-US":"Stock","vi-VN":"Kho"},"extra":{"en-US":"x"}},{"name":{"en-US":"Only"}}]`)},
	}

	rows, err := Load(fsys, schemaName, model.LanguageCodeViVn)

	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "Stock", rows[0]["code"])
	assert.Equal(t, "Kho", rows[0]["name"])
	assert.Equal(t, map[string]any{"en-US": "Stock", "vi-VN": "Kho"}, rows[0]["label"])
	assert.Equal(t, map[string]any{"en-US": "x"}, rows[0]["extra"])
	assert.Equal(t, "Only", rows[1]["name"])

	rows, err = Load(fsys, schemaName, model.LanguageCodeEnUs)
	require.NoError(t, err)
	assert.Equal(t, "Stock", rows[0]["name"])
}

func TestLoad_MissingOrInvalidFile(t *testing.T) {
	fsys := fstest.MapFS{"broken.json": {Data: []byte(`{`)}}

	_, err := Load(fsys, "absent", model.LanguageCodeViVn)
	assert.Error(t, err)
	_, err = Load(fsys, "broken", model.LanguageCodeViVn)
	assert.Error(t, err)
}

func TestCreateAll_ResolvesAliasesInOrder(t *testing.T) {
	rows := []dmodel.DynamicFields{
		{"$alias": "root", "code": "WH"},
		{"$alias": "stock", "code": "Stock", "parent_id": "@root"},
		{"code": "Scrap", "parent_id": "@root", "tags": []any{"@stock", "plain"}},
	}
	aliases := NewAliases()
	created := []dmodel.DynamicFields{}
	nextId := 0

	err := aliases.CreateAll(rows, "ORG-1", func(row dmodel.DynamicFields) (model.Id, error) {
		nextId++
		created = append(created, row)
		return model.Id([]string{"", "ID-1", "ID-2", "ID-3"}[nextId]), nil
	})

	require.NoError(t, err)
	require.Len(t, created, 3)
	assert.NotContains(t, created[0], AliasKey)
	assert.Equal(t, "ORG-1", created[0]["org_id"])
	assert.Equal(t, "ID-1", created[1]["parent_id"])
	assert.Equal(t, []any{"ID-2", "plain"}, created[2]["tags"])
	stockId, found := aliases.Id("stock")
	assert.True(t, found)
	assert.Equal(t, model.Id("ID-2"), stockId)
}

func TestCreateAll_UnknownAliasFails(t *testing.T) {
	rows := []dmodel.DynamicFields{{"parent_id": "@nowhere"}}

	err := NewAliases().CreateAll(rows, "ORG-1", func(dmodel.DynamicFields) (model.Id, error) {
		return "ID", nil
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "nowhere")
}
