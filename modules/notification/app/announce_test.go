package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
)

// Which channels a notification goes to is what decides whether the browser is pushed to.
//
// Before the channel seam existed the stream was written to unconditionally, so a send naming only
// another channel still reached the browser. These cases pin the corrected behaviour (AC03).

func TestTheWebIsPushedToWhenItIsADestination(t *testing.T) {
	assert.True(t, isResolved([]string{"web"}, modconstants.ChannelWeb))
	assert.True(t, isResolved([]string{"telegram", "web"}, modconstants.ChannelWeb))
}

// The behaviour change: asking for another channel alone no longer pushes to the browser.
func TestTheWebIsNotPushedToWhenItIsNotADestination(t *testing.T) {
	assert.False(t, isResolved([]string{"telegram"}, modconstants.ChannelWeb))
	assert.False(t, isResolved([]string{"telegram", "mobile"}, modconstants.ChannelWeb))
}

// Nothing resolved means nothing delivered. The notification is still in the inbox, which is the
// source of truth; only the push is withheld.
func TestNothingResolvedPushesNothing(t *testing.T) {
	assert.False(t, isResolved(nil, modconstants.ChannelWeb))
	assert.False(t, isResolved([]string{}, modconstants.ChannelWeb))
}

// A name that merely contains the channel's is not that channel.
func TestAPartialNameIsNotAMatch(t *testing.T) {
	assert.False(t, isResolved([]string{"webapp"}, modconstants.ChannelWeb))
	assert.False(t, isResolved([]string{"WEB"}, modconstants.ChannelWeb))
}
