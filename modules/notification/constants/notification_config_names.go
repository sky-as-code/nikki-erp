package constants

import core "github.com/sky-as-code/nikki-erp/modules/core/constants"

// Application Configuration keys for Notification.
//
// Each key must also exist in config.default.yaml. The configuration service type-asserts its
// fallback argument, so a missing key panics rather than quietly falling back.
const (
	// StreamTopicPrefix is the pub/sub topic the wake-up signals are published under. A prefix
	// rather than one fixed topic, so that an operator can separate environments sharing a broker.
	StreamTopicPrefix core.ConfigName = "NOTIFICATION.PUBSUB.STREAM_TOPIC_PREFIX"

	// StreamHeartbeatIntervalSecs is the fallback heartbeat interval, used when an organization
	// has set none of its own. A gateway's idle timeout must stay well above it (BR-FS 9).
	StreamHeartbeatIntervalSecs core.ConfigName = "NOTIFICATION.STREAM.HEARTBEAT_INTERVAL_SECS"

	// StreamMaxBufferedEvents caps how far one slow client may fall behind before its stream is
	// closed (BR-FS 19). A system limit, not a per-organization preference: it protects the
	// instance's memory, so it is not overridable upward through settings.
	StreamMaxBufferedEvents core.ConfigName = "NOTIFICATION.STREAM.MAX_BUFFERED_EVENTS"

	// StreamReplayPageSize bounds one catch-up read, so that a client returning after a long
	// absence cannot ask the instance to materialize an unbounded backlog in one go.
	StreamReplayPageSize core.ConfigName = "NOTIFICATION.STREAM.REPLAY_PAGE_SIZE"

	// InboxDefaultPageSize and InboxMaxPageSize bound the REST inbox listing.
	InboxDefaultPageSize core.ConfigName = "NOTIFICATION.INBOX.DEFAULT_PAGE_SIZE"
	InboxMaxPageSize     core.ConfigName = "NOTIFICATION.INBOX.MAX_PAGE_SIZE"
)
