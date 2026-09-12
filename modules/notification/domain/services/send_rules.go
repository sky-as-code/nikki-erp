package services

import (
	"strings"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// NormalizedSend is a send request after the rules of BR 11 have been applied to it: the fields
// are the ones that will be written, not the ones the caller happened to send.
type NormalizedSend struct {
	RecipientUserIds []model.Id
	Title            string
	Message          string
	Severity         models.Severity
	DistributionMode models.DistributionMode

	// Channels is what was asked for. It is resolved against the organization's configuration
	// separately, because that needs the settings and this does not.
	Channels []modconstants.ChannelName

	IdempotencyKey *string
}

// registeredChannels is the channel registry of BR 5, kept in code rather than in a table: adding
// a channel means writing an adapter, which a row could not do on its own.
//
// Phase 1 registers only the web channel. A name outside this set is rejected rather than ignored
// (BR 11.8), so that a sender's typo surfaces at the call instead of silently delivering nowhere.
var registeredChannels = map[modconstants.ChannelName]bool{
	modconstants.ChannelWeb: true,
}

// IsRegisteredChannel reports whether a channel name is one the system knows.
func IsRegisteredChannel(name string) bool {
	return registeredChannels[modconstants.ChannelName(name)]
}

// NormalizeSend applies BR 11's rules and collects every problem it finds.
//
// Every rule is checked before returning rather than failing on the first, because a caller fixing
// one field at a time against a service that answers one error at a time is a slow way to find out
// it had three.
func NormalizeSend(request it.SendNotificationRequest) (*NormalizedSend, ft.ClientErrors) {
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

	mode, channels, channelErrs := normalizeChannels(request.Channels)
	errs = append(errs, channelErrs...)

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
		IdempotencyKey:   normalizeIdempotencyKey(request.IdempotencyKey),
	}, nil
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
func normalizeChannels(
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

		if !IsRegisteredChannel(trimmed) {
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
