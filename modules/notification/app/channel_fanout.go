package app

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

// deliveryTimeout bounds one attempt. It is applied by the channel to whatever transport it uses;
// the fan-out only carries it, because how long is a policy and not a channel's to decide.
const deliveryTimeout = 30 * time.Second

// ChannelRegistry is the fan-out's view of the attached channels.
type ChannelRegistry interface {
	ChannelFor(name modconstants.ChannelName) itExt.NotificationChannel
}

// ChannelFanOut delivers a committed notification to each of its channels, and is the only writer
// of notification_deliveries.
//
// It runs after the commit, never inside it: an external call inside the transaction would hold a
// row lock for the length of a network timeout, and a rolled-back transaction would leave a message
// already delivered for a notification that does not exist (BR 11.12, BR 11.13).
type ChannelFanOut struct {
	channels   ChannelRegistry
	deliveries it.DeliveryRepository
	logger     logging.LoggerService
}

func NewChannelFanOut(
	channels ChannelRegistry, deliveries it.DeliveryRepository, logger logging.LoggerService,
) *ChannelFanOut {
	return &ChannelFanOut{
		channels:   channels,
		deliveries: deliveries,
		logger:     logger,
	}
}

// Deliver writes and drives one delivery per recipient per channel.
//
// Nothing here can fail the send. The notification is already durable, and a channel that cannot
// be reached is a delivery problem rather than a business one (BR 15, BR 32).
func (this *ChannelFanOut) Deliver(
	ctx corectx.Context, orgId model.Id, data it.SendNotificationResultData,
) {
	if data.Duplicate {
		// A replay created nothing. Delivering again would send a second message for a
		// notification the sender was told already existed (BR 12).
		return
	}

	// Detached from the request on purpose: the caller's context is cancelled as soon as it is
	// answered, and a delivery cut short by that would be recorded as a failure that never
	// happened. A nil context is not a caller's mistake to be punished with a panic -- this runs
	// after the answer, where nothing is left to report it to.
	detached := context.Background()
	if ctx != nil {
		detached = context.WithoutCancel(ctx)
	}

	for index, recipientId := range data.RecipientIds {
		for _, name := range data.ResolvedChannels {
			channel := this.channels.ChannelFor(modconstants.ChannelName(name))
			if channel == nil {
				// The channel was attached when this was sent and is not now. The row is
				// committed and the sender is gone, so this is a deployment change to be recorded
				// rather than a request to refuse.
				this.warnf("notification: channel '%s' is no longer attached; delivery skipped", name)
				continue
			}

			this.deliverOne(ctx, detached, orgId, recipientId, userIdAt(data, index), channel, data)
		}
	}
}

// deliverOne records an attempt and makes it.
//
// The row is written BEFORE the attempt, so that a process dying mid-delivery leaves something a
// sweeper can find. Writing it afterwards would lose the whole attempt on a crash, which is the
// one case durable delivery tracking exists for (BR 15).
func (this *ChannelFanOut) deliverOne(
	ctx corectx.Context,
	detached context.Context,
	orgId model.Id,
	recipientId model.Id,
	recipientUserId model.Id,
	channel itExt.NotificationChannel,
	data it.SendNotificationResultData,
) {
	delivery, err := this.insertPending(ctx, orgId, recipientId, channel.Name())
	if err != nil {
		this.errorf("notification: a delivery row could not be written", err)
		return
	}

	if enabled, err := channel.IsEnabledFor(ctx, orgId); err != nil || !enabled {
		// A channel switched off for this organization is a skip, not a failure: BR 14 requires
		// the notification to stay successful either way.
		this.finish(ctx, delivery, itExt.DeliveryOutcome{
			Skipped:    true,
			SkipReason: skipReasonChannelDisabled,
		})
		return
	}

	outcome := channel.Deliver(detached, itExt.DeliveryInput{
		OrgId:           orgId,
		NotificationId:  data.NotificationId,
		RecipientId:     recipientId,
		RecipientUserId: recipientUserId,
		Args:            this.argsFor(data, channel.Name()),
		AttemptNumber:   1,
		Timeout:         deliveryTimeout,
	})

	this.finish(ctx, delivery, outcome)
}

