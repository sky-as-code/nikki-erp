// Package message is the port for Sales' outbound integration events. It must not be confused with
// modules/core/event, which is internal to the server; this one leaves it, so its payloads are a
// public contract: versioned, self-contained, and not changed without a version bump.
//
// No domain service publishes through this port directly. A domain service writes an outbox row
// inside its own transaction and the sweep in app/ calls this.
package message

import (
	"context"
)

// IntegrationEventPublisher pushes one event to the external broker. It takes a plain
// context.Context because the sweep runs in the background, long after the request that produced the
// event committed, so there is no request context left to carry.
type IntegrationEventPublisher interface {
	// Publish sends one event. An error means the broker did not take it, and the caller leaves the
	// row unpublished so the next sweep tries again; swallowing it would silently drop events.
	Publish(ctx context.Context, event IntegrationEvent) error
}

// IntegrationEvent is one event as it goes over the wire. Envelope and payload are separate fields
// so a consumer can route on the type and check the version without parsing a payload it may not
// understand.
type IntegrationEvent struct {
	// EventId is stable across republication and is what a consumer deduplicates on. Delivery is
	// at-least-once by construction: a sweep that publishes and then fails before marking the row
	// publishes again.
	EventId string

	EventType string

	// AggregateId is what the event is about, carried in the envelope as well as the payload so a
	// consumer can partition or filter on it without decoding.
	AggregateId string

	// SchemaVersion tells a consumer which shape Payload is in, so one written today can refuse a
	// payload from a future version rather than misreading it.
	SchemaVersion string

	// OccurredAt is when the business event happened, not when it was published: a consumer ordering
	// by publication time would reorder events the business produced in sequence.
	OccurredAt int64

	// OrgId scopes the event. Carried explicitly because the background sweep has no request context
	// to derive it from.
	OrgId string

	// Payload is self-contained by design: a consumer reading back into Sales to interpret it would
	// couple the two at read time and would see today's state, not the state the event described.
	Payload map[string]any
}
