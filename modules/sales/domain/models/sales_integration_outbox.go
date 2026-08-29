package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The integration event outbox. A row written in the same transaction as the business change cannot
// disagree with it, whereas a direct publish either announces a sale that rolls back or loses one
// that happened. A separate sweep then moves rows to the broker, turning an atomicity problem into a
// retryable delivery one. Delivery is at-least-once — a sweep can publish and fail before marking —
// so consumers deduplicate on the stable event_id.
//
// This is the EXTERNAL kind of event, not the internal Watermill bus: its payloads are a public
// contract with consumers Sales does not control, which is why they are versioned and self-contained.

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

// SalesEventSchemaVersion is the shape every payload written today is in. A string rather than a
// counter so a consumer can accept 1.x and refuse 2.0: bump the minor for a field added, the major
// for one removed, renamed or given a new meaning.
const SalesEventSchemaVersion = "1.0"

// The integration event types, named after what happened in the past tense, never after what a
// consumer should do about it: an instruction-shaped name would be Sales coupling itself to another
// module, which is what the outbox exists to avoid.
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

// SalesEventTypes lists every event this module publishes. Exported so a test can assert the set; a
// type published but absent from this list is one no consumer was told to expect.
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

// GetPayload returns the event's facts. Typed as `any` because domain/models must not import another
// module, and a stored jsonmap comes back as whatever the JSON decoder chose.
func (this SalesIntegrationOutboxEntry) GetPayload() any {
	return this.GetFieldData().GetAny(SalesOutboxFieldPayload)
}

// IsPublished reports whether the event has reached the broker. An unpublished row is the queue:
// there is no separate status column, because a status and a timestamp could disagree about the
// same fact.
func (this SalesIntegrationOutboxEntry) IsPublished() bool {
	return this.GetPublishedAt() != nil
}
