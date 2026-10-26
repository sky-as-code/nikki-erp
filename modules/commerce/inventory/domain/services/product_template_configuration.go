package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// One reconcile path serves every entry point (plan D2): template create and update with a
// configuration payload, the preview dry run, generate_variants and template unarchive. Each loads
// the stored configuration, validates the desired one, plans against every variant (archived
// included), and either reports the plan or writes the configuration and executes it.

// reconcileRequest is one reconcile. input nil means "the current configuration, unchanged".
type reconcileRequest struct {
	templateId    string
	input         *itProduct.TemplateConfigurationInput
	deleteAllowed bool
	dryRun        bool
}

// reconcileOutcome is the plan, the delete-or-archive verdicts and, after an apply, what was done.
type reconcileOutcome struct {
	template  *models.ProductTemplate
	state     *templateConfigState
	desired   *desiredConfig
	plan      *VariantPlan
	variants  *templateVariants
	decisions map[string]invalidDecision
	result    ReconcileResult
}

func (this *ProductTemplateDomainServiceImpl) reconcile(
	ctx corectx.Context, req reconcileRequest,
) (*reconcileOutcome, ft.ClientErrors, error) {
	template, err := this.fetchTemplate(ctx, req.templateId)
	if err != nil || template == nil {
		return nil, nil, err
	}
	orgId := derefString(template.GetOrgId())
	state, err := loadTemplateConfigState(ctx, req.templateId)
	if err != nil {
		return nil, nil, err
	}
	input := itProduct.TemplateConfigurationInput{}
	if req.input != nil {
		input = *req.input
	}
	desired, cErrs, err := resolveDesiredConfig(ctx, state, orgId, input)
	if err != nil || cErrs.Count() > 0 {
		return nil, cErrs, err
	}
	outcome, cErrs, err := this.planReconcile(ctx, template, state, desired, req.deleteAllowed)
	if err != nil || cErrs.Count() > 0 || req.dryRun {
		return outcome, cErrs, err
	}
	cErrs, err = this.executeReconcile(ctx, outcome, state, desired, req.input != nil)
	return outcome, cErrs, err
}

func (this *ProductTemplateDomainServiceImpl) planReconcile(
	ctx corectx.Context, template *models.ProductTemplate, state *templateConfigState,
	desired *desiredConfig, deleteAllowed bool,
) (*reconcileOutcome, ft.ClientErrors, error) {
	variants, err := loadTemplateVariants(ctx, state)
	if err != nil {
		return nil, nil, err
	}
	archived := template.IsArchived()
	plan := PlanTemplateVariants(PlanInput{
		Previous:         state.previousPlanConfig(),
		Next:             desired.planConfig(),
		Existing:         variants.existing,
		TemplateArchived: archived != nil && *archived,
	})
	if plan.HasViolations() {
		return nil, remapPlanErrors(plan.ClientErrors, desired), nil
	}
	decider := usageDecider{
		dispatcher: this.usageDispatcher, orgId: derefString(template.GetOrgId()),
		deleteAllowed: deleteAllowed, children: variants.children,
	}
	decisions, cErrs, err := decider.decide(ctx, plan.Invalid)
	if err != nil || cErrs.Count() > 0 {
		return nil, cErrs, err
	}
	return &reconcileOutcome{
		template: template, state: state, desired: desired, plan: plan, variants: variants, decisions: decisions,
	}, nil, nil
}

func (this *ProductTemplateDomainServiceImpl) executeReconcile(
	ctx corectx.Context, outcome *reconcileOutcome, state *templateConfigState,
	desired *desiredConfig, writeConfig bool,
) (ft.ClientErrors, error) {
	templateId := derefString(outcome.template.GetId())
	orgId := derefString(outcome.template.GetOrgId())
	if writeConfig {
		written, cErrs, err := writeConfiguration(ctx, configWriter{templateId: templateId, orgId: orgId, state: state}, desired)
		if err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
		state = written
	}
	archived := outcome.template.IsArchived()
	executor := &planExecutor{
		templateId: templateId, orgId: orgId, templateActive: archived == nil || !*archived,
		refs: state.valueRefs(), variants: outcome.variants, decisions: outcome.decisions,
	}
	cErrs, err := executor.run(ctx, outcome.plan)
	outcome.result = executor.result
	return cErrs, err
}

// remapPlanErrors points a planner violation naming an attribute at that attribute's payload row.
func remapPlanErrors(cErrs ft.ClientErrors, desired *desiredConfig) ft.ClientErrors {
	out := make(ft.ClientErrors, 0, len(cErrs))
	for _, item := range cErrs {
		if attributeId, ok := item.Vars["attribute_id"].(string); ok {
			if row := desired.rowOfAttribute(attributeId); row >= 0 {
				item.Field = linePath(row, "attribute_id")
			}
		}
		out = append(out, item)
	}
	return out
}
