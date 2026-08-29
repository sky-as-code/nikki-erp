package models

import (
	"slices"
	"testing"
)

// These pin the order schema against the business rules, so an edit that quietly changes a
// constraint fails here rather than in production data.

// TestSalesOrderKeepsFourIndependentStatuses: the four are never collapsed into one. An order can be
// fully paid and undelivered, or delivered and unpaid, or complete with its VAT invoice rejected.
func TestSalesOrderKeepsFourIndependentStatuses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderSchemaBuilder())
	for _, name := range []string{
		SalesOrderFieldStatus,
		SalesOrderFieldPaymentStatus,
		SalesOrderFieldFulfillmentStatus,
		SalesOrderFieldInvoiceStatus,
	} {
		field := fieldOf(t, schema, name)
		if !field.IsRequiredForCreate() {
			t.Errorf("%s must be required: an order with an unknown status in any dimension is "+
				"a row no query can filter correctly", name)
		}
		if !field.IsNoUpdate() {
			t.Errorf("%s must be no_update: a status is moved by a lifecycle operation that "+
				"validates the transition, never by a plain update", name)
		}
	}
}

// The idempotency mechanism is this index. Strict rather than loose: the loose variant would also
// emit a unique over sales_channel_id alone where the key IS NULL, capping each channel at one
// key-less order — and most orders carry no key.
func TestSalesOrderIdempotencyKeyIsScopedToChannel(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderSchemaBuilder())

	found := false
	for _, unique := range schema.PartialUniquesStrict() {
		if unique.NullableField != SalesOrderFieldIdempotencyKey {
			continue
		}
		if slices.Contains(unique.NotNullFields, SalesOrderFieldSalesChannelId) {
			found = true
		}
	}
	if !found {
		t.Error("sales_orders needs a STRICT partial unique on (sales_channel_id, " +
			"idempotency_key): without it a timed-out caller that retries creates a second order")
	}
}

// order_number is unique tenant-wide rather than per channel: a customer reading a number off a
// receipt has no idea which channel sold it.
func TestSalesOrderNumberIsGloballyUniqueAndImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderSchemaBuilder())
	number := fieldOf(t, schema, SalesOrderFieldOrderNumber)

	if !number.IsUnique() {
		t.Error("order_number must be plainly unique, not scoped to a channel: it is quoted off " +
			"a receipt with no channel in hand")
	}
	if !number.IsNoUpdate() {
		t.Error("order_number must be no_update: it has already been printed")
	}
	if !number.IsRequiredForCreate() {
		t.Error("order_number must be required: an order nobody can quote cannot be supported")
	}
}

// Both are NOT NULL and immutable. sales_channel_id is denormalised from the sales point and written
// only from the point's own channel, never from the request payload, which makes the consistency
// invariant true by construction; immutability keeps it true afterwards.
func TestSalesOrderChannelAndPointAreRequiredAndImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderSchemaBuilder())
	for _, name := range []string{SalesOrderFieldSalesChannelId, SalesOrderFieldSalesPointId} {
		field := fieldOf(t, schema, name)
		if !field.IsRequiredForCreate() {
			t.Errorf("%s must be NOT NULL: a nullable one would leak 'unknown' into every "+
				"query that groups by it (D-19)", name)
		}
		if !field.IsNoUpdate() {
			t.Errorf("%s must be immutable: moving an order between channels would rewrite "+
				"which channel the sale happened in (D-20)", name)
		}
	}
}

// A customer is optional because anonymous sale is the default, not the exception: a vending kiosk
// never knows who is buying.
func TestSalesOrderCustomerIsOptional(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderSchemaBuilder())
	if fieldOf(t, schema, SalesOrderFieldCustomerReference).IsRequiredForCreate() {
		t.Error("customer_reference must be optional: anonymous sale is the default (BR 87.6)")
	}
}

// A line is identified within its order by line_number, and that pair must be unique: returns and
// fiscal documents name lines by number, so a shared number makes both ambiguous.
func TestSalesOrderLineNumberIsUniquePerOrder(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderLineSchemaBuilder())

	found := false
	for _, unique := range schema.CompositeUniques() {
		if slices.Contains(unique.Fields, SalesOrderLineFieldSalesOrderId) &&
			slices.Contains(unique.Fields, SalesOrderLineFieldLineNumber) {
			found = true
		}
	}
	if !found {
		t.Error("sales_order_lines must be unique on (sales_order_id, line_number): a return " +
			"and a fiscal document both name a line by its number")
	}
}

// All three quantities are required: partial fulfilment is normal, and a single quantity could not
// express a line where two of three items were dispensed and the third jammed.
func TestSalesOrderLineCarriesThreeQuantities(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderLineSchemaBuilder())
	for _, name := range []string{
		SalesOrderLineFieldOrderedQuantity,
		SalesOrderLineFieldFulfilledQuantity,
		SalesOrderLineFieldReturnedQuantity,
	} {
		if !fieldOf(t, schema, name).IsRequiredForCreate() {
			t.Errorf("%s must be required (BR 87.8): a nil quantity cannot be added up", name)
		}
	}
}

// product_variant_id is nullable because a combo parent line sells no single variant — its
// components do. Requiring it would make a combo unrepresentable.
func TestSalesOrderLineVariantIsNullableForCombos(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderLineSchemaBuilder())
	if fieldOf(t, schema, SalesOrderLineFieldProductVariantId).IsRequiredForCreate() {
		t.Error("product_variant_id must be nullable: a combo parent line sells no single " +
			"variant, its components do (BR 17)")
	}
}

// Every snapshot field must actually exist on the schema. SnapshotFields is the single definition of
// "frozen after confirm", so a name that drifted from the schema would silently stop being enforced.
func TestSnapshotFieldsAllExist(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderLineSchemaBuilder())
	for _, name := range SnapshotFields {
		if _, ok := schema.Fields()[name]; !ok {
			t.Errorf("SnapshotFields names %q, which sales_order_line does not declare: the "+
				"immutability rule would silently not apply to it", name)
		}
	}
}

// uom_id is required on every line, which is what makes the UomUsageProbe necessary: without it
// Essential would let somebody edit a unit that sales history depends on, and an old receipt would
// come to mean a different amount of goods.
func TestSalesOrderLineRequiresUom(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesOrderLineSchemaBuilder())
	uom := fieldOf(t, schema, SalesOrderLineFieldUomId)
	if !uom.IsRequiredForCreate() {
		t.Error("uom_id must be required: a quantity without its unit is not a quantity")
	}
	if !uom.IsNoUpdate() {
		t.Error("uom_id must be immutable: re-expressing a sold quantity in another unit would " +
			"change what was sold")
	}
}
