package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// schemaWithTenantKey builds a minimal DDL-emitting schema whose columns cover everything the
// index-naming tests need. tenantKey mirrors what coremart's base model contributes: when set,
// the query builder prepends that column to every unique key, which is what pushes generated
// names past PostgreSQL's 63-byte identifier limit.
func schemaWithTenantKey(tableName string, tenantKey bool) *dmodel.ModelSchemaBuilder {
	builder := dmodel.DefineModel("test_" + tableName).
		TableName(tableName).
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").
			DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
		Field(dmodel.DefineField().Name("code").
			DataType(dmodel.FieldDataTypeString(1, 100)).RequiredForCreate()).
		Field(dmodel.DefineField().Name("name").
			DataType(dmodel.FieldDataTypeString(1, 100)).RequiredForCreate()).
		Field(dmodel.DefineField().Name("org_id").
			DataType(dmodel.FieldDataTypeUlid()))
	if tenantKey {
		builder.Field(dmodel.DefineField().Name("tenant_id").
			DataType(dmodel.FieldDataTypeUlid()).UseTypeDefault().TenantKey())
	}
	return builder
}

func createTableSqls(t *testing.T, builder *dmodel.ModelSchemaBuilder) ([]string, error) {
	t.Helper()
	// These schemas declare no relations, so the registry is only along for the ride.
	schema := builder.Build()
	sqls, _, err := NewPgQueryBuilder().SqlCreateTable(schema, dmodel.GetSchemaRegistry())
	return sqls, err
}

func TestCompositeUnique_DerivesNameFromTableAndColumns(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
		CompositeUnique(dmodel.CompositeUniqueParam{Fields: []string{"code", "name"}}))
	require.NoError(t, err)
	assert.Contains(t, sqls[0], `CONSTRAINT "widgets_code_name_ukey" UNIQUE`)
}

func TestCompositeUnique_PrependsTenantKeyToDerivedName(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", true).
		CompositeUnique(dmodel.CompositeUniqueParam{Fields: []string{"code", "name"}}))
	require.NoError(t, err)
	assert.Contains(t, sqls[0], `CONSTRAINT "widgets_tenant_id_code_name_ukey" UNIQUE`)
}

// An explicit IndexName replaces the whole derived stem, tenant prefix included; only the
// "_ukey" suffix is still appended.
func TestCompositeUnique_HonoursExplicitIndexName(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", true).
		CompositeUnique(dmodel.CompositeUniqueParam{
			IndexName: "widgets_tid_code_name",
			Fields:    []string{"code", "name"},
		}))
	require.NoError(t, err)
	assert.Contains(t, sqls[0], `CONSTRAINT "widgets_tid_code_name_ukey" UNIQUE`)
	assert.NotContains(t, sqls[0], "widgets_tenant_id_code_name_ukey")
}

func TestPartialUnique_HonoursExplicitIndexName(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", true).
		PartialUniqueLoose(dmodel.PartialUniqueParam{
			IndexName:     "widgets_tid_name_org",
			NotNullFields: []string{"name"},
			NullableField: "org_id",
		}))
	require.NoError(t, err)
	joined := strings.Join(sqls, "\n")
	assert.Contains(t, joined, `CREATE UNIQUE INDEX "widgets_tid_name_org_ukey_notnull"`)
	assert.Contains(t, joined, `CREATE UNIQUE INDEX "widgets_tid_name_org_ukey_null"`)
}

// Unique index names must not pick up the "_idx" suffix reserved for non-unique search
// indexes; they already carry a "_ukey*" suffix.
func TestUniqueIndexNames_DoNotGainIdxSuffix(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
		CompositeUnique(dmodel.CompositeUniqueParam{
			IndexName: "widgets_code",
			Fields:    []string{"code"},
		}).
		PartialUniqueLoose(dmodel.PartialUniqueParam{
			IndexName:     "widgets_name_org",
			NotNullFields: []string{"name"},
			NullableField: "org_id",
		}))
	require.NoError(t, err)
	assert.NotContains(t, strings.Join(sqls, "\n"), "_idx_ukey")
}

// Search indexes keep the "_idx" convention.
func TestSearchIndex_KeepsIdxSuffix(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
		SearchIndexGroup(dmodel.SearchIndexGroupParam{
			IndexName: "widgets_code",
			Fields:    []string{"code"},
		}))
	require.NoError(t, err)
	assert.Contains(t, strings.Join(sqls, "\n"), `CREATE INDEX "widgets_code_idx"`)
}

func TestIndexName_OverLimitIsRejected(t *testing.T) {
	longName := strings.Repeat("a", PgMaxIdentifierLen) // + "_ukey" overflows

	t.Run("composite unique", func(t *testing.T) {
		_, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
			CompositeUnique(dmodel.CompositeUniqueParam{
				IndexName: longName,
				Fields:    []string{"code"},
			}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "identifier limit")
	})

	t.Run("partial unique", func(t *testing.T) {
		_, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
			PartialUniqueLoose(dmodel.PartialUniqueParam{
				IndexName:     longName,
				NotNullFields: []string{"name"},
				NullableField: "org_id",
			}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "identifier limit")
	})

	t.Run("search index", func(t *testing.T) {
		_, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
			SearchIndexGroup(dmodel.SearchIndexGroupParam{
				IndexName: longName,
				Fields:    []string{"code"},
			}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "identifier limit")
	})
}

// A name derived from a long table plus many columns overflows just as an explicit one does.
func TestDerivedIndexName_OverLimitIsRejected(t *testing.T) {
	_, err := createTableSqls(t, schemaWithTenantKey(strings.Repeat("w", 55), true).
		CompositeUnique(dmodel.CompositeUniqueParam{Fields: []string{"code", "name"}}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "identifier limit")
}
