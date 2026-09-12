package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// NewNotificationDomainService is handed the composable default by the notification onion.
func NewNotificationDomainService(
	base composable.CrudDomainService,
	notifications it.NotificationRepository,
	recipients it.RecipientRepository,
	deliveries it.DeliveryRepository,
	normalizer *SendNormalizer,
) it.NotificationDomainService {
	return &NotificationDomainServiceImpl{
		CrudDomainService: base,
		notifications:     notifications,
		recipients:        recipients,
		deliveries:        deliveries,
		normalizer:        normalizer,
	}
}

type NotificationDomainServiceImpl struct {
	composable.CrudDomainService

	notifications it.NotificationRepository
	recipients    it.RecipientRepository
	deliveries    it.DeliveryRepository
	normalizer    *SendNormalizer
}

// Send raises one notification for one or more people.
//
// The whole of BR 11's ordering lives here: validate, then look for a replay, then write the
// notification and every recipient in ONE transaction (BR 11.11). Nothing external is called
// inside it (BR 11.12) — the wake-up signal is published by the application service after the
// commit, because a reader woken before the commit would find nothing.
func (this *NotificationDomainServiceImpl) Send(
	ctx corectx.Context, orgId model.Id, request it.SendNotificationRequest,
) (*it.SendNotificationResult, error) {
	normalized, vErrs := this.normalizer.NormalizeSend(ctx, orgId, request)
	if len(vErrs) > 0 {
		return &it.SendNotificationResult{ClientErrors: vErrs}, nil
	}

	if existing, err := this.findReplay(ctx, orgId, request.SourceModule, normalized); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	return this.persist(ctx, orgId, request, normalized)
}

// findReplay returns the earlier notification when this send has already been handled.
//
// It is the fast half of the replay guard. The slow half is the partial unique index, which
// catches the two-deliveries-at-once race this lookup cannot see (BR 11.15, BR 12).
func (this *NotificationDomainServiceImpl) findReplay(
	ctx corectx.Context, orgId model.Id, sourceModule string, normalized *NormalizedSend,
) (*it.SendNotificationResult, error) {
	if normalized.IdempotencyKey == nil {
		return nil, nil
	}

	existingId, err := this.notifications.FindByIdempotencyKey(
		ctx, orgId, strings.TrimSpace(sourceModule), *normalized.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existingId == nil {
		return nil, nil
	}

	return &it.SendNotificationResult{
		HasData: true,
		Data: it.SendNotificationResultData{
			NotificationId:   *existingId,
			RecipientUserIds: idStrings(normalized.RecipientUserIds),
			ResolvedChannels: channelStrings(normalized.ResolvedChannels),
			Duplicate:        true,
		},
	}, nil
}

// persist writes the notification and its recipients as one unit.
//
// A recipient row without its notification, or a notification nobody receives, are both worse than
// no notification at all: the first is unreadable and the second is a silent loss. One transaction
// is what rules both out (BR 11.10, BR 11.11).
func (this *NotificationDomainServiceImpl) persist(
	ctx corectx.Context, orgId model.Id, request it.SendNotificationRequest, normalized *NormalizedSend,
) (*it.SendNotificationResult, error) {
	tranx, err := this.notifications.BeginTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Send")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	notificationId, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "Send")
	}

	notification := buildNotification(*notificationId, orgId, request, normalized)
	if _, err := this.notifications.Insert(tranxCtx, notification.GetFieldData()); err != nil {
		// A unique violation here is the race the lookup could not see: the other delivery of the
		// same command won. Its notification is the one that exists, so this call reports it
		// rather than failing (BR 12).
		if replay, replayErr := this.recoverFromRace(ctx, orgId, request, normalized, err); replayErr != nil {
			return nil, replayErr
		} else if replay != nil {
			return replay, nil
		}
		return nil, errors.Wrap(err, "Send")
	}

	recipientIds, err := this.insertRecipients(tranxCtx, orgId, *notificationId, normalized)
	if err != nil {
		return nil, err
	}

	if err := tranx.Commit(); err != nil {
		return nil, errors.Wrap(err, "Send")
	}

	return &it.SendNotificationResult{
		HasData: true,
		Data: it.SendNotificationResultData{
			NotificationId:   *notificationId,
			CreatedAt:        model.NewModelDateTime().String(),
			RecipientUserIds: idStrings(normalized.RecipientUserIds),
			RecipientIds:     recipientIds,
			ResolvedChannels: channelStrings(normalized.ResolvedChannels),
			ChannelArgs:      request.ChannelArgs,
		},
	}, nil
}

