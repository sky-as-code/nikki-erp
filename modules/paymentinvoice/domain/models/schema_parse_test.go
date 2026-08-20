package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The JSON model files are parsed at start-up by RegisterModels; a malformed file panics the whole
// app rather than failing the build. These tests turn that into a test failure instead, and pin
// the properties the module's money and lifecycle rules depend on.

func requireBaseSchemasRegistered(t *testing.T) {
	t.Helper()
	// Normally done by CoreModule.RegisterModels during app start-up.
	_ = basemodel.RegisterJsonBaseSchemas()
}

func requireField(t *testing.T, schema *dmodel.ModelSchema, fieldName string) *dmodel.ModelField {
	t.Helper()
	field, ok := schema.Fields()[fieldName]
	require.Truef(t, ok, "schema %q has no field %q", schema.Name(), fieldName)
	return field
}

// The schema name in the JSON is what the engine, the permission check and the frontend all key
// on; the Go constant is what this module's own code passes around. They are declared in separate
// files, and a drift between them denies every request with nothing pointing at the cause.
func TestEverySchemaParsesUnderItsConstantName(t *testing.T) {
	requireBaseSchemasRegistered(t)

	cases := []struct {
		schemaName string
		tableName  string
		build      func() *dmodel.ModelSchemaBuilder
	}{
		{PaymentMethodSchemaName, "paymentinvoice_payment_methods", PaymentMethodSchemaBuilder},
		{PaymentProfileSchemaName, "paymentinvoice_payment_profiles", PaymentProfileSchemaBuilder},
		{OrderSchemaName, "paymentinvoice_orders", OrderSchemaBuilder},
		{TransactionSchemaName, "paymentinvoice_transactions", TransactionSchemaBuilder},
		{InvoiceSchemaName, "paymentinvoice_invoices", InvoiceSchemaBuilder},
		{InvoiceLineSchemaName, "paymentinvoice_invoice_lines", InvoiceLineSchemaBuilder},
	}

	for _, testCase := range cases {
		t.Run(testCase.schemaName, func(t *testing.T) {
			schema := testCase.build().Build()

			assert.Equal(t, testCase.schemaName, schema.Name())
			assert.Equal(t, testCase.tableName, schema.TableName())
			// Without a record label field the frontend relation picker shows raw ULIDs.
			assert.NotEmpty(t, schema.RecordLabelField())
		})
	}
}

// Money is decimal, never an integer type. How many minor units a currency has is a property of
// the currency — VND has none, KWD has three — so amounts held as a count of them would push a
// per-currency scale into every caller that reads one. A percentage is decimal for a related
// reason: an integer column truncates a rate like 8.5 rather than rejecting it.
func TestMoneyAndPercentageFieldsAreDecimal(t *testing.T) {
	requireBaseSchemasRegistered(t)

	decimalFields := map[*dmodel.ModelSchema][]string{
		OrderSchemaBuilder().Build():       {OrderFieldAmount, OrderFieldRefundAmount},
		TransactionSchemaBuilder().Build(): {TransactionFieldAmount},
		InvoiceSchemaBuilder().Build(): {
			InvoiceFieldSubtotalAmount,
			InvoiceFieldTaxAmount,
			InvoiceFieldTotalAmount,
		},
		InvoiceLineSchemaBuilder().Build(): {
			InvoiceLineFieldUnitPrice,
			InvoiceLineFieldAmount,
			InvoiceLineFieldTaxRatePercent,
		},
		PaymentMethodSchemaBuilder().Build(): {
			PaymentMethodFieldMinAmount,
			PaymentMethodFieldMaxAmount,
		},
	}

	for schema, fieldNames := range decimalFields {
		for _, fieldName := range fieldNames {
			field := requireField(t, schema, fieldName)
			assert.Equalf(t, dmodel.FieldDataTypeNameDecimal, field.DataType().String(),
				"%s.%s must be decimal", schema.Name(), fieldName)
		}
	}
}

// An integer field's width is chosen from the largest value it can legitimately hold. A line
// quantity is bounded at a million, so int64 would reserve eight bytes a row to represent values
// no line can carry, and would tell the next reader the bound had not been considered.
func TestBoundedCountersAreInt32(t *testing.T) {
	requireBaseSchemasRegistered(t)

	quantity := requireField(t, InvoiceLineSchemaBuilder().Build(), InvoiceLineFieldQuantity)
	assert.Equal(t, dmodel.FieldDataTypeNameInt32, quantity.DataType().String())
}

