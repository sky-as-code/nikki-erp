package baserepo

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func fkTestSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("test_order_line").
		TableName("test_order_lines").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("uom_id").DataType(dmodel.FieldDataTypeUlid())).
		Build()
}

func pgFkError(constraint string) error {
	return &pq.Error{Code: pq.ErrorCode(pgForeignKeyViolation), Constraint: constraint}
}

// An insert naming a parent that does not exist is the writer's mistake, and the client can fix
// it if told which field was wrong.
func TestNormalizeFk_InsertNamesTheOffendingField(t *testing.T) {
	clientErrs := normalizeFkViolation(
		pgFkError("test_order_lines_uom_id_fkey"), fkTestSchema(), false)

	require.NotNil(t, clientErrs)
	require.Equal(t, 1, clientErrs.Count())
	assert.True(t, clientErrs.Has("uom_id"), "the constraint name identifies the column")
}

// Deleting a referenced parent is a refusal, not a bad request, and the constraint names the
// CHILD table rather than a column of the row being deleted.
func TestNormalizeFk_DeleteReportsTheReferencingTable(t *testing.T) {
	clientErrs := normalizeFkViolation(
		pgFkError("sales_order_lines_uom_id_fkey"), fkTestSchema(), true)

	require.NotNil(t, clientErrs)
	require.Equal(t, 1, clientErrs.Count())
	assert.Contains(t, clientErrs.ToError().Error(), "sales_order_lines")
	assert.Contains(t, clientErrs.ToError().Error(), "still referenced")
}

// The key a delete refusal reports must match what usagecheck reports before the statement runs.
// One condition must not look like two different problems depending on which layer caught it.
func TestNormalizeFk_DeleteUsesTheSameKeyAsTheBusinessCheck(t *testing.T) {
	assert.Equal(t, "resource_in_use", ErrorKeyResourceInUse)
}

// PostgreSQL truncates identifiers at 63 bytes, so a long table plus column loses the shape the
// field is parsed from. Reporting no field is right; reporting a wrong one is not.
func TestNormalizeFk_UnparseableConstraintYieldsNoField(t *testing.T) {
	clientErrs := normalizeFkViolation(
		pgFkError("inventory_product_variant_attribute_values_template_attribute_v"),
		fkTestSchema(), false)

	require.NotNil(t, clientErrs)
	require.Equal(t, 1, clientErrs.Count())
	assert.False(t, clientErrs.Has("uom_id"))
}

// A column named by the driver beats one parsed out of a constraint name.
func TestNormalizeFk_PrefersTheDriverReportedColumn(t *testing.T) {
	err := &pq.Error{
		Code:       pq.ErrorCode(pgForeignKeyViolation),
		Constraint: "test_order_lines_uom_id_fkey",
		Column:     "uom_id",
	}

	clientErrs := normalizeFkViolation(err, fkTestSchema(), false)

	require.NotNil(t, clientErrs)
	assert.True(t, clientErrs.Has("uom_id"))
}

// Everything that is not a foreign key violation must pass through untouched, or a real failure
// would be reported to the client as an ordinary refusal.
func TestNormalizeFk_IgnoresOtherErrors(t *testing.T) {
	assert.Nil(t, normalizeFkViolation(errors.New("connection refused"), fkTestSchema(), false))
	assert.Nil(t, normalizeFkViolation(
		&pq.Error{Code: "23505", Constraint: "test_order_lines_pkey"}, fkTestSchema(), false),
		"a unique violation is a different condition with its own handling")
	assert.Nil(t, normalizeFkViolation(nil, fkTestSchema(), false))
}

// The driver error may arrive wrapped by a layer between the statement and here.
func TestNormalizeFk_FindsAWrappedDriverError(t *testing.T) {
	wrapped := errors.Wrap(pgFkError("test_order_lines_uom_id_fkey"), "DeleteOne")

	clientErrs := normalizeFkViolation(wrapped, fkTestSchema(), false)

	require.NotNil(t, clientErrs)
	assert.True(t, clientErrs.Has("uom_id"))
}
