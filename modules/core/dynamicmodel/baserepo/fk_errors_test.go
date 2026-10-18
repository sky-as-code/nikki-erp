package baserepo

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules"
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

// Embedding the interface leaves every method but the two under test unimplemented, which is all
// the namespace lookup reads.
type fkTestModule struct {
	modules.InCodeModule
	name   string
	prefix string
}

func (this fkTestModule) Name() string        { return this.name }
func (this fkTestModule) ModelPrefix() string { return this.prefix }

func withLoadedModules(t *testing.T, mods ...modules.InCodeModule) {
	original := loadedModules
	loadedModules = func() []modules.InCodeModule { return mods }
	t.Cleanup(func() { loadedModules = original })
}

// The referencing table is sent as a translation key in the namespace of the module that owns
// it, so a client can name it in the user's language rather than show a raw table name.
func TestNormalizeFk_DeleteReportsANamespacedTableKey(t *testing.T) {
	withLoadedModules(t,
		fkTestModule{name: "iam", prefix: "iam"},
		fkTestModule{name: "vendingmachine", prefix: "vdmc"})

	clientErrs := normalizeFkViolation(
		&pq.Error{Code: pq.ErrorCode(pgForeignKeyViolation), Table: "vdmc_kiosks"}, fkTestSchema(), true)

	require.NotNil(t, clientErrs)
	item := (*clientErrs)[0]
	assert.Equal(t, "common:resource_in_use", item.Key)
	assert.Equal(t, "vendingmachine:vdmc_kiosks", item.Vars["referenced_by"],
		"the namespace is the module name, not its model prefix")
}

func TestNormalizeFk_InsertReportsTheCommonKey(t *testing.T) {
	clientErrs := normalizeFkViolation(
		pgFkError("test_order_lines_uom_id_fkey"), fkTestSchema(), false)

	require.NotNil(t, clientErrs)
	assert.Equal(t, "common:foreign_key_violation", (*clientErrs)[0].Key)
}

// "inventory_x" must not be claimed by a module prefixed "inv" when "inventory" is loaded too,
// and a prefix only counts up to its separator.
func TestLanguageNamespace_LongestPrefixWins(t *testing.T) {
	withLoadedModules(t,
		fkTestModule{name: "short", prefix: "inv"},
		fkTestModule{name: "inventory", prefix: "inventory"})

	assert.Equal(t, "inventory", languageNamespaceOf("inventory_products"))
	assert.Equal(t, "short", languageNamespaceOf("inv_items"))
}

func TestLanguageNamespace_UnknownPrefixFallsBackToCommon(t *testing.T) {
	withLoadedModules(t, fkTestModule{name: "iam", prefix: "iam"})

	assert.Equal(t, "common", languageNamespaceOf("sales_order_lines"))
	assert.Equal(t, "common", languageNamespaceOf("iamx_users"))
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
