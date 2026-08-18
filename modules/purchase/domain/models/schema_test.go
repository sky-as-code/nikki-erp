package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// Every schema is built here because ParseModelJson PANICS on a malformed declaration rather than
// returning an error — an unknown data type, a sized type missing its bounds, a bad index name. In
// an application that panic happens at start-up, so without these tests the first sign of a typo is
// a server that will not boot.
func TestAllSchemasParse(t *testing.T) {
	requireBaseSchemasRegistered(t)

	testCases := []struct {
		name    string
		builder func() *dmodel.ModelSchemaBuilder
		table   string
	}{
		{ConfigurationSchemaName, ConfigurationSchemaBuilder, "purchase_configurations"},
		{SourcingGroupSchemaName, SourcingGroupSchemaBuilder, "purchase_sourcing_groups"},
		{AgreementSchemaName, AgreementSchemaBuilder, "purchase_agreements"},
		{AgreementLineSchemaName, AgreementLineSchemaBuilder, "purchase_agreement_lines"},
		{PurchaseOrderSchemaName, PurchaseOrderSchemaBuilder, "purchase_orders"},
		{PurchaseOrderLineSchemaName, PurchaseOrderLineSchemaBuilder, "purchase_order_lines"},
		{AuditEventSchemaName, AuditEventSchemaBuilder, "purchase_audit_events"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			schema := testCase.builder().Build()

			assert.Equal(t, testCase.name, schema.Name())
			assert.Equal(t, testCase.table, schema.TableName())
		})
	}
}

// PUR-R2. The order is the one resource here that must NOT be archivable: its lifecycle is its
// status, and an archived-but-open order would be a document withdrawn from view while still
// committing the business to a purchase.
func TestPurchaseOrderIsNotArchivable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	_, ok := PurchaseOrderSchemaBuilder().Build().Field(basemodel.FieldIsArchived)

	assert.False(t, ok, "purchase_order must not extend the archivable mixin")
}

// The agreement, by contrast, is archivable, and status and is_archived are independent: a closed
// agreement may be archived or not and both combinations are meaningful (BR 34).
func TestAgreementIsArchivable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	_, ok := AgreementSchemaBuilder().Build().Field(basemodel.FieldIsArchived)

	assert.True(t, ok, "purchase_agreement must be archivable")
}

// PUR-R3: the lifecycle fields are writable only by their own actions. Expressing that as no_update
// in the schema is what makes the engine reject them, so there is no hand-written guard to forget.
//
// This is the security property of the whole module: without it, a plain PATCH could approve an
// order, mark it acknowledged by a vendor who never saw it, or move it to PURCHASE_ORDER without
// ever passing the approval threshold.
func TestPurchaseOrderLifecycleFieldsAreNoUpdate(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := PurchaseOrderSchemaBuilder().Build()

	for _, fieldName := range []string{
		PurchaseOrderFieldStatus,
		PurchaseOrderFieldIsLocked,
		PurchaseOrderFieldVendorAcknowledged,
		PurchaseOrderFieldConfirmedAt,
		PurchaseOrderFieldApprovalRequired,
		PurchaseOrderFieldApprovedBy,
		PurchaseOrderFieldApprovedAt,
	} {
		field, ok := schema.Field(fieldName)
		require.True(t, ok, "%s missing", fieldName)
		assert.True(t, field.IsNoUpdate(), "%s must be no_update", fieldName)
	}
}

// PUR-R4: the header totals are server-computed. A client-supplied total is ignored, not trusted,
// so that the figure can never disagree with the lines it claims to total.
func TestPurchaseOrderTotalsAreNoUpdate(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := PurchaseOrderSchemaBuilder().Build()

	for _, fieldName := range []string{
		PurchaseOrderFieldUntaxedAmount,
		PurchaseOrderFieldTaxAmount,
		PurchaseOrderFieldTotalAmount,
	} {
		field, ok := schema.Field(fieldName)
		require.True(t, ok, "%s missing", fieldName)
		assert.True(t, field.IsNoUpdate(), "%s must be no_update", fieldName)
	}
}

