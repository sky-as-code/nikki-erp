package external

import (
	"fmt"
	"sort"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

// NotificationChannels is every attached channel, keyed by the name it answers to.
//
// A map rather than a slice: dispatch is a lookup, and two channels claiming one name is not
// something the rest of the code should have to consider.
type NotificationChannels map[modconstants.ChannelName]itExt.NotificationChannel

// ChannelDispatcher is the one place that knows which channels exist.
//
// Validation and delivery both read this same set, so a channel can no longer be accepted when a
// notification is sent and then found missing when it is delivered -- which is exactly what the
// separate registry of names allowed before it was replaced.
type ChannelDispatcher struct {
	channels NotificationChannels
	logger   logging.LoggerService
}

func NewChannelDispatcher(
	channels NotificationChannels, logger logging.LoggerService,
) *ChannelDispatcher {
	if len(channels) == 0 && logger != nil {
		// Not fatal. A deployment that wants only the durable inbox is a legitimate one: a
		// notification is readable whether or not anything ever delivers it.
		logger.Warnf("notification: no distribution channel is attached; nothing will be delivered")
	}

	return &ChannelDispatcher{
		channels: channels,
		logger:   logger,
	}
}

// Names lists every attached channel, sorted so that a resolved channel list and the errors that
// mention it do not change order between runs.
func (this *ChannelDispatcher) Names() []modconstants.ChannelName {
	names := make([]modconstants.ChannelName, 0, len(this.channels))
	for name := range this.channels {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })

	return names
}

func (this *ChannelDispatcher) IsRegistered(name string) bool {
	_, exists := this.channels[modconstants.ChannelName(name)]
	return exists
}

// ChannelFor returns the channel of that name, or nil when none is attached.
func (this *ChannelDispatcher) ChannelFor(name modconstants.ChannelName) itExt.NotificationChannel {
	return this.channels[name]
}

// EnabledFor is the channels an organization has switched on, and is what resolves a send that
// named no channels at all (BR 14).
//
// A channel that cannot answer is left out rather than failing the send: its setting being
// unreadable is not the sender's problem, and refusing would make an unrelated outage look like a
// bad request.
func (this *ChannelDispatcher) EnabledFor(
	ctx corectx.Context, orgId model.Id,
) []modconstants.ChannelName {
	enabled := make([]modconstants.ChannelName, 0, len(this.channels))
	for _, name := range this.Names() {
		isEnabled, err := this.channels[name].IsEnabledFor(ctx, orgId)
		if err != nil {
			if this.logger != nil {
				this.logger.Error(
					fmt.Sprintf("notification: channel '%s' could not be read for org '%s'", name, orgId),
					err)
			}
			continue
		}
		if isEnabled {
			enabled = append(enabled, name)
		}
	}

	return enabled
}

// ValidateArgs hands each channel its own arguments and collects what comes back.
//
// requested is what the send will actually deliver to. Both rules below reject rather than ignore,
// for the reason the unknown-channel rule already gives: a sender's mistake should surface at the
// call instead of quietly delivering nowhere.
func (this *ChannelDispatcher) ValidateArgs(
	ctx corectx.Context,
	args map[string]itExt.ChannelArgs,
	requested []modconstants.ChannelName,
) ft.ClientErrors {
	errs := make(ft.ClientErrors, 0)

	isRequested := make(map[modconstants.ChannelName]bool, len(requested))
	for _, name := range requested {
		isRequested[name] = true
	}

	for _, key := range sortedKeys(args) {
		field := fmt.Sprintf("channel_args.%s", key)
		name := modconstants.ChannelName(key)

		channel, exists := this.channels[name]
		if !exists {
			errs = append(errs, *ft.NewValidationError(
				field, "err_channel_args_unknown_channel",
				fmt.Sprintf("'%s' is not a known channel; known channels are %s",
					key, joinNames(this.Names()))))
			continue
		}

		if !isRequested[name] {
			// The requested set is named in the message because under distribution mode "all" the
			// sender never wrote it down: it is resolved from what the organization has enabled,
			// and without seeing it the rejection cannot be acted on.
			errs = append(errs, *ft.NewValidationError(
				field, "err_channel_args_not_requested",
				fmt.Sprintf("arguments were given for '%s', which this notification does not go to; it goes to %s",
					key, joinNames(requested))))
			continue
		}

		errs = append(errs, channel.ValidateArgs(ctx, args[name.String()])...)
	}

	// A channel is asked to check its arguments even when none were sent, so that one requiring an
	// argument can say so. Without this a required argument would only be missed at delivery.
	for _, name := range requested {
		if _, given := args[name.String()]; given {
			continue
		}
		if channel, exists := this.channels[name]; exists {
			errs = append(errs, channel.ValidateArgs(ctx, nil)...)
		}
	}

	return errs
}

func sortedKeys(args map[string]itExt.ChannelArgs) []string {
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

func joinNames(names []modconstants.ChannelName) string {
	if len(names) == 0 {
		return "no channel"
	}

	joined := ""
	for i, name := range names {
		if i > 0 {
			joined += ", "
		}
		joined += string(name)
	}

	return joined
}
