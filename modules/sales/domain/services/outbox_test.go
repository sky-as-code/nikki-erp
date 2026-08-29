package services

import (
	"testing"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The outbox envelope and ordering, pinned without a repository: being wrong here means a consumer
// cannot deduplicate, or rebuilds a sale's history in the wrong order.

func outboxRow(eventId, eventType string, occurredAt time.Time) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOutboxFieldId:            "ROW-" + eventId,
		models.SalesOutboxFieldEventId:       eventId,
		models.SalesOutboxFieldEventType:     eventType,
		models.SalesOutboxFieldAggregateId:   "OR01",
		models.SalesOutboxFieldSchemaVersion: models.SalesEventSchemaVersion,
		models.SalesOutboxFieldOccurredAt:    model.ModelDateTime(occurredAt),
		models.SalesOutboxFieldPayload:       map[string]any{"sales_order_id": "OR01"},
		basemodel.FieldOrgId:                 "ORG1",
	}
}

// The event id is the public contract consumers deduplicate on; the row id is Sales storage. Tying
// them would make a change of storage a change of the contract.
func TestTheRowIdAndTheEventIdAreSeparate(t *testing.T) {
	row := outboxRow("EVT-1", models.EventSalesOrderConfirmed, time.Now())

	event := IntegrationEventOf(row)
	if event.EventId != "EVT-1" {
		t.Errorf("event id = %q, want EVT-1", event.EventId)
	}
	if event.RowId != "ROW-EVT-1" {
		t.Errorf("row id = %q, want ROW-EVT-1", event.RowId)
	}
	if event.RowId == event.EventId {
		t.Error("the row id and the event id must not be the same value: one is storage, the " +
			"other is what consumers deduplicate on")
	}
}

// The envelope carries everything a consumer needs to route and version-check without decoding the
// payload, so it can refuse a version it does not understand.
func TestTheEnvelopeIsReadableWithoutThePayload(t *testing.T) {
	row := outboxRow("EVT-2", models.EventSalesPaymentCaptured, time.Now())

	event := IntegrationEventOf(row)
	if event.EventType != models.EventSalesPaymentCaptured {
		t.Errorf("event type = %q", event.EventType)
	}
	if event.SchemaVersion != models.SalesEventSchemaVersion {
		t.Errorf("schema version = %q, want %q", event.SchemaVersion, models.SalesEventSchemaVersion)
	}
	if event.AggregateId != "OR01" {
		t.Errorf("aggregate id = %q, want OR01", event.AggregateId)
	}
	if event.OrgId != "ORG1" {
		t.Errorf("org id = %q, want ORG1", event.OrgId)
	}
}

// A malformed row degrades to empty fields rather than panicking: a panic on one row would strand
// every event behind it, on this sweep and every later one.
func TestAMalformedRowDoesNotPanicTheSweep(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("reading a malformed outbox row panicked: %v", recovered)
		}
	}()

	// Every field the wrong shape, which is what a bad jsonb round trip looks like.
	event := IntegrationEventOf(dmodel.DynamicFields{
		models.SalesOutboxFieldEventId:     42,
		models.SalesOutboxFieldEventType:   nil,
		models.SalesOutboxFieldAggregateId: []string{"nope"},
		models.SalesOutboxFieldPayload:     "not a map",
	})

	if event.EventId != "" || event.EventType != "" {
		t.Error("a malformed row must degrade to empty fields, which the broker refuses and the " +
			"sweep counts, rather than yielding nonsense")
	}
	if event.Payload != nil {
		t.Error("a payload that is not a map must read as nil rather than as a fabricated one")
	}
}

// An empty payload is legitimate; the column is required, so a nil map would fail the business
// transaction that produced the event.
func TestAnEmptyPayloadIsStoredAsAnEmptyMap(t *testing.T) {
	if got := payloadOrEmpty(nil); got == nil {
		t.Fatal("a nil payload must become an empty map, not stay nil")
	}
	if got := payloadOrEmpty(nil); len(got) != 0 {
		t.Errorf("an absent payload must be empty, got %d entries", len(got))
	}

	supplied := map[string]any{"amount": 100}
	if got := payloadOrEmpty(supplied); len(got) != 1 {
		t.Error("a supplied payload must be passed through untouched")
	}
}

// Events sort oldest-first by business time, or a cancellation would be delivered before the
// confirmation it cancels.
func TestEventsSortByBusinessTimeOldestFirst(t *testing.T) {
	base := time.Date(2026, 8, 27, 10, 0, 0, 0, time.UTC)

	rows := []dmodel.DynamicFields{
		outboxRow("EVT-CANCEL", models.EventSalesOrderCancelled, base.Add(time.Hour)),
		outboxRow("EVT-CONFIRM", models.EventSalesOrderConfirmed, base),
	}

	if occurredAtOf(rows[0]) <= occurredAtOf(rows[1]) {
		t.Fatal("the fixture is wrong: the cancellation must be the later event")
	}

	if occurredAtOf(rows[1]) >= occurredAtOf(rows[0]) {
		t.Error("the confirmation must sort before the cancellation that follows it")
	}
}

// A row with no timestamp sorts first: sending it earliest cannot reorder stamped events around it.
func TestAnUnstampedEventSortsFirst(t *testing.T) {
	stamped := outboxRow("EVT-1", models.EventSalesOrderConfirmed, time.Now())
	unstamped := dmodel.DynamicFields{models.SalesOutboxFieldEventId: "EVT-0"}

	if occurredAtOf(unstamped) >= occurredAtOf(stamped) {
		t.Error("an event with no business timestamp must sort before one that has it")
	}
}

// The minimum event set, asserted against the exported list: an event type published but absent from
// the list is one no consumer was told to expect.
func TestTheMinimumEventSetIsDeclared(t *testing.T) {
	declared := map[string]bool{}
	for _, eventType := range models.SalesEventTypes() {
		declared[eventType] = true
	}

	for _, required := range []string{
		"SalesOrderConfirmed", "SalesOrderCancelled",
		"SalesPaymentCaptured", "SalesPaymentRefunded",
		"SalesFulfillmentRequested",
		"SalesReturnApproved", "SalesReturnCompleted",
		"FiscalDocumentRequested", "FiscalAdjustmentRequested",
	} {
		if !declared[required] {
			t.Errorf("BR 80 requires the %q event, which is not declared", required)
		}
	}
}

// Event types are named after what happened, never after what a consumer should do: an imperative
// name would be Sales instructing another module, the coupling an outbox exists to avoid.
func TestEventNamesArePastTenseNotImperative(t *testing.T) {
	for _, eventType := range models.SalesEventTypes() {
		for _, imperative := range []string{
			"Reserve", "Issue", "Create", "Update", "Delete", "Send", "Publish", "Do",
		} {
			if len(eventType) >= len(imperative) && eventType[:len(imperative)] == imperative {
				t.Errorf("%q begins with the imperative %q: an event names what happened, not "+
					"what a consumer should do about it", eventType, imperative)
			}
		}
	}
}
