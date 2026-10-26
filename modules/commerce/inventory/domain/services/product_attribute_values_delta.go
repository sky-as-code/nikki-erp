package services

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const (
	valuesDeltaFieldAdd    = itProduct.ProductAttributeFieldValues + ".add"
	valuesDeltaFieldUpdate = itProduct.ProductAttributeFieldValues + ".update"
	valuesDeltaFieldRemove = itProduct.ProductAttributeFieldValues + ".remove"
)

// valuesDeltaRowFields are the value fields a delta row may set. attribute_id and org_id are the
// parent's, so a row naming them is ignored rather than trusted.
var valuesDeltaRowFields = []string{
	models.ProductAttributeValueFieldCode,
	models.ProductAttributeValueFieldName,
	models.ProductAttributeValueFieldSequence,
	models.ProductAttributeValueFieldColorHex,
}

func valuesDeltaRowPath(list string, index int) string {
	return fmt.Sprintf("%s[%d]", list, index)
}

// prefixRowErrors points the errors a value write reported at the delta row that caused them.
func prefixRowErrors(cErrs ft.ClientErrors, prefix string) ft.ClientErrors {
	out := make(ft.ClientErrors, 0, len(cErrs))
	for _, item := range cErrs {
		if item.Field == "" {
			item.Field = prefix
		} else {
			item.Field = prefix + "." + item.Field
		}
		out = append(out, item)
	}
	return out
}

// pickDeltaRow copies the writable fields of a delta row into a fresh command.
func pickDeltaRow(row dmodel.DynamicFields, extra ...string) dmodel.DynamicFields {
	out := dmodel.DynamicFields{}
	for _, field := range append(append([]string{}, valuesDeltaRowFields...), extra...) {
		if raw, ok := row[field]; ok {
			out[field] = raw
		}
	}
	return out
}

// validateValuesDeltaShape checks what can be checked without reading the database: list sizes,
// the create-only-adds rule, and ids that are missing or named twice.
func validateValuesDeltaShape(delta itProduct.ProductAttributeValuesDelta, isCreate bool) ft.ClientErrors {
	out := ft.ClientErrors{}
	maxItems := itProduct.ProductAttributeValuesDeltaMaxItems
	for field, size := range map[string]int{
		valuesDeltaFieldAdd: len(delta.Add), valuesDeltaFieldUpdate: len(delta.Update),
		valuesDeltaFieldRemove: len(delta.Remove),
	} {
		if size > maxItems {
			out.Append(*ft.NewValidationError(field, ft.ErrorKey("product_attribute.values_too_many", "inventory"),
				"at most {{max}} values can be sent in one list", map[string]any{"max": maxItems}))
		}
	}
	if isCreate && (len(delta.Update) > 0 || len(delta.Remove) > 0) {
		out.Append(*ft.NewValidationError(itProduct.ProductAttributeFieldValues,
			ft.ErrorKey("product_attribute.values_create_add_only", "inventory"),
			"a new attribute has no values to update or remove; send add only"))
	}
	if out.Count() > 0 {
		return out
	}
	return validateDeltaIds(delta)
}

func validateDeltaIds(delta itProduct.ProductAttributeValuesDelta) ft.ClientErrors {
	out := ft.ClientErrors{}
	seen := map[model.Id]bool{}
	check := func(id model.Id, path string) {
		if id == "" {
			out.Append(*ft.NewValidationError(path, ft.ErrorKey("err_missing_required_field"), "id is required"))
			return
		}
		if seen[id] {
			out.Append(*ft.NewValidationError(path, ft.ErrorKey("product_attribute.values_duplicate_id", "inventory"),
				"value {{id}} appears more than once in the delta", map[string]any{"id": string(id)}))
		}
		seen[id] = true
	}
	for i, row := range delta.Update {
		check(deltaRowId(row), valuesDeltaRowPath(valuesDeltaFieldUpdate, i)+"."+basemodel.FieldId)
	}
	for i, id := range delta.Remove {
		check(id, valuesDeltaRowPath(valuesDeltaFieldRemove, i))
	}
	return out
}

