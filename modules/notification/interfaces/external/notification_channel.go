package external

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
)

// ChannelArgs is one channel's argument object, exactly as the sender wrote it.
//
// The core never indexes into this map. Only the channel named by the key it was stored under
// reads inside it, which is what lets a channel be attached or detached without the core learning
// any of its vocabulary (BR 5).
type ChannelArgs map[string]any

// DeliveryInput is everything a channel needs for one attempt at one recipient.
type DeliveryInput struct {
	OrgId           model.Id
	NotificationId  model.Id
	RecipientId     model.Id
	RecipientUserId model.Id

	Title    string
	Message  string
	Severity string

	SourceModule       string
	SourceResourceName *string
	SourceResourceKey  map[string]any
	Metadata           map[string]any

	// Args is the sub-map the sender stored under this channel's own name, or nil when it sent
	// none. A channel that requires an argument refuses in ValidateArgs, at send time, rather
	// than discovering the absence here where there is no caller left to tell.
	Args ChannelArgs

	AttemptNumber int
	Timeout       time.Duration
}

// DeliveryOutcome is what one attempt produced.
//
// It is returned instead of an error because every failure here is a fact to record on the
// delivery row, not a fault in the caller: a provider refusing a message is data, and propagating
// it as an error would let a channel fail a send that is already durable (BR 15, BR 32).
type DeliveryOutcome struct {
	Succeeded bool

	// Skipped means the attempt was deliberately not made -- no subscription, channel switched
	// off, notification already expired. It is kept apart from failure because the two are
	// different answers: a skip is correct behaviour and a failure is not, and BR 14 requires a
	// skipped channel to leave the notification successful.
	Skipped    bool
	SkipReason string

	// Retryable says only whether trying again could succeed. How many times is policy, and
	// belongs to the dispatcher, not here (BR 15).
	Retryable bool

	ErrorCode string

	// ErrorMessage must carry no credential, token or endpoint: these are written to a column
	// that is read back over the API, and BR 33 forbids them in logs for the same reason.
	ErrorMessage string
}

// NotificationChannel delivers one notification to one recipient over one transport.
//
// Implementations are registered in code rather than in a table (BR 4): attaching a channel is a
// new adapter and one line of wiring, and a row could not carry a Deliver.
type NotificationChannel interface {
	// Name is the discriminator throughout: the key in channel_args, a value in the requested
	// channels, and what is written to notification_deliveries.channel_name.
	Name() modconstants.ChannelName

	// ValidateArgs checks this channel's own arguments, at send time, before anything is
	// persisted. It returns field-scoped client errors so the sender is answered with the reason
	// rather than a 500.
	//
	// It is called with nil when the sender supplied no arguments for this channel, which is how
	// a channel with a required argument refuses. The core passes the map through untouched and
	// has no opinion about what is in it.
	ValidateArgs(ctx corectx.Context, args ChannelArgs) ft.ClientErrors

	// IsEnabledFor reports whether an organization has this channel switched on. Each channel
	// reads its own setting, so attaching one adds a setting without the core learning its name.
	//
	// A false answer produces a skipped delivery and never a failed one (BR 14).
	IsEnabledFor(ctx corectx.Context, orgId model.Id) (bool, error)

	// Deliver makes exactly one attempt.
	//
	// It takes a plain context rather than the request context because it runs after the commit,
	// detached from the request that caused it, for the reason given at the fan-out call site.
	Deliver(ctx context.Context, in DeliveryInput) DeliveryOutcome
}
