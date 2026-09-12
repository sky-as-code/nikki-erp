package channels

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modsettings "github.com/sky-as-code/nikki-erp/modules/notification/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

func validateLink(t *testing.T, link any) []string {
	t.Helper()

	errs := NewWebChannel(nil).ValidateArgs(nil, itExt.ChannelArgs{ArgAppLink: link})
	keys := make([]string, 0, len(errs))
	for _, item := range errs {
		keys = append(keys, item.Key)
	}

	return keys
}

// A notification that points nowhere is the ordinary case: the bell simply shows it.
func TestNoArgumentsIsValid(t *testing.T) {
	assert.Empty(t, NewWebChannel(nil).ValidateArgs(nil, nil))
	assert.Empty(t, NewWebChannel(nil).ValidateArgs(nil, itExt.ChannelArgs{}))
}

func TestAnInternalLinkIsAccepted(t *testing.T) {
	assert.Empty(t, validateLink(t, "nikkiweb://notification/notifications"))
}

func TestAnExternalLinkIsAccepted(t *testing.T) {
	assert.Empty(t, validateLink(t, "https://example.com/orders/1"))
}

// The link ends up in an href, so the scheme is checked against an allow-list rather than a list of
// things to keep out: "javascript:" is the obvious one, but so is every scheme nobody has thought
// of yet.
func TestAScriptLinkIsRefused(t *testing.T) {
	assert.Contains(t, validateLink(t, "javascript:alert(1)"), "err_app_link_scheme_unsupported")
}

func TestAnUnknownSchemeIsRefused(t *testing.T) {
	assert.Contains(t, validateLink(t, "ftp://example.com/x"), "err_app_link_scheme_unsupported")
}

// A bare path has no scheme, so there is nothing to check it against and nothing that could resolve
// it: a relative link is exactly the mistake this rejects.
func TestARelativeLinkIsRefused(t *testing.T) {
	assert.Contains(t, validateLink(t, "/notification/notifications"), "err_app_link_invalid")
}

func TestANonStringLinkIsRefused(t *testing.T) {
	assert.Contains(t, validateLink(t, 42), "err_app_link_invalid")
}

func TestAnEmptyLinkIsRefused(t *testing.T) {
	assert.Contains(t, validateLink(t, "   "), "err_app_link_invalid")
}

func TestAnOverlongLinkIsRefused(t *testing.T) {
	long := "https://example.com/"
	for len(long) <= maxAppLinkLength {
		long += "x"
	}

	assert.Contains(t, validateLink(t, long), "err_app_link_too_long")
}

// The channel reads only its own argument. Anything else the sender put in the object is none of
// its business and must not make the send fail.
func TestUnknownArgumentsAreLeftAlone(t *testing.T) {
	errs := NewWebChannel(nil).ValidateArgs(nil, itExt.ChannelArgs{"something_else": "value"})

	assert.Empty(t, errs)
}

func TestTheWebChannelDeliversByHavingNothingToDo(t *testing.T) {
	outcome := NewWebChannel(nil).Deliver(nil, itExt.DeliveryInput{})

	require.True(t, outcome.Succeeded)
	assert.False(t, outcome.Skipped)
}

// stubSettings answers the one question the channel asks.
type stubSettings struct {
	value  bool
	asked  string
	called int
}

func (this *stubSettings) Bool(ctx corectx.Context, name string, fallback bool) bool {
	this.asked = name
	this.called++
	return this.value
}

// The channel reads its own setting and no other. Each channel owning its own switch is what lets
// one be attached without the core learning that the setting exists.
func TestTheChannelReadsItsOwnSetting(t *testing.T) {
	settings := &stubSettings{value: true}

	enabled, err := NewWebChannel(settings).IsEnabledFor(nil, model.Id("org-1"))

	require.NoError(t, err)
	assert.True(t, enabled)
	assert.Equal(t, modsettings.OrgSettingWebEnabled, settings.asked)
	assert.Equal(t, 1, settings.called)
}

func TestASwitchedOffChannelReportsItself(t *testing.T) {
	settings := &stubSettings{value: false}

	enabled, err := NewWebChannel(settings).IsEnabledFor(nil, model.Id("org-1"))

	require.NoError(t, err)
	assert.False(t, enabled)
}

// With nothing to read from, the channel falls back to the schema's own default rather than
// refusing: a channel that could not load its configuration must not take the send down with it.
func TestTheChannelFallsBackToTheDeclaredDefault(t *testing.T) {
	enabled, err := NewWebChannel(nil).IsEnabledFor(nil, model.Id("org-1"))

	require.NoError(t, err)
	assert.Equal(t, modsettings.DefaultWebEnabled, enabled)
}
