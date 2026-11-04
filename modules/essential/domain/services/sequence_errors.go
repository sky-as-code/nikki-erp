package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
)

// The Sequence error contract (BR §12.1). Each key is the BR code, lowercased and namespaced, so a
// client recognizes the condition without matching message text.
var (
	ErrKeySeqInvalidInput         = seqErrorKey("err_sequence_invalid_input")
	ErrKeySeqInvalidTemplate      = seqErrorKey("err_sequence_invalid_template")
	ErrKeySeqInvalidTimezone      = seqErrorKey("err_sequence_invalid_timezone")
	ErrKeySeqInvalidIdempotencyId = seqErrorKey("err_sequence_invalid_idempotency_id")
	ErrKeySeqConfigurationLocked  = seqErrorKey("err_sequence_configuration_locked")
	ErrKeySeqArchived             = seqErrorKey("err_sequence_archived")
	ErrKeySeqExhausted            = seqErrorKey("err_sequence_exhausted")
	ErrKeySeqCapacityExceeded     = seqErrorKey("err_sequence_capacity_exceeded")
	ErrKeySeqIdempotencyConflict  = seqErrorKey("err_sequence_idempotency_conflict")
	ErrKeySeqIdempotencyExpired   = seqErrorKey("err_sequence_idempotency_expired")
	ErrKeySeqRegistrationConflict = seqErrorKey("err_sequence_registration_conflict")
	ErrKeySeqPadding              = seqErrorKey("err_sequence_padding")
	ErrKeySeqReadOnlyField        = seqErrorKey("err_sequence_read_only_field")
)

func seqErrorKey(key string) string {
	return ft.ErrorKey(key, modconstants.EssentialModuleName)
}

func newSeqValidation(field string, key string, message string, vars ...map[string]any) *ft.ClientErrorItem {
	if field == "" {
		return ft.NewAnonymousValidationError(key, message, vars...)
	}
	return ft.NewValidationError(field, key, message, vars...)
}

func newSeqViolation(field string, key string, message string, vars ...map[string]any) *ft.ClientErrorItem {
	if field == "" {
		return ft.NewAnonymousBusinessViolation(key, message, vars...)
	}
	return ft.NewBusinessViolation(field, key, message, vars...)
}

// singleError wraps one item into the ClientErrors a result carries.
func singleError(item *ft.ClientErrorItem) ft.ClientErrors {
	return ft.ClientErrors{*item}
}
