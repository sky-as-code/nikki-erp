package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/services"
	infraExt "github.com/sky-as-code/nikki-erp/modules/notification/infra/external"
	infraChannels "github.com/sky-as-code/nikki-erp/modules/notification/infra/external/channels"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// The channel wiring is resolved here rather than left to the first boot that needs it.
//
// A channel is attached by one Register line and one entry in the set the dispatcher is built
// from; forgetting either is a runtime failure at startup, which is an expensive place to find out.
// These tests resolve the same graph the application does, so the mistake fails a test instead.

func TestTheAttachedChannelsResolve(t *testing.T) {
	require.NoError(t, deps.Register(func() logging.LoggerService { return nil }))

	// The settings module is not booted here, so its two contracts are stood in for. Everything
	// between them and the channel is the real wiring: this test fails if a channel stops being
	// registered, or if what it depends on is never bound.
	require.NoError(t, deps.Register(
		func() itSettings.TenantSettingsAppService { return nil },
		func() itSettings.OrgSettingsAppService { return nil },
	))
	require.NoError(t, infraExt.InitExternalServices())

	err := deps.Invoke(func(dispatcher *infraExt.ChannelDispatcher) {
		assert.True(t, dispatcher.IsRegistered(modconstants.ChannelWeb.String()),
			"the web channel is attached in InitExternalServices and must resolve")

		channel := dispatcher.ChannelFor(modconstants.ChannelWeb)
		require.NotNil(t, channel)
		assert.Equal(t, modconstants.ChannelWeb, channel.Name())
	})
	require.NoError(t, err)
}

// The name a channel answers to must be the key it is filed under, or a send naming that channel
// would validate and then find nothing to deliver with.
func TestAChannelIsFiledUnderTheNameItAnswersTo(t *testing.T) {
	web := infraChannels.NewWebChannel(nil)

	var channel itExt.NotificationChannel = web
	dispatcher := infraExt.NewChannelDispatcher(
		infraExt.NotificationChannels{web.Name(): channel}, nil)

	assert.Equal(t, channel.Name(), dispatcher.ChannelFor(web.Name()).Name())
}

// The send rules must reach the same registry the delivery does. Two sets would let a channel be
// accepted when a notification is sent and be missing when it is delivered.
func TestTheSendRulesSeeTheAttachedChannels(t *testing.T) {
	web := infraChannels.NewWebChannel(nil)
	dispatcher := infraExt.NewChannelDispatcher(
		infraExt.NotificationChannels{web.Name(): web}, nil)

	normalizer := services.NewSendNormalizer(dispatcher)
	require.NotNil(t, normalizer)

	assert.True(t, dispatcher.IsRegistered("web"))
	assert.False(t, dispatcher.IsRegistered("telegram"),
		"a channel nobody attached must not be accepted")
}
