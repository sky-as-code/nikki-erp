package services

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

// planReconciler decides the outcome of each existing variant against the next configuration.
// taken holds every key a record of that kind already holds, archived included: such a key is
// never created again. claimed holds target keys already satisfied by keep, restore or extend.
type planReconciler struct {
	input     PlanInput
	next      *planState
	prev      *planState
	targets   planTargets
	extension map[string]string
	taken     map[models.VariantKind]map[string]bool
	claimed   map[models.VariantKind]map[string]bool
}

var planKindOrder = []models.VariantKind{models.VariantKindBase, models.VariantKindComplete}

func newPlanReconciler(input PlanInput, next *planState, targets planTargets) *planReconciler {
	prevConfig := PlanConfig{}
	if input.Previous != nil {
		prevConfig = *input.Previous
	}
	prev := newPlanState(prevConfig)
	reconciler := &planReconciler{
		input: input, next: next, prev: prev, targets: targets,
		extension: detectExtension(prev, next),
		taken:     map[models.VariantKind]map[string]bool{},
		claimed:   map[models.VariantKind]map[string]bool{},
	}
	for _, kind := range planKindOrder {
		reconciler.taken[kind] = map[string]bool{}
		reconciler.claimed[kind] = map[string]bool{}
	}
	for _, variant := range input.Existing {
		reconciler.taken[planKindOf(variant)][variant.Key] = true
	}
	return reconciler
}

func planKindOf(variant PlanVariant) models.VariantKind {
	if variant.Kind == models.VariantKindBase {
		return models.VariantKindBase
	}
	return models.VariantKindComplete
}

// detectExtension returns the pairs to append to existing variants (CR01 §6.2, CR02 §7.2): every
// previous generating attribute stays with the same mode, and every new generating attribute is
// instant with exactly one selected value. Adding a dynamic attribute never extends.
func detectExtension(prev *planState, next *planState) map[string]string {
	for _, dim := range prev.generating() {
		nextDim, found := next.lines[dim.attributeId]
		if !found || nextDim.mode != dim.mode {
			return nil
		}
	}
	extension := map[string]string{}
	for _, dim := range next.generating() {
		if prevDim, found := prev.lines[dim.attributeId]; found && prevDim.mode.CreatesVariants() {
			continue
		}
		if dim.mode != models.VariantCreationModeInstant || len(dim.valueIds) != 1 {
			return nil
		}
		extension[dim.attributeId] = dim.valueIds[0]
	}
	if len(extension) == 0 {
		return nil
	}
	return extension
}

func (this *planReconciler) run(plan *VariantPlan) {
	for _, kind := range planKindOrder {
		for _, variant := range this.input.Existing {
			if planKindOf(variant) == kind {
				plan.add(this.classify(plan, variant))
			}
		}
	}
	for _, key := range this.targets.keys {
		if this.taken[this.targets.kind][key] || this.claimed[this.targets.kind][key] {
			continue
		}
		plan.add(PlannedVariant{Kind: this.targets.kind, Key: key, Outcome: PlanOutcomeCreate})
	}
}

func (this *planReconciler) classify(plan *VariantPlan, variant PlanVariant) PlannedVariant {
	kind := planKindOf(variant)
	item := PlannedVariant{
		VariantId: variant.Id, Kind: kind, Key: variant.Key,
		IsArchived: variant.IsArchived, ArchiveSource: variant.ArchiveSource,
	}
	valid := this.next.isValid(kind, variant.Key)
	switch {
	case !variant.IsArchived && valid:
		item.Outcome = PlanOutcomeKeep
		this.claimed[kind][variant.Key] = true
	case !variant.IsArchived:
		item.Outcome = this.tryExtend(plan, variant, &item)
	case valid:
		item.Outcome = this.classifyArchivedValid(plan, variant)
	case variant.ArchiveSource == models.ArchiveSourceTemplateCascade:
		// Configuration now invalidates it, so the executor restamps it system_sync (plan D3).
		item.Outcome = PlanOutcomeInvalid
	default:
		item.Outcome = PlanOutcomeUntouched
	}
	this.fillBaseKey(&item)
	return item
}

// classifyArchivedValid restores only a system_sync archive of an active template. A user archive
// is never auto-restored and a template_cascade archive is left to the template unarchive.
func (this *planReconciler) classifyArchivedValid(plan *VariantPlan, variant PlanVariant) PlanOutcome {
	if variant.ArchiveSource != models.ArchiveSourceSystemSync || this.input.TemplateArchived {
		return PlanOutcomeUntouched
	}
	if this.rolesDiffer(variant) {
		plan.violate(ReasonPlanRoleConflict,
			"variant {{variant_id}} was created with different attribute roles; reconcile it manually",
			map[string]any{"variant_id": variant.Id, "combination_key": variant.Key})
		return PlanOutcomeConflict
	}
	this.claimed[planKindOf(variant)][variant.Key] = true
	return PlanOutcomeRestore
}

func (this *planReconciler) rolesDiffer(variant PlanVariant) bool {
	for attributeId, mode := range variant.Roles {
		line, found := this.next.lines[attributeId]
		if !found || line.mode != mode {
			return true
		}
	}
	return false
}

// tryExtend appends the extension pairs to an active variant that matched the previous
// configuration, when the extended key is valid and no record of the same kind holds it.
func (this *planReconciler) tryExtend(plan *VariantPlan, variant PlanVariant, item *PlannedVariant) PlanOutcome {
	kind := planKindOf(variant)
	if this.extension == nil || !this.prev.isValid(kind, variant.Key) {
		return PlanOutcomeInvalid
	}
	pairs := ParseKeyToPairs(variant.Key)
	for attributeId, valueId := range this.extension {
		pairs[attributeId] = valueId
	}
	newKey := BuildKeyFromPairs(pairs)
	if !this.next.isValid(kind, newKey) {
		return PlanOutcomeInvalid
	}
	item.NewKey = newKey
	if this.taken[kind][newKey] {
		plan.violate(ReasonPlanCombinationConflict,
			"variant {{variant_id}} cannot take combination {{combination_key}}: another variant holds it",
			map[string]any{"variant_id": variant.Id, "combination_key": newKey})
		return PlanOutcomeConflict
	}
	this.claimed[kind][newKey] = true
	return PlanOutcomeExtend
}

func (this *planReconciler) fillBaseKey(item *PlannedVariant) {
	if item.Kind != models.VariantKindComplete || !this.next.isMixed() {
		return
	}
	switch item.Outcome {
	case PlanOutcomeKeep, PlanOutcomeRestore:
		item.BaseKey = this.next.instantSubKey(item.Key)
	case PlanOutcomeExtend:
		item.BaseKey = this.next.instantSubKey(item.NewKey)
	}
}
