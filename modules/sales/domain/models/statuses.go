package models

// The status values of the lifecycle-bearing Sales resources, and the enums that go with them.
//
// They are declared here rather than inline at their use sites so that the schema JSON and the code
// reading it cannot drift: a typo in a comparison would otherwise be a condition that is silently
// never true, which no compiler catches.
//
// Stored lower-case, matching the strings the JSON schemas declare.

// SalesChannelStatus is the business lifecycle of a sales channel.
//
// It is deliberately separate from is_archived, which is the system lifecycle. CR §9 and §10 keep
// them apart because they answer different questions: suspended means "not selling right now",
// archived means "no longer part of the catalogue of channels". Collapsing them would lose the
// difference between a channel paused for the season and one retired for good.
type SalesChannelStatus string

const (
	// SalesChannelStatusActive permits new sales points, new orders and integration requests.
	SalesChannelStatusActive = SalesChannelStatus("active")
	// SalesChannelStatusSuspended stops all three, while leaving reads, returns, refunds and
	// fiscal adjustments of existing transactions working.
	SalesChannelStatusSuspended = SalesChannelStatus("suspended")
)

// SalesPointStatus is the business lifecycle of a sales point.
type SalesPointStatus string

const (
	// SalesPointStatusActive permits new sales orders at this point.
	SalesPointStatusActive = SalesPointStatus("active")
	// SalesPointStatusSuspended stops new orders but keeps history, returns and refunds available
	// — what a temporarily offline kiosk needs (CR §14, §52).
	SalesPointStatusSuspended = SalesPointStatus("suspended")
)
