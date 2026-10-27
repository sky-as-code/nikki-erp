package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func sharingView() *dmodel.ModelSchema {
	return dmodel.DefineModel("place_a_view").
		TableName("place_as").
		SharesTable().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("ref_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("ref").
			ManyToOne("refact_parent", dmodel.DynamicFields{"ref_id": "id"}).
			OnDelete(dmodel.RelationCascadeSetNull)).
		Build()
}

func viewReader() *dmodel.ModelSchema {
	return placementSchema("place_reader", "place_readers", "place_a_view", false)
}

func TestSharesTable_EmitsNoSql(t *testing.T) {
	owner := placementSchema("place_a", "place_as", "refact_parent", false)
	without := refActionRegistry(t, refActionParent(), owner)
	with := refActionRegistry(t, refActionParent(),
		placementSchema("place_a", "place_as", "refact_parent", false), sharingView())

	expected, err := GenCreateSql(without, DialectPostgres)
	require.NoError(t, err)
	actual, err := GenCreateSql(with, DialectPostgres)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestSharesTable_KeepsDbMetadata(t *testing.T) {
	view := sharingView()
	assert.True(t, view.SharesTable())
	assert.Equal(t, []string{"id"}, view.PrimaryKeys())
	assert.Equal(t, "place_as", view.TableName())
}

// The constraint to the shared table belongs to the table's owner schema, never to the view.
func TestSharesTable_NoForeignKeyTowardsView(t *testing.T) {
	reader := viewReader()
	registry := refActionRegistry(t, refActionParent(),
		placementSchema("place_a", "place_as", "refact_parent", false), sharingView(), reader)

	sqls, _, err := NewPgQueryBuilder().SqlCreateTable(reader, registry)
	require.NoError(t, err)
	assert.NotContains(t, sqls[0], "FOREIGN KEY")
}

func TestSharesTable_ParsedFromJson(t *testing.T) {
	schema := dmodel.ParseModelJson(`{
		"name": "place_json_view", "table_name": "place_as", "shares_table": true,
		"fields": [{ "name": "id", "data_type": "ulid", "primary_key": true }]
	}`).Build()
	assert.True(t, schema.SharesTable())
	assert.Equal(t, []string{"id"}, schema.PrimaryKeys())
}
