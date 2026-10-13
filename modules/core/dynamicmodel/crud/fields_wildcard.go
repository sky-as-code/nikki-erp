package crud

import (
	"slices"

	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// FieldsWildcard in a `fields` selection stands for every readable field of the root schema:
// persisted columns and computed fields alike, edges excluded — the same as sending no `fields`.
// It lets a client add edge paths (`*`, `template.name`) without enumerating the record's own
// fields, which a non-empty selection otherwise requires because it replaces the default.
const FieldsWildcard = string(orm.SelectAllColumns)

// NormalizeFieldsWildcard prepares a selection holding FieldsWildcard for the read pipeline: the
// token stays (the query builder renders it as `t0.*`), the root's physical columns it already
// covers are dropped, and every computed field is named explicitly so the computed-field layer
// evaluates them. Other entries (edge paths, bare edges) are kept in order. It is idempotent, and
// returns the input as-is when it holds no wildcard.
func NormalizeFieldsWildcard(schema *dmodel.ModelSchema, fields []string) []string {
	if schema == nil || !slices.Contains(fields, FieldsWildcard) {
		return fields
	}
	readable := schema.ReadableFields()
	seen := datastructure.Set[string]{}
	result := make([]string, 0, len(fields)+len(readable))
	appendOnce := func(name string) {
		if !seen.Contains(name) {
			seen.Add(name)
			result = append(result, name)
		}
	}
	appendOnce(FieldsWildcard)
	for _, name := range fields {
		if field, ok := schema.Field(name); ok && !field.IsVirtual() {
			continue
		}
		appendOnce(name)
	}
	for _, field := range readable {
		if field.IsVirtual() {
			appendOnce(field.Name())
		}
	}
	return result
}

// ExpandFieldsWildcard replaces each FieldsWildcard with the schema's readable field names, in
// place, dropping duplicates and keeping every other entry untouched. It names what a wildcard
// selection returns, for `desired_fields`; the input is returned as-is when it holds no wildcard.
func ExpandFieldsWildcard(schema *dmodel.ModelSchema, fields []string) []string {
	if schema == nil || !slices.Contains(fields, FieldsWildcard) {
		return fields
	}
	seen := datastructure.Set[string]{}
	result := make([]string, 0, len(fields)+len(schema.ReadableFields()))
	appendOnce := func(name string) {
		if !seen.Contains(name) {
			seen.Add(name)
			result = append(result, name)
		}
	}
	for _, name := range fields {
		if name != FieldsWildcard {
			appendOnce(name)
			continue
		}
		for _, field := range schema.ReadableFields() {
			appendOnce(field.Name())
		}
	}
	return result
}
