package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The template's attribute configuration is applied by the template's own create and update (the
// update permission) and previewed by its own action. Whether reconcile may delete an unused
// invalid variant is the caller's product_variant delete power, checked here without failing:
// lacking it turns a delete into an archive, never a refusal.

// mayDeleteVariants reports whether the caller holds delete on product variants in the org the
// params name.
func (this *ProductTemplateApplicationServiceImpl) mayDeleteVariants(ctx corectx.Context, params dmodel.DynamicFields) bool {
	orgId := composable.ParamOrgId(params)
	perm := requestguard.PermFor(composable.PermissionDelete, models.ProductVariantSchemaName, requestguard.ResourceScopeOrg)
	return requestguard.AssertPermission(ctx, perm.InOrg(orgId)) == nil
}

// stampVariantDeleteAllowed overwrites whatever a client sent under the flag with the checked
// answer, so the domain service never trusts the payload for it.
func (this *ProductTemplateApplicationServiceImpl) stampVariantDeleteAllowed(ctx corectx.Context, cmd dmodel.DynamicFields) {
	cmd[itProduct.ProductTemplateFieldVariantDeleteAllowed] = this.mayDeleteVariants(ctx, cmd)
}

// PreviewAttributeConfiguration dry-runs a configuration save of the template:
// {org_id, etag, attribute_lines, exclusions}, the same body a PUT would carry.
func (this *ProductTemplateApplicationServiceImpl) PreviewAttributeConfiguration(
	ctx corectx.Context, cmd itProduct.PreviewAttributeConfigurationCommand,
) (*itProduct.PreviewAttributeConfigurationResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionPreviewAttributeConfiguration, cmd); err != nil || cErrs != nil {
		return anyFailure(cErrs, err)
	}
	previewer, ok := this.productSvc.(itProduct.ProductTemplateConfigurationService)
	if !ok {
		return nil, errors.New("PreviewAttributeConfiguration: the product service cannot preview configurations")
	}
	query, cErrs := buildPreviewQuery(cmd)
	if cErrs.Count() > 0 {
		return anyFailure(&cErrs, nil)
	}
	query.DeleteAllowed = this.mayDeleteVariants(ctx, cmd)
	result, err := previewer.PreviewTemplateConfiguration(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "PreviewAttributeConfiguration")
	}
	if result.ClientErrors.Count() > 0 {
		return anyFailure(&result.ClientErrors, nil)
	}
	if !result.HasData {
		return &itProduct.PreviewAttributeConfigurationResult{}, nil
	}
	return anyResult(result.Data), nil
}

func buildPreviewQuery(cmd dmodel.DynamicFields) (itProduct.PreviewTemplateConfigurationQuery, ft.ClientErrors) {
	copied := dmodel.DynamicFields{}
	for key, value := range cmd {
		copied[key] = value
	}
	input, _, malformed := itProduct.LiftTemplateConfiguration(copied)
	if malformed != "" {
		return itProduct.PreviewTemplateConfigurationQuery{}, ft.ClientErrors{*ft.NewValidationError(malformed,
			ft.ErrorKey("err_malformed_request"),
			"attribute_lines must be [{attribute_id, value_ids, sequence}] and exclusions [{left_value_id, right_value_id}]")}
	}
	query := itProduct.PreviewTemplateConfigurationQuery{
		TemplateId: readStringField(cmd, paramRecordId),
		Etag:       readStringField(cmd, basemodel.FieldEtag),
	}
	if input != nil {
		query.Input = *input
	}
	return query, nil
}
