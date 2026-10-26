package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// TemplateLineConfig is one enabled template line as the variant rules see it: the attribute, the
// mode the line was applied under, and the catalogue values still selected on it.
type TemplateLineConfig struct {
	AttributeId      string
	Mode             models.VariantCreationMode
	SelectedValueIds map[string]bool
}

// TemplateConfiguration is a template's current enabled configuration plus its active exclusion
// pairs. Every value id is a CATALOGUE value id, the id combination keys carry; the exclusion rows
// store template value ids and are translated on load.
type TemplateConfiguration struct {
	Lines      []TemplateLineConfig
	Exclusions [][2]string
}

// HasOnDemandLine reports whether any enabled line is dynamic or never. Such a template may
// legitimately have no active variant, because the selectable products are resolved on demand.
func (this TemplateConfiguration) HasOnDemandLine() bool {
	for _, line := range this.Lines {
		if line.Mode == models.VariantCreationModeDynamic || line.Mode == models.VariantCreationModeNever {
			return true
		}
	}
	return false
}

// IsCombinationValid reports whether a variant's combination is still one the configuration
// produces: every pair names an enabled generating line and a value selected on it, the set of
// attributes is exactly the generating set for its kind, and no active exclusion pair is inside
// it. A complete is keyed by the instant and dynamic lines; a base by the instant lines only, and
// exists only while the template mixes instant with dynamic.
func (this TemplateConfiguration) IsCombinationValid(combinationKey string, kind models.VariantKind) bool {
	required := map[string]TemplateLineConfig{}
	hasDynamic := false
	for _, line := range this.Lines {
		switch {
		case line.Mode == models.VariantCreationModeInstant:
			required[line.AttributeId] = line
		case line.Mode == models.VariantCreationModeDynamic:
			hasDynamic = true
			if kind == models.VariantKindComplete {
				required[line.AttributeId] = line
			}
		}
	}
	if kind == models.VariantKindBase && (!hasDynamic || len(required) == 0) {
		return false
	}

	pairs := ParseCombinationKey(combinationKey)
	usedAttributes, usedValues := map[string]bool{}, map[string]bool{}
	for _, pair := range pairs {
		line, ok := required[pair.AttributeId]
		if !ok || usedAttributes[pair.AttributeId] || !line.SelectedValueIds[pair.ValueId] {
			return false
		}
		usedAttributes[pair.AttributeId] = true
		usedValues[pair.ValueId] = true
	}
	if len(pairs) != len(required) {
		return false
	}
	return !this.excludesAny(usedValues)
}

func (this TemplateConfiguration) excludesAny(valueIds map[string]bool) bool {
	for _, exclusion := range this.Exclusions {
		if valueIds[exclusion[0]] && valueIds[exclusion[1]] {
			return true
		}
	}
	return false
}

// LoadTemplateConfiguration reads a template's enabled lines, their selected values and its
// active exclusions. A line's mode is its snapshot when one was taken, else the attribute's
// current mode.
func LoadTemplateConfiguration(ctx corectx.Context, templateId string) (TemplateConfiguration, error) {
	config := TemplateConfiguration{}
	lineRepo, err := repoFor(models.ProductTemplateAttributeSchemaName)
	if err != nil {
		return config, err
	}
	rows, err := models.FindTemplateAttributes(ctx, lineRepo, templateId, maxTemplateVariants)
	if err != nil {
		return config, errors.Wrap(err, "LoadTemplateConfiguration")
	}
	catalogueOf := map[string]string{}
	for _, row := range rows {
		line := models.NewProductTemplateAttributeFrom(row)
		if enabled := line.GetIsEnabled(); enabled != nil && !*enabled {
			continue
		}
		loaded, err := loadLineConfig(ctx, line, catalogueOf)
		if err != nil {
			return config, err
		}
		config.Lines = append(config.Lines, loaded)
	}
	templatePairs, err := loadActiveExclusions(ctx, templateId)
	config.Exclusions = toCataloguePairs(templatePairs, catalogueOf)
	return config, err
}

// toCataloguePairs translates exclusion pairs from template value ids to catalogue value ids. A
// pair naming a value of a disabled line is dropped: it cannot be inside any valid combination.
func toCataloguePairs(templatePairs [][2]string, catalogueOf map[string]string) [][2]string {
	pairs := make([][2]string, 0, len(templatePairs))
	for _, pair := range templatePairs {
		left, right := catalogueOf[pair[0]], catalogueOf[pair[1]]
		if left != "" && right != "" {
			pairs = append(pairs, [2]string{left, right})
		}
	}
	return pairs
}

// loadLineConfig reads one line; catalogueOf collects template value id -> catalogue value id.
func loadLineConfig(
	ctx corectx.Context, line *models.ProductTemplateAttribute, catalogueOf map[string]string,
) (TemplateLineConfig, error) {
	loaded := TemplateLineConfig{
		AttributeId:      derefId(line.GetAttributeId()),
		SelectedValueIds: map[string]bool{},
	}
	if snapshot := line.GetVariantCreationSnapshot(); snapshot != nil && *snapshot != "" {
		loaded.Mode = *snapshot
	} else {
		attribute, err := findRecord(ctx, models.ProductAttributeSchemaName, models.ProductAttributeFieldId, loaded.AttributeId)
		if err != nil {
			return loaded, err
		}
		if attribute != nil {
			if mode := models.NewProductAttributeFrom(attribute).GetVariantCreationMode(); mode != nil {
				loaded.Mode = *mode
			}
		}
	}

	valueRepo, err := repoFor(models.ProductTemplateAttributeValueSchemaName)
	if err != nil {
		return loaded, err
	}
	values, err := models.FindTemplateAttributeValues(ctx, valueRepo, derefId(line.GetId()), maxTemplateVariants)
	if err != nil {
		return loaded, errors.Wrap(err, "loadLineConfig")
	}
	for _, row := range values {
		value := models.NewProductTemplateAttributeValueFrom(row)
		catalogueId := derefId(value.GetAttributeValueId())
		catalogueOf[derefId(value.GetId())] = catalogueId
		if selected := value.GetIsSelected(); selected == nil || *selected {
			loaded.SelectedValueIds[catalogueId] = true
		}
	}
	return loaded, nil
}

func loadActiveExclusions(ctx corectx.Context, templateId string) ([][2]string, error) {
	repo, err := repoFor(models.ProductTemplateCombinationExclusionSchemaName)
	if err != nil {
		return nil, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.ProductTemplateCombinationExclusionFieldProductTemplateId, dmodel.Equals, templateId),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
	)
	found, err := repo.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: maxTemplateVariants})
	if err != nil {
		return nil, errors.Wrap(err, "loadActiveExclusions")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "loadActiveExclusions")
	}
	pairs := make([][2]string, 0, len(found.Data.Items))
	for _, row := range found.Data.Items {
		exclusion := models.NewProductTemplateCombinationExclusionFrom(row)
		pairs = append(pairs, [2]string{derefId(exclusion.GetLeftValueId()), derefId(exclusion.GetRightValueId())})
	}
	return pairs, nil
}
