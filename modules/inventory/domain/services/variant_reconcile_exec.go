package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The reconcile executor applies a VariantPlan inside the template's transaction. Order matters:
// invalid completes go before invalid bases (base_variant_id is RESTRICT), then bases are
// restored, extended and created before completes, so a complete can point at a base the same
// plan made. Variant deletes go to the repository directly: the variant domain Delete guards a
// user's delete (last complete, base with children), which reconcile has already decided.

// templateVariants are every variant of a template, archived ones included, with the attribute
// roles each was created under.
type templateVariants struct {
	byId     map[string]*models.ProductVariant
	existing []PlanVariant
	children map[string][]string
}

// loadTemplateVariants reads the template's variants and their join rows. Roles come from the
// snapshot of the line each join row's template value belongs to.
func loadTemplateVariants(ctx corectx.Context, state *templateConfigState) (*templateVariants, error) {
	rows, err := searchConfigRows(ctx, models.ProductVariantSchemaName,
		models.ProductVariantFieldProductTemplateId, []string{state.templateId})
	if err != nil {
		return nil, err
	}
	loaded := &templateVariants{byId: map[string]*models.ProductVariant{}, children: map[string][]string{}}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		variant := models.NewProductVariantFrom(row)
		id := derefString(variant.GetId())
		loaded.byId[id] = variant
		ids = append(ids, id)
		if baseId := derefString(variant.GetBaseVariantId()); baseId != "" {
			loaded.children[baseId] = append(loaded.children[baseId], id)
		}
	}
	roles, err := loadVariantRoles(ctx, state, ids)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		loaded.existing = append(loaded.existing, toPlanVariant(loaded.byId[id], roles[id]))
	}
	return loaded, nil
}

func toPlanVariant(variant *models.ProductVariant, roles map[string]models.VariantCreationMode) PlanVariant {
	archived := variant.IsArchived()
	return PlanVariant{
		Id:            derefString(variant.GetId()),
		Kind:          variantKindOf(variant),
		Key:           derefString(variant.GetCombinationKey()),
		IsArchived:    archived != nil && *archived,
		ArchiveSource: EffectiveArchiveSource(variant),
		BaseVariantId: derefString(variant.GetBaseVariantId()),
		Roles:         roles,
	}
}

func loadVariantRoles(
	ctx corectx.Context, state *templateConfigState, variantIds []string,
) (map[string]map[string]models.VariantCreationMode, error) {
	roles := map[string]map[string]models.VariantCreationMode{}
	if len(variantIds) == 0 {
		return roles, nil
	}
	rows, err := searchConfigRows(ctx, models.ProductVariantAttributeValueSchemaName,
		models.ProductVariantAttributeValueFieldProductVariantId, variantIds)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		join := models.NewProductVariantAttributeValueFrom(row)
		info, found := state.tavIndex[derefString(join.GetTemplateAttributeValueId())]
		if !found || info.line.snapshot == "" {
			continue
		}
		variantId := derefString(join.GetProductVariantId())
		if roles[variantId] == nil {
			roles[variantId] = map[string]models.VariantCreationMode{}
		}
		roles[variantId][info.line.attributeId] = info.line.snapshot
	}
	return roles, nil
}

// planExecutor writes one plan. refs must be loaded after the configuration rows were written.
type planExecutor struct {
	templateId     string
	orgId          string
	templateActive bool
	refs           TemplateValueRefs
	variants       *templateVariants
	decisions      map[string]invalidDecision
	baseIds        map[string]string
	result         ReconcileResult
}

// ReconcileResult reports what an apply did, by variant id.
type ReconcileResult struct {
	Kept     []string `json:"kept_variant_ids"`
	Extended []string `json:"extended_variant_ids"`
	Restored []string `json:"restored_variant_ids"`
	Created  []string `json:"created_variant_ids"`
	Archived []string `json:"archived_variant_ids"`
	Deleted  []string `json:"deleted_variant_ids"`
}

func (this *planExecutor) run(ctx corectx.Context, plan *VariantPlan) (ft.ClientErrors, error) {
	if cErrs, err := this.removeInvalid(ctx, plan.Invalid); err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	this.collectBaseIds(plan)
	for _, kind := range planKindOrder {
		if cErrs, err := this.applyKind(ctx, plan, kind); err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
	}
	return nil, nil
}

// removeInvalid deletes or archives the invalid records, completes first.
func (this *planExecutor) removeInvalid(ctx corectx.Context, invalid []PlannedVariant) (ft.ClientErrors, error) {
	for _, kind := range []models.VariantKind{models.VariantKindComplete, models.VariantKindBase} {
		for _, item := range invalid {
			if item.Kind != kind {
				continue
			}
			if cErrs, err := this.removeOne(ctx, item); err != nil || cErrs.Count() > 0 {
				return cErrs, err
			}
		}
	}
	return nil, nil
}

