package models

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// What the fiscal request schema must guarantee, and why each one is worth a test.

// The idempotency key is the only thing standing between a timeout and a duplicate VAT invoice, so
// it is unique and it cannot be edited. Editable, it would let a caller rewrite the key of a request
// already sent to a provider and then legitimately issue a second document for one sale.
func TestFiscalRequestIdempotencyKeyIsUniqueAndImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFiscalRequestSchemaBuilder())

	field := fieldOf(t, schema, SalesFiscalRequestFieldIdempotencyKey)
	if !field.IsRequiredForCreate() {
		t.Error("idempotency_key must be required: a request without one cannot be retried safely, " +
			"and an unsafe retry issues a second legal document")
	}
	if !field.IsNoUpdate() {
		t.Error("idempotency_key must be no_update: editing the key of a request already sent " +
			"would let the same sale be invoiced twice")
	}

	found := false
	for _, unique := range schema.CompositeUniques() {
		if len(unique.Fields) == 1 && unique.Fields[0] == SalesFiscalRequestFieldIdempotencyKey {
			found = true
		}
	}
	if !found {
		t.Error("idempotency_key must carry a unique index; it is the Sales-side half of the " +
			"duplicate-invoice guard, the other half being the key travelling to the provider")
	}
}

// The bill and the intent are immutable. A request that could be re-pointed would leave an issued
// legal document describing a sale it was never issued for; one that could change its intent would
// leave the document it already caused unexplained.
func TestFiscalRequestBillAndIntentAreImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFiscalRequestSchemaBuilder())
	for _, name := range []string{
		SalesFiscalRequestFieldSalesBillId,
		SalesFiscalRequestFieldIntent,
	} {
		field := fieldOf(t, schema, name)
		if !field.IsRequiredForCreate() {
			t.Errorf("%s must be required for create", name)
		}
		if !field.IsNoUpdate() {
			t.Errorf("%s must be no_update: an issued document cannot retrospectively be said to "+
				"have been about something else", name)
		}
	}
}

// The intent enum carries BUSINESS INTENT and nothing else (BR 48, 49).
//
// The assertion that matters is the ABSENCE: no value here names a document type, because deciding
// between an invoice, a credit note and an adjustment declaration is invoice law, and BR 46 and
// BR 94.26 put it on the provider's side of the port.
func TestFiscalIntentCarriesBusinessIntentNotDocumentTypes(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFiscalRequestSchemaBuilder())
	values := enumSetOf(t, schema, SalesFiscalRequestFieldIntent)

	for _, wanted := range []string{
		string(SalesFiscalIntentIssueOriginal),
		string(SalesFiscalIntentAdjustForFullReturn),
		string(SalesFiscalIntentAdjustForPartialReturn),
		string(SalesFiscalIntentAdjustPrice),
	} {
		if !values[wanted] {
			t.Errorf("the intent enum must offer %q", wanted)
		}
	}

	for _, forbidden := range []string{
		"invoice", "credit_note", "debit_note", "adjustment_declaration", "vat_invoice",
	} {
		if values[forbidden] {
			t.Errorf("the intent enum must NOT offer %q: naming a document type here encodes "+
				"invoice law in Sales, which BR 46 forbids", forbidden)
		}
	}
	if len(values) != 4 {
		t.Errorf("the intent enum has %d values, want exactly the 4 business intents", len(values))
	}
}

// PENDING IS NOT ISSUED (BR 77). Both must exist as distinct states, because a request in flight and
// a confirmed document are the two things a customer must never see conflated: one of them can be
// deducted and the other does not exist.
func TestFiscalStatusSeparatesPendingFromIssued(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFiscalRequestSchemaBuilder())
	values := enumSetOf(t, schema, SalesFiscalRequestFieldStatus)

	for _, wanted := range []string{
		string(SalesFiscalStatusPending),
		string(SalesFiscalStatusIssued),
		string(SalesFiscalStatusFailed),
		string(SalesFiscalStatusCancelled),
	} {
		if !values[wanted] {
			t.Errorf("the status enum must offer %q", wanted)
		}
	}

	field := fieldOf(t, schema, SalesFiscalRequestFieldStatus)
	if field.IsNoUpdate() {
		t.Error("status must be updatable: unlike the order statuses it is moved by the provider's " +
			"answer arriving, which is not a client edit but is still an update to this row")
	}
}

// provider_reference is nullable, and that is the assertion. Required, it would have to be invented
// at creation time — before any provider has issued anything — and an invented reference to a legal
// document is worse than none: it points at nothing while looking like it points somewhere.
func TestProviderReferenceIsNullUntilIssued(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFiscalRequestSchemaBuilder())

	field := fieldOf(t, schema, SalesFiscalRequestFieldProviderRef)
	if field.IsRequiredForCreate() {
		t.Error("provider_reference must not be required: it does not exist until a provider has " +
			"confirmed the document, and inventing one would point at a document that is not there")
	}
}

// The adjustment link is nullable and self-referential (BR 58): null on an original, set on every
// adjustment. Held here rather than as a flag on the original, so one invoice can be adjusted more
// than once — which is exactly what a sale returned in two parts produces.
func TestAdjustmentLinkIsOptionalAndSelfReferential(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFiscalRequestSchemaBuilder())

	field := fieldOf(t, schema, SalesFiscalRequestFieldOriginalId)
	if field.IsRequiredForCreate() {
		t.Error("original_fiscal_request_id must be optional: an ISSUE_ORIGINAL adjusts nothing")
	}
}

