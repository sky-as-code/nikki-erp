package app

import (
	"context"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// NewNotificationApplicationService is handed the composable default by the notification onion.
func NewNotificationApplicationService(
	base composable.CrudApplicationService,
	broker it.RealtimeNotificationBroker,
	logger logging.LoggerService,
) it.NotificationApplicationService {
	domainSvc, ok := base.DomainService().(it.NotificationDomainService)
	if !ok {
		panic(errors.New("the notification onion must be built with NewNotificationDomainService"))
	}
	return &NotificationApplicationServiceImpl{
		CrudApplicationService: base,
		domainSvc:              domainSvc,
		broker:                 broker,
		logger:                 logger,
	}
}

type NotificationApplicationServiceImpl struct {
	composable.CrudApplicationService

	domainSvc it.NotificationDomainService
	broker    it.RealtimeNotificationBroker
	logger    logging.LoggerService
}

// SendNotification is the single business use case (BR 2). The direct caller lands here, and so
// does the command handler; neither has a copy of the rules.
//
// It is deliberately not reachable over REST (BR 29): sending is something a module does on a
// user's behalf, and exposing it would let anyone write into anyone else's inbox.
func (this *NotificationApplicationServiceImpl) SendNotification(
	ctx corectx.Context, request it.SendNotificationRequest,
) (*it.SendNotificationResult, error) {
	// AssertAction resolves the organization AND checks membership, so a caller cannot name an
	// organization it has no business writing into. The org travels in params because that is
	// where ResolveOrgScope reads it from.
	params := dmodel.DynamicFields{basemodel.FieldOrgId: request.OrgId}
	orgId, cErrs := this.AssertAction(ctx, modconstants.ActionSend, params)
	if cErrs != nil {
		return &it.SendNotificationResult{ClientErrors: *cErrs}, nil
	}
	if orgId == nil {
		errs := ft.ClientErrors{*ft.NewValidationError(
			basemodel.FieldOrgId, "err_org_id_required",
			"a notification belongs to an organization")}
		return &it.SendNotificationResult{ClientErrors: errs}, nil
	}

	result, err := this.domainSvc.Send(ctx, *orgId, request)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}

	// After the commit, never inside it (BR 11.13): an instance woken earlier would look for a
	// row that its own transaction cannot see yet and find nothing.
	this.announce(ctx, *orgId, result.Data)
	return result, nil
}

// announce wakes whichever instances hold a stream for these people.
//
// A failure here is logged and swallowed on purpose. The notification is already durable, and the
// client reconciles through the inbox or an after_seq reconnect, so failing the send would turn a
// missed realtime push into a lost business action (BR 11.14, BR 32).
func (this *NotificationApplicationServiceImpl) announce(
	ctx corectx.Context, orgId model.Id, data it.SendNotificationResultData,
) {
	if data.Duplicate {
		// A replay created nothing, so there is nothing new to wake anyone for.
		return
	}

	for _, userId := range data.RecipientUserIds {
		signal := it.WakeUpSignal{OrgId: orgId, UserId: model.Id(userId)}
		if err := this.broker.Publish(context.WithoutCancel(ctx), signal); err != nil {
			this.logger.Error("notification wake-up signal was not published", err)
		}
	}
}