func deltaRowId(row dmodel.DynamicFields) model.Id {
	return composable.ParamId(row)
}

// finalValueEntry is one value of the attribute as it will stand after the delta. path is empty
// for an untouched existing value, which is never reported: a clash between two stored rows
// predates this request.
type finalValueEntry struct {
	value attributeValueName
	path  string
}

// validateValuesDeltaFinalState checks the delta against the attribute's stored values (archived
// ones included): every updated or removed id must belong to the attribute, and no name may
// clash, per locale after normalization, with any other value of the final state.
func validateValuesDeltaFinalState(
	existing []attributeValueName, delta itProduct.ProductAttributeValuesDelta,
) ft.ClientErrors {
	entries, out := buildFinalValueState(existing, delta)
	if out.Count() > 0 {
		return out
	}
	values := make([]attributeValueName, len(entries))
	for i, entry := range entries {
		values[i] = entry.value
	}
	for _, entry := range entries {
		if entry.path == "" {
			continue
		}
		if clash := findNameClash(entry.value, values); clash != nil {
			out.Append(*newValueNameDuplicateError(entry.path, *clash))
		}
	}
	return out
}

func buildFinalValueState(
	existing []attributeValueName, delta itProduct.ProductAttributeValuesDelta,
) ([]finalValueEntry, ft.ClientErrors) {
	out := ft.ClientErrors{}
	byId := make(map[model.Id]*finalValueEntry, len(existing))
	order := make([]model.Id, 0, len(existing))
	for _, value := range existing {
		byId[value.id] = &finalValueEntry{value: value}
		order = append(order, value.id)
	}
	for i, id := range delta.Remove {
		if _, ok := byId[id]; !ok {
			out.Append(*newValueNotInAttributeError(valuesDeltaRowPath(valuesDeltaFieldRemove, i), id))
			continue
		}
		delete(byId, id)
	}
	for i, row := range delta.Update {
		id := deltaRowId(row)
		entry, ok := byId[id]
		if !ok {
			out.Append(*newValueNotInAttributeError(valuesDeltaRowPath(valuesDeltaFieldUpdate, i)+"."+basemodel.FieldId, id))
			continue
		}
		if name := row.GetLangJson(models.ProductAttributeValueFieldName); name != nil {
			entry.value.name = *name
			entry.path = valuesDeltaRowPath(valuesDeltaFieldUpdate, i) + "." + models.ProductAttributeValueFieldName
		}
	}
	entries := make([]finalValueEntry, 0, len(byId)+len(delta.Add))
	for _, id := range order {
		if entry, ok := byId[id]; ok {
			entries = append(entries, *entry)
		}
	}
	for i, row := range delta.Add {
		entries = append(entries, newAddedEntry(row, i))
	}
	return entries, out
}

// newAddedEntry gives an added row a synthetic id, so the name comparison can tell it from its
// siblings the same way it tells stored rows apart.
func newAddedEntry(row dmodel.DynamicFields, index int) finalValueEntry {
	value := attributeValueName{id: model.Id(fmt.Sprintf("+add#%d", index))}
	if name := row.GetLangJson(models.ProductAttributeValueFieldName); name != nil {
		value.name = *name
	}
	return finalValueEntry{
		value: value,
		path:  valuesDeltaRowPath(valuesDeltaFieldAdd, index) + "." + models.ProductAttributeValueFieldName,
	}
}

func newValueNameDuplicateError(field string, clash nameClash) *ft.ClientErrorItem {
	return ft.NewValidationError(field, ft.ErrorKey("product_attribute_value.name_duplicate", "inventory"),
		"another value of this attribute, archived ones included, is already named '{{name}}' ({{locale}})",
		map[string]any{"name": clash.name, "locale": string(clash.locale)})
}

func newValueNotInAttributeError(field string, id model.Id) *ft.ClientErrorItem {
	return ft.NewBusinessViolation(field, ft.ErrorKey("product_attribute_value.not_in_attribute", "inventory"),
		"value {{id}} does not belong to this attribute", map[string]any{"id": string(id)})
}
