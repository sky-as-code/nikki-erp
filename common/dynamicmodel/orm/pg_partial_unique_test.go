package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The difference between the two partial-unique kinds, stated as SQL.
//
// Loose emits two indexes, so it also constrains the rows whose nullable column is NULL: at most
// one per combination of the not-null columns. That is right when the nullable column is the
// SCOPE — "this name is unique per organization, and unique among the rows belonging to none".
//
// Strict emits only the IS NOT NULL index, leaving NULL rows alone. That is right when the
// nullable column is the VALUE — an optional external reference or display code that must be
// unique when present, but whose absence says nothing about the row.
//
// Picking loose for a value column is a data-loss bug rather than a stylistic slip: it silently
// caps the table at one NULL-valued row per scope, and the second insert fails with a duplicate
// key error naming a column the caller never wrote to.

func TestPartialUniqueStrict_EmitsOnlyTheNotNullIndex(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
		PartialUniqueStrict(dmodel.PartialUniqueParam{
			IndexName:     "widgets_name_org",
			NotNullFields: []string{"name"},
			NullableField: "org_id",
		}))
	require.NoError(t, err)
	joined := strings.Join(sqls, "\n")

	assert.Contains(t, joined,
		`CREATE UNIQUE INDEX "widgets_name_org_ukey" ON "widgets" ("name", "org_id") WHERE "org_id" IS NOT NULL`)
	assert.NotContains(t, joined, `IS NULL`,
		"a strict group must not constrain the rows whose nullable column is NULL")
	assert.NotContains(t, joined, "_ukey_notnull",
		"a strict group owns the plain _ukey suffix; _ukey_notnull would imply a missing _ukey_null sibling")
}

func TestPartialUniqueLoose_EmitsBothIndexes(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
		PartialUniqueLoose(dmodel.PartialUniqueParam{
			IndexName:     "widgets_name_org",
			NotNullFields: []string{"name"},
			NullableField: "org_id",
		}))
	require.NoError(t, err)
	joined := strings.Join(sqls, "\n")

	assert.Contains(t, joined,
		`CREATE UNIQUE INDEX "widgets_name_org_ukey_notnull" ON "widgets" ("name", "org_id") WHERE "org_id" IS NOT NULL`)
	assert.Contains(t, joined,
		`CREATE UNIQUE INDEX "widgets_name_org_ukey_null" ON "widgets" ("name") WHERE "org_id" IS NULL`)
}

// The tenant key is prepended to both kinds, so a strict index is still scoped per tenant.
func TestPartialUniqueStrict_PrependsTenantKey(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", true).
		PartialUniqueStrict(dmodel.PartialUniqueParam{
			IndexName:     "widgets_tid_name_org",
			NotNullFields: []string{"name"},
			NullableField: "org_id",
		}))
	require.NoError(t, err)
	assert.Contains(t, strings.Join(sqls, "\n"),
		`ON "widgets" ("tenant_id", "name", "org_id") WHERE "org_id" IS NOT NULL`)
}

// Both kinds are length-checked, but against the suffixes they actually emit: a name that fits
// under "_ukey" may still overflow under "_ukey_notnull".
func TestPartialUniqueStrict_OverLimitIsRejected(t *testing.T) {
	longName := strings.Repeat("w", 60)
	_, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
		PartialUniqueStrict(dmodel.PartialUniqueParam{
			IndexName:     longName,
			NotNullFields: []string{"name"},
			NullableField: "org_id",
		}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "identifier limit")
}

// A schema may mix the two kinds; each keeps its own suffix scheme.
func TestPartialUniqueMixed_KeepsSeparateSuffixes(t *testing.T) {
	sqls, err := createTableSqls(t, schemaWithTenantKey("widgets", false).
		PartialUniqueLoose(dmodel.PartialUniqueParam{
			IndexName:     "widgets_name_org",
			NotNullFields: []string{"name"},
			NullableField: "org_id",
		}).
		PartialUniqueStrict(dmodel.PartialUniqueParam{
			IndexName:     "widgets_code_org",
			NotNullFields: []string{"code"},
			NullableField: "org_id",
		}))
	require.NoError(t, err)
	joined := strings.Join(sqls, "\n")

	assert.Contains(t, joined, `"widgets_name_org_ukey_notnull"`)
	assert.Contains(t, joined, `"widgets_name_org_ukey_null"`)
	assert.Contains(t, joined, `"widgets_code_org_ukey"`)
	assert.NotContains(t, joined, `"widgets_code_org_ukey_null"`,
		"the strict group must not gain the loose group's NULL index")
}

// The schema accessors partition the groups, so a caller can ask which kind it is looking at.
func TestPartialUniqueAccessors_PartitionByKind(t *testing.T) {
	schema := schemaWithTenantKey("widgets", false).
		PartialUniqueLoose(dmodel.PartialUniqueParam{
			NotNullFields: []string{"name"},
			NullableField: "org_id",
		}).
		PartialUniqueStrict(dmodel.PartialUniqueParam{
			NotNullFields: []string{"code"},
			NullableField: "org_id",
		}).
		Build()

	assert.Len(t, schema.PartialUniques(), 2)
	require.Len(t, schema.PartialUniquesLoose(), 1)
	require.Len(t, schema.PartialUniquesStrict(), 1)
	assert.Equal(t, []string{"name"}, schema.PartialUniquesLoose()[0].NotNullFields)
	assert.Equal(t, []string{"code"}, schema.PartialUniquesStrict()[0].NotNullFields)
	assert.False(t, schema.PartialUniquesLoose()[0].Strict)
	assert.True(t, schema.PartialUniquesStrict()[0].Strict)
}
