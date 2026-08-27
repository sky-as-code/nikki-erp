// Package message is the port for Sales' OUTBOUND integration events (BR 80).
//
// This is one of two messaging systems in the codebase and they must not be conflated:
//
//   - This one LEAVES THE SERVER. It publishes to the external broker for consumers Sales does not
//     control, so its payloads are a public contract: versioned, self-contained, and not changed
//     without a version bump.
//   - modules/core/event is INTERNAL. It carries "something changed" between parts of one module
//     over the Watermill bus and never reaches an external consumer.
//
// See the package comment on vending_machine_new/interfaces/message, which draws the same line for
// the same reason.
//
// # Nothing publishes directly through this port from a domain service
//
// A domain service writes an outbox ROW inside its own transaction; the sweep in app/ is what calls
// this. That indirection is the whole point of BR 80 — see the package comment on
// domain/models/sales_integration_outbox.go for why publishing inline cannot be made correct.
package message

import (
	"context"
)

// IntegrationEventPublisher pushes one event to the external broker.
//
// It takes a plain context.Context rather than corectx.Context, and for the reason
// OrderPaymentResultPublisher does in vending: the sweep runs in the background, long after the
// request that produced the event committed and returned, so there is no request context left to
// carry. An event may also be published by a retry hours later, when the original request's actor,
// org and locale are meaningless.
type IntegrationEventPublisher interface {
	// Publish sends one event. An error means the broker did not take it, and the caller leaves the
	// row unpublished so the next sweep tries again — which is why this returns an error rather than
	// swallowing one: a publisher that reported success it did not have would silently drop events.
	Publish(ctx context.Context, event IntegrationEvent) error
}

// IntegrationEvent is one event as it goes over the wire.
//
// The envelope and the payload are separate fields rather than one merged map, so a consumer can
// route on the type and check the version WITHOUT parsing a payload it may not understand. Merging
// them would mean a consumer had to decode the thing it was trying to decide whether it could
// decode.
type IntegrationEvent struct {
	// EventId is stable across republication, and is what a consumer deduplicates on. Delivery is
	// at-least-once by construction — a sweep that publishes and then fails before marking the row
	// will publish again — so this is what makes acceptance 94.34's idempotency requirement
	// satisfiable rather than merely asked for.
	EventId string

	EventType string

	// AggregateId is what the event is about, carried in the envelope as well as the payload so a
	// consumer can partition or filter on it without decoding.
	AggregateId string

	// SchemaVersion tells a consumer which shape Payload is in, so one written today can refuse a
	// payload from a future version rather than misreading it.
	SchemaVersion string

	// OccurredAt is when the BUSINESS event happened, not when it was published. The two differ by
	// however long the row waited, and a consumer ordering by publication time would reorder events
	// the business produced in sequence.
	OccurredAt int64

	// OrgId scopes the event. Carried explicitly because the background sweep has no request context
	// to derive it from.
	OrgId string

	// Payload is the event's facts, self-contained by design: a consumer must not have to read back
	// into Sales to interpret it, both because that would couple the two at read time and because
	// the answer would be today's state rather than the state the event described.
	Payload map[string]any
}