// The line's computed amounts are no_update for the same reason — except tax_amount, which is
// deliberately an INPUT because there is no tax engine (D9). That exception is asserted rather than
// left as an absence, so that adding a tax engine later is a visible change here.
func TestPurchaseOrderLineTaxIsAnInputAndTheRestAreNot(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := PurchaseOrderLineSchemaBuilder().Build()

	for _, fieldName := range []string{
		PurchaseOrderLineFieldSubtotal,
		PurchaseOrderLineFieldTotal,
		PurchaseOrderLineFieldInventoryQuantity,
	} {
		field, ok := schema.Field(fieldName)
		require.True(t, ok, "%s missing", fieldName)
		assert.True(t, field.IsNoUpdate(), "%s must be computed, not supplied", fieldName)
	}

	tax, ok := schema.Field(PurchaseOrderLineFieldTaxAmount)
	require.True(t, ok)
	assert.False(t, tax.IsNoUpdate(),
		"a line's tax is supplied by the client: there is no tax engine to compute it")
}

// BR-UOM-PUR-004: the quantity and unit the buyer chose are never overwritten by the conversion.
// Only the derived inventory_quantity is computed, which is why these two must stay writable.
func TestPurchaseOrderLineKeepsTheOriginalQuantityWritable(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := PurchaseOrderLineSchemaBuilder().Build()

	quantity, ok := schema.Field(PurchaseOrderLineFieldQuantity)
	require.True(t, ok)
	assert.False(t, quantity.IsNoUpdate(), "the ordered quantity is the buyer's, not derived")

	uom, ok := schema.Field(PurchaseOrderLineFieldUomId)
	require.True(t, ok)
	assert.False(t, uom.IsNoUpdate(), "the ordered unit is the buyer's, not derived")
}

// D5a: no foreign key crosses a module boundary. A vendor lives in Contacts, a product and a unit
// in Inventory and Essential; an edge to any of them would make this module's schema depend on
// another module's tables, which the migration tooling cannot even generate.
func TestNoCrossModuleForeignKeys(t *testing.T) {
	requireBaseSchemasRegistered(t)

	foreign := map[string]bool{
		PurchaseOrderFieldVendorId:             true,
		PurchaseOrderFieldCurrencyId:           true,
		PurchaseOrderFieldBuyerId:              true,
		PurchaseOrderLineFieldProductVariantId: true,
		PurchaseOrderLineFieldUomId:            true,
	}

	for _, builder := range []func() *dmodel.ModelSchemaBuilder{
		PurchaseOrderSchemaBuilder,
		PurchaseOrderLineSchemaBuilder,
		AgreementSchemaBuilder,
		AgreementLineSchemaBuilder,
	} {
		schema := builder().Build()
		for _, relation := range schema.ToRelations() {
			assert.False(t, foreign[relation.SrcField],
				"%s.%s must not be a foreign key: it points outside this module",
				schema.Name(), relation.SrcField)
		}
	}
}

// The intra-module edges, which SHOULD exist: a line cannot outlive its parent, and an orphaned
// line would keep contributing to a total nobody can read.
func TestLinesCascadeFromTheirParents(t *testing.T) {
	requireBaseSchemasRegistered(t)

	testCases := []struct {
		builder func() *dmodel.ModelSchemaBuilder
		dest    string
	}{
		{PurchaseOrderLineSchemaBuilder, PurchaseOrderSchemaName},
		{AgreementLineSchemaBuilder, AgreementSchemaName},
	}

	for _, testCase := range testCases {
		schema := testCase.builder().Build()
		relations := schema.ToRelations()
		require.Len(t, relations, 1, "%s", schema.Name())
		assert.Equal(t, testCase.dest, relations[0].DestSchemaName)
		assert.Equal(t, dmodel.RelationCascadeCascade, relations[0].OnDelete,
			"%s must cascade from its parent", schema.Name())
	}
}

