// Package delivery declares the contracts for raising and reading notifications.
//
// The send command arrives over the CQRS bus rather than as an HTTP call, because a source module
// runs in the same process and a loopback request would need a routable URL and an authenticated
// caller for work that has neither.
package delivery

import (
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// The compile guard: assigning the command to the interface here turns a missing CqrsRequestType
// into a build error rather than a runtime panic on the first dispatch, when the sender is already
// live and nobody is watching.
func init() {
	var _ cqrs.Request = (*SendNotificationCommand)(nil)
}

// The BR names this command "notification.send.v1". That exact string is not expressible here:
// RequestType renders as "{module}_{submodule}.{action}" and the scheduler's executor validates
// that shape. Versioning therefore lives in the payload, not in the name.
var sendNotificationCommandType = cqrs.RequestType{
	Module: "notification", Submodule: "delivery", Action: "sendNotification",
}

// SendNotificationCommand asks Notification to raise a notification for one or more users.
//
// CommandId is required and is what makes redelivery safe: the bus may deliver the same command
// more than once, and the same command id must yield the same notification rather than a second
// one (BR 12). It becomes the idempotency key when the sender supplies none of its own.
//
// OrgId is accepted here, unlike on the direct call, because a command may be dispatched by a
// background worker that has no user context. It is still validated against the caller's
// principal before anything is written (BR 11).
type SendNotificationCommand struct {
	CommandId    string `json:"command_id"`
	OrgId        string `json:"org_id"`
	SourceModule string `json:"source_module"`

	RecipientUserIds []string `json:"recipients"`

	Title    string `json:"title"`
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"`

	// Channels omitted means every enabled channel; an empty array is rejected (BR 11.7).
	Channels []string `json:"channels,omitempty"`

	// ChannelArgs holds one argument object per channel, keyed by channel name (BR 5).
	ChannelArgs map[string]map[string]any `json:"channel_args,omitempty"`

	SourceResourceName *string        `json:"source_resource_name,omitempty"`
	SourceResourceKey  map[string]any `json:"source_resource_key,omitempty"`

	Metadata  map[string]any `json:"metadata,omitempty"`
	ExpiresAt *string        `json:"expires_at,omitempty"`

	// IdempotencyKey overrides the default of CommandId, for a sender that already has a natural
	// key of its own.
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

func (SendNotificationCommand) CqrsRequestType() cqrs.RequestType {
	return sendNotificationCommandType
}

// SendNotificationCommandResult reports what the send produced. Duplicate distinguishes "created"
// from "this command had already been handled", which is the signal a redelivery happened.
type SendNotificationCommandResult struct {
	NotificationId string `json:"notification_id"`
	CreatedAt      string `json:"created_at"`
	Duplicate      bool   `json:"duplicate"`

	ResolvedChannels []string `json:"resolved_channels"`
}
