package services

import (
	"encoding/json"
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ReasonCopyCatalogueArchived refuses a copy whose configuration names archived catalogue rows.
const ReasonCopyCatalogueArchived = "product_template.copy_catalogue_archived"

// copyNameSuffix marks the copy's default name in every language the source carries.
const copyNameSuffix = " (copy)"

// templateCopyFields are the template columns a copy carries. Everything else is the source's own
// identity or history: id, etag, audit, archive state, import identity (external_id,
// source_system) and the computed read model. The image file key is not copied: a file belongs to
// one record, and deleting or replacing it on either template would remove it from the other.
var templateCopyFields = []string{
	models.ProductTemplateFieldName,
	models.ProductTemplateFieldShortName,
	models.ProductTemplateFieldProductType,
	models.ProductTemplateFieldIsCombo,
	models.ProductTemplateFieldCategoryId,
	models.ProductTemplateFieldBrandId,
	models.ProductTemplateFieldUomId,
	models.ProductTemplateFieldSaleOk,
	models.ProductTemplateFieldPurchaseOk,
	models.ProductTemplateFieldDescription,
	models.ProductTemplateFieldSalesDescription,
	models.ProductTemplateFieldPurchaseDescription,
	models.ProductTemplateFieldBaseSalesPrice,
	models.ProductTemplateFieldDefaultWeight,
	models.ProductTemplateFieldDefaultLength,
	models.ProductTemplateFieldDefaultWidth,
	models.ProductTemplateFieldDefaultHeight,
	models.ProductTemplateFieldStatus,
}

var _ itProduct.TemplateCopier = (*ProductTemplateDomainServiceImpl)(nil)

// BuildTemplateCopy builds the create payload of a template's copy (CR01 §11): its own fields and
// its current configuration only - enabled lines with their selected values, and active exclusions
// between them - in catalogue ids, so the normal create path plans the variants. Variants, stock,
// SKUs and archive reasons are not copied. An archived source is allowed; an archived catalogue
// attribute or value refuses the copy and is listed, never restored on the user's behalf.
func (this *ProductTemplateDomainServiceImpl) BuildTemplateCopy(
	ctx corectx.Context, query itProduct.CopyTemplateQuery,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	template, err := this.fetchTemplate(ctx, query.TemplateId)
	if err != nil || template == nil {
		return &dyn.OpResult[dmodel.DynamicFields]{}, err
	}
	state, err := loadTemplateConfigState(ctx, query.TemplateId)
	if err != nil {
		return nil, err
	}
	lines, exclusions := copyConfiguration(state)

	refused, err := assertCatalogueActive(ctx, lines)
	if err != nil || refused != nil {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: refused}, err
	}

	payload, err := copyTemplateFields(template.GetFieldData(), query.Name)
	if err != nil {
		return nil, err
	}
	payload[itProduct.ProductTemplateFieldAttributeLines] = lines
	payload[itProduct.ProductTemplateFieldExclusions] = exclusions
	return &dyn.OpResult[dmodel.DynamicFields]{Data: payload, HasData: true}, nil
}

// copyConfiguration reads the enabled lines with their selected values, and the active exclusions
// whose two values are both still selected.
func copyConfiguration(
	state *templateConfigState,
) ([]itProduct.AttributeLineInput, []itProduct.ExclusionInput) {
	lines := []itProduct.AttributeLineInput{}
	for _, line := range state.enabledLines() {
		valueIds := line.selectedValueIds()
		if len(valueIds) == 0 {
			continue
		}
		lines = append(lines, itProduct.AttributeLineInput{
			AttributeId: line.attributeId, ValueIds: valueIds, Sequence: line.sequence,
		})
	}
	refs := state.valueRefs()
	exclusions := []itProduct.ExclusionInput{}
	for _, exclusion := range state.planConfig().Exclusions {
		_, leftSelected := refs[exclusion.LeftValueId]
		_, rightSelected := refs[exclusion.RightValueId]
		if leftSelected && rightSelected {
			exclusions = append(exclusions, itProduct.ExclusionInput{
				LeftValueId: exclusion.LeftValueId, RightValueId: exclusion.RightValueId,
			})
		}
	}
	return lines, exclusions
}

// copyTemplateFields copies the carried columns, round-tripped through JSON so the payload has the
// shape a REST client sends, and names the copy. An explicit name replaces the default suffix.
func copyTemplateFields(source dmodel.DynamicFields, name any) (dmodel.DynamicFields, error) {
	carried := map[string]any{}
	for _, field := range templateCopyFields {
		if value, found := source[field]; found && value != nil {
			carried[field] = value
		}
	}
	encoded, err := json.Marshal(carried)
	if err != nil {
		return nil, errors.Wrap(err, "copyTemplateFields")
	}
	payload := dmodel.DynamicFields{}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, errors.Wrap(err, "copyTemplateFields")
	}
	if name != nil {
		payload[models.ProductTemplateFieldName] = name
		return payload, nil
	}
	if names, ok := payload[models.ProductTemplateFieldName].(map[string]any); ok {
		for language, text := range names {
			if value, isText := text.(string); isText && value != "" {
				names[language] = value + copyNameSuffix
			}
		}
	}
	return payload, nil
}

// assertCatalogueActive lists the archived catalogue attributes and values the copy would name.
func assertCatalogueActive(
	ctx corectx.Context, lines []itProduct.AttributeLineInput,
) (ft.ClientErrors, error) {
	attributeIds, valueIds := []string{}, []string{}
	for _, line := range lines {
		attributeIds = append(attributeIds, line.AttributeId)
		valueIds = append(valueIds, line.ValueIds...)
	}
	archivedAttributes, err := findArchivedIds(ctx, models.ProductAttributeSchemaName, attributeIds)
	if err != nil {
		return nil, err
	}
	archivedValues, err := findArchivedIds(ctx, models.ProductAttributeValueSchemaName, valueIds)
	if err != nil {
		return nil, err
	}
	if len(archivedAttributes) == 0 && len(archivedValues) == 0 {
		return nil, nil
	}
	return ft.ClientErrors{*ft.NewBusinessViolation(models.ProductTemplateFieldId,
		ft.ErrorKey(ReasonCopyCatalogueArchived, "inventory"),
		"the product uses archived attributes or values; unarchive them or remove them from the product first",
		map[string]any{"attribute_ids": archivedAttributes, "value_ids": archivedValues})}, nil
}

// findArchivedIds returns, sorted, which of the ids name archived rows of a schema.
func findArchivedIds(ctx corectx.Context, schemaName string, ids []string) ([]string, error) {
	archived := []string{}
	if len(ids) == 0 {
		return archived, nil
	}
	repo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldId, dmodel.In, attrStringsToAny(ids)...),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, true),
	)
	found, err := repo.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: len(ids)})
	if err != nil {
		return nil, errors.Wrap(err, "findArchivedIds")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "findArchivedIds")
	}
	for _, row := range found.Data.Items {
		if id := derefId(basemodel.NewDynamicModel(row).GetId()); id != "" {
			archived = append(archived, id)
		}
	}
	sort.Strings(archived)
	return archived, nil
}
