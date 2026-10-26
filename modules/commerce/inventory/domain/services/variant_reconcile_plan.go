package services

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

// planTargets are the records a configuration pre-creates: completes for every valid instant
// combination when there is no dynamic dimension, bases for every instant combination with a valid
// dynamic completion in a mixed template, nothing for a dynamic-only template. keys stays nil when
// count passes the pre-create limit, so a large product is never materialized.
type planTargets struct {
	kind  models.VariantKind
	keys  []string
	count int
}

// PlanTemplateVariants is the pure variant planner shared by preview, apply, generate_variants and
// template unarchive. It never touches the database: the caller loads the configuration and the
// template's variants (archived ones included) and executes the returned plan in one transaction.
// A plan with violations must be refused as a whole.
func PlanTemplateVariants(input PlanInput) *VariantPlan {
	next := newPlanState(input.Next)
	plan := &VariantPlan{InactiveExclusions: next.inactive}
	if reportDuplicateLines(plan, next) {
		return plan
	}
	maxPreCreated := positiveOr(input.Next.MaxPreCreated, maxTemplateVariants)
	countCap := max(positiveOr(input.CountCap, DefaultPossibleCountCap), maxPreCreated)

	possible := next.countLeaves(next.generating(), nil, countCap)
	targets := next.buildTargets(maxPreCreated, countCap)
	fillPlanCounts(plan, targets, possible, countCap)
	if possible == 0 {
		plan.violate(ReasonPlanNoValidCombination,
			"the configuration leaves no valid combination; archive the template instead", nil)
		return plan
	}
	if targets.count > maxPreCreated {
		plan.violate(ReasonPlanTooManyVariants,
			"the configuration would pre-create {{count}} variants, over the limit of {{limit}}",
			map[string]any{"count": targets.count, "limit": maxPreCreated, "kind": targets.kind})
		return plan
	}
	newPlanReconciler(input, next, targets).run(plan)
	return plan
}

func reportDuplicateLines(plan *VariantPlan, state *planState) bool {
	for _, attributeId := range state.duplicates {
		plan.violate(ReasonPlanDuplicateLine, "attribute {{attribute_id}} appears on more than one line",
			map[string]any{"attribute_id": attributeId})
	}
	return len(state.duplicates) > 0
}

func fillPlanCounts(plan *VariantPlan, targets planTargets, possible int, countCap int) {
	plan.PreCreateKind = targets.kind
	plan.PreCreateCount = targets.count
	if possible > countCap {
		plan.PossibleOverThreshold = true
		return
	}
	plan.PossibleCompleteCount = &possible
}

func positiveOr(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func (this *planState) buildTargets(maxPreCreated int, countCap int) planTargets {
	if len(this.dynamic) == 0 {
		return this.completeTargets(maxPreCreated, countCap)
	}
	if len(this.instant) == 0 {
		return planTargets{}
	}
	return this.baseTargets(maxPreCreated, countCap)
}

func (this *planState) completeTargets(maxPreCreated int, countCap int) planTargets {
	targets := planTargets{kind: models.VariantKindComplete}
	targets.count = this.countLeaves(this.instant, nil, countCap)
	if targets.count > maxPreCreated {
		return targets
	}
	this.walkLeaves(this.instant, nil, func(values []string) bool {
		targets.keys = append(targets.keys, this.keyOfValues(values))
		return true
	})
	return targets
}

// baseTargets walks the instant combinations once, counting those with a valid dynamic completion
// and keeping their keys only while within the pre-create limit.
func (this *planState) baseTargets(maxPreCreated int, countCap int) planTargets {
	targets := planTargets{kind: models.VariantKindBase}
	this.walkLeaves(this.instant, nil, func(values []string) bool {
		if this.countLeaves(this.dynamic, values, 0) == 0 {
			return true
		}
		targets.count++
		if targets.count <= maxPreCreated {
			targets.keys = append(targets.keys, this.keyOfValues(values))
		}
		return targets.count <= countCap
	})
	if targets.count > maxPreCreated {
		targets.keys = nil
	}
	return targets
}
