package json

import (
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)


const (
	// ErrKeyMalformed is reported when the input is not parseable JSON at all.
	ErrKeyMalformed = "dynamicmodel.json.malformed"

	// ErrKeyBadSchema is reported when the JSON Schema itself is unusable.
	ErrKeyBadSchema = "dynamicmodel.json.badSchema"

	// ErrKeyInvalid is reported for each JSON Schema constraint the input violates.
	ErrKeyInvalid = "dynamicmodel.json.invalid"
)

// ValidateSchemaJson validates jsonStr against jsonSchema.
// It returns an empty (non-nil) ClientErrors when the input is valid, and one item per
// violation otherwise, with Field set to the dotted path of the offending value.
func ValidateSchemaJson(jsonStr string, jsonSchema string) ft.ClientErrors {
	clientErrs := ft.ClientErrors{}

	compiled, err := compileSchema(jsonSchema)
	if err != nil {
		clientErrs.Append(*ft.NewAnonymousValidationError(ErrKeyBadSchema, err.Error()))
		return clientErrs
	}

	instance, err := jsonschema.UnmarshalJSON(strings.NewReader(jsonStr))
	if err != nil {
		clientErrs.Append(*ft.NewAnonymousValidationError(ErrKeyMalformed, err.Error()))
		return clientErrs
	}

	if err := compiled.Validate(instance); err != nil {
		validationErr, ok := err.(*jsonschema.ValidationError)
		if !ok {
			clientErrs.Append(*ft.NewAnonymousValidationError(ErrKeyInvalid, err.Error()))
			return clientErrs
		}
		collectLeafErrors(validationErr.BasicOutput(), &clientErrs)
	}

	return clientErrs
}

// collectLeafErrors walks a basic output tree and appends only the leaves, which carry
// the actual constraint that failed. Parent units merely restate that a subtree failed.
func collectLeafErrors(unit *jsonschema.OutputUnit, clientErrs *ft.ClientErrors) {
	if unit == nil || unit.Valid {
		return
	}

	if len(unit.Errors) == 0 {
		message := ""
		if unit.Error != nil {
			// Use String(), which carries the printer BasicOutput attached; reading
			// Kind.LocalizedString(nil) directly panics on a nil printer.
			message = unit.Error.String()
		}
		clientErrs.Append(*ft.NewValidationError(
			pointerToPath(unit.InstanceLocation),
			ErrKeyInvalid,
			message,
		))
		return
	}

	for i := range unit.Errors {
		collectLeafErrors(&unit.Errors[i], clientErrs)
	}
}

// pointerToPath turns a JSON Pointer ("/fields/2/data_type") into the dotted path
// ClientErrorItem.Field uses elsewhere in the codebase ("fields.2.data_type").
func pointerToPath(pointer string) string {
	trimmed := strings.TrimPrefix(pointer, "/")
	if trimmed == "" {
		return ""
	}

	segments := strings.Split(trimmed, "/")
	for i, segment := range segments {
		// Unescape per RFC 6901: ~1 is "/" and ~0 is "~", in that order.
		segment = strings.ReplaceAll(segment, "~1", "/")
		segments[i] = strings.ReplaceAll(segment, "~0", "~")
	}

	return strings.Join(segments, ".")
}
