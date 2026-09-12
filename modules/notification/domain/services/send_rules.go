package services

import (
	"strings"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
)

// NormalizedSend is a send request after the rules of BR 11 have been applied to it: the fields
// are the ones that will be written, not the ones the caller happened to send.
type NormalizedSend struct {
	RecipientUserIds []model.Id
	Title            string
	Message          string
	Severity         models.Severity
	DistributionMode models.DistributionMode

	// Channels is what was asked for, and is nil under distribution mode "all": the sender named
	// nothing, so there is nothing to record as requested.
	Channels []modconstants.ChannelName

	// ResolvedChannels is what the notification will actually be delivered to -- the requested
	// channels under mode "explicit", and the organization's enabled ones under mode "all".
	ResolvedChannels []modconstants.ChannelName

	IdempotencyKey *string
}

// ChannelRegistry is what the normalizer needs to know about the attached channels (BR 5).
//
// It is an interface here, and satisfied by the dispatcher, so that these rules can be tested
// against a set of channels made up for the test rather than whichever ones happen to be wired.
type ChannelRegistry interface {
	Names() []modconstants.ChannelName
	IsRegistered(name string) bool
	EnabledFor(ctx corectx.Context, orgId model.Id) []modconstants.ChannelName
	ValidateArgs(
		ctx corectx.Context,
		args map[string]itExt.ChannelArgs,
		requested []modconstants.ChannelName,
	) ft.ClientErrors
}

// SendNormalizer applies BR 11's rules.
//
// It holds the registry rather than reading a package-level one, because which channels exist is a
// property of how the application was assembled, not of this file.
type SendNormalizer struct {
	channels ChannelRegistry
}

func NewSendNormalizer(channels ChannelRegistry) *SendNormalizer {
	return &SendNormalizer{channels: channels}
}

// NormalizeSend applies BR 11's rules and collects every problem it finds.
//
// Every rule is checked before returning rather than failing on the first, because a caller fixing
// one field at a time against a service that answers one error at a time is a slow way to find out
// it had three.
//
// orgId is needed because resolving "every enabled channel" depends on the organization: the same
// request can be valid for one and not for another.
func (this *SendNormalizer) NormalizeSend(
	ctx corectx.Context, orgId model.Id, request it.SendNotificationRequest,
) (*NormalizedSend, ft.ClientErrors) {
	errs := make(ft.ClientErrors, 0)

	recipients := dedupeIds(request.RecipientUserIds)
	if len(recipients) == 0 {
		errs = append(errs, *ft.NewValidationError(
			"recipients", "err_recipients_required",
			"a notification must have at least one recipient"))
	}

	title := strings.TrimSpace(request.Title)
	if title == "" {
		errs = append(errs, *ft.NewValidationError(
			"title", "err_title_required", "title is required"))
	}

	message := strings.TrimSpace(request.Message)
	if message == "" {
		errs = append(errs, *ft.NewValidationError(
			"message", "err_message_required", "message is required"))
	}

	if strings.TrimSpace(request.SourceModule) == "" {
		errs = append(errs, *ft.NewValidationError(
			"source_module", "err_source_module_required", "source module is required"))
	}

	severity, severityErr := normalizeSeverity(request.Severity)
	if severityErr != nil {
		errs = append(errs, *severityErr)
	}

	mode, channels, channelErrs := this.normalizeChannels(request.Channels)
	errs = append(errs, channelErrs...)

	// Resolution has to happen before the arguments are checked: under mode "all" the sender named
	// no channels, so the set the arguments are checked against is the one the organization has
	// enabled, not one the request carries.
	resolved := channels
	if mode == models.DistributionModeAll {
		resolved = this.channels.EnabledFor(ctx, orgId)
	}

	errs = append(errs, this.channels.ValidateArgs(ctx, toChannelArgs(request.ChannelArgs), resolved)...)

	if len(errs) > 0 {
		return nil, errs
	}

	return &NormalizedSend{
		RecipientUserIds: recipients,
		Title:            title,
		Message:          message,
		Severity:         severity,
		DistributionMode: mode,
		Channels:         channels,
		ResolvedChannels: resolved,
		IdempotencyKey:   normalizeIdempotencyKey(request.IdempotencyKey),
	}, nil
}

// toChannelArgs re-types the request's plain maps as the channels' own argument type. The shape is
// identical; the named type is what stops the core from treating one channel's arguments as
// something it may read.
func toChannelArgs(raw map[string]map[string]any) map[string]itExt.ChannelArgs {
	if len(raw) == 0 {
		return nil
	}

	args := make(map[string]itExt.ChannelArgs, len(raw))
	for name, values := range raw {
		args[name] = itExt.ChannelArgs(values)
	}

	return args
}

// normalizeSeverity defaults an omitted severity to info (BR 11.4) and rejects an unknown one.
func normalizeSeverity(raw string) (models.Severity, *ft.ClientErrorItem) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return models.SeverityInfo, nil
	}

	severity := models.Severity(trimmed)
	switch severity {
	case models.SeverityInfo, models.SeveritySuccess, models.SeverityWarning, models.SeverityDanger:
		return severity, nil
	}
	return "", ft.NewValidationError(
		"severity", "err_severity_unknown", "unknown severity: "+trimmed)
}

// normalizeChannels decides the distribution mode and validates the names.
//
// The distinction that matters is nil versus empty: a caller that named no channels wants every
// enabled one (BR 11.5), while a caller that sent an empty list has asked for delivery to nothing,
// which is a mistake rather than an instruction (BR 11.7).
func (this *SendNormalizer) normalizeChannels(
	raw []string,
) (models.DistributionMode, []modconstants.ChannelName, ft.ClientErrors) {
	if raw == nil {
		return models.DistributionModeAll, nil, nil
	}

	errs := make(ft.ClientErrors, 0)
	if len(raw) == 0 {
		errs = append(errs, *ft.NewValidationError(
			"channels", "err_channels_empty",
			"channels must name at least one channel; omit it to use every enabled channel"))
		return "", nil, errs
	}

	seen := make(map[string]bool, len(raw))
	channels := make([]modconstants.ChannelName, 0, len(raw))
	for _, name := range raw {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" || seen[trimmed] {
			// Duplicates collapse rather than erroring (BR 11.9): asking for the same channel
			// twice is unambiguous about what was wanted.
			continue
		}
		seen[trimmed] = true

		if !this.channels.IsRegistered(trimmed) {
			errs = append(errs, *ft.NewValidationError(
				"channels", "err_channel_unknown", "unknown channel: "+trimmed))
			continue
		}
		channels = append(channels, modconstants.ChannelName(trimmed))
	}

	if len(errs) > 0 {
		return "", nil, errs
	}
	if len(channels) == 0 {
		errs = append(errs, *ft.NewValidationError(
			"channels", "err_channels_empty", "channels must name at least one channel"))
		return "", nil, errs
	}
	return models.DistributionModeExplicit, channels, nil
}

// dedupeIds collapses repeats while keeping the caller's order (BR 11.2). Sending the same person
// twice must produce one recipient, not two rows that would each show up in their inbox.
func dedupeIds(raw []string) []model.Id {
	seen := make(map[string]bool, len(raw))
	ids := make([]model.Id, 0, len(raw))
	for _, id := range raw {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		ids = append(ids, model.Id(trimmed))
	}
	return ids
}

// normalizeIdempotencyKey treats a blank key as no key at all, so that a caller sending an empty
// string does not claim the one "" slot per organization and module.
func normalizeIdempotencyKey(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
