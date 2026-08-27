package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The integration event outbox (BR 80, acceptance 94.34).
//
// # Why a table and not a direct publish
//
// Publishing straight from a domain service makes the broker part of the transaction without any of
// the guarantees of one. The two failures are symmetric and both happen: publish-then-commit
// announces a sale that then rolls back, and commit-then-publish loses the announcement of a sale
// that really happened. Neither is detectable afterwards.
//
// A row written in the SAME transaction as the business change cannot disagree with it. Either both
// landed or neither did. A separate sweep then moves rows to the broker, which turns an atomicity
// problem into a delivery problem — and delivery problems are the ones retrying actually fixes.
//
// # At-least-once, therefore consumers deduplicate on event_id
//
// A sweep that publishes and then fails before marking the row will publish again. That is the
// correct trade: the alternative, marking before publishing, loses events. event_id is stable across
// those retries, which is what makes acceptance 94.34's idempotency requirement satisfiable by a
// consumer rather than merely asked of it.
//
// # This is the EXTERNAL kind of event
//
// The codebase runs two messaging systems and they must not be conflated (see the package comment on
// vending_machine_new/interfaces/message). The internal Watermill bus carries "something changed"
// between parts of one module and never leaves the server. This table feeds the external broker, and
// its payloads are a public contract with consumers Sales does not control — which is why they are
// versioned and self-contained.

const (
	SalesIntegrationOutboxSchemaName = "sales_integration_outbox"

	SalesOutboxFieldId            = basemodel.FieldId
	SalesOutboxFieldOrgId         = basemodel.FieldOrgId
	SalesOutboxFieldEventId       = "event_id"
	SalesOutboxFieldAggregateId   = "aggregate_id"
	SalesOutboxFieldEventType     = "event_type"
	SalesOutboxFieldSchemaVersion = "schema_version"
	SalesOutboxFieldPayload       = "payload"
	SalesOutboxFieldOccurredAt    = "occurred_at"
	SalesOutboxFieldPublishedAt   = "published_at"
	SalesOutboxFieldAttemptCount  = "attempt_count"
	SalesOutboxFieldLastError     = "last_error"
)

// SalesEventSchemaVersion is the shape every payload written today is in.
//
// A string like accounting's SnapshotSchemaVersion rather than a counter, so a compatible addition
// and a breaking change can be told apart: a consumer can accept 1.x and refuse 2.0. Bump the minor
// for a field added, the major for one removed, renamed or given a new meaning.
const SalesEventSchemaVersion = "1.0"

// The integration event types (BR 80).
//
// Named after what HAPPENED, in the past tense, never after what a consumer should do about it. An
// event called SalesOrderConfirmed can be consumed by anybody for any reason; one called
// ReserveStockForOrder would be Sales instructing another module, which is the coupling the outbox
// exists to avoid.
const (
	EventSalesOrderConfirmed  = "SalesOrderConfirmed"
	EventSalesOrderCancelled  = "SalesOrderCancelled"
	EventSalesPaymentCaptured = "SalesPaymentCaptured"
	EventSalesPaymentRefunded = "SalesPaymentRefunded"

	EventSalesFulfillmentRequested = "SalesFulfillmentRequested"

	EventSalesReturnApproved  = "SalesReturnApproved"
	EventSalesReturnCompleted = "SalesReturnCompleted"

	EventFiscalDocumentRequested   = "FiscalDocumentRequested"
	EventFiscalAdjustmentRequested = "FiscalAdjustmentRequested"
)

// SalesEventTypes lists every event this module publishes.
//
// Exported so a test can assert the set against BR 80's minimum rather than trusting that somebody
// remembered. A type published but absent from this list is one no consumer was told to expect.
func SalesEventTypes() []string {
	return []string{
		EventSalesOrderConfirmed,
		EventSalesOrderCancelled,
		EventSalesPaymentCaptured,
		EventSalesPaymentRefunded,
		EventSalesFulfillmentRequested,
		EventSalesReturnApproved,
		EventSalesReturnCompleted,
		EventFiscalDocumentRequested,
		EventFiscalAdjustmentRequested,
	}
}

//go:embed sales_integration_outbox.json
var salesIntegrationOutboxSchemaJson string

func SalesIntegrationOutboxSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesIntegrationOutboxSchemaJson)
}

// SalesIntegrationOutboxEntry is one event awaiting, or past, publication.
type SalesIntegrationOutboxEntry struct {
	basemodel.DynamicModelBase
}

func NewSalesIntegrationOutboxEntryFrom(src dmodel.DynamicFields) *SalesIntegrationOutboxEntry {
	return &SalesIntegrationOutboxEntry{basemodel.NewDynamicModel(src)}
}

func (this SalesIntegrationOutboxEntry) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOutboxFieldId)
}

func (this SalesIntegrationOutboxEntry) GetEventId() *string {
	return this.GetFieldData().GetString(SalesOutboxFieldEventId)
}

func (this SalesIntegrationOutboxEntry) GetEventType() *string {
	return this.GetFieldData().GetString(SalesOutboxFieldEventType)
}

func (this SalesIntegrationOutboxEntry) GetAggregateId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOutboxFieldAggregateId)
}

func (this SalesIntegrationOutboxEntry) GetSchemaVersion() *string {
	return this.GetFieldData().GetString(SalesOutboxFieldSchemaVersion)
}

func (this SalesIntegrationOutboxEntry) GetPublishedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesOutboxFieldPublishedAt)
}

// GetPayload returns the event's facts.
//
// Typed as `any` for the same reason the tax and buyer snapshots are: domain/models must not import
// another module, and a stored jsonmap comes back as whatever the JSON decoder chose.
func (this SalesIntegrationOutboxEntry) GetPayload() any {
	return this.GetFieldData().GetAny(SalesOutboxFieldPayload)
}

// IsPublished reports whether the event has reached the broker.
//
// The one question the sweep asks. An unpublished row IS the queue — there is no separate status
// column, because a status and a timestamp that could disagree about the same fact is one more thing
// to reconcile, and the timestamp is the one an operator actually wants to see.
func (this SalesIntegrationOutboxEntry) IsPublished() bool {
	return this.GetPublishedAt() != nil
}
