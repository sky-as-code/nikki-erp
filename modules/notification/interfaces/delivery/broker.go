package delivery

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
)

// StreamEvent is one line on a client's stream. It is the wire contract of BR-FS 6: every event
// carries a type the client may not recognise and must then ignore, so that new event types can be
// added without breaking a deployed client.
type StreamEvent struct {
	// Seq is the recipient's stream_seq. Absent on a heartbeat, which is why it is a pointer:
	// a heartbeat is not a position in the stream and must never advance the client's cursor
	// (BR-FS 9).
	Seq *int64 `json:"seq,omitempty"`

	Type string `json:"type"`

	Data any `json:"data,omitempty"`
}

// Event types carried on the stream.
const (
	StreamEventNotificationCreated = "notification.created"
	StreamEventHeartbeat           = "heartbeat"
)

// WakeUpSignal tells an instance that a user has something new waiting. It carries only enough to
// route: the instance then reads the notification from the database.
//
// THE PAYLOAD IS NOT THE NOTIFICATION and must never be treated as one (BR 17, BR-FS 18). The
// broker is a transport with no durability guarantee; the database is the source of truth. An
// instance that trusted this payload would serve whatever a lost or reordered message happened to
// contain.
type WakeUpSignal struct {
	OrgId  model.Id `json:"org_id"`
	UserId model.Id `json:"user_id"`

	// Seq lets a receiver skip a read it has already done. A hint, never a guarantee.
	Seq int64 `json:"seq"`
}

// RealtimeNotificationBroker wakes the backend instances that may be holding a stream for a user.
//
// It exists because a stream is held by ONE instance while the notification may be created on
// ANOTHER (BR 17). An in-memory registry alone is correct only while exactly one instance runs,
// and the service is published in ingress mode, so scaling past one is a one-line change that
// would otherwise silently break realtime delivery.
type RealtimeNotificationBroker interface {
	// Publish signals interest in a user, AFTER the notification has been committed. Publishing
	// inside the transaction would wake a reader that cannot yet see the row.
	Publish(ctx context.Context, signal WakeUpSignal) error

	// Subscribe returns every signal this instance receives, until the context is cancelled.
	Subscribe(ctx context.Context) (<-chan WakeUpSignal, error)
}

// StreamRegistry tracks the streams this instance is currently serving. It is deliberately
// per-instance and in-memory: a stream is a live HTTP response, which cannot be shared or handed
// to another process. The broker is what makes the set of instances behave as one.
type StreamRegistry interface {
	// Register returns the channel the handler writes to, and a function that removes it. The
	// channel is buffered; Dispatch drops rather than blocks when it is full (BR-FS 19).
	Register(orgId model.Id, userId model.Id, buffer int) (<-chan StreamEvent, func())

	// Dispatch offers an event to every stream this instance holds for the user. It reports how
	// many streams accepted it, and never blocks on a slow reader.
	Dispatch(orgId model.Id, userId model.Id, event StreamEvent) int

	// Count is the active stream total, for the observability metric of BR 33.
	Count() int

	// HasStreams reports whether this instance is serving anyone for that person, so a wake-up
	// can be dropped without a database read when nobody here is listening.
	HasStreams(orgId model.Id, userId model.Id) bool
}