func (this *planExecutor) removeOne(ctx corectx.Context, item PlannedVariant) (ft.ClientErrors, error) {
	variant := this.variants.byId[item.VariantId]
	if variant == nil {
		return nil, nil
	}
	if !this.decisions[item.VariantId].delete {
		changed, err := ArchiveVariantForConfiguration(ctx, variant)
		if changed {
			this.result.Archived = append(this.result.Archived, item.VariantId)
		}
		return nil, err
	}
	variantRepo, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}
	deleted, err := variantRepo.DeleteOne(ctx, dmodel.DynamicFields{models.ProductVariantFieldId: item.VariantId})
	if err != nil {
		return nil, errors.Wrap(err, "delete invalid variant")
	}
	if deleted.ClientErrors.Count() > 0 {
		return deleted.ClientErrors, nil
	}
	this.result.Deleted = append(this.result.Deleted, item.VariantId)
	return nil, nil
}

// collectBaseIds indexes the bases that will exist after the plan by their (new) key; created
// bases are added as they are written.
func (this *planExecutor) collectBaseIds(plan *VariantPlan) {
	this.baseIds = map[string]string{}
	for _, list := range [][]PlannedVariant{plan.Keep, plan.Restore, plan.Extend} {
		for _, item := range list {
			if item.Kind != models.VariantKindBase {
				continue
			}
			key := item.Key
			if item.NewKey != "" {
				key = item.NewKey
			}
			this.baseIds[key] = item.VariantId
		}
	}
}

func (this *planExecutor) applyKind(ctx corectx.Context, plan *VariantPlan, kind models.VariantKind) (ft.ClientErrors, error) {
	steps := []struct {
		items []PlannedVariant
		apply func(corectx.Context, PlannedVariant) (ft.ClientErrors, error)
	}{
		{plan.Restore, this.restore}, {plan.Extend, this.extend}, {plan.Create, this.create}, {plan.Keep, this.keep},
	}
	for _, step := range steps {
		for _, item := range step.items {
			if item.Kind != kind {
				continue
			}
			if cErrs, err := step.apply(ctx, item); err != nil || cErrs.Count() > 0 {
				return cErrs, err
			}
		}
	}
	return nil, nil
}

func (this *planExecutor) restore(ctx corectx.Context, item PlannedVariant) (ft.ClientErrors, error) {
	variant := this.variants.byId[item.VariantId]
	if variant == nil || !this.templateActive {
		return nil, nil
	}
	restored, err := RestoreSyncArchivedVariant(ctx, variant)
	if err != nil {
		return nil, err
	}
	if restored {
		this.result.Restored = append(this.result.Restored, item.VariantId)
	}
	return nil, this.relinkBase(ctx, item)
}

func (this *planExecutor) extend(ctx corectx.Context, item PlannedVariant) (ft.ClientErrors, error) {
	cErrs, err := ExtendVariantKey(ctx, item.VariantId, this.orgId, item.Key, item.NewKey, this.refs)
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	this.result.Extended = append(this.result.Extended, item.VariantId)
	return nil, this.relinkBase(ctx, item)
}

func (this *planExecutor) create(ctx corectx.Context, item PlannedVariant) (ft.ClientErrors, error) {
	variantId, cErrs, err := MaterializeVariant(ctx, VariantMaterializeSpec{
		TemplateId: this.templateId, OrgId: this.orgId, Kind: item.Kind, Key: item.Key,
		BaseVariantId: this.baseIds[item.BaseKey], ValueRefs: this.refs,
	})
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	if item.Kind == models.VariantKindBase {
		this.baseIds[item.Key] = variantId
	}
	this.result.Created = append(this.result.Created, variantId)
	return nil, nil
}

func (this *planExecutor) keep(ctx corectx.Context, item PlannedVariant) (ft.ClientErrors, error) {
	this.result.Kept = append(this.result.Kept, item.VariantId)
	return nil, this.relinkBase(ctx, item)
}

// relinkBase points a complete of a mixed template at the base holding its instant sub-key when
// it points elsewhere (or nowhere). base_variant_id is no_update for clients; the system keeps it.
func (this *planExecutor) relinkBase(ctx corectx.Context, item PlannedVariant) error {
	if item.Kind != models.VariantKindComplete || item.BaseKey == "" {
		return nil
	}
	baseId := this.baseIds[item.BaseKey]
	variant := this.variants.byId[item.VariantId]
	if baseId == "" || variant == nil || derefString(variant.GetBaseVariantId()) == baseId {
		return nil
	}
	return writeVariantFields(ctx, dmodel.DynamicFields{
		models.ProductVariantFieldId: item.VariantId, models.ProductVariantFieldBaseVariantId: baseId,
	}, "relinkBase")
}
