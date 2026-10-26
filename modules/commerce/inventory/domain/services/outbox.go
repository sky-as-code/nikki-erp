package services

import (
	"sort"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itMessage "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/message"
)

// Writing integration events.
//
// RecordEvent must be called inside the caller's transaction: it does no transaction management of
// its own, so called outside one it silently degrades to the commit-then-publish race the outbox
// exists to avoid. Nothing here touches the broker — the row is the handoff and app/ drains it.

// OutboxDrain publishes whatever is waiting in the outbox, now.
//
// It is a package-level hook rather than an injected dependency for the same reason the cron
// scoping strategy is (modules/core/job.SetSweepScoper): the writers here are package functions
// with no container behind them, and draining is a deployment concern, not one any single write
// path should have to be handed.
type OutboxDrain func(ctx corectx.Context)

// outboxDrain is nil until app/ installs one, and a nil drain is a working system: the cron sweep
// is what guarantees delivery, so a build that never installs this is simply the slower one.
var outboxDrain OutboxDrain

// SetOutboxDrain installs the drain, from the module's OnAppStarted.
func SetOutboxDrain(drain OutboxDrain) {
	if drain != nil {
		outboxDrain = drain
	}
}

// DrainOutboxNow asks for the events just written to go out without waiting for the next sweep.
//
// It returns immediately; the drain itself is somebody else's goroutine.
//
// A call from inside a transaction does nothing, deliberately rather than as a guard against
// misuse: the rows are not committed yet, so a drain reading through that same transaction would
// publish an event whose write may still roll back. That is the commit-then-publish race the outbox
// exists to avoid, arriving from the other direction. Those events wait for the sweep, which is the
// guarantee in every case anyway.
func DrainOutboxNow(ctx corectx.Context) {
	if outboxDrain == nil || ctx.GetDbTranx() != nil {
		return
	}
	outboxDrain(ctx)
}

// RecordEventParams is one integration event to publish.
type RecordEventParams struct {
	// EventType is one of the constants in domain/models, named after what happened rather than what
	// a consumer should do about it.
	EventType string

	// AggregateId is the reservation the event concerns.
	AggregateId string

	// OrgId scopes the row. Taken explicitly rather than from the context, because the org of the
	// record the caller already loaded is the authoritative one.
	OrgId string

	// Payload is the event's facts, self-contained so a consumer never reads back into Inventory.
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

	engineRepo, err := repoFor(models.InventoryIntegrationOutboxSchemaName)
	if err != nil {
		return "", err
	}

	record := dmodel.DynamicFields{
		models.InventoryOutboxFieldId:            string(*id),
		models.InventoryOutboxFieldEventId:       string(*eventId),
		models.InventoryOutboxFieldAggregateId:   params.AggregateId,
		models.InventoryOutboxFieldEventType:     params.EventType,
		models.InventoryOutboxFieldSchemaVersion: models.InventoryEventSchemaVersion,
		models.InventoryOutboxFieldPayload:       payloadOrEmpty(params.Payload),
		models.InventoryOutboxFieldOccurredAt:    model.ModelDateTime(time.Unix(occurredAt, 0).UTC()),

		basemodel.FieldOrgId: params.OrgId,
	}

	if _, err := engineRepo.Insert(ctx, record); err != nil {
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
// never delivered before the creation it releases.
//
// Sorting happens here rather than in the query because RepoSearchParam carries no sort. Which rows
// land in the page is still the repository's choice, so a backlog larger than one page can drain
// slightly out of order across sweeps; consumers must be idempotent and order by occurred_at anyway.
func UnpublishedEvents(ctx corectx.Context, limit int) ([]dmodel.DynamicFields, error) {
	engineRepo, err := repoFor(models.InventoryIntegrationOutboxSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		models.InventoryOutboxFieldPublishedAt, dmodel.IsNotSet))

	found, err := engineRepo.Search(ctx, dyn.RepoSearchParam{
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
	at := dateTimeOf(record, models.InventoryOutboxFieldOccurredAt)
	if at == nil {
		return 0
	}
	return at.GoTime().UnixNano()
}

// MarkEventPublished stamps a row that reached the broker. Called after a successful publish, never
// before: marking first would lose any event whose publish then failed, while marking after can only
// republish one, which consumers deduplicate on event_id.
func MarkEventPublished(ctx corectx.Context, rowId string) error {
	engineRepo, err := repoFor(models.InventoryIntegrationOutboxSchemaName)
	if err != nil {
		return err
	}
	_, err = engineRepo.Update(ctx, dmodel.DynamicFields{
		models.InventoryOutboxFieldId:          rowId,
		models.InventoryOutboxFieldPublishedAt: model.ModelDateTime(time.Now().UTC()),
	})
	return err
}

// RecordPublishFailure counts a failed attempt and keeps the reason. The row stays unpublished so the
// next sweep retries it, and attempt_count is incremented here so a retry loop cannot forget to count.
func RecordPublishFailure(ctx corectx.Context, row dmodel.DynamicFields, message string) error {
	engineRepo, err := repoFor(models.InventoryIntegrationOutboxSchemaName)
	if err != nil {
		return err
	}
	_, err = engineRepo.Update(ctx, dmodel.DynamicFields{
		models.InventoryOutboxFieldId:           stringOf(row, models.InventoryOutboxFieldId),
		models.InventoryOutboxFieldAttemptCount: int32Of(row, models.InventoryOutboxFieldAttemptCount) + 1,
		models.InventoryOutboxFieldLastError:    truncateError(message),
	})
	return err
}

// OutboxEvent is a stored event ready to publish, paired with the row it came from. The row id stays
// outside the wire event: it is Inventory storage, while the event id is the public contract.
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
	payload, _ := row[models.InventoryOutboxFieldPayload].(map[string]any)

	return OutboxEvent{
		RowId: stringOf(row, models.InventoryOutboxFieldId),
		IntegrationEvent: itMessage.IntegrationEvent{
			EventId:       stringOf(row, models.InventoryOutboxFieldEventId),
			EventType:     stringOf(row, models.InventoryOutboxFieldEventType),
			AggregateId:   stringOf(row, models.InventoryOutboxFieldAggregateId),
			SchemaVersion: stringOf(row, models.InventoryOutboxFieldSchemaVersion),
			OccurredAt:    occurredAtOf(row) / int64(time.Second),
			OrgId:         stringOf(row, basemodel.FieldOrgId),
			Payload:       payload,
		},
	}
}
