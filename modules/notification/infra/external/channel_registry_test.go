package external

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

// stubChannel is a channel that records what it was asked and answers as the test tells it to.
type stubChannel struct {
	name     modconstants.ChannelName
	enabled  bool
	argsErr  string
	seenArgs []itExt.ChannelArgs
}

func (this *stubChannel) Name() modconstants.ChannelName { return this.name }

func (this *stubChannel) ValidateArgs(
	ctx corectx.Context, args itExt.ChannelArgs,
) ft.ClientErrors {
	this.seenArgs = append(this.seenArgs, args)

	errs := make(ft.ClientErrors, 0)
	if this.argsErr != "" {
		errs = append(errs, *ft.NewValidationError("args", this.argsErr, "refused"))
	}
	return errs
}

func (this *stubChannel) IsEnabledFor(ctx corectx.Context, orgId model.Id) (bool, error) {
	return this.enabled, nil
}

func (this *stubChannel) Deliver(
	ctx context.Context, in itExt.DeliveryInput,
) itExt.DeliveryOutcome {
	return itExt.DeliveryOutcome{Succeeded: true}
}

func dispatcherWith(chans ...*stubChannel) (*ChannelDispatcher, NotificationChannels) {
	registry := NotificationChannels{}
	for _, channel := range chans {
		registry[channel.name] = channel
	}

	return NewChannelDispatcher(registry, nil), registry
}

func stub(name string, enabled bool) *stubChannel {
	return &stubChannel{name: modconstants.ChannelName(name), enabled: enabled}
}

// Sorted, so that a resolved channel list and the errors naming it do not shuffle between runs.
func TestNamesAreStable(t *testing.T) {
	dispatcher, _ := dispatcherWith(stub("web", true), stub("mobile", true), stub("telegram", true))

	assert.Equal(t, []modconstants.ChannelName{"mobile", "telegram", "web"}, dispatcher.Names())
}

func TestOnlyEnabledChannelsResolve(t *testing.T) {
	dispatcher, _ := dispatcherWith(stub("web", true), stub("mobile", false))

	assert.Equal(t, []modconstants.ChannelName{"web"}, dispatcher.EnabledFor(nil, model.Id("org-1")))
}

// The dispatcher hands each object to its owner and to nobody else. This is the property that lets
// a channel be attached without the core learning anything about what is inside.
func TestEachChannelSeesOnlyItsOwnArguments(t *testing.T) {
	web := stub("web", true)
	mobile := stub("mobile", true)
	dispatcher, _ := dispatcherWith(web, mobile)

	errs := dispatcher.ValidateArgs(nil,
		map[string]itExt.ChannelArgs{
			"web":    {"app_link": "nikkiweb://x"},
			"mobile": {"app_link": "coremart://y"},
		},
		[]modconstants.ChannelName{"web", "mobile"})

	require.Empty(t, errs)
	require.Len(t, web.seenArgs, 1)
	require.Len(t, mobile.seenArgs, 1)
	assert.Equal(t, "nikkiweb://x", web.seenArgs[0]["app_link"])
	assert.Equal(t, "coremart://y", mobile.seenArgs[0]["app_link"])
}

// A channel requiring an argument can only refuse if it is asked when none was sent.
func TestAChannelIsAskedWhenNoArgumentsWereSent(t *testing.T) {
	web := stub("web", true)
	dispatcher, _ := dispatcherWith(web)

	errs := dispatcher.ValidateArgs(nil, nil, []modconstants.ChannelName{"web"})

	require.Empty(t, errs)
	require.Len(t, web.seenArgs, 1)
	assert.Nil(t, web.seenArgs[0])
}

func TestAChannelsRefusalIsReported(t *testing.T) {
	web := stub("web", true)
	web.argsErr = "err_app_link_scheme_unsupported"
	dispatcher, _ := dispatcherWith(web)

	errs := dispatcher.ValidateArgs(nil,
		map[string]itExt.ChannelArgs{"web": {"app_link": "javascript:x"}},
		[]modconstants.ChannelName{"web"})

	require.Len(t, errs, 1)
	assert.Equal(t, "err_app_link_scheme_unsupported", errs[0].Key)
}

func TestUnknownChannelArgumentsAreRefused(t *testing.T) {
	dispatcher, _ := dispatcherWith(stub("web", true))

	errs := dispatcher.ValidateArgs(nil,
		map[string]itExt.ChannelArgs{"webbapp": {}},
		[]modconstants.ChannelName{"web"})

	require.Len(t, errs, 1)
	assert.Equal(t, "err_channel_args_unknown_channel", errs[0].Key)
	// The known names are listed, because a rejection that does not say what was expected cannot
	// be acted on.
	assert.Contains(t, errs[0].Message, "web")
}

func TestArgumentsForAChannelThatIsNotADestinationAreRefused(t *testing.T) {
	dispatcher, _ := dispatcherWith(stub("web", true), stub("mobile", true))

	errs := dispatcher.ValidateArgs(nil,
		map[string]itExt.ChannelArgs{"mobile": {}},
		[]modconstants.ChannelName{"web"})

	require.Len(t, errs, 1)
	assert.Equal(t, "err_channel_args_not_requested", errs[0].Key)
	assert.Contains(t, errs[0].Message, "web")
}

// An empty registry must be inert rather than fatal: a deployment that wants only the durable
// inbox is a legitimate one.
func TestAnEmptyRegistryIsUsable(t *testing.T) {
	dispatcher, _ := dispatcherWith()

	assert.Empty(t, dispatcher.Names())
	assert.False(t, dispatcher.IsRegistered("web"))
	assert.Nil(t, dispatcher.ChannelFor("web"))
	assert.Empty(t, dispatcher.EnabledFor(nil, model.Id("org-1")))
}
