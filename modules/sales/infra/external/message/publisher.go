// Package message speaks the broker's protocol on behalf of Sales' integration event port.
//
// The ONLY package in Sales that knows events are JSON over a topic. Everything above depends on
// interfaces/message, so changing the encoding or the broker changes this file and nothing else.
package message

import (
	"context"
	"encoding/json"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"

	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
)

// topicPrefix namespaces Sales' events on the broker.
//
// One topic PER EVENT TYPE rather than a single sales topic, so a consumer subscribes to what it
// cares about instead of receiving everything and discarding most of it. An accounting consumer
// wanting payments should not have to decode every fulfilment event to find out it does not want it.
const topicPrefix = "sales/events/"

// TopicOf names the topic one event type is published on.
//
// Exported because a consumer needs to know what to subscribe to, and deriving the name in two
// places is how a publisher and a subscriber come to disagree about it silently.
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

// wireEvent is the envelope as it appears on the broker.
//
// A struct with explicit JSON tags rather than marshalling the port type directly, because this
// shape is a PUBLIC CONTRACT with consumers Sales does not control: a Go field renamed for internal
// tidiness must not silently rename a JSON key that somebody's parser depends on. The separation is
// what makes schema_version meaningful rather than decorative.
type wireEvent struct {
	EventId       string `json:"event_id"`
	EventType     string `json:"event_type"`
	AggregateId   string `json:"aggregate_id"`
	SchemaVersion string `json:"schema_version"`
	OccurredAt    int64  `json:"occurred_at"`
	OrgId         string `json:"org_id"`

	Payload map[string]any `json:"payload"`
}

// Publish sends one event.
//
// An encoding failure and a broker failure are both returned as errors, and the caller treats them
// the same way: the row stays unpublished and the next sweep retries. They are not the same
// underneath — a malformed payload will fail identically forever while a broker recovers — which is
// what attempt_count and last_error on the row exist to let an operator tell apart.
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
