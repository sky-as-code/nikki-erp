package services

import (
	"sort"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
)

// Writing integration events.
//
// RecordEvent must be called inside the caller's transaction: it does no transaction management of
// its own, so called outside one it silently degrades to the commit-then-publish race the outbox
// exists to avoid. Nothing here touches the broker — the row is the handoff and app/ drains it.

// RecordEventParams is one integration event to publish.
type RecordEventParams struct {
	// EventType is one of the constants in domain/models, named after what happened rather than what
	// a consumer should do about it.
	EventType string

	// AggregateId is the order, bill or return the event concerns.
	AggregateId string

	// OrgId scopes the row. Taken explicitly rather than from the context, because the org of the
	// record the caller already loaded is the authoritative one.
	OrgId string

	// Payload is the event's facts, self-contained so a consumer never reads back into Sales.
	Payload map[string]any

	// OccurredAt is when the business event happened; zero defaults to now, so a backfill must pass it.
	OccurredAt int64
}

// RecordEvent writes one integration event into the outbox. Must be called inside the caller's
// transaction.
func RecordEvent(ctx corectx.Context, params RecordEventParams) (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", err
	}

	// The event id is generated separately from the row id: consumers deduplicate on the event id, so
	// tying it to a storage identifier would make a change of storage a change of the public contract.
	eventId, err := model.NewId()
	if err != nil {
		return "", err
	}

	occurredAt := params.OccurredAt
	if occurredAt == 0 {
		occurredAt = NowUnix()
	}

	engine, err := engineFor(models.SalesIntegrationOutboxSchemaName)
	if err != nil {
		return "", err
	}

	record := dmodel.DynamicFields{
		models.SalesOutboxFieldId:            string(*id),
		models.SalesOutboxFieldEventId:       string(*eventId),
		models.SalesOutboxFieldAggregateId:   params.AggregateId,
		models.SalesOutboxFieldEventType:     params.EventType,
		models.SalesOutboxFieldSchemaVersion: models.SalesEventSchemaVersion,
		models.SalesOutboxFieldPayload:       payloadOrEmpty(params.Payload),
		models.SalesOutboxFieldOccurredAt:    model.ModelDateTime(time.Unix(occurredAt, 0).UTC()),

		basemodel.FieldOrgId: params.OrgId,
	}

	if _, err := engine.ResourceRepository().Insert(ctx, record); err != nil {
		return "", err
	}
	return string(*eventId), nil
}

// payloadOrEmpty guarantees a non-nil map: the column is required, and a nil map would fail the write
// and take the business transaction down with it, though carrying no extra facts is legitimate.
func payloadOrEmpty(payload map[string]any) map[string]any {
	if payload == nil {
		return map[string]any{}
	}
	return payload
}

// UnpublishedEvents reads the events still waiting for the broker, oldest first, so a cancellation is
// never delivered before the confirmation it cancels.
//
// Sorting happens here rather than in the query because RepoSearchParam carries no sort. Which rows
// land in the page is still the repository's choice, so a backlog larger than one page can drain
// slightly out of order across sweeps; consumers must be idempotent and order by occurred_at anyway.
func UnpublishedEvents(ctx corectx.Context, limit int) ([]dmodel.DynamicFields, error) {
	engine, err := engineFor(models.SalesIntegrationOutboxSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		models.SalesOutboxFieldPublishedAt, dmodel.IsNotSet))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  limit,
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	events := found.Data.Items
	sort.SliceStable(events, func(first, second int) bool {
		return occurredAtOf(events[first]) < occurredAtOf(events[second])
	})
	return events, nil
}

// occurredAtOf reads an event's business timestamp as a comparable number. Absent sorts first, since
// sending an unstamped row earliest is the reading that cannot reorder events around it.
func occurredAtOf(record dmodel.DynamicFields) int64 {
	at := dateTimeOf(record, models.SalesOutboxFieldOccurredAt)
	if at == nil {
		return 0
	}
	return at.GoTime().UnixNano()
}

// MarkEventPublished stamps a row that reached the broker. Called after a successful publish, never
// before: marking first would lose any event whose publish then failed, while marking after can only
// republish one, which consumers deduplicate on event_id.
func MarkEventPublished(ctx corectx.Context, rowId string) error {
	engine, err := engineFor(models.SalesIntegrationOutboxSchemaName)
	if err != nil {
		return err
	}
	_, err = engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.SalesOutboxFieldId:          rowId,
		models.SalesOutboxFieldPublishedAt: model.ModelDateTime(time.Now().UTC()),
	})
	return err
}

// RecordPublishFailure counts a failed attempt and keeps the reason. The row stays unpublished so the
// next sweep retries it, and attempt_count is incremented here so a retry loop cannot forget to count.
func RecordPublishFailure(ctx corectx.Context, row dmodel.DynamicFields, message string) error {
	engine, err := engineFor(models.SalesIntegrationOutboxSchemaName)
	if err != nil {
		return err
	}
	_, err = engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.SalesOutboxFieldId:           stringOf(row, models.SalesOutboxFieldId),
		models.SalesOutboxFieldAttemptCount: int32Of(row, models.SalesOutboxFieldAttemptCount) + 1,
		models.SalesOutboxFieldLastError:    truncateError(message),
	})
	return err
}

// OutboxEvent is a stored event ready to publish, paired with the row it came from. The row id stays
// outside the wire event: it is Sales storage, while the event id is the public contract.
type OutboxEvent struct {
	itMessage.IntegrationEvent

	// RowId identifies the outbox row to mark once the broker has taken the event.
	RowId string
}

// IntegrationEventOf turns a stored row into the event that goes over the wire. Fields are read
// defensively rather than type-asserted, because a row that went through a jsonb column arrives as
// whatever the JSON decoder chose; a malformed row degrades to empty fields the broker refuses,
// rather than panicking the sweep and stranding the page.
func IntegrationEventOf(row dmodel.DynamicFields) OutboxEvent {
	payload, _ := row[models.SalesOutboxFieldPayload].(map[string]any)

	return OutboxEvent{
		RowId: stringOf(row, models.SalesOutboxFieldId),
		IntegrationEvent: itMessage.IntegrationEvent{
			EventId:       stringOf(row, models.SalesOutboxFieldEventId),
			EventType:     stringOf(row, models.SalesOutboxFieldEventType),
			AggregateId:   stringOf(row, models.SalesOutboxFieldAggregateId),
			SchemaVersion: stringOf(row, models.SalesOutboxFieldSchemaVersion),
			OccurredAt:    occurredAtOf(row) / int64(time.Second),
			OrgId:         stringOf(row, basemodel.FieldOrgId),
			Payload:       payload,
		},
	}
}
