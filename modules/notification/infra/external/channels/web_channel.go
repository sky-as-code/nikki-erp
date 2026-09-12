// Package channels holds the distribution channels themselves.
//
// Each one is self-contained: it owns the meaning of its own arguments, reads its own setting, and
// knows how to reach exactly one transport. Nothing here is imported by the core, which reaches a
// channel only through the NotificationChannel interface.
package channels

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	modsettings "github.com/sky-as-code/nikki-erp/modules/notification/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

// ArgAppLink is the argument this channel understands: where the notification points when someone
// clicks it in the browser.
//
// The name is declared here, in the channel, and nowhere in the core. That is the whole point of
// the arrangement -- a channel that wanted a different argument would not touch anything else.
const ArgAppLink = "app_link"

// webLinkSchemes is what an app link may start with.
//
// An allow-list rather than a deny-list, because this value ends up in an href: "javascript:" is
// the obvious thing to keep out, but so is every scheme nobody has thought of yet.
var webLinkSchemes = map[string]bool{
	"nikkiweb": true,
	"http":     true,
	"https":    true,
}

// maxAppLinkLength bounds the stored argument. Long enough for a real URL with query parameters,
// short enough that the column cannot be used as free storage.
const maxAppLinkLength = 2000

// WebChannel delivers to the browser.
//
// It sends nothing itself. A notification reaches a connected browser because the send publishes a
// wake-up signal and StreamDispatcher turns that into an event on whichever instance holds the
// stream; the notification is in the inbox either way, whether or not anyone is connected
// (BR-FS 22). So this channel's delivery is the record that the web was a destination -- and its
// real work is validating the argument the browser will act on.
type WebChannel struct {
	settings OrgSettingsReader
}

// OrgSettingsReader is the slice of the settings reader this channel needs.
//
// Declared here rather than imported so that the channel depends on the one question it asks, and
// so that a test can answer it without standing up the settings module.
type OrgSettingsReader interface {
	Bool(ctx corectx.Context, name string, fallback bool) bool
}

func NewWebChannel(settings OrgSettingsReader) *WebChannel {
	return &WebChannel{settings: settings}
}

func (this *WebChannel) Name() modconstants.ChannelName {
	return modconstants.ChannelWeb
}

func (this *WebChannel) ValidateArgs(
	ctx corectx.Context, args itExt.ChannelArgs,
) ft.ClientErrors {
	errs := make(ft.ClientErrors, 0)

	// No arguments at all is valid: a notification that points nowhere in particular is the
	// ordinary case, and the bell simply shows it without a destination.
	if len(args) == 0 {
		return errs
	}

	field := fmt.Sprintf("channel_args.%s.%s", modconstants.ChannelWeb, ArgAppLink)

	raw, given := args[ArgAppLink]
	if !given {
		return errs
	}

	link, isString := raw.(string)
	if !isString {
		errs = append(errs, *ft.NewValidationError(
			field, "err_app_link_invalid", "app link must be a string"))
		return errs
	}

	link = strings.TrimSpace(link)
	if link == "" {
		errs = append(errs, *ft.NewValidationError(
			field, "err_app_link_invalid", "app link must not be empty"))
		return errs
	}

	if len(link) > maxAppLinkLength {
		errs = append(errs, *ft.NewValidationError(
			field, "err_app_link_too_long",
			fmt.Sprintf("app link must be at most %d characters", maxAppLinkLength)))
		return errs
	}

	parsed, err := url.Parse(link)
	if err != nil || parsed.Scheme == "" {
		errs = append(errs, *ft.NewValidationError(
			field, "err_app_link_invalid",
			"app link must be an absolute link, such as 'nikkiweb://notification/notifications'"))
		return errs
	}

	if !webLinkSchemes[strings.ToLower(parsed.Scheme)] {
		errs = append(errs, *ft.NewValidationError(
			field, "err_app_link_scheme_unsupported",
			fmt.Sprintf("'%s' is not a supported link scheme", parsed.Scheme)))
	}

	return errs
}

// IsEnabledFor reads this channel's own setting, and only its own.
//
// orgId is not passed to the settings module: the organization a read applies to is the one the
// request is acting as, which is what stops any caller reading another organization's
// configuration by naming its id. The two agree here because the send resolved and authorized the
// organization before reaching this point.
func (this *WebChannel) IsEnabledFor(ctx corectx.Context, orgId model.Id) (bool, error) {
	if this.settings == nil {
		return modsettings.DefaultWebEnabled, nil
	}

	return this.settings.Bool(
		ctx, modsettings.OrgSettingWebEnabled, modsettings.DefaultWebEnabled), nil
}

// Deliver records that the web was a destination for this notification.
//
// There is no call to make. The notification is already durable and already announced, so the
// honest outcome is success: the browser will have it from the stream if someone is connected, and
// from the inbox if not.
func (this *WebChannel) Deliver(
	ctx context.Context, in itExt.DeliveryInput,
) itExt.DeliveryOutcome {
	return itExt.DeliveryOutcome{Succeeded: true}
}
