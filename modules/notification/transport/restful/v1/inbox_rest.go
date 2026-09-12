package v1

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// InboxRest serves a signed-in person their own notifications.
//
// None of these routes accepts a user id. Whose inbox is being read is decided by the request
// context alone, which is what makes it impossible to ask for somebody else's (BR 18, BR 29).
type InboxRest struct {
	inbox it.InboxApplicationService
}

func NewInboxRest(inbox it.InboxApplicationService) *InboxRest {
	return &InboxRest{inbox: inbox}
}

func (this *InboxRest) GetInbox(echoCtx *echo.Context, payload map[string]any) error {
	query := parseInboxQuery(*echoCtx)

	call := func(ctx corectx.Context, _ dmodel.DynamicFields) (*dyn.OpResult[it.InboxResultData], error) {
		return this.inbox.GetInbox(ctx, query)
	}
	return composable.ServeAction(echoCtx, "get notification inbox", payload, call, toInboxPageDto)
}

func (this *InboxRest) GetUnreadCount(echoCtx *echo.Context, payload map[string]any) error {
	call := func(ctx corectx.Context, _ dmodel.DynamicFields) (*dyn.OpResult[it.UnreadCountResultData], error) {
		return this.inbox.GetUnreadCount(ctx)
	}
	return composable.ServeAction(echoCtx, "get unread notification count", payload, call, toUnreadCountDto)
}

// MarkOneRead marks the notification named in the path.
//
// It is the single-id form of MarkRead rather than its own code path, so that marking one and
// marking many cannot drift apart in their idempotency (BR 20.6).
func (this *InboxRest) MarkOneRead(echoCtx *echo.Context, payload map[string]any) error {
	notificationId := (*echoCtx).Param("id")

	call := func(ctx corectx.Context, _ dmodel.DynamicFields) (*dyn.OpResult[it.MarkReadResultData], error) {
		return this.inbox.MarkRead(ctx, it.MarkReadCommand{NotificationIds: []string{notificationId}})
	}
	return composable.ServeAction(echoCtx, "mark notification read", payload, call, toMarkReadDto)
}

func (this *InboxRest) MarkRead(echoCtx *echo.Context, payload map[string]any) error {
	ids := parseNotificationIds(payload)

	call := func(ctx corectx.Context, _ dmodel.DynamicFields) (*dyn.OpResult[it.MarkReadResultData], error) {
		return this.inbox.MarkRead(ctx, it.MarkReadCommand{NotificationIds: ids})
	}
	return composable.ServeAction(echoCtx, "mark notifications read", payload, call, toMarkReadDto)
}

// parseInboxQuery reads the filters off the query string.
//
// An unreadable value is ignored rather than refused: these are filters, and the honest failure
// for a malformed one is to return the unfiltered page rather than an error the client cannot act
// on. Note there is no user_id here, and adding one would be a security bug, not a feature.
func parseInboxQuery(echoCtx echo.Context) it.InboxQuery {
	query := it.InboxQuery{}

	if raw := strings.TrimSpace(echoCtx.QueryParam("is_read")); raw != "" {
		if value, err := strconv.ParseBool(raw); err == nil {
			query.IsRead = &value
		}
	}
	if raw := strings.TrimSpace(echoCtx.QueryParam("severity")); raw != "" {
		query.Severity = &raw
	}
	if raw := strings.TrimSpace(echoCtx.QueryParam("source_module")); raw != "" {
		query.SourceModule = &raw
	}
	if raw := strings.TrimSpace(echoCtx.QueryParam("cursor")); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
			query.Cursor = &value
		}
	}
	if raw := strings.TrimSpace(echoCtx.QueryParam("limit")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			query.Limit = value
		}
	}
	return query
}

// parseNotificationIds reads the id list from the request body, accepting the JSON shapes a client
// actually sends: a list of strings, or of objects carrying an id.
func parseNotificationIds(payload map[string]any) []string {
	raw, found := payload["notification_ids"]
	if !found {
		return nil
	}
	values, ok := raw.([]any)
	if !ok {
		return nil
	}

	ids := make([]string, 0, len(values))
	for _, value := range values {
		if id, ok := value.(string); ok && strings.TrimSpace(id) != "" {
			ids = append(ids, strings.TrimSpace(id))
		}
	}
	return ids
}
