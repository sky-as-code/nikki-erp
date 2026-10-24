package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

const (
	// usageScanLimit bounds the rows an attribute or value usage check reads. A usage message
	// needs only a count and a few ids, so a larger fan-out is reported as at least this many.
	usageScanLimit = 1000
	// usageListedTemplates is how many template ids a refusal names.
	usageListedTemplates = 5
	// maxAttributeValuesScanned bounds the siblings the value-name rule compares against.
	maxAttributeValuesScanned = 5000
)

// searchAttributeRows runs a bounded search on another resource's repository. IncludeArchived stays nil,
// so archived rows take part: every usage rule here counts them.
func searchAttributeRows(
	ctx corectx.Context, schemaName string, graph *dmodel.SearchGraph, fields []string, limit int,
) ([]dmodel.DynamicFields, error) {
	repo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	found, err := repo.Search(ctx, dyn.RepoSearchParam{Graph: graph, Fields: fields, Page: 0, Size: limit})
	if err != nil {
		return nil, errors.Wrap(err, "search "+schemaName)
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "search "+schemaName)
	}
	if !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}

func attrAndGraph(conditions ...dmodel.SearchNode) *dmodel.SearchGraph {
	graph := &dmodel.SearchGraph{}
	graph.And(conditions...)
	return graph
}

func attrCondition(field string, operator dmodel.Operator, values ...any) dmodel.SearchNode {
	return *dmodel.NewSearchNode().NewCondition(field, operator, values...)
}

// findAttributeLines returns lines applying the attribute. enabledOnly nil means every line,
// disabled history included.
func findAttributeLines(
	ctx corectx.Context, attributeId string, enabledOnly *bool, limit int,
) ([]dmodel.DynamicFields, error) {
	nodes := []dmodel.SearchNode{
		attrCondition(models.ProductTemplateAttributeFieldAttributeId, dmodel.Equals, attributeId),
	}
	if enabledOnly != nil {
		nodes = append(nodes, attrCondition(models.ProductTemplateAttributeFieldIsEnabled, dmodel.Equals, *enabledOnly))
	}
	return searchAttributeRows(ctx, models.ProductTemplateAttributeSchemaName, attrAndGraph(nodes...),
		[]string{basemodel.FieldId, models.ProductTemplateAttributeFieldProductTemplateId}, limit)
}

// findTemplatesSelectingValue returns the distinct templates whose enabled line has the value
// selected: the uses that archiving the value would strand.
func findTemplatesSelectingValue(ctx corectx.Context, valueId string) ([]string, error) {
	selected, err := searchAttributeRows(ctx, models.ProductTemplateAttributeValueSchemaName, attrAndGraph(
		attrCondition(models.ProductTemplateAttributeValueFieldAttributeValueId, dmodel.Equals, valueId),
		attrCondition(models.ProductTemplateAttributeValueFieldIsSelected, dmodel.Equals, true),
	), []string{basemodel.FieldId, models.ProductTemplateAttributeValueFieldTemplateAttributeId}, usageScanLimit)
	if err != nil || len(selected) == 0 {
		return nil, err
	}
	lineIds := attrDistinctStrings(selected, models.ProductTemplateAttributeValueFieldTemplateAttributeId)
	lines, err := searchAttributeRows(ctx, models.ProductTemplateAttributeSchemaName, attrAndGraph(
		attrCondition(basemodel.FieldId, dmodel.In, attrStringsToAny(lineIds)...),
		attrCondition(models.ProductTemplateAttributeFieldIsEnabled, dmodel.Equals, true),
	), []string{basemodel.FieldId, models.ProductTemplateAttributeFieldProductTemplateId}, usageScanLimit)
	if err != nil {
		return nil, err
	}
	return attrDistinctStrings(lines, models.ProductTemplateAttributeFieldProductTemplateId), nil
}

// valueHasTemplateHistory reports whether any template line, selected or not, ever offered the
// value. Such a value can only be archived: variants were built from that row.
func valueHasTemplateHistory(ctx corectx.Context, valueId string) (bool, error) {
	rows, err := searchAttributeRows(ctx, models.ProductTemplateAttributeValueSchemaName, attrAndGraph(
		attrCondition(models.ProductTemplateAttributeValueFieldAttributeValueId, dmodel.Equals, valueId),
	), []string{basemodel.FieldId}, 1)
	return len(rows) > 0, err
}

// loadAttributeValueNames reads every value of an attribute, archived ones included.
func loadAttributeValueNames(ctx corectx.Context, attributeId string) ([]attributeValueName, error) {
	rows, err := searchAttributeRows(ctx, models.ProductAttributeValueSchemaName, attrAndGraph(
		attrCondition(models.ProductAttributeValueFieldAttributeId, dmodel.Equals, attributeId),
	), []string{basemodel.FieldId, models.ProductAttributeValueFieldName}, maxAttributeValuesScanned)
	if err != nil {
		return nil, err
	}
	out := make([]attributeValueName, 0, len(rows))
	for _, row := range rows {
		value := attributeValueName{id: composable.ParamId(row)}
		if name := row.GetLangJson(models.ProductAttributeValueFieldName); name != nil {
			value.name = *name
		}
		out = append(out, value)
	}
	return out, nil
}

func attrDistinctStrings(rows []dmodel.DynamicFields, field string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, row := range rows {
		text := composable.ParamString(row, field)
		if text == "" || seen[text] {
			continue
		}
		seen[text] = true
		out = append(out, text)
	}
	return out
}

func attrStringsToAny(values []string) []any {
	out := make([]any, len(values))
	for i, value := range values {
		out[i] = value
	}
	return out
}

// templateUsageVars are the message variables of a usage refusal: the template count and the
// first few ids, which is as much as a caller holding only attribute permission should learn.
func templateUsageVars(templateIds []string) map[string]any {
	listed := templateIds
	if len(listed) > usageListedTemplates {
		listed = listed[:usageListedTemplates]
	}
	return map[string]any{"count": len(templateIds), "template_ids": strings.Join(listed, ", ")}
}

func newTemplateUsageViolation(field string, key string, message string, templateIds []string) ft.ClientErrorItem {
	return *ft.NewBusinessViolation(field, ft.ErrorKey(key, "inventory"), message, templateUsageVars(templateIds))
}
