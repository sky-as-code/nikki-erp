package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The stored display_name of a variant: the template name, then the variant's value names in the
// template's line order, in every enabled language. It is written when the variant is created and
// rewritten when the template is renamed or its configuration changes. A rename of an attribute
// value does not reach it (see docs/problems/inventory).

// variantNamer names variants of one template from its stored name, line order and value names.
type variantNamer struct {
	templateName model.LangJson
	rank         map[string]int
	valueNames   map[string]model.LangJson
	locales      []model.LanguageCode
}

// loadVariantNamer reads what naming the given keys needs: the template name, its lines (for the
// order) and the names of every value the keys hold, a fixed number of reads whatever the count.
func loadVariantNamer(
	ctx corectx.Context, templateName model.LangJson, templateId string, keys []string,
) (*variantNamer, error) {
	locales, err := dyn.ResolveEnabledLanguages(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "loadVariantNamer")
	}
	namer := &variantNamer{
		templateName: templateName, rank: map[string]int{}, valueNames: map[string]model.LangJson{}, locales: locales,
	}
	valueIds := keysValueIds(keys)
	if len(valueIds) == 0 {
		return namer, nil
	}
	states := map[string]*templateConfigState{templateId: {templateId: templateId, tavIndex: map[string]tavInfo{}}}
	if _, err := loadConfigLines(ctx, states, []string{templateId}); err != nil {
		return nil, err
	}
	namer.rank = storedLineRanks(states[templateId])
	rows, err := searchConfigRows(ctx, models.ProductAttributeValueSchemaName, basemodel.FieldId, valueIds)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		value := models.NewProductAttributeValueFrom(row)
		if name := value.GetName(); name != nil {
			namer.valueNames[derefId(value.GetId())] = *name
		}
	}
	return namer, nil
}

// name renders one key; fallback (the SKU) stands in for a template with no name.
func (this *variantNamer) name(key string, fallback string) model.LangJson {
	valueIds := orderedKeyValueIds(key, this.rank)
	values := make([]model.LangJson, 0, len(valueIds))
	for _, valueId := range valueIds {
		if name, ok := this.valueNames[valueId]; ok {
			values = append(values, name)
		}
	}
	return models.BuildVariantDisplayName(this.templateName, values, this.locales, fallback)
}

func keysValueIds(keys []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, key := range keys {
		for _, pair := range ParseCombinationKey(key) {
			if pair.ValueId != "" && !seen[pair.ValueId] {
				seen[pair.ValueId] = true
				out = append(out, pair.ValueId)
			}
		}
	}
	return out
}

// readTemplateNaming reads the template fields a new variant starts from: its name and its cost.
func readTemplateNaming(ctx corectx.Context, templateId string) (*models.ProductTemplate, error) {
	repo, err := repoFor(models.ProductTemplateSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := repo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.ProductTemplateFieldId: templateId},
		Fields: []string{models.ProductTemplateFieldId, models.ProductTemplateFieldName, models.ProductTemplateFieldCost},
	})
	if err != nil {
		return nil, errors.Wrap(err, "readTemplateNaming")
	}
	if !found.HasData {
		return nil, nil
	}
	return models.NewProductTemplateFrom(found.Data), nil
}

// RefreshTemplateVariantNames rewrites the display_name of every variant of a template whose name
// no longer matches its template name, line order or values, in the caller's transaction. It goes
// through the variant repository, as the archive cascade does: display_name carries no
// track_history, so the rewrite adds nothing to the history.
func RefreshTemplateVariantNames(ctx corectx.Context, templateId string) error {
	template, err := readTemplateNaming(ctx, templateId)
	if err != nil || template == nil {
		return err
	}
	variantRepo, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return err
	}
	rows, err := models.FindTemplateVariants(ctx, variantRepo, templateId, MaxCascadeVariants)
	if err != nil {
		return errors.Wrap(err, "RefreshTemplateVariantNames")
	}
	variants := make([]*models.ProductVariant, 0, len(rows))
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		variant := models.NewProductVariantFrom(row)
		variants = append(variants, variant)
		keys = append(keys, derefString(variant.GetCombinationKey()))
	}
	namer, err := loadVariantNamer(ctx, langJsonOrEmpty(template.GetName()), templateId, keys)
	if err != nil {
		return err
	}
	for _, variant := range variants {
		if err := rewriteVariantName(ctx, variant, namer); err != nil {
			return err
		}
	}
	return nil
}

func rewriteVariantName(ctx corectx.Context, variant *models.ProductVariant, namer *variantNamer) error {
	name := namer.name(derefString(variant.GetCombinationKey()), derefString(variant.GetSku()))
	if stored := variant.GetDisplayName(); stored != nil && stored.IsEqual(name) {
		return nil
	}
	update := dmodel.DynamicFields{models.ProductVariantFieldId: derefString(variant.GetId())}
	update.SetLangJson(models.ProductVariantFieldDisplayName, &name)
	return writeVariantFields(ctx, update, "RefreshTemplateVariantNames")
}

func langJsonOrEmpty(value *model.LangJson) model.LangJson {
	if value == nil {
		return model.LangJson{}
	}
	return *value
}

// fillVariantFromTemplate names a new variant and, when no cost was sent, gives it the template's
// cost. Every creation path runs it -- reconcile, on-demand resolve, the variant API, import -- so no
// variant is stored nameless. A display_name in the payload is replaced: it is server-maintained.
func fillVariantFromTemplate() dynamicresource.BeforeValidationFn {
	return func(
		ctx corectx.Context, entity *dynamicresource.DynamicEntity, _ *ft.ClientErrors,
	) (*dynamicresource.DynamicEntity, error) {
		data := entity.GetFieldData()
		delete(data, models.ProductVariantFieldDisplayName)
		templateId := readStringParam(data, models.ProductVariantFieldProductTemplateId)
		if templateId == "" {
			return entity, nil
		}
		template, err := readTemplateNaming(ctx, templateId)
		if err != nil || template == nil {
			return entity, err
		}
		if data[models.ProductVariantFieldCost] == nil && template.GetCost() != nil {
			data.SetDecimal(models.ProductVariantFieldCost, template.GetCost())
		}
		key := readStringParam(data, models.ProductVariantFieldCombinationKey)
		namer, err := loadVariantNamer(ctx, langJsonOrEmpty(template.GetName()), templateId, []string{key})
		if err != nil {
			return nil, err
		}
		name := namer.name(key, readStringParam(data, models.ProductVariantFieldSku))
		data.SetLangJson(models.ProductVariantFieldDisplayName, &name)
		return entity, nil
	}
}

// dropServerManagedVariantFields discards a display_name sent on update. It is dropped rather than
// refused because a form saving a whole record sends back what it read.
func dropServerManagedVariantFields() dynamicresource.BeforeValidationFn {
	return func(
		_ corectx.Context, entity *dynamicresource.DynamicEntity, _ *ft.ClientErrors,
	) (*dynamicresource.DynamicEntity, error) {
		delete(entity.GetFieldData(), models.ProductVariantFieldDisplayName)
		return entity, nil
	}
}
