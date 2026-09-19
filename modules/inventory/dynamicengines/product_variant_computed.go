package dynamicengines

import (
	"fmt"
	"sort"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

const (
	// VariantDisplayNameFn and VariantAttributeSummaryFn are the function names
	// product_variant.json declares. A schema naming one nobody registers fails the boot check,
	// so these constants are the single spelling shared by declaration and registration.
	VariantDisplayNameFn      = "inventory.variant_display_name"
	VariantAttributeSummaryFn = "inventory.variant_attribute_summary"

	// maxAttributesPerVariant bounds the page size of the attribute reads. A variant with more
	// attributes than this would lose the surplus from its display name; the cap is far above the
	// handful of axes a sellable product is built from.
	maxAttributesPerVariant = 20
)

// variantAttributes is one row of the attribute read: which variant it belongs to, what the
// attribute is called, what the variant's value for it is called, and where it sorts.
type variantAttributes struct {
	variantId string
	attribute string
	value     string
	sequence  int64
}

// newVariantComputedFunctions builds the "function"-kind computed fields of product_variant.
//
// Both read the same three-edge walk, so they share one batched fetch: the engine hands the whole
// page to each function, and resolving per row would turn a listing into the N+1 this module
// forbids by name. The values are fetched once per call rather than cached, because an attribute
// rename must show on the next read.
func newVariantComputedFunctions(
	lookupOnion func(schemaName string) (composable.DynamicResourceEngineOnion, bool),
) map[string]composable.ComputedFieldFn {
	displayName := func(ctx corectx.Context, req composable.ComputeFnRequest) ([]any, error) {
		byVariant, err := fetchVariantAttributes(ctx, lookupOnion, req.Models)
		if err != nil {
			return nil, err
		}
		out := make([]any, len(req.Models))
		for i, row := range req.Models {
			out[i] = buildDisplayName(row, byVariant[readVariantId(row)])
		}
		return out, nil
	}

	attributeSummary := func(ctx corectx.Context, req composable.ComputeFnRequest) ([]any, error) {
		byVariant, err := fetchVariantAttributes(ctx, lookupOnion, req.Models)
		if err != nil {
			return nil, err
		}
		out := make([]any, len(req.Models))
		for i, row := range req.Models {
			pairs := byVariant[readVariantId(row)]
			labels := make([]string, 0, len(pairs))
			for _, pair := range pairs {
				labels = append(labels, fmt.Sprintf("%s: %s", pair.attribute, pair.value))
			}
			out[i] = labels
		}
		return out, nil
	}

	return map[string]composable.ComputedFieldFn{
		VariantDisplayNameFn:      displayName,
		VariantAttributeSummaryFn: attributeSummary,
	}
}

// buildDisplayName renders "Coca Cola (No sugar, Big)".
//
// A variant with no attributes is the ordinary single-variant case, not a defect, so it is named
// by its template alone rather than gaining an empty pair of brackets.
func buildDisplayName(row dmodel.DynamicFields, attributes []variantAttributes) string {
	name := readLangText(row[models.ProductVariantFieldTemplateName])
	if name == "" {
		// The template name is what identifies the variant to a person; with none, the SKU is the
		// only other thing on the row that does, and an empty cell would identify nothing.
		name = readString(row[models.ProductVariantFieldSku])
	}
	if len(attributes) == 0 {
		return name
	}
	values := make([]string, 0, len(attributes))
	for _, attribute := range attributes {
		values = append(values, attribute.value)
	}
	return fmt.Sprintf("%s (%s)", name, strings.Join(values, ", "))
}

// fetchVariantAttributes reads every attribute value of the whole page in one search, grouped by
// variant and ordered the way the template declares them.
func fetchVariantAttributes(
	ctx corectx.Context,
	lookupOnion func(schemaName string) (composable.DynamicResourceEngineOnion, bool),
	rows []dmodel.DynamicFields,
) (map[string][]variantAttributes, error) {
	byVariant := map[string][]variantAttributes{}
	variantIds := distinctVariantIds(rows)
	if len(variantIds) == 0 {
		return byVariant, nil
	}
	// Walked one edge at a time rather than as a single `template_attribute_value.attribute_value
	// .attribute.name` selection: a `fields=` projection resolves exactly one dot
	// (orm.MaxSelectGraphColumnDots), and asking for more is refused outright. Search *conditions*
	// reach five, which is what makes each hop's id filter legal — the two limits are not the same.
	//
	// Three queries for the whole page, not per row: each hop filters on the ids the previous one
	// returned.
	joinRows, err := searchByIds(ctx, lookupOnion,
		models.ProductVariantAttributeValueSchemaName,
		models.ProductVariantAttributeValueFieldProductVariantId, variantIds,
		[]string{
			models.ProductVariantAttributeValueFieldProductVariantId,
			models.ProductVariantAttributeValueFieldTemplateAttributeValueId,
		},
		len(variantIds)*maxAttributesPerVariant)
	if err != nil {
		return nil, err
	}
	if len(joinRows) == 0 {
		return byVariant, nil
	}

	templateValueRows, err := searchByIds(ctx, lookupOnion,
		models.ProductTemplateAttributeValueSchemaName,
		basemodel.FieldId,
		distinctStrings(joinRows, models.ProductVariantAttributeValueFieldTemplateAttributeValueId),
		[]string{
			basemodel.FieldId,
			models.ProductTemplateAttributeValueFieldAttributeValueId,
			models.ProductTemplateAttributeValueFieldSequence,
		},
		len(joinRows))
	if err != nil {
		return nil, err
	}

	valueRows, err := searchByIds(ctx, lookupOnion,
		models.ProductAttributeValueSchemaName,
		basemodel.FieldId,
		distinctStrings(templateValueRows, models.ProductTemplateAttributeValueFieldAttributeValueId),
		[]string{
			basemodel.FieldId,
			models.ProductAttributeValueFieldName,
			models.ProductAttributeValueFieldAttributeId,
		},
		len(templateValueRows))
	if err != nil {
		return nil, err
	}

	attributeRows, err := searchByIds(ctx, lookupOnion,
		models.ProductAttributeSchemaName,
		basemodel.FieldId,
		distinctStrings(valueRows, models.ProductAttributeValueFieldAttributeId),
		[]string{basemodel.FieldId, models.ProductAttributeFieldName},
		len(valueRows))
	if err != nil {
		return nil, err
	}

	templateValueById := indexById(templateValueRows)
	valueById := indexById(valueRows)
	attributeNameById := map[string]string{}
	for id, row := range indexById(attributeRows) {
		attributeNameById[id] = readLangText(row[models.ProductAttributeFieldName])
	}

	for _, join := range joinRows {
		variantId := readString(join[models.ProductVariantAttributeValueFieldProductVariantId])
		if variantId == "" {
			continue
		}
		templateValue := templateValueById[readString(
			join[models.ProductVariantAttributeValueFieldTemplateAttributeValueId])]
		if templateValue == nil {
			continue
		}
		value := valueById[readString(
			templateValue[models.ProductTemplateAttributeValueFieldAttributeValueId])]
		if value == nil {
			continue
		}
		byVariant[variantId] = append(byVariant[variantId], variantAttributes{
			variantId: variantId,
			attribute: attributeNameById[readString(value[models.ProductAttributeValueFieldAttributeId])],
			value:     readLangText(value[models.ProductAttributeValueFieldName]),
			sequence:  readInt(templateValue[models.ProductTemplateAttributeValueFieldSequence]),
		})
	}
	for id := range byVariant {
		entries := byVariant[id]
		// The template declares the order its attributes read in, so "Coca Cola (No sugar, Big)"
		// keeps that order rather than whatever the join returned.
		sort.SliceStable(entries, func(a, b int) bool {
			return entries[a].sequence < entries[b].sequence
		})
		byVariant[id] = entries
	}
	return byVariant, nil
}

// searchByIds reads one hop: the rows of schemaName whose idField is among ids, projecting only
// flat columns. Returns no rows for an empty id list rather than searching for everything.
func searchByIds(
	ctx corectx.Context,
	lookupOnion func(schemaName string) (composable.DynamicResourceEngineOnion, bool),
	schemaName string, idField string, ids []string, fields []string, size int,
) ([]dmodel.DynamicFields, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	onion, ok := lookupOnion(schemaName)
	if !ok {
		return nil, errors.Errorf("computed %q: no engine serves %q", VariantDisplayNameFn, schemaName)
	}
	keys := make([]any, len(ids))
	for i, id := range ids {
		keys[i] = id
	}
	graph := dmodel.NewSearchGraph()
	graph.NewCondition(idField, dmodel.In, keys...)

	result, err := onion.Repository().Search(ctx, dyn.RepoSearchParam{
		Fields: fields,
		Graph:  graph,
		Page:   0,
		Size:   size,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "computed %q: read %q", VariantDisplayNameFn, schemaName)
	}
	if result == nil {
		return nil, nil
	}
	if result.ClientErrors.Count() > 0 {
		return nil, errors.Errorf("computed %q: read %q: %v",
			VariantDisplayNameFn, schemaName, result.ClientErrors.ToError())
	}
	return result.Data.Items, nil
}

func indexById(rows []dmodel.DynamicFields) map[string]dmodel.DynamicFields {
	byId := make(map[string]dmodel.DynamicFields, len(rows))
	for _, row := range rows {
		if id := readString(row[basemodel.FieldId]); id != "" {
			byId[id] = row
		}
	}
	return byId
}

func distinctStrings(rows []dmodel.DynamicFields, field string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		value := readString(row[field])
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func distinctVariantIds(rows []dmodel.DynamicFields) []string {
	seen := map[string]bool{}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		id := readVariantId(row)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}

func readVariantId(row dmodel.DynamicFields) string {
	return readString(row[basemodel.FieldId])
}

func readString(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	if text, ok := value.(*string); ok && text != nil {
		return *text
	}
	return ""
}

// readLangText renders a name that may be a plain string or a per-locale map. A LangJson
// interpolated directly would print its Go representation, never the name.
func readLangText(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case model.LangJson:
		return firstLangValue(typed)
	case map[string]string:
		converted := model.LangJson{}
		for key, text := range typed {
			converted[key] = text
		}
		return firstLangValue(converted)
	case map[string]any:
		converted := model.LangJson{}
		for key, text := range typed {
			converted[key] = readString(text)
		}
		return firstLangValue(converted)
	}
	return readString(value)
}

// firstLangValue picks a locale deterministically, so the same row does not change name between
// reads. The repository resolves the request's language when one is set; this is the fallback for
// the internal read here, which asks for none.
func firstLangValue(lang model.LangJson) string {
	keys := make([]string, 0, len(lang))
	for key := range lang {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if lang[key] != "" {
			return lang[key]
		}
	}
	return ""
}

func readInt(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case float64:
		return int64(typed)
	}
	return 0
}
