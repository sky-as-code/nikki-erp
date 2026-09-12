package delivery

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type SendNotificationResult = dyn.OpResult[SendNotificationResultData]
type InboxResult = dyn.OpResult[InboxResultData]
type UnreadCountResult = dyn.OpResult[UnreadCountResultData]
type MarkReadResult = dyn.OpResult[MarkReadResultData]

// InboxResultData is one page of a user's inbox, newest first (BR 18).
type InboxResultData struct {
	Items []InboxItem

	// NextCursor is the stream_seq to pass back for the following page, or nil at the end.
	NextCursor *int64
}

type UnreadCountResultData struct {
	Count int
}

// InboxQuery reads the calling user's own notifications. It carries NO user id and NO org id on
// purpose: both come from the request context, so that no client can ask for another person's
// inbox by passing a different id (BR 18, BR 29).
type InboxQuery struct {
	IsRead       *bool
	Severity     *string
	SourceModule *string

	// Cursor is the stream_seq of the last item already seen; the page returned is strictly older
	// than it. Absent means the newest page.
	Cursor *int64
	Limit  int
}

// StreamQuery opens a stream for the calling user. AfterSeq replays everything newer than a
// sequence the client has already processed (BR-FS 10).
type StreamQuery struct {
	AfterSeq *int64
}

// StreamScope is whose notifications a stream carries. Both values come from the request
// context, never from the request itself.
type StreamScope struct {
	OrgId  model.Id
	UserId model.Id
}

// MarkReadCommand marks one or many notifications read for the calling user only.
type MarkReadCommand struct {
	NotificationIds []string
}

// NotificationRepository reads and writes notification rows.
type NotificationRepository interface {
	composable.CrudRepository

	// FindByIdempotencyKey returns the id of an existing notification raised under this key, or
	// nil when there is none.
	FindByIdempotencyKey(
		ctx corectx.Context, orgId model.Id, sourceModule string, key string,
	) (*model.Id, error)
}

// RecipientRepository reads and writes recipient rows, and answers the inbox queries. The inbox
// reads join notification to recipient, which is why they live here rather than on the
// notification repository.
type RecipientRepository interface {
	composable.CrudRepository

	NextStreamSeq(ctx corectx.Context) (int64, error)
	Inbox(ctx corectx.Context, orgId model.Id, userId model.Id, query InboxQuery) ([]InboxItem, error)
	Replay(ctx corectx.Context, orgId model.Id, userId model.Id, afterSeq *int64, limit int) ([]InboxItem, error)
	UnreadCount(ctx corectx.Context, orgId model.Id, userId model.Id) (int, error)
	MarkRead(ctx corectx.Context, orgId model.Id, userId model.Id, notificationIds []model.Id) (MarkReadResultData, error)
}

// DeliveryRepository reads and writes delivery rows.
type DeliveryRepository interface {
	composable.CrudRepository
}

// NotificationDomainService owns the send rules of BR 11, and is the only writer of a
// notification and its recipients.
type NotificationDomainService interface {
	composable.CrudDomainService

	// Send writes the notification. orgId is resolved and authorized by the application
	// service before it gets here: a domain service never decides which organization it is
	// acting in.
	Send(ctx corectx.Context, orgId model.Id, request SendNotificationRequest) (*SendNotificationResult, error)
}

// NotificationApplicationService is the authorized surface. SendNotification is the single
// business use case both entry paths reach (BR 2): the command handler maps onto it rather than
// reimplementing anything.
type NotificationApplicationService interface {
	composable.CrudApplicationService

	SendNotification(ctx corectx.Context, request SendNotificationRequest) (*SendNotificationResult, error)
}

// InboxApplicationService is what a signed-in person calls about their own notifications. Every
// method resolves the user and organization from the context (BR 29).
type InboxApplicationService interface {
	GetInbox(ctx corectx.Context, query InboxQuery) (*InboxResult, error)
	GetUnreadCount(ctx corectx.Context) (*UnreadCountResult, error)
	MarkRead(ctx corectx.Context, command MarkReadCommand) (*MarkReadResult, error)

	// Replay answers the stream's catch-up read. It is on the application service, not the
	// repository, because it must assert the caller may read this stream at all.
	Replay(ctx corectx.Context, query StreamQuery, limit int) (*InboxResult, error)

	// AuthorizeStream resolves whose stream the caller may open, refusing when they may not open
	// one at all. It exists so the stream handler authorizes through the same path as every other
	// inbox read instead of inspecting the context itself.
	AuthorizeStream(ctx corectx.Context) (*StreamScope, error)
}
