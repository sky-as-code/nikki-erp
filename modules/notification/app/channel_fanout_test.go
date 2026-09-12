package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

// recordingDeliveries captures the rows the fan-out writes, so a test can read back what was
// recorded without a database.
type recordingDeliveries struct {
	composable.CrudRepository

	inserted []dmodel.DynamicFields
	updated  []dmodel.DynamicFields
}

// Both record a COPY. The fan-out hands the same field map to the insert and then to the update,
// which is how the repository layer is written; keeping the reference would let the later call
// rewrite what the earlier one is supposed to have recorded.
func (this *recordingDeliveries) Insert(
	ctx corectx.Context, fields dmodel.DynamicFields,
) (*dyn.OpResult[int], error) {
	this.inserted = append(this.inserted, snapshot(fields))
	return nil, nil
}

func (this *recordingDeliveries) Update(
	ctx corectx.Context, fields dmodel.DynamicFields,
) (*composable.MutateResult, error) {
	this.updated = append(this.updated, snapshot(fields))
	return nil, nil
}

func snapshot(fields dmodel.DynamicFields) dmodel.DynamicFields {
	copied := make(dmodel.DynamicFields, len(fields))
	for key, value := range fields {
		copied[key] = value
	}
	return copied
}

// lastUpdated is the row as the fan-out last left it.
func (this *recordingDeliveries) lastUpdated() *models.Delivery {
	if len(this.updated) == 0 {
		return nil
	}
	return models.NewDeliveryFrom(this.updated[len(this.updated)-1])
}

// scriptedChannel answers with whatever outcome the test set.
type scriptedChannel struct {
	name     modconstants.ChannelName
	enabled  bool
	outcome  itExt.DeliveryOutcome
	seenArgs itExt.ChannelArgs
	delivers int
}

func (this *scriptedChannel) Name() modconstants.ChannelName { return this.name }

func (this *scriptedChannel) ValidateArgs(
	ctx corectx.Context, args itExt.ChannelArgs,
) ft.ClientErrors {
	return nil
}

func (this *scriptedChannel) IsEnabledFor(ctx corectx.Context, orgId model.Id) (bool, error) {
	return this.enabled, nil
}

func (this *scriptedChannel) Deliver(
	ctx context.Context, in itExt.DeliveryInput,
) itExt.DeliveryOutcome {
	this.seenArgs = in.Args
	this.delivers++
	return this.outcome
}

type stubRegistry struct {
	channels map[modconstants.ChannelName]itExt.NotificationChannel
}

func (this *stubRegistry) ChannelFor(
	name modconstants.ChannelName,
) itExt.NotificationChannel {
	return this.channels[name]
}

func fanOutWith(chans ...*scriptedChannel) (*ChannelFanOut, *recordingDeliveries) {
	registry := &stubRegistry{channels: map[modconstants.ChannelName]itExt.NotificationChannel{}}
	for _, channel := range chans {
		registry.channels[channel.name] = channel
	}

	deliveries := &recordingDeliveries{}
	return NewChannelFanOut(registry, deliveries, nil), deliveries
}

func sendResult(channels ...string) it.SendNotificationResultData {
	return it.SendNotificationResultData{
		NotificationId:   model.Id("notif-1"),
		RecipientIds:     []model.Id{model.Id("recip-1")},
		RecipientUserIds: []string{"user-1"},
		ResolvedChannels: channels,
	}
}

func succeeding(name string) *scriptedChannel {
	return &scriptedChannel{
		name:    modconstants.ChannelName(name),
		enabled: true,
		outcome: itExt.DeliveryOutcome{Succeeded: true},
	}
}

// The first rows this table has ever held: one per recipient per channel.
func TestOneRowIsWrittenPerRecipientPerChannel(t *testing.T) {
	fanOut, deliveries := fanOutWith(succeeding("web"), succeeding("mobile"))

	data := sendResult("web", "mobile")
	data.RecipientIds = []model.Id{model.Id("recip-1"), model.Id("recip-2")}
	data.RecipientUserIds = []string{"user-1", "user-2"}

	fanOut.Deliver(nil, model.Id("org-1"), data)

	assert.Len(t, deliveries.inserted, 4, "two recipients over two channels")
}

// The row exists before the attempt is made, so a process dying mid-delivery leaves something a
// sweeper can find rather than nothing at all.
func TestTheRowIsWrittenBeforeTheAttempt(t *testing.T) {
	channel := succeeding("web")
	fanOut, deliveries := fanOutWith(channel)

	fanOut.Deliver(nil, model.Id("org-1"), sendResult("web"))

	require.Len(t, deliveries.inserted, 1)
	pending := models.NewDeliveryFrom(deliveries.inserted[0])
	require.NotNil(t, pending.GetStatus())
	assert.Equal(t, models.DeliveryStatusPending, *pending.GetStatus(),
		"the row starts pending, before anything has been tried")
	assert.Equal(t, int32(0), *pending.GetAttemptCount())
}

