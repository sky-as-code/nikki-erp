package services

import (
	"encoding/json"
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ReasonDuplicateCatalogueArchived refuses a duplicate whose configuration names archived catalogue
// rows.
const ReasonDuplicateCatalogueArchived = "product_template.duplicate_catalogue_archived"

// duplicateNameSuffix marks the duplicate's default name in every language the source carries.
const duplicateNameSuffix = " (copy)"

// templateDuplicateFields are the template columns a duplicate carries. Everything else is the
// source's own identity or history: id, etag, audit, archive state, import identity (external_id,
// source_system) and the computed read model. The preview image is not carried as is: a file
// belongs to one record, so the duplicate gets its own clone (duplicatePreviewImage).
var templateDuplicateFields = []string{
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

var _ itProduct.TemplateDuplicator = (*ProductTemplateDomainServiceImpl)(nil)

// BuildTemplateDuplicate builds the create payload of a template's duplicate (CR01 §11): its own
// fields, a clone of its preview image, and its current configuration only - enabled lines with
// their selected values, and active exclusions between them - in catalogue ids, so the normal
// create path plans the variants. Variants, stock, SKUs and archive reasons are not carried. An
// archived source is allowed; an archived catalogue attribute or value refuses the duplicate and is
// listed, never restored on the user's behalf.
func (this *ProductTemplateDomainServiceImpl) BuildTemplateDuplicate(
	ctx corectx.Context, query itProduct.DuplicateTemplateQuery,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	template, err := this.fetchTemplate(ctx, query.TemplateId)
	if err != nil || template == nil {
		return &dyn.OpResult[dmodel.DynamicFields]{}, err
	}
	state, err := loadTemplateConfigState(ctx, query.TemplateId)
	if err != nil {
		return nil, err
	}
	lines, exclusions := duplicateConfiguration(state)

	refused, err := assertCatalogueActive(ctx, lines)
	if err != nil || refused != nil {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: refused}, err
	}

	payload, err := duplicateTemplateFields(template.GetFieldData(), query.Name)
	if err != nil {
		return nil, err
	}
	if err := this.duplicatePreviewImage(ctx, template, payload); err != nil {
		return nil, err
	}
	payload[itProduct.ProductTemplateFieldAttributeLines] = lines
	payload[itProduct.ProductTemplateFieldExclusions] = exclusions
	return &dyn.OpResult[dmodel.DynamicFields]{Data: payload, HasData: true}, nil
}

// duplicatePreviewImage clones the source's image into a new pending file of the same org, which
// the create of the duplicate then marks in use. It runs last, after every refusal, so a refused
// duplicate clones nothing; a clone whose create fails stays pending and its object expires.
func (this *ProductTemplateDomainServiceImpl) duplicatePreviewImage(
	ctx corectx.Context, template *models.ProductTemplate, payload dmodel.DynamicFields,
) error {
	fileId := derefString(template.GetPreviewImageFileId())
	if fileId == "" {
		return nil
	}
	if this.files == nil {
		return errors.New("duplicatePreviewImage: no storage consumer to clone the image with")
	}
	cloneId, err := this.files.DuplicateFile(ctx, storage.DuplicateFileCommand{
		FileId: model.Id(fileId), OrgId: template.GetOrgId(),
	})
	if err != nil {
		return errors.Wrap(err, "duplicatePreviewImage")
	}
	if cloneId != nil {
		payload[models.ProductTemplateFieldPreviewImageFileId] = string(*cloneId)
	}
	return nil
}

// duplicateConfiguration reads the enabled lines with their selected values, and the active exclusions
// whose two values are both still selected.
func duplicateConfiguration(
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

// duplicateTemplateFields copies the carried columns, round-tripped through JSON so the payload has the
// shape a REST client sends, and names the duplicate. An explicit name replaces the default suffix.
func duplicateTemplateFields(source dmodel.DynamicFields, name any) (dmodel.DynamicFields, error) {
	carried := map[string]any{}
	for _, field := range templateDuplicateFields {
		if value, found := source[field]; found && value != nil {
			carried[field] = value
		}
	}
	encoded, err := json.Marshal(carried)
	if err != nil {
		return nil, errors.Wrap(err, "duplicateTemplateFields")
	}
	payload := dmodel.DynamicFields{}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, errors.Wrap(err, "duplicateTemplateFields")
	}
	if name != nil {
		payload[models.ProductTemplateFieldName] = name
		return payload, nil
	}
	if names, ok := payload[models.ProductTemplateFieldName].(map[string]any); ok {
		for language, text := range names {
			if value, isText := text.(string); isText && value != "" {
				names[language] = value + duplicateNameSuffix
			}
		}
	}
	return payload, nil
}

// assertCatalogueActive lists the archived catalogue attributes and values the duplicate would name.
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
		ft.ErrorKey(ReasonDuplicateCatalogueArchived, "inventory"),
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
