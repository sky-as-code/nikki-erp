package baserepo

import (
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func uniqueTestSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("test_location").
		TableName("test_locations").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("code").DataType(dmodel.FieldDataTypeString(1, 50)).RequiredForCreate()).
		Field(dmodel.DefineField().Name("org_id").DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate()).
		Field(dmodel.DefineField().Name("warehouse_id").DataType(dmodel.FieldDataTypeUlid())).
		PartialUniqueLoose(dmodel.PartialUniqueParam{
			IndexName: "test_locs_code_org_wh", NotNullFields: []string{"code", "org_id"}, NullableField: "warehouse_id",
		}).
		Build()
}

func uniqueGroupsOf(t *testing.T, constraint string) any {
	t.Helper()
	clientErrs := normalizeUniqueViolation(
		&pq.Error{Code: pq.ErrorCode(pgUniqueViolation), Constraint: constraint}, uniqueTestSchema())
	require.NotNil(t, clientErrs)
	require.Equal(t, 1, clientErrs.Count())
	item := (*clientErrs)[0]
	assert.Equal(t, "common:err_unique_constraint_violated", item.Key)
	return item.Vars["uniques"]
}

// A partial unique is not covered by the pre-insert collision check, so its violation reaches the
// database; it must still answer the same client error rather than a 500.
func TestNormalizeUnique_NullBranchNamesTheNotNullFields(t *testing.T) {
	assert.Equal(t, [][]string{{"code", "org_id"}}, uniqueGroupsOf(t, "test_locs_code_org_wh_ukey_null"))
}

func TestNormalizeUnique_NotNullBranchAddsTheNullableField(t *testing.T) {
	assert.Equal(t, [][]string{{"code", "org_id", "warehouse_id"}},
		uniqueGroupsOf(t, "test_locs_code_org_wh_ukey_notnull"))
}

func TestNormalizeUnique_UnknownConstraintReportsNoGroup(t *testing.T) {
	assert.Equal(t, [][]string{}, uniqueGroupsOf(t, "something_else_ukey"))
}

func TestNormalizeUnique_IgnoresOtherErrors(t *testing.T) {
	assert.Nil(t, normalizeUniqueViolation(errors.New("connection refused"), uniqueTestSchema()))
	assert.Nil(t, normalizeUniqueViolation(pgFkError("test_locations_org_id_fkey"), uniqueTestSchema()))
}
