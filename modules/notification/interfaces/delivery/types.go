package delivery

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// SendNotificationRequest is the one shape both entry paths carry: a direct application service
// call and a command bus message (BR 11, BR 12). Optional fields are pointers so that "omitted"
// and "empty" stay distinguishable — the difference decides distribution_mode (BR 11.5, BR 11.6).
type SendNotificationRequest struct {
	// OrgId is the organization the notification belongs to. It is checked against the caller's
	// membership before anything is written, so naming one the caller has no business in is
	// refused rather than honoured (BR 11, BR 29). A direct caller passes the org it is already
	// acting in; a command passes the one on the command.
	OrgId string

	// RecipientUserIds must hold at least one user, and duplicates collapse to one recipient
	// (BR 11.1, BR 11.2).
	RecipientUserIds []string

	Title   string
	Message string

	// Severity defaults to "info" when empty (BR 11.4).
	Severity string

	// Channels nil means "every channel enabled for the organization"; a non-nil empty slice is
	// invalid rather than a synonym for it (BR 11.7).
	Channels []string

	// ChannelArgs carries per-channel arguments, keyed by channel name. Each channel reads only
	// the object stored under its own name and the core reads none of them (BR 5).
	//
	// Every key must name an attached channel that this notification actually goes to; both are
	// checked at send, where the sender is still there to be told.
	ChannelArgs map[string]map[string]any

	SourceModule       string
	SourceResourceName *string
	SourceResourceKey  map[string]any

	Metadata  map[string]any
	ExpiresAt *string

	// IdempotencyKey makes a repeat send return the first notification instead of raising a
	// second one (BR 11.15). The command handler defaults it to the command id (BR 12).
	IdempotencyKey *string
}

// SendNotificationResultData is returned once the notification is durable, and deliberately not
// once it has been delivered: waiting on a channel would make a send fail for reasons that are
// not the sender's and not recoverable by it (BR 11 Return, BR 32).
type SendNotificationResultData struct {
	NotificationId   model.Id
	CreatedAt        string
	RecipientUserIds []string

	// RecipientIds are the recipient rows just written, in the same order as RecipientUserIds. A
	// delivery row points at one of these, so the fan-out would otherwise have to read back rows
	// the send has just created.
	RecipientIds []model.Id

	ResolvedChannels []string

	// ChannelArgs is what the sender wrote, carried through so the fan-out can hand each channel
	// its own object without reading the notification back.
	ChannelArgs map[string]map[string]any

	// Duplicate reports that an existing notification was returned rather than a new one created.
	// The caller sees success either way; this is for the observability counter (BR 33).
	Duplicate bool
}

// InboxItem is one notification as its recipient sees it. It joins the notification to that
// person's own recipient row, which is where read state lives (BR 7).
type InboxItem struct {
	NotificationId     model.Id
	RecipientId        model.Id
	StreamSeq          int64
	Title              string
	Message            string
	Severity           string
	SourceModule       string
	SourceResourceName *string
	SourceResourceKey  map[string]any
	Metadata           map[string]any
	CreatedAt          string

	// ReadAt is nil while unread. Calculated into IsRead for the client rather than stored twice.
	ReadAt *string
}

// IsRead is the calculated field the API exposes (BR 7).
func (this InboxItem) IsRead() bool {
	return this.ReadAt != nil
}

// MarkReadResultData reports what a bulk mark-read actually changed, which a caller cannot infer
// from the request: ids belonging to another user are silently not updated (BR 21).
type MarkReadResultData struct {
	RequestedCount   int
	UpdatedCount     int
	AlreadyReadCount int
}