// D6: the order code is unique per organization, not globally. Two organizations in one deployment
// each numbering their orders from the same series is normal, not a conflict.
func TestCodesAreUniquePerOrg(t *testing.T) {
	requireBaseSchemasRegistered(t)

	testCases := []struct {
		builder func() *dmodel.ModelSchemaBuilder
		field   string
	}{
		{PurchaseOrderSchemaBuilder, PurchaseOrderFieldCode},
		{AgreementSchemaBuilder, AgreementFieldCode},
	}

	for _, testCase := range testCases {
		schema := testCase.builder().Build()
		composites := schema.CompositeUniques()
		require.Len(t, composites, 1, "%s", schema.Name())
		assert.Equal(t, []string{testCase.field, basemodel.FieldOrgId}, composites[0].Fields)
	}
}

// One configuration per organization. Two rows would make "does this order need approval?" depend
// on which was read.
func TestConfigurationIsOnePerOrg(t *testing.T) {
	requireBaseSchemasRegistered(t)

	composites := ConfigurationSchemaBuilder().Build().CompositeUniques()

	require.Len(t, composites, 1)
	assert.Equal(t, []string{basemodel.FieldOrgId}, composites[0].Fields)
}

// PUR-R6: an audit event is immutable. It extends the readonly auditable mixin, so it carries
// created_at but no updated_at — there is no such thing as editing one.
func TestAuditEventIsImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := AuditEventSchemaBuilder().Build()

	_, hasCreated := schema.Field(basemodel.FieldCreatedAt)
	assert.True(t, hasCreated, "an audit event must record when it happened")

	_, hasUpdated := schema.Field(basemodel.FieldUpdatedAt)
	assert.False(t, hasUpdated, "an audit event is never updated, so it has no updated_at")

	_, hasArchived := schema.Field(basemodel.FieldIsArchived)
	assert.False(t, hasArchived, "an audit event is never archived")
}

// The audit event deliberately holds no foreign key to what it describes: it must outlive the
// record, and a cascade would delete exactly the history someone is auditing.
func TestAuditEventHasNoEdgeToItsSubject(t *testing.T) {
	requireBaseSchemasRegistered(t)

	assert.Empty(t, AuditEventSchemaBuilder().Build().ToRelations(),
		"an audit event must survive the deletion of the record it describes")
}

// Status values are lower-case in the database (D7). The requirement writes them upper-case, which
// is presentation; every enum in this codebase stores lower-case.
func TestStatusValuesAreLowerCase(t *testing.T) {
	requireBaseSchemasRegistered(t)

	order, ok := PurchaseOrderSchemaBuilder().Build().Field(PurchaseOrderFieldStatus)
	require.True(t, ok)
	assert.Equal(t,
		[]string{"rfq", "rfq_sent", "to_approve", "purchase_order", "cancelled"},
		enumValuesOf(t, order))

	agreement, ok := AgreementSchemaBuilder().Build().Field(AgreementFieldStatus)
	require.True(t, ok)
	assert.Equal(t,
		[]string{"draft", "confirmed", "closed", "cancelled"},
		enumValuesOf(t, agreement))
}

// enumValuesOf reads the declared values of an enum field. They live in the data type's options
// map rather than on the field itself, keyed "enumValues".
func enumValuesOf(t *testing.T, field *dmodel.ModelField) []string {
	t.Helper()
	raw, ok := field.DataType().Options()["enumValues"]
	require.True(t, ok, "field is not an enum")
	values, ok := raw.([]string)
	require.True(t, ok, "enumValues is not a []string")
	return values
}

func requireBaseSchemasRegistered(t *testing.T) {
	t.Helper()
	// Normally done by CoreModule.RegisterModels during app start-up.
	_ = basemodel.RegisterJsonBaseSchemas()
}
