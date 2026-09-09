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

// NewUnauthenticatedError reports that no principal was established for an operation that
// requires one.
//
// Distinct from insufficient permissions on purpose: that one means "we know who you are and you
// may not", this one means "nothing authenticated you". Reporting the former for the latter sends
// an administrator looking for an entitlement to grant when the real fault is a caller - usually
// a job or a consumer - that never established a principal.
func NewUnauthenticatedError() *ClientErrorItem {
	return NewAuthorizationError(
		ErrorKey("err_unauthenticated"),
		"This operation requires an authenticated principal.",
	)
}

// NewInvalidOrganizationContextError reports that the organization the caller is acting in does
// not permit the record being reached for.
//
// Authorization is entitlement AND organization scope: holding `update:product:org` says nothing
// about a product in an org the caller is not acting within.
func NewInvalidOrganizationContextError(orgId string) *ClientErrorItem {
	return NewAuthorizationError(
		ErrorKey("err_invalid_organization_context"),
		"The current organization context does not permit this operation: {{org_id}}",
		map[string]any{
			"org_id": orgId,
		},
	)
}
