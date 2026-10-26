package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// The template's own create, update and delete carry its attribute configuration and its
// variants in one transaction (plan D2, D7), following the nested-delta pattern of the IAM role:
// lift the nested fields, write the template row first so its etag guards the configuration, then
// the configuration and the reconcile. Any refusal rolls the template change back too.

// liftConfigurationOrRefuse lifts the configuration payload and checks its shape.
func liftConfigurationOrRefuse(
	cmd dmodel.DynamicFields,
) (*itProduct.TemplateConfigurationInput, bool, ft.ClientErrors) {
	input, deleteAllowed, malformed := itProduct.LiftTemplateConfiguration(cmd)
	if malformed != "" {
		return nil, false, ft.ClientErrors{*ft.NewValidationError(malformed, ft.ErrorKey("err_malformed_request"),
			"attribute_lines must be [{attribute_id, value_ids, sequence}] and exclusions "+
				"[{left_value_id, right_value_id}]")}
	}
	if input == nil {
		return nil, deleteAllowed, nil
	}
	return input, deleteAllowed, validateConfigurationShape(*input)
}

// Create writes the template and its variants: the configuration's pre-created variants, or the
// single empty-key default complete when there is no generating attribute (CR01 §6.1).
func (this *ProductTemplateDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	input, deleteAllowed, cErrs := liftConfigurationOrRefuse(cmd)
	if cErrs.Count() > 0 {
		return &composable.CreateResult{ClientErrors: cErrs}, nil
	}
	opts := safe.GetOptional(options, composable.CreateOptions{})
	var result *composable.CreateResult
	refused, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			if result, err = this.CrudDomainService.Create(tranxCtx, cmd, opts); err != nil {
				return nil, err
			}
			if result.ClientErrors.Count() > 0 || !result.HasData {
				return &result.ClientErrors, nil
			}
			_, cErrs, err := this.reconcile(tranxCtx, reconcileRequest{
				templateId: readStringParam(result.Data, basemodel.FieldId), input: input, deleteAllowed: deleteAllowed,
			})
			return &cErrs, err
		})
	if err != nil {
		return nil, errors.Wrap(err, "create product template")
	}
	if refused != nil {
		return &composable.CreateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

// Update applies the template's own columns and, when the body carries `attribute_lines` or
// `exclusions`, the configuration with its reconcile. A configuration-only save still moves the
// template's etag, which is the configuration's version. A replaced default image is removed from
// IT Storage once the whole update has committed.
func (this *ProductTemplateDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	return this.updateWithConfiguration(ctx, cmd, safe.GetOptional(options, composable.UpdateOptions{}))
}

func (this *ProductTemplateDomainServiceImpl) updateWithConfiguration(
	ctx corectx.Context, cmd composable.UpdateCommand, opts composable.UpdateOptions,
) (*composable.MutateResult, error) {
	input, deleteAllowed, cErrs := liftConfigurationOrRefuse(cmd)
	if cErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: cErrs}, nil
	}
	if input == nil {
		return this.CrudDomainService.Update(ctx, cmd, opts)
	}
	var result *composable.MutateResult
	refused, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			if result, err = this.CrudDomainService.Update(tranxCtx, cmd, opts); err != nil {
				return nil, err
			}
			if result.ClientErrors.Count() > 0 || !result.HasData {
				return &result.ClientErrors, nil
			}
			_, cErrs, err := this.reconcile(tranxCtx, reconcileRequest{
				templateId: string(composable.ParamId(cmd)), input: input, deleteAllowed: deleteAllowed,
			})
			return &cErrs, err
		})
	if err != nil {
		return nil, errors.Wrap(err, "update product template")
	}
	if refused != nil {
		return &composable.MutateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

// Delete removes the template together with its variants when none of them was ever used
// (Inventory history and every dependant module) and no module names the template itself;
// otherwise it refuses and suggests archiving. Lines, values and exclusions cascade.
func (this *ProductTemplateDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	var result *composable.MutateResult
	refused, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			cErrs, err := this.deleteTemplateVariants(tranxCtx, cmd)
			if err != nil || cErrs.Count() > 0 {
				return &cErrs, err
			}
			if result, err = this.CrudDomainService.Delete(tranxCtx, cmd, options...); err != nil {
				return nil, err
			}
			return &result.ClientErrors, nil
		})
	if err != nil {
		return nil, errors.Wrap(err, "delete product template")
	}
	if refused != nil {
		return &composable.MutateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

func (this *ProductTemplateDomainServiceImpl) deleteTemplateVariants(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (ft.ClientErrors, error) {
	templateId := string(composable.ParamId(cmd))
	template, err := this.fetchTemplate(ctx, templateId)
	if err != nil || template == nil {
		return nil, err
	}
	vErrs := ft.ClientErrors{}
	if this.usageDispatcher != nil {
		params := dmodel.DynamicFields{basemodel.FieldId: templateId, basemodel.FieldOrgId: derefString(template.GetOrgId())}
		if err := AssertTemplateNotUsedExternally(withoutTransaction(ctx), this.usageDispatcher, params, &vErrs); err != nil {
			return nil, err
		}
		if vErrs.Count() > 0 {
			return vErrs, nil
		}
	}
	state, err := loadTemplateConfigState(ctx, templateId)
	if err != nil {
		return nil, err
	}
	variants, err := loadTemplateVariants(ctx, state)
	if err != nil {
		return nil, err
	}
	return this.deleteUnusedVariants(ctx, template, variants)
}

// deleteUnusedVariants deletes every variant, completes before bases, after one batched usage
// decision; a single used variant refuses the whole delete.
func (this *ProductTemplateDomainServiceImpl) deleteUnusedVariants(
	ctx corectx.Context, template *models.ProductTemplate, variants *templateVariants,
) (ft.ClientErrors, error) {
	all := make([]PlannedVariant, 0, len(variants.existing))
	for _, variant := range variants.existing {
		all = append(all, PlannedVariant{VariantId: variant.Id, Kind: variant.Kind, Key: variant.Key})
	}
	decider := usageDecider{
		dispatcher: this.usageDispatcher, orgId: derefString(template.GetOrgId()),
		deleteAllowed: true, children: variants.children,
	}
	decisions, cErrs, err := decider.decide(ctx, all)
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	used := 0
	for _, decision := range decisions {
		if !decision.delete {
			used++
		}
	}
	if used > 0 {
		return ft.ClientErrors{TemplateHasUsedVariantsViolation(used)}, nil
	}
	executor := &planExecutor{variants: variants, decisions: decisions}
	return executor.removeInvalid(ctx, all)
}