// The order carries no column per gateway. What one gateway needs at create time — a terminal id
// for a card reader — is meaningless to the others, so a column named for it would be dead weight
// on every order paid another way and adding a gateway would mean a migration.
func TestOrderCarriesNoGatewaySpecificColumn(t *testing.T) {
	requireBaseSchemasRegistered(t)

	fields := OrderSchemaBuilder().Build().Fields()
	_, hasPosId := fields["pos_id"]
	assert.False(t, hasPosId, "gateway-specific inputs belong in metadata, not in a column")

	_, hasMetadata := fields[OrderFieldMetadata]
	assert.True(t, hasMetadata)
}

// Currency and payment method are reference data, not enums: a deployment must be able to add a
// gateway account or withdraw one without a release. Holding them as ulid references is also what
// lets the frontend render a relation picker rather than a hardcoded list.
func TestCurrencyAndMethodAreReferences(t *testing.T) {
	requireBaseSchemasRegistered(t)

	order := OrderSchemaBuilder().Build()
	assert.Equal(t, dmodel.FieldDataTypeNameUlid,
		requireField(t, order, OrderFieldCurrencyId).DataType().String())
	assert.Equal(t, dmodel.FieldDataTypeNameUlid,
		requireField(t, order, OrderFieldPaymentMethodId).DataType().String())

	assert.Equal(t, dmodel.FieldDataTypeNameUlid,
		requireField(t, InvoiceSchemaBuilder().Build(), InvoiceFieldCurrencyId).DataType().String())
}

// Status is advanced only by this module's own actions and by the gateway callbacks. Leaving it
// client-writable would let a caller declare an order paid that no gateway ever settled, and the
// records would look identical to genuine ones.
func TestLifecycleFieldsRejectClientUpdates(t *testing.T) {
	requireBaseSchemasRegistered(t)

	assert.True(t, requireField(t, OrderSchemaBuilder().Build(), OrderFieldStatus).IsNoUpdate())
	assert.True(t, requireField(t, TransactionSchemaBuilder().Build(), TransactionFieldStatus).IsNoUpdate())

	invoice := InvoiceSchemaBuilder().Build()
	assert.True(t, requireField(t, invoice, InvoiceFieldStatus).IsNoUpdate())
	// The number and the totals are assigned by issue, together and from the lines.
	assert.True(t, requireField(t, invoice, InvoiceFieldNumber).IsNoUpdate())
	assert.True(t, requireField(t, invoice, InvoiceFieldTotalAmount).IsNoUpdate())
}

// An order is found by order_code on every gateway callback and by order_id whenever support or
// the ordering system quotes one. Both must be unique, or a callback could settle the wrong order.
func TestOrderIdentifiersAreRequiredAndImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := OrderSchemaBuilder().Build()
	for _, fieldName := range []string{OrderFieldOrderId, OrderFieldOrderCode} {
		field := requireField(t, schema, fieldName)
		assert.Truef(t, field.IsRequiredForCreate(), "%s must be required for create", fieldName)
		assert.Truef(t, field.IsNoUpdate(), "%s must be immutable", fieldName)
	}
}

// An order records which merchant account collected it, and that record has to outlive the
// account: the order is evidence of money that moved, so a profile being withdrawn must not take
// it down with it. Optional, because an order collected with the deployment's own credentials
// names no profile at all — which is every order taken before profiles existed.
func TestOrderRecordsItsPaymentProfileWithoutDependingOnIt(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := OrderSchemaBuilder().Build()
	field := requireField(t, schema, OrderFieldPaymentProfileId)

	assert.Equal(t, dmodel.FieldDataTypeNameUlid, field.DataType().String())
	assert.False(t, field.IsRequiredForCreate(), "an order may be collected without a profile")
	assert.True(t, field.IsNoUpdate(),
		"the account that took the money cannot be rewritten after the fact")

	_, hasEdge := schema.Fields()["payment_profile"]
	assert.False(t, hasEdge, "held as a plain id: an edge would cascade a withdrawal into the orders")
}

// Every record this module writes belongs to an organization, and the schema enforces it. This is
// here because it was not enforced anywhere else for a while: the payment flow composed an order
// without org_id, and the only thing that noticed was the schema — as a Go error out of the
// create, which reads as a server fault rather than as the missing field it is.
func TestOrgIdIsRequiredOnEveryRecordThePaymentFlowWrites(t *testing.T) {
	requireBaseSchemasRegistered(t)

	for _, build := range []func() *dmodel.ModelSchemaBuilder{
		OrderSchemaBuilder, TransactionSchemaBuilder,
	} {
		schema := build().Build()
		field := requireField(t, schema, "org_id")

		assert.Truef(t, field.IsRequiredForCreate(),
			"%s must refuse a record that names no organization", schema.Name())
	}
}