func TestASuccessfulDeliveryIsRecordedAsSent(t *testing.T) {
	fanOut, deliveries := fanOutWith(succeeding("web"))

	fanOut.Deliver(nil, model.Id("org-1"), sendResult("web"))

	row := deliveries.lastUpdated()
	require.NotNil(t, row)
	assert.Equal(t, models.DeliveryStatusSent, *row.GetStatus())
	assert.NotNil(t, row.GetSentAt())
	assert.Equal(t, int32(1), *row.GetAttemptCount())
}

// A channel switched off for this organization is a skip, and a skip is not an attempt: counting it
// would let a channel that was never tried exhaust a retry budget.
func TestADisabledChannelIsSkippedRatherThanFailed(t *testing.T) {
	channel := succeeding("web")
	channel.enabled = false
	fanOut, deliveries := fanOutWith(channel)

	fanOut.Deliver(nil, model.Id("org-1"), sendResult("web"))

	row := deliveries.lastUpdated()
	require.NotNil(t, row)
	assert.Equal(t, models.DeliveryStatusSkipped, *row.GetStatus())
	require.NotNil(t, row.GetSkipReason())
	assert.Equal(t, skipReasonChannelDisabled, *row.GetSkipReason())
	assert.Equal(t, int32(0), *row.GetAttemptCount(), "a skip is not an attempt")
	assert.Equal(t, 0, channel.delivers, "a disabled channel is never asked to deliver")
}

// A channel's own skip reason is carried through without the core knowing what it means.
func TestAChannelsOwnSkipReasonIsKept(t *testing.T) {
	channel := succeeding("web")
	channel.outcome = itExt.DeliveryOutcome{Skipped: true, SkipReason: "no_device_token"}
	fanOut, deliveries := fanOutWith(channel)

	fanOut.Deliver(nil, model.Id("org-1"), sendResult("web"))

	row := deliveries.lastUpdated()
	require.NotNil(t, row.GetSkipReason())
	assert.Equal(t, "no_device_token", *row.GetSkipReason())
}

// Retryable means another attempt could help, so the row goes back to pending for the sweeper.
func TestARetryableFailureReturnsToPending(t *testing.T) {
	channel := succeeding("web")
	channel.outcome = itExt.DeliveryOutcome{Retryable: true, ErrorCode: "HTTP_503"}
	fanOut, deliveries := fanOutWith(channel)

	fanOut.Deliver(nil, model.Id("org-1"), sendResult("web"))

	row := deliveries.lastUpdated()
	assert.Equal(t, models.DeliveryStatusPending, *row.GetStatus())
	assert.Equal(t, int32(1), *row.GetAttemptCount())
}

// A permanent failure is terminal: retrying an invalid destination forever helps nobody.
func TestAPermanentFailureIsTerminal(t *testing.T) {
	channel := succeeding("web")
	channel.outcome = itExt.DeliveryOutcome{ErrorCode: "INVALID_DESTINATION"}
	fanOut, deliveries := fanOutWith(channel)

	fanOut.Deliver(nil, model.Id("org-1"), sendResult("web"))

	row := deliveries.lastUpdated()
	assert.Equal(t, models.DeliveryStatusFailed, *row.GetStatus())
}

// Each channel is handed the object stored under its own name, and nothing else.
func TestEachChannelReceivesOnlyItsOwnArguments(t *testing.T) {
	web := succeeding("web")
	mobile := succeeding("mobile")
	fanOut, _ := fanOutWith(web, mobile)

	data := sendResult("web", "mobile")
	data.ChannelArgs = map[string]map[string]any{
		"web":    {"app_link": "nikkiweb://x"},
		"mobile": {"app_link": "coremart://y"},
	}

	fanOut.Deliver(nil, model.Id("org-1"), data)

	assert.Equal(t, "nikkiweb://x", web.seenArgs["app_link"])
	assert.Equal(t, "coremart://y", mobile.seenArgs["app_link"])
}

// A replay created nothing, so delivering again would send a second message for a notification the
// sender was told already existed.
func TestAReplayDeliversNothing(t *testing.T) {
	fanOut, deliveries := fanOutWith(succeeding("web"))

	data := sendResult("web")
	data.Duplicate = true
	fanOut.Deliver(nil, model.Id("org-1"), data)

	assert.Empty(t, deliveries.inserted)
}

// A channel detached between the send and the delivery is recorded and stepped over, not fatal.
func TestADetachedChannelIsSteppedOver(t *testing.T) {
	fanOut, deliveries := fanOutWith(succeeding("web"))

	fanOut.Deliver(nil, model.Id("org-1"), sendResult("web", "telegram"))

	assert.Len(t, deliveries.inserted, 1, "only the attached channel gets a row")
}