// recoverFromRace turns a lost insert race into the winner's result.
//
// Only a replay can be recovered this way: without an idempotency key there is no earlier
// notification to point at, and the error is a real one.
func (this *NotificationDomainServiceImpl) recoverFromRace(
	ctx corectx.Context,
	orgId model.Id,
	request it.SendNotificationRequest,
	normalized *NormalizedSend,
	cause error,
) (*it.SendNotificationResult, error) {
	if normalized.IdempotencyKey == nil {
		return nil, nil
	}
	replay, err := this.findReplay(ctx, orgId, request.SourceModule, normalized)
	if err != nil {
		// Report the original failure rather than this one: the insert is what went wrong, and
		// the recovery attempt failing on top of it is noise.
		return nil, errors.Wrap(cause, "Send")
	}
	return replay, nil
}

// insertRecipients writes one row per person and returns the ids it minted, in the order the
// recipients were given, so that the fan-out can address them without reading them back.
func (this *NotificationDomainServiceImpl) insertRecipients(
	ctx corectx.Context, orgId model.Id, notificationId model.Id, normalized *NormalizedSend,
) ([]model.Id, error) {
	recipientIds := make([]model.Id, 0, len(normalized.RecipientUserIds))

	for _, userId := range normalized.RecipientUserIds {
		seq, err := this.recipients.NextStreamSeq(ctx)
		if err != nil {
			return nil, err
		}

		recipientId, err := model.NewId()
		if err != nil {
			return nil, errors.Wrap(err, "Send")
		}
		recipientIds = append(recipientIds, *recipientId)

		recipient := models.NewRecipient()
		recipient.SetId(recipientId)
		recipient.SetOrgId(&orgId)
		recipient.SetNotificationId(&notificationId)
		recipient.SetRecipientUserId(&userId)
		recipient.SetStreamSeq(&seq)

		if _, err := this.recipients.Insert(ctx, recipient.GetFieldData()); err != nil {
			return nil, errors.Wrap(err, "Send")
		}
	}
	return recipientIds, nil
}

func buildNotification(
	id model.Id, orgId model.Id, request it.SendNotificationRequest, normalized *NormalizedSend,
) *models.Notification {
	notification := models.NewNotification()
	notification.SetId(&id)
	notification.SetOrgId(&orgId)

	sourceModule := strings.TrimSpace(request.SourceModule)
	notification.SetSourceModule(&sourceModule)
	notification.SetSourceResourceName(request.SourceResourceName)
	notification.SetSourceResourceKey(request.SourceResourceKey)
	notification.SetMetadata(request.Metadata)

	notification.SetTitle(&normalized.Title)
	notification.SetMessage(&normalized.Message)
	notification.SetSeverity(&normalized.Severity)
	notification.SetDistributionMode(&normalized.DistributionMode)
	notification.SetIdempotencyKey(normalized.IdempotencyKey)

	if len(normalized.Channels) > 0 {
		notification.SetRequestedChannels(map[string]any{
			"channels": channelStrings(normalized.Channels),
		})
	}
	if len(request.ChannelArgs) > 0 {
		// Stored whole, exactly as it arrived. Keeping the arguments of a channel that is not a
		// destination this time costs nothing and leaves the record faithful to what was sent.
		args := make(map[string]any, len(request.ChannelArgs))
		for name, channelArgs := range request.ChannelArgs {
			args[name] = channelArgs
		}
		notification.SetChannelArgs(args)
	}
	if request.ExpiresAt != nil {
		if expiresAt, err := model.ParseModelDateTime(*request.ExpiresAt); err == nil {
			notification.SetExpiresAt(&expiresAt)
		}
	}
	return notification
}

func idStrings(ids []model.Id) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

func channelStrings(channels []modconstants.ChannelName) []string {
	out := make([]string, 0, len(channels))
	for _, channel := range channels {
		out = append(out, channel.String())
	}
	return out
}
