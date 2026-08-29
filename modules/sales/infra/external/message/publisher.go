// Package message speaks the broker's protocol for Sales' integration event port. It is the only
// package in Sales that knows events are JSON over a topic, so changing the encoding or the broker
// changes this file and nothing else.
package message

import (
	"context"
	"encoding/json"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"

	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
)

// topicPrefix namespaces Sales' events on the broker. There is one topic per event type rather than
// a single sales topic, so a consumer subscribes to what it cares about instead of decoding
// everything.
const topicPrefix = "sales/events/"

// TopicOf names the topic one event type is published on. Exported so a consumer subscribes by
// calling this rather than deriving the name separately and silently disagreeing.
func TopicOf(eventType string) string {
	return topicPrefix + eventType
}

// AppPublisher publishes Sales integration events to the broker.
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

	Payload map[string]any `json:"payload"`
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
		Payload:       event.Payload,
	})
	if err != nil {
		return errors.Wrapf(err, "encoding the sales integration event '%s'", event.EventId)
	}

	return errors.Wrapf(
		this.publisher.Publish(ctx, TopicOf(event.EventType), encoded),
		"publishing the sales integration event '%s'", event.EventId)
}
