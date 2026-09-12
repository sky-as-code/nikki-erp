package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

// fakeRegistry stands in for the attached channels.
//
// The rules are tested against a set made up here rather than whichever channels happen to be
// wired into the application, so that attaching or detaching a real one cannot quietly change what
// these tests mean.
type fakeRegistry struct {
	registered []modconstants.ChannelName
	enabled    []modconstants.ChannelName

	// argErrors is returned for the named channel, standing in for a channel that refuses its own
	// arguments. The core never looks inside an argument object, so a fake refusal proves as much
	// as a real one.
	argErrors map[modconstants.ChannelName]string

	// validatedWith records every call, so a test can show that a channel was asked about its own
	// arguments even when none were sent.
	validatedWith map[modconstants.ChannelName]int
}

func newFakeRegistry(names ...modconstants.ChannelName) *fakeRegistry {
	return &fakeRegistry{
		registered:    names,
		enabled:       names,
		argErrors:     map[modconstants.ChannelName]string{},
		validatedWith: map[modconstants.ChannelName]int{},
	}
}

func (this *fakeRegistry) Names() []modconstants.ChannelName {
	return this.registered
}

func (this *fakeRegistry) IsRegistered(name string) bool {
	for _, registered := range this.registered {
		if string(registered) == name {
			return true
		}
	}
	return false
}

func (this *fakeRegistry) EnabledFor(
	ctx corectx.Context, orgId model.Id,
) []modconstants.ChannelName {
	return this.enabled
}

func (this *fakeRegistry) ValidateArgs(
	ctx corectx.Context,
	args map[string]itExt.ChannelArgs,
	requested []modconstants.ChannelName,
) ft.ClientErrors {
	errs := make(ft.ClientErrors, 0)

	isRequested := map[modconstants.ChannelName]bool{}
	for _, name := range requested {
		isRequested[name] = true
	}

	for key := range args {
		name := modconstants.ChannelName(key)
		if !this.IsRegistered(key) {
			errs = append(errs, *ft.NewValidationError(
				"channel_args."+key, "err_channel_args_unknown_channel", "unknown channel"))
			continue
		}
		if !isRequested[name] {
			errs = append(errs, *ft.NewValidationError(
				"channel_args."+key, "err_channel_args_not_requested", "channel not a destination"))
			continue
		}
		this.validatedWith[name]++
		if key, refused := this.argErrors[name]; refused {
			errs = append(errs, *ft.NewValidationError("channel_args."+string(name), key, "refused"))
		}
	}

	for _, name := range requested {
		if _, given := args[string(name)]; given {
			continue
		}
		this.validatedWith[name]++
		if key, refused := this.argErrors[name]; refused {
			errs = append(errs, *ft.NewValidationError("channel_args."+string(name), key, "refused"))
		}
	}

	return errs
}

// webOnly is the registry most of these tests want: one attached, enabled channel.
func webOnly() *fakeRegistry {
	return newFakeRegistry(modconstants.ChannelWeb)
}

func normalizeWith(
	registry ChannelRegistry, request it.SendNotificationRequest,
) (*NormalizedSend, ft.ClientErrors) {
	return NewSendNormalizer(registry).NormalizeSend(nil, model.Id("org-1"), request)
}

func NormalizeSend(request it.SendNotificationRequest) (*NormalizedSend, ft.ClientErrors) {
	return normalizeWith(webOnly(), request)
}

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

// BR 5: the channels an organization has enabled are what a send with no named channels goes to.
// Before this, mode "all" resolved to nothing at all and the resolved list came back empty.
func TestDistributeToAllResolvesTheOrgsEnabledChannels(t *testing.T) {
	registry := newFakeRegistry(modconstants.ChannelWeb, modconstants.ChannelName("mobile"))

	normalized, errs := normalizeWith(registry, validRequest())

	require.Empty(t, errs)
	assert.Equal(t, models.DistributionModeAll, normalized.DistributionMode)
	assert.Nil(t, normalized.Channels, "nothing was requested, so nothing is recorded as requested")
	assert.Equal(t,
		[]modconstants.ChannelName{modconstants.ChannelWeb, modconstants.ChannelName("mobile")},
		normalized.ResolvedChannels)
}

// A channel that is attached but switched off for this organization is not a destination.
func TestDistributeToAllSkipsADisabledChannel(t *testing.T) {
	registry := newFakeRegistry(modconstants.ChannelWeb, modconstants.ChannelName("mobile"))
	registry.enabled = []modconstants.ChannelName{modconstants.ChannelWeb}

	normalized, errs := normalizeWith(registry, validRequest())

	require.Empty(t, errs)
	assert.Equal(t, []modconstants.ChannelName{modconstants.ChannelWeb}, normalized.ResolvedChannels)
}

