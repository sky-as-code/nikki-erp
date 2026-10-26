// Package message speaks the broker's protocol for Inventory's integration event port. It is the only
// package in Inventory that knows events are JSON over a topic, so changing the encoding or the broker
// changes this file and nothing else.
package message

import (
	"context"
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"

	itMessage "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/message"
)

// topicPrefix namespaces Inventory's events on the broker. There is one topic per event type rather than
// a single inventory topic, so a consumer subscribes to what it cares about instead of decoding
// everything.
const topicPrefix = "inventory/events/"

// TopicOf names the topic one event type is published on. Exported so a consumer subscribes by
// calling this rather than deriving the name separately and silently disagreeing.
func TopicOf(eventType string) string {
	return topicPrefix + eventType
}

// AppPublisher publishes Inventory integration events to the broker.
type AppPublisher struct {
	publisher pubsub.Publisher
}

func NewPublisher(publisher pubsub.Publisher) *AppPublisher {
	return &AppPublisher{publisher: publisher}
}

// wireEvent is the envelope as it appears on the broker. It has explicit JSON tags rather than
// marshalling the port type directly, because this shape is a public contract: renaming a Go field
// must not rename a JSON key a consumer's parser depends on.
type wireEvent struct {
	EventId       string `json:"event_id"`
	EventType     string `json:"event_type"`
	AggregateId   string `json:"aggregate_id"`
	SchemaVersion string `json:"schema_version"`
	OccurredAt    int64  `json:"occurred_at"`
	OrgId         string `json:"org_id"`

	// Scope is what the publishing context was scoped by - in a multi-tenant deployment, the tenant
	// whose outbox row this is. It sits on the ENVELOPE, once, rather than inside each event type's
	// payload: a consumer subscribes on a goroutine with no request behind it, so every event that
	// crosses this broker needs it, and an event type that had to remember to carry it is an event
	// type that will eventually forget.
	//
	// The internal event bus carries the same thing in message metadata (coreEvent.EventDelivery).
	// It cannot here: this broker's messages are raw bytes with nowhere to put metadata.
	//
	// Omitted entirely in a single-tenant build, where there is nothing to scope by.
	Scope dmodel.DynamicFields `json:"scope,omitempty"`

	Payload map[string]any `json:"payload"`
}

// scopeOf reads the publishing context's scope.
//
// A plain context.Context is enough: the constraints live in the inner context's values, so this
// sees them whether it is handed a corectx.Context or the detached context the sweep derives from
// one.
func scopeOf(ctx context.Context) dmodel.DynamicFields {
	constraints, ok := ctx.Value(corectx.CtxKeyDomainConstraints).(dmodel.DynamicFields)
	if !ok || len(constraints) == 0 {
		return nil
	}
	return constraints
}

// Publish sends one event. Encoding and broker failures both return an error and the row stays
// unpublished for the next sweep, but a malformed payload fails forever while a broker recovers —
// attempt_count and last_error on the row are what let an operator tell them apart.
func (this *AppPublisher) Publish(
	ctx context.Context, event itMessage.IntegrationEvent,
) error {
	encoded, err := json.Marshal(wireEvent{
		EventId:       event.EventId,
		EventType:     event.EventType,
		AggregateId:   event.AggregateId,
		SchemaVersion: event.SchemaVersion,
		OccurredAt:    event.OccurredAt,
		OrgId:         event.OrgId,
		Scope:         scopeOf(ctx),
		Payload:       event.Payload,
	})
	if err != nil {
		return errors.Wrapf(err, "encoding the inventory integration event '%s'", event.EventId)
	}

	return errors.Wrapf(
		this.publisher.Publish(ctx, TopicOf(event.EventType), encoded),
		"publishing the inventory integration event '%s'", event.EventId)
}
