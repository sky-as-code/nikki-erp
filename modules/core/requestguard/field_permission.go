package requestguard

import (
	"sort"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Field-level authorization. Grants are additive: a resource-level grant `A:R:scope` covers every
// field of R, and a field-level grant `A:R.f:scope` covers field f alone. Every function here first
// asks the resource-level question, so a caller holding the whole resource pays one lookup.

// HasPermission is AssertPermission as a yes/no answer.
func HasPermission(ctx corectx.Context, requiredPerm Perm) bool {
	return AssertPermission(ctx, requiredPerm) == nil
}

// AssertFieldsPermission authorizes perm on every key of fields - the field map a caller writes or
// asks for. perm names the action, resource and scope; its FieldName is ignored. The error lists
// every missing field expression at once, so one 403 tells an administrator everything to grant.
func AssertFieldsPermission(ctx corectx.Context, perm Perm, fields dmodel.DynamicFields) *ft.ClientErrors {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	return AssertFieldNamesPermission(ctx, perm, names)
}

// AssertFieldNamesPermission is AssertFieldsPermission over a list of field names. No name means
// nothing to authorize beyond an authenticated principal: deciding whether the action itself is
// allowed is AssertPermission's or AssertAnyFieldPermission's job.
func AssertFieldNamesPermission(ctx corectx.Context, perm Perm, fieldNames []string) *ft.ClientErrors {
	resourcePerm := perm.ForField("")
	refused := AssertPermission(ctx, resourcePerm)
	if refused == nil || ctx.GetPermissions().Principal.IsZero() {
		return refused
	}

	missing := make([]string, 0)
	for _, name := range sortedUnique(fieldNames) {
		fieldPerm := resourcePerm.ForField(name)
		if !HasPermission(ctx, fieldPerm) {
			missing = append(missing, exactExpression(fieldPerm))
		}
	}
	if len(missing) == 0 {
		return nil
	}
	cErrs := ft.NewClientErrors()
	cErrs.Append(*ft.NewInsufficientPermissionsError(missing))
	return cErrs
}

// ReadableFields returns the subset of fieldNames the caller holds perm on, in input order. It is
// named for reads, its main use, but answers for whatever action perm names.
func ReadableFields(ctx corectx.Context, perm Perm, fieldNames []string) []string {
	resourcePerm := perm.ForField("")
	if HasPermission(ctx, resourcePerm) {
		return fieldNames
	}
	result := make([]string, 0, len(fieldNames))
	for _, name := range fieldNames {
		if HasPermission(ctx, resourcePerm.ForField(name)) {
			result = append(result, name)
		}
	}
	return result
}

// AssertAnyFieldPermission is the action gate of a field-aware operation: granted when the caller
// holds perm on the whole resource or on at least one of fieldNames. What the caller may then
// touch is narrowed per field by AssertFieldNamesPermission / ReadableFields. A refusal reports the
// resource-level expression, which is the one grant that would answer every field.
func AssertAnyFieldPermission(ctx corectx.Context, perm Perm, fieldNames []string) *ft.ClientErrors {
	resourcePerm := perm.ForField("")
	refused := AssertPermission(ctx, resourcePerm)
	if refused == nil || ctx.GetPermissions().Principal.IsZero() {
		return refused
	}
	for _, name := range fieldNames {
		if HasPermission(ctx, resourcePerm.ForField(name)) {
			return nil
		}
	}
	return refused
}

func sortedUnique(items []string) []string {
	result := dedupe(items)
	sort.Strings(result)
	return result
}
