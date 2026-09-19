package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The integration event outbox, the same mechanism Sales uses: a row written in the same
// transaction as the stock change cannot disagree with it, and a separate sweep moves rows to the
// broker. Delivery is at-least-once, so consumers deduplicate on the stable event_id.

const (
	InventoryIntegrationOutboxSchemaName = "inventory_integration_outbox"

	InventoryOutboxFieldId            = basemodel.FieldId
	InventoryOutboxFieldOrgId         = basemodel.FieldOrgId
	InventoryOutboxFieldEventId       = "event_id"
	InventoryOutboxFieldAggregateId   = "aggregate_id"
	InventoryOutboxFieldEventType     = "event_type"
	InventoryOutboxFieldSchemaVersion = "schema_version"
	InventoryOutboxFieldPayload       = "payload"
	InventoryOutboxFieldOccurredAt    = "occurred_at"
	InventoryOutboxFieldPublishedAt   = "published_at"
	InventoryOutboxFieldAttemptCount  = "attempt_count"
	InventoryOutboxFieldLastError     = "last_error"
)

// InventoryEventSchemaVersion is the shape every payload written today is in. Bump the minor for
// a field added, the major for one removed, renamed or given a new meaning.
const InventoryEventSchemaVersion = "1.0"

// Reservation lifecycle events, named after what happened. Expiry carries both the deadline it
// took effect at (effective_at) and when it was written (recorded_at): a consumer must decide
// validity by the former, never by when the message arrived.
const (
	EventInventoryReservationCreated           = "InventoryReservationCreated"
	EventInventoryReservationPartiallyConsumed = "InventoryReservationPartiallyConsumed"
	EventInventoryReservationConsumed          = "InventoryReservationConsumed"
	EventInventoryReservationReleased          = "InventoryReservationReleased"
	EventInventoryReservationExpired           = "InventoryReservationExpired"
	EventInventoryReservationPaidProtected     = "InventoryReservationPaidProtected"
)

// InventoryEventTypes lists every event this module publishes, so a test can assert the set.
func InventoryEventTypes() []string {
	return []string{
		EventInventoryReservationCreated,
		EventInventoryReservationPartiallyConsumed,
		EventInventoryReservationConsumed,
		EventInventoryReservationReleased,
		EventInventoryReservationExpired,
		EventInventoryReservationPaidProtected,
	}
}

//go:embed inventory_integration_outbox.json
var inventoryIntegrationOutboxSchemaJson string

func InventoryIntegrationOutboxSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(inventoryIntegrationOutboxSchemaJson)
}

// InventoryIntegrationOutboxEntry is one event awaiting, or past, publication.
type InventoryIntegrationOutboxEntry struct {
	basemodel.DynamicModelBase
}

func NewInventoryIntegrationOutboxEntryFrom(src dmodel.DynamicFields) *InventoryIntegrationOutboxEntry {
	return &InventoryIntegrationOutboxEntry{basemodel.NewDynamicModel(src)}
}

func (this InventoryIntegrationOutboxEntry) GetId() *model.Id {
	return this.GetFieldData().GetModelId(InventoryOutboxFieldId)
}

func (this InventoryIntegrationOutboxEntry) GetEventId() *string {
	return this.GetFieldData().GetString(InventoryOutboxFieldEventId)
}

func (this InventoryIntegrationOutboxEntry) GetEventType() *string {
	return this.GetFieldData().GetString(InventoryOutboxFieldEventType)
}

func (this InventoryIntegrationOutboxEntry) GetAggregateId() *model.Id {
	return this.GetFieldData().GetModelId(InventoryOutboxFieldAggregateId)
}

func (this InventoryIntegrationOutboxEntry) GetSchemaVersion() *string {
	return this.GetFieldData().GetString(InventoryOutboxFieldSchemaVersion)
}

func (this InventoryIntegrationOutboxEntry) GetPublishedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(InventoryOutboxFieldPublishedAt)
}

// GetPayload is typed as `any` because a stored jsonmap comes back as whatever the JSON decoder
// chose, and this package must not import the module that interprets it.
func (this InventoryIntegrationOutboxEntry) GetPayload() any {
	return this.GetFieldData().GetAny(InventoryOutboxFieldPayload)
}

// IsPublished reports whether the event has reached the broker; an unpublished row is the queue.
func (this InventoryIntegrationOutboxEntry) IsPublished() bool {
	return this.GetPublishedAt() != nil
}