// enumSetOf reads the permitted values of an enum field as a set, for membership assertions.
//
// A thin wrapper over the package's enumValuesOf, which returns a slice: these tests assert
// presence and absence rather than order, and a set says that plainly.
func enumSetOf(t *testing.T, schema *dmodel.ModelSchema, name string) map[string]bool {
	t.Helper()

	values := map[string]bool{}
	for _, value := range enumValuesOf(t, fieldOf(t, schema, name)) {
		values[value] = true
	}
	if len(values) == 0 {
		t.Fatalf("%s parsed with no enum values; the field type or the reader changed", name)
	}
	return values
}

// The outbox's event id is unique and immutable, because it is what consumers deduplicate on. An
// event that could change its identity could be processed twice under two names, which defeats the
// only mechanism making at-least-once delivery safe (acceptance 94.34).
func TestOutboxEventIdIsUniqueAndImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesIntegrationOutboxSchemaBuilder())

	field := fieldOf(t, schema, SalesOutboxFieldEventId)
	if !field.IsRequiredForCreate() {
		t.Error("event_id must be required: a consumer with nothing to deduplicate on cannot " +
			"tolerate the redelivery this design guarantees")
	}
	if !field.IsNoUpdate() {
		t.Error("event_id must be no_update: an event that changed identity could be processed " +
			"twice under two names")
	}

	found := false
	for _, unique := range schema.CompositeUniques() {
		if len(unique.Fields) == 1 && unique.Fields[0] == SalesOutboxFieldEventId {
			found = true
		}
	}
	if !found {
		t.Error("event_id must carry a unique index")
	}
}

// published_at is nullable, and that IS the queue. Required, every row would be born published and
// the sweep would have nothing to find.
func TestPublishedAtIsNullableBecauseItIsTheQueue(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesIntegrationOutboxSchemaBuilder())

	if fieldOf(t, schema, SalesOutboxFieldPublishedAt).IsRequiredForCreate() {
		t.Error("published_at must be nullable: an unpublished row is the queue, and a required " +
			"column would mean every event was born already delivered")
	}

	// occurred_at, by contrast, is always known: it is when the business event happened, which is
	// the moment the row is written.
	if !fieldOf(t, schema, SalesOutboxFieldOccurredAt).IsRequiredForCreate() {
		t.Error("occurred_at must be required: consumers order by it, and an event with no " +
			"business time cannot be placed in the sequence")
	}
}

// The payload is immutable. An event is a statement about a moment; editing one rewrites history
// that consumers have already acted on, and the ones that already consumed it never find out.
func TestTheOutboxPayloadIsImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesIntegrationOutboxSchemaBuilder())
	for _, name := range []string{
		SalesOutboxFieldPayload,
		SalesOutboxFieldEventType,
		SalesOutboxFieldAggregateId,
		SalesOutboxFieldOccurredAt,
	} {
		if !fieldOf(t, schema, name).IsNoUpdate() {
			t.Errorf("%s must be no_update: an event states what happened at a moment, and "+
				"editing it rewrites history consumers have already acted on", name)
		}
	}
}

// A quotation's number is its own, unique and immutable, and NOT derived from the order sequence.
// Deriving it would either leave holes in that sequence when a quotation lapses, or reuse numbers —
// and fiscal systems read the order sequence, so the holes are not cosmetic.
func TestQuotationNumberIsIndependentOfTheOrderSequence(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesQuotationSchemaBuilder())

	field := fieldOf(t, schema, SalesQuotationFieldNumber)
	if !field.IsRequiredForCreate() {
		t.Error("quotation_number must be required")
	}
	if !field.IsNoUpdate() {
		t.Error("quotation_number must be no_update: it is what the customer quotes back")
	}

	found := false
	for _, unique := range schema.CompositeUniques() {
		if len(unique.Fields) == 1 && unique.Fields[0] == SalesQuotationFieldNumber {
			found = true
		}
	}
	if !found {
		t.Error("quotation_number must be unique")
	}
}

// The quotation line deliberately LACKS the order line's fulfilment columns. A quotation moves no
// goods, and a fulfilled_quantity on an offer would be a number nothing could ever set.
func TestQuotationLinesCarryNoFulfilmentColumns(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesQuotationLineSchemaBuilder())

	for _, absent := range []string{
		"fulfilled_quantity", "returned_quantity", "requires_fulfillment",
	} {
		if _, present := schema.Fields()[absent]; present {
			t.Errorf("a quotation line must not carry %q: a quotation moves no goods, so the "+
				"column would be one nothing could ever set", absent)
		}
	}

	// And it must carry what an offer actually needs.
	for _, present := range []string{
		SalesQuotationLineFieldQuantity,
		SalesQuotationLineFieldUnitPrice,
		SalesQuotationLineFieldFinalAmount,
	} {
		if _, declared := schema.Fields()[present]; !declared {
			t.Errorf("a quotation line must carry %q", present)
		}
	}
}

// converted_sales_order_id is nullable and carries NO foreign key. Nullable because most quotations
// never convert; no FK so that an order removed by a retention sweep does not take its quotation's
// history with it.
func TestTheConversionLinkIsOptional(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesQuotationSchemaBuilder())

	if fieldOf(t, schema, SalesQuotationFieldConvertedOrder).IsRequiredForCreate() {
		t.Error("converted_sales_order_id must be optional: a quotation is created before it is " +
			"accepted, and most are never accepted at all")
	}
}
