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

	rows, err := Load(fsys, "inventory_warehouse")

	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "WH", rows[0]["code"])
}

func TestLoad_MissingOrInvalidFile(t *testing.T) {
	fsys := fstest.MapFS{"broken.json": {Data: []byte(`{`)}}

	_, err := Load(fsys, "absent")
	assert.Error(t, err)
	_, err = Load(fsys, "broken")
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