func (this *ChannelFanOut) insertPending(
	ctx corectx.Context, orgId model.Id, recipientId model.Id, name modconstants.ChannelName,
) (*models.Delivery, error) {
	id, err := model.NewId()
	if err != nil {
		return nil, err
	}

	channelName := name.String()
	status := models.DeliveryStatusPending
	attempts := int32(0)

	delivery := models.NewDelivery()
	delivery.SetId(id)
	delivery.SetOrgId(&orgId)
	delivery.SetRecipientId(&recipientId)
	delivery.SetChannelName(&channelName)
	delivery.SetStatus(&status)
	delivery.SetAttemptCount(&attempts)

	if _, err := this.deliveries.Insert(ctx, delivery.GetFieldData()); err != nil {
		return nil, err
	}

	return delivery, nil
}

// finish records what the attempt produced.
//
// A skip does not count as an attempt: nothing was tried, and counting it would let a channel that
// is merely switched off exhaust a retry budget it never used.
func (this *ChannelFanOut) finish(
	ctx corectx.Context, delivery *models.Delivery, outcome itExt.DeliveryOutcome,
) {
	now := model.NewModelDateTime()

	switch {
	case outcome.Succeeded:
		status := models.DeliveryStatusSent
		delivery.SetStatus(&status)
		delivery.SetSentAt(&now)
		delivery.SetLastAttemptAt(&now)
		delivery.SetAttemptCount(attemptCount(1))

	case outcome.Skipped:
		status := models.DeliveryStatusSkipped
		reason := outcome.SkipReason
		delivery.SetStatus(&status)
		if reason != "" {
			delivery.SetSkipReason(&reason)
		}

	default:
		// Retryable decides only whether trying again could help. How many times is policy, and
		// belongs to the sweeper that will pick this up, not to the channel that answered.
		status := models.DeliveryStatusFailed
		if outcome.Retryable {
			status = models.DeliveryStatusPending
		}
		delivery.SetStatus(&status)
		delivery.SetLastAttemptAt(&now)
		delivery.SetAttemptCount(attemptCount(1))

		if outcome.ErrorCode != "" {
			code := outcome.ErrorCode
			delivery.SetLastErrorCode(&code)
		}
		if outcome.ErrorMessage != "" {
			message := outcome.ErrorMessage
			delivery.SetLastErrorMessage(&message)
		}
	}

	if _, err := this.deliveries.Update(ctx, delivery.GetFieldData()); err != nil {
		this.errorf("notification: a delivery outcome could not be recorded", err)
	}
}

// argsFor is the argument object stored under one channel's name, or nil when the sender wrote
// none. The fan-out does not look inside it.
func (this *ChannelFanOut) argsFor(
	data it.SendNotificationResultData, name modconstants.ChannelName,
) itExt.ChannelArgs {
	if data.ChannelArgs == nil {
		return nil
	}

	args, given := data.ChannelArgs[name.String()]
	if !given {
		return nil
	}

	return itExt.ChannelArgs(args)
}

func attemptCount(n int32) *int32 {
	return &n
}

func userIdAt(data it.SendNotificationResultData, index int) model.Id {
	if index < len(data.RecipientUserIds) {
		return model.Id(data.RecipientUserIds[index])
	}
	return ""
}

func (this *ChannelFanOut) warnf(format string, args ...any) {
	if this.logger != nil {
		this.logger.Warnf(format, args...)
	}
}

func (this *ChannelFanOut) errorf(message string, err error) {
	if this.logger != nil {
		this.logger.Error(message, err)
	}
}

// skipReasonChannelDisabled is the core's own reason, as opposed to the free-form ones a channel
// may return for conditions only it knows about.
const skipReasonChannelDisabled = "channel_disabled"
