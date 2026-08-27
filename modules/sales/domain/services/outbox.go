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

// Writing integration events (BR 80, acceptance 94.34, SALES-037).
//
// # RecordEvent must be called INSIDE the caller's transaction
//
// That is the entire mechanism, and it is easy to lose by accident. A row written in the same
// transaction as the business change cannot disagree with it: either the sale confirmed and the
// event exists, or neither happened. Called outside, this degrades to exactly the commit-then-
// publish race the outbox was built to avoid, and the degradation is silent — the code still runs,
// the events still mostly arrive, and the ones lost are lost only when something else fails.
//
// So RecordEvent takes the context it is given and does no transaction management of its own. A
// caller already inside withTransaction gets the guarantee; a caller outside one does not.
//
// # The sweep publishes, and this file does not
//
// Nothing here touches the broker. The row is the handoff, and app/ drains it — see
// interfaces/message for the port and why it takes a plain context.

// RecordEventParams is one integration event to publish.
type RecordEventParams struct {
	// EventType is one of the constants in domain/models. Named after what HAPPENED, never after
	// what a consumer should do about it.
	EventType string

	// AggregateId is the order, bill or return the event concerns.
	AggregateId string

	// OrgId scopes the row. Taken explicitly rather than from the context, because the caller has
	// already loaded the record the event is about and its org is the authoritative one.
	OrgId string

	// Payload is the event's facts. Self-contained: a consumer must not have to read back into Sales.
	Payload map[string]any

	// OccurredAt is when the business event happened. Defaulted to now when zero, which is right for
	// an event recorded as it happens and wrong for one backfilled — so a backfill passes it.
	OccurredAt int64
}

// RecordEvent writes one integration event into the outbox.
//
// MUST be called inside the caller's transaction. See the package comment.
func RecordEvent(ctx corectx.Context, params RecordEventParams) (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", err
	}

	// The event id is separate from the row id, and generated separately. They could be the same
	// value today, but a consumer deduplicates on the event id, so tying it to a storage identifier
	// would make a change of storage a change of the public contract.
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

// payloadOrEmpty guarantees a non-nil map.
//
// The column is required, and a nil map would fail the write — turning "this event carried no extra
// facts", which is legitimate, into a failure of the business transaction that produced it. An event
// with an empty payload still tells a consumer that something happened and to what.
func payloadOrEmpty(payload map[string]any) map[string]any {
	if payload == nil {
		return map[string]any{}
	}
	return payload
}

// UnpublishedEvents reads the events still waiting for the broker, oldest first.
//
// Oldest first because consumers see business order that way. Publishing newest-first would deliver
// a cancellation before the confirmation it cancels, and a consumer reconstructing state from the
// stream would end up with the sale still open.
//
// The ordering is applied HERE rather than in the query, because RepoSearchParam carries no sort.
// Sorting one bounded page in Go is exact for what it returns and cheap at this size — but note the
// consequence: which rows land in the page is still the repository's choice, so a backlog larger
// than one page can be drained slightly out of order across sweeps. That is acceptable because
// consumers must be idempotent and event-ordered by occurred_at anyway (BR 80 carries the timestamp
// for exactly this), and unacceptable to fix by widening the page until it holds everything.
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

// occurredAtOf reads an event's business timestamp as a comparable number.
//
// Absent sorts first rather than last: a row whose timestamp did not survive the round trip is older
// than anything correctly stamped, and sending it earliest is the reading that cannot reorder events
// around it.
func occurredAtOf(record dmodel.DynamicFields) int64 {
	at := dateTimeOf(record, models.SalesOutboxFieldOccurredAt)
	if at == nil {
		return 0
	}
	return at.GoTime().UnixNano()
}

// MarkEventPublished stamps a row that reached the broker.
//
// Called AFTER a successful publish, never before. The ordering is the deliberate half of the
// at-least-once trade: marking first would lose any event whose publish then failed, while marking
// after can only republish one — and a duplicate is what the consumer's event_id deduplication
// already handles.
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

// RecordPublishFailure counts a failed attempt and keeps the reason.
//
// The row stays unpublished, so the next sweep retries it. attempt_count is incremented from what
// was read rather than left to the caller, so a retry loop cannot forget to count and an operator
// can tell a broker that is briefly unreachable from an event that will never go.
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

// OutboxEvent is a stored event ready to publish, paired with the row it came from.
//
// The row id is carried alongside rather than inside the wire event, and that separation is the
// point: the row id is Sales storage and the event id is the public contract, so a consumer never
// sees the former and a change of storage never changes the latter.
type OutboxEvent struct {
	itMessage.IntegrationEvent

	// RowId identifies the outbox row to mark once the broker has taken the event.
	RowId string
}

// IntegrationEventOf turns a stored row into the event that goes over the wire.
//
// Every field is read defensively rather than type-asserted, because a row that has been through a
// jsonb column and back arrives as whatever the JSON decoder chose - the same reason the shared
// readers in this package exist. A malformed row degrades to an event with empty fields, which a
// broker refuses and the sweep counts, rather than panicking the whole sweep and stranding the page.
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
