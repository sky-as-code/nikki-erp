package cqrs

import (
	"context"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// requestContextOf builds the context a command is executed under.
//
// The bus carries no principal: a command is a message, not an authenticated HTTP request, so the
// caller's identity does not survive the hop. The handler therefore acts as a SERVICE principal
// scoped to the organization named on the command.
//
// That principal is not a bypass. PrincipalKindService authorizes on the entitlements it was
// granted, exactly like a person, and the org on it is what mayActInOrg checks the command's
// organization against — so a command naming an organization this deployment has not granted
// Notification access to is refused, not honoured.
func requestContextOf(
	ctx context.Context, command *it.SendNotificationCommand,
) (corectx.Context, error) {
	orgId := strings.TrimSpace(command.OrgId)
	if orgId == "" {
		return nil, errors.New("SendNotificationCommand: org_id is required")
	}
	if strings.TrimSpace(command.CommandId) == "" {
		// Without it a redelivery cannot be recognised, and the bus would raise a second
		// notification for the same event (BR 12).
		return nil, errors.New("SendNotificationCommand: command_id is required")
	}

	org := model.Id(orgId)
	requestCtx := corectx.NewRequestContextM(ctx, modconstants.NotificationModuleName)
	requestCtx.SetPermissions(corectx.ContextPermissions{
		Principal: corectx.Principal{
			Kind:        corectx.PrincipalKindService,
			Id:          model.Id(modconstants.NotificationModuleName),
			OrgId:       &org,
			DisplayName: "notification command bus",
		},
	})
	return requestCtx, nil
}
