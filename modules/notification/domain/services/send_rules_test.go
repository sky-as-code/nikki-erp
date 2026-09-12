package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

func validRequest() it.SendNotificationRequest {
	return it.SendNotificationRequest{
		RecipientUserIds: []string{"user-1"},
		Title:            "Order confirmed",
		Message:          "SO-10020 was confirmed.",
		SourceModule:     "sales",
	}
}

func errorKeys(request it.SendNotificationRequest) []string {
	_, errs := NormalizeSend(request)
	keys := make([]string, 0, len(errs))
	for _, item := range errs {
		keys = append(keys, item.Key)
	}
	return keys
}

func TestSeverityDefaultsToInfo(t *testing.T) {
	normalized, errs := NormalizeSend(validRequest())

	require.Empty(t, errs)
	assert.Equal(t, models.SeverityInfo, normalized.Severity)
}

// BR 11.5: no channels named means every enabled channel, resolved later against the org.
func TestOmittedChannelsMeanDistributeToAll(t *testing.T) {
	normalized, errs := NormalizeSend(validRequest())

	require.Empty(t, errs)
	assert.Equal(t, models.DistributionModeAll, normalized.DistributionMode)
	assert.Nil(t, normalized.Channels)
}

// BR 11.6: naming channels is an explicit distribution.
func TestNamedChannelsAreExplicit(t *testing.T) {
	request := validRequest()
	request.Channels = []string{"web"}

	normalized, errs := NormalizeSend(request)

	require.Empty(t, errs)
	assert.Equal(t, models.DistributionModeExplicit, normalized.DistributionMode)
	assert.Equal(t, []modconstants.ChannelName{modconstants.ChannelWeb}, normalized.Channels)
}

// BR 11.7: an empty list is a mistake, not a synonym for "all". The distinction is nil versus
// empty, which is why the field is a slice and not a count.
func TestEmptyChannelListIsRejected(t *testing.T) {
	request := validRequest()
	request.Channels = []string{}

	assert.Contains(t, errorKeys(request), "err_channels_empty")
}

// BR 11.8: an unregistered channel is refused rather than dropped, so a typo does not look like a
// successful send that reached nobody.
func TestUnknownChannelIsRejected(t *testing.T) {
	request := validRequest()
	request.Channels = []string{"telegram"}

	assert.Contains(t, errorKeys(request), "err_channel_unknown")
}

// BR 11.9: asking for one channel twice is unambiguous, so it collapses instead of erroring.
func TestDuplicateChannelsCollapse(t *testing.T) {
	request := validRequest()
	request.Channels = []string{"web", "web"}

	normalized, errs := NormalizeSend(request)

	require.Empty(t, errs)
	assert.Equal(t, []modconstants.ChannelName{modconstants.ChannelWeb}, normalized.Channels)
}

// BR 11.2: one person named twice gets one notification, not two inbox rows.
func TestDuplicateRecipientsCollapseKeepingOrder(t *testing.T) {
	request := validRequest()
	request.RecipientUserIds = []string{"user-2", "user-1", "user-2", "  ", "user-3"}

	normalized, errs := NormalizeSend(request)

	require.Empty(t, errs)
	assert.Equal(t,
		[]model.Id{model.Id("user-2"), model.Id("user-1"), model.Id("user-3")},
		normalized.RecipientUserIds)
}

// BR 11.1, 11.3: the fields without which a notification means nothing.
func TestRequiredFieldsAreEnforced(t *testing.T) {
	request := it.SendNotificationRequest{
		RecipientUserIds: []string{"   "},
		Title:            "  ",
		Message:          "",
		SourceModule:     "",
	}

	keys := errorKeys(request)

	assert.Contains(t, keys, "err_recipients_required")
	assert.Contains(t, keys, "err_title_required")
	assert.Contains(t, keys, "err_message_required")
	assert.Contains(t, keys, "err_source_module_required")
}

// Every problem is reported at once rather than one call at a time.
func TestEveryViolationIsReportedTogether(t *testing.T) {
	request := it.SendNotificationRequest{
		RecipientUserIds: nil,
		Title:            "",
		Message:          "",
		SourceModule:     "sales",
		Severity:         "catastrophic",
	}

	assert.Len(t, errorKeys(request), 4)
}

func TestUnknownSeverityIsRejected(t *testing.T) {
	request := validRequest()
	request.Severity = "catastrophic"

	assert.Contains(t, errorKeys(request), "err_severity_unknown")
}

// A blank key must not claim the single "no key" slot the partial unique leaves per organization
// and module.
func TestBlankIdempotencyKeyIsTreatedAsAbsent(t *testing.T) {
	request := validRequest()
	blank := "   "
	request.IdempotencyKey = &blank

	normalized, errs := NormalizeSend(request)

	require.Empty(t, errs)
	assert.Nil(t, normalized.IdempotencyKey)
}

func TestIdempotencyKeyIsTrimmed(t *testing.T) {
	request := validRequest()
	key := "  cmd-123  "
	request.IdempotencyKey = &key

	normalized, errs := NormalizeSend(request)

	require.Empty(t, errs)
	require.NotNil(t, normalized.IdempotencyKey)
	assert.Equal(t, "cmd-123", *normalized.IdempotencyKey)
}