func TestExplicitChannelsAreWhatIsResolved(t *testing.T) {
	registry := newFakeRegistry(modconstants.ChannelWeb, modconstants.ChannelName("mobile"))
	request := validRequest()
	request.Channels = []string{"web"}

	normalized, errs := normalizeWith(registry, request)

	require.Empty(t, errs)
	assert.Equal(t, []modconstants.ChannelName{modconstants.ChannelWeb}, normalized.ResolvedChannels)
}

// The core passes an argument object through without reading it: whatever is inside is the
// channel's business, and a channel that accepts it is the only thing that decides so.
func TestChannelArgsReachTheChannelUntouched(t *testing.T) {
	request := validRequest()
	request.Channels = []string{"web"}
	request.ChannelArgs = map[string]map[string]any{
		"web": {"app_link": "nikkiweb://notification/notifications", "anything": 42},
	}

	normalized, errs := NormalizeSend(request)

	require.Empty(t, errs)
	assert.Equal(t, []modconstants.ChannelName{modconstants.ChannelWeb}, normalized.ResolvedChannels)
}

// A channel is asked about its arguments even when none were sent, so that one requiring an
// argument can refuse at the call rather than at delivery, where no caller is left to tell.
func TestAChannelIsAskedEvenWhenNoArgumentsWereSent(t *testing.T) {
	registry := webOnly()

	_, errs := normalizeWith(registry, validRequest())

	require.Empty(t, errs)
	assert.Equal(t, 1, registry.validatedWith[modconstants.ChannelWeb])
}

func TestAChannelMayRefuseItsOwnArguments(t *testing.T) {
	registry := webOnly()
	registry.argErrors[modconstants.ChannelWeb] = "err_app_link_scheme_unsupported"

	request := validRequest()
	request.Channels = []string{"web"}
	request.ChannelArgs = map[string]map[string]any{
		"web": {"app_link": "javascript:alert(1)"},
	}

	_, errs := normalizeWith(registry, request)

	require.Len(t, errs, 1)
	assert.Equal(t, "err_app_link_scheme_unsupported", errs[0].Key)
}

// A name nobody answers to is a mistake, and is refused for the same reason an unknown name in
// `channels` is: the sender is still there to be told, and delivering nowhere is worse.
func TestArgumentsForAnUnknownChannelAreRefused(t *testing.T) {
	request := validRequest()
	request.Channels = []string{"web"}
	request.ChannelArgs = map[string]map[string]any{
		"webbapp": {"app_link": "nikkiweb://x"},
	}

	assert.Contains(t, errorKeys(request), "err_channel_args_unknown_channel")
}

// Arguments for a channel this notification does not go to are refused rather than ignored: the
// sender asked for something that will not happen, and silence would hide it.
func TestArgumentsForAnUnrequestedChannelAreRefused(t *testing.T) {
	registry := newFakeRegistry(modconstants.ChannelWeb, modconstants.ChannelName("mobile"))
	request := validRequest()
	request.Channels = []string{"web"}
	request.ChannelArgs = map[string]map[string]any{
		"mobile": {"app_link": "coremart://order/1"},
	}

	_, errs := normalizeWith(registry, request)

	require.Len(t, errs, 1)
	assert.Equal(t, "err_channel_args_not_requested", errs[0].Key)
}

// Under mode "all" the sender named no channels, so "requested" means what the organization has
// enabled -- and arguments for a channel outside that set are still refused.
func TestUnrequestedArgumentsAreJudgedAgainstTheResolvedSetUnderModeAll(t *testing.T) {
	registry := newFakeRegistry(modconstants.ChannelWeb, modconstants.ChannelName("mobile"))
	registry.enabled = []modconstants.ChannelName{modconstants.ChannelWeb}

	request := validRequest()
	request.ChannelArgs = map[string]map[string]any{
		"mobile": {"app_link": "coremart://order/1"},
	}

	_, errs := normalizeWith(registry, request)

	require.Len(t, errs, 1)
	assert.Equal(t, "err_channel_args_not_requested", errs[0].Key)
}

// Detaching a channel must not need a change anywhere else: the same send that was valid becomes a
// plain rejection, and nothing panics for want of an adapter that is no longer there.
func TestDetachingAChannelLeavesTheRulesIntact(t *testing.T) {
	registry := newFakeRegistry()

	request := validRequest()
	request.Channels = []string{"web"}

	_, errs := normalizeWith(registry, request)

	require.NotEmpty(t, errs)
	assert.Equal(t, "err_channel_unknown", errs[0].Key)
}
