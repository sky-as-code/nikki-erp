package fault

func NewNotFoundError(field string) *ClientErrorItem {
	return NewBusinessViolation(
		field,
		ErrorKey("err_not_found"),
		"The desired data could not be found",
	)
}

func NewNotFoundValError[T any](values []T) *ClientErrorItem {
	return NewAnonymousBusinessViolation(
		ErrorKey("err_value_not_found"),
		"Value(s) could not be found: {{values}}",
		map[string]any{
			"values": values,
		},
	)
}

func NewAnonymousNotFoundError() *ClientErrorItem {
	return NewAnonymousBusinessViolation(
		ErrorKey("err_not_found"),
		"The desired data could not be found",
	)
}

func NewEtagMismatchedError() *ClientErrorItem {
	return NewBusinessViolation(
		"etag",
		ErrorKey("err_etag_mismatched"),
		"This data has been modified by someone else",
	)
}

func NewExclusiveFieldsError(conflictFields []string) *ClientErrorItem {
	return NewAnonymousBusinessViolation(
		ErrorKey("err_exclusive_fields"),
		"Only one of these fields can have a value: {{fields}}",
		map[string]any{
			"fields": conflictFields,
		},
	)
}

func NewExclusiveFieldsMissingError(missingFields []string) *ClientErrorItem {
	return NewAnonymousBusinessViolation(
		ErrorKey("err_exclusive_fields_missing"),
		"One of these fields (not all of them) is required: {{fields}}",
		map[string]any{
			"fields": missingFields,
		},
	)
}

func NewOverlappedFieldsError(overlappedFields []string) *ClientErrorItem {
	return NewAnonymousBusinessViolation(
		ErrorKey("err_overlapped_fields"),
		"These fields must not have overlapping values: {{fields}}",
		map[string]any{
			"fields": overlappedFields,
		},
	)
}

func NewInsufficientPermissionsError(requiredEntitlements []string) *ClientErrorItem {
	return NewAuthorizationError(
		ErrorKey("err_insufficient_permissions"),
		"Insufficient permissions. Request following entitlement(s) to perform this action: {{entitlements}}",
		map[string]any{
			"entitlements": requiredEntitlements,
		},
	)
}
