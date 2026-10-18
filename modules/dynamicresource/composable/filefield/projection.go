package filefield

import (
	"slices"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

const paramFields = "fields"

// appendHiddenKeys makes a read carry every file field's key, so its URL can be presigned even when
// the projection leaves the key out — the default columns of a list, say, which show the URL and
// not the storage key. It returns the keys it added; the caller strips them from the response.
//
// fallback is the projection a read without `fields` gets: nil for a single-record read, which
// returns every column already.
func appendHiddenKeys(query dmodel.DynamicFields, fields []FileField, fallback []string) []string {
	projection := requestedFields(query)
	if len(projection) == 0 {
		projection = fallback
	}
	if len(projection) == 0 || slices.Contains(projection, corecrud.FieldsWildcard) {
		return nil
	}

	var added []string
	for _, field := range fields {
		if !slices.Contains(projection, field.KeyField) {
			added = append(added, field.KeyField)
		}
	}
	if len(added) > 0 {
		// Concat, not append: fallback may be the schema's shared default field list.
		query[paramFields] = slices.Concat(projection, added)
	}
	return added
}

func stripHiddenKeys(records []dmodel.DynamicFields, desired []string, added []string) []string {
	if len(added) == 0 {
		return desired
	}
	for _, record := range records {
		for _, key := range added {
			delete(record, key)
		}
	}
	return slices.DeleteFunc(slices.Clone(desired), func(name string) bool {
		return slices.Contains(added, name)
	})
}

func requestedFields(query dmodel.DynamicFields) []string {
	switch typed := query[paramFields].(type) {
	case []string:
		return typed
	case []any:
		names := make([]string, 0, len(typed))
		for _, item := range typed {
			if name, ok := item.(string); ok {
				names = append(names, name)
			}
		}
		return names
	default:
		return nil
	}
}
