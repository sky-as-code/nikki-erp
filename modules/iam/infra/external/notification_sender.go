package external

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itNotif "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// notificationSender reaches Notification over the command bus, the contract it offers source
// modules; iam does not depend on the Notification module being loaded.
type notificationSender struct {
	bus cqrs.CqrsBus
}

func (this *notificationSender) NotifyUsers(ctx corectx.Context, cmd itExt.NotifyUsersCommand) error {
	request := &itNotif.SendNotificationCommand{
		CommandId:        cmd.CommandId,
		OrgId:            string(cmd.OrgId),
		SourceModule:     c.IamModuleName,
		RecipientUserIds: make([]string, 0, len(cmd.RecipientUserIds)),
		Title:            cmd.Title,
		Message:          cmd.Message,
		Severity:         cmd.Severity,
		Metadata:         cmd.Metadata,
	}
	for _, userId := range cmd.RecipientUserIds {
		request.RecipientUserIds = append(request.RecipientUserIds, string(userId))
	}
	if len(request.RecipientUserIds) == 0 {
		return errors.New("notification has no recipient")
	}
	if !this.bus.IsRequestTypeRegistered(request.CqrsRequestType().String()) {
		return errors.New("the notification module is not loaded")
	}

	result := itNotif.SendNotificationCommandResult{}
	if err := this.bus.Request(ctx, request, &result); err != nil {
		return errors.Wrap(err, "send notification")
	}
	// Notification answers a refused command with an empty reply rather than an error.
	if result.NotificationId == "" {
		return errors.New("notification was refused; see the notification module's log")
	}
	return nil
}
