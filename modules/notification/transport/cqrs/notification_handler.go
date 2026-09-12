// Package cqrs receives the commands other modules dispatch into Notification.
//
// A handler here is thin by design: it maps the command onto the application service and turns the
// result into a reply. Every rule about what may be sent lives in the domain, so that a
// notification raised over the bus and one raised by a direct call behave identically (BR 2).
package cqrs

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

type NotificationHandler struct {
	notifications it.NotificationApplicationService
	logger        logging.LoggerService
}

func NewNotificationHandler(
	notifications it.NotificationApplicationService,
	logger logging.LoggerService,
) *NotificationHandler {
	return &NotificationHandler{notifications: notifications, logger: logger}
}

// SendNotification raises the notification a command asks for.
//
// The command id becomes the idempotency key when the sender supplied none, and that is what makes
// redelivery safe: the bus may deliver the same command more than once, and every delivery after
// the first finds the notification already there and returns it (BR 12).
//
// The reply is sent only after the notification is durable. Delivery to a channel is not waited
// for — a command that waited would fail for reasons the sender cannot act on (BR 12, BR 32).
func (this *NotificationHandler) SendNotification(
	ctx context.Context, packet *cqrs.RequestPacket[it.SendNotificationCommand],
) (*cqrs.Reply[it.SendNotificationCommandResult], error) {
	command := packet.Request()
	if command == nil {
		return nil, nil
	}

	requestCtx, err := requestContextOf(ctx, command)
	if err != nil {
		return nil, err
	}

	result, err := this.notifications.SendNotification(requestCtx, mapCommand(*command))
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		// A refused command is not a transport failure. Returning an error would have the bus
		// redeliver a command that will be refused identically every time.
		this.logger.Warn("notification command was refused", map[string]any{
			"command_id": command.CommandId,
			"errors":     result.ClientErrors,
		})
		return &cqrs.Reply[it.SendNotificationCommandResult]{}, nil
	}

	return &cqrs.Reply[it.SendNotificationCommandResult]{
		Result: it.SendNotificationCommandResult{
			NotificationId:   string(result.Data.NotificationId),
			CreatedAt:        result.Data.CreatedAt,
			Duplicate:        result.Data.Duplicate,
			ResolvedChannels: result.Data.ResolvedChannels,
		},
	}, nil
}

// mapCommand turns the wire shape into the one business request. It adds no rule of its own.
func mapCommand(command it.SendNotificationCommand) it.SendNotificationRequest {
	idempotencyKey := command.IdempotencyKey
	if idempotencyKey == nil {
		// The command id, so that a redelivery of this exact command is recognised as one
		// (BR 12). A sender with a natural key of its own overrides it.
		key := command.CommandId
		idempotencyKey = &key
	}

	return it.SendNotificationRequest{
		OrgId:              command.OrgId,
		RecipientUserIds:   command.RecipientUserIds,
		Title:              command.Title,
		Message:            command.Message,
		Severity:           command.Severity,
		Channels:           command.Channels,
		ChannelArgs:        command.ChannelArgs,
		SourceModule:       command.SourceModule,
		SourceResourceName: command.SourceResourceName,
		SourceResourceKey:  command.SourceResourceKey,
		Metadata:           command.Metadata,
		ExpiresAt:          command.ExpiresAt,
		IdempotencyKey:     idempotencyKey,
	}
}

// InitCqrsHandlers subscribes Notification's command handlers.
func InitCqrsHandlers() error {
	if err := deps.Register(NewNotificationHandler); err != nil {
		return err
	}
	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *NotificationHandler) error {
		return cqrsBus.SubscribeRequests(
			context.Background(),
			cqrs.NewHandler(handler.SendNotification),
		)
	})
}
