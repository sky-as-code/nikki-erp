package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// Turning an attribute selection into a concrete variant, and bringing a template's variants in
// step with its attribute configuration. Methods on the same type as the service core.

// ResolveProductSelection is the earlier lookup-by-selection port. On-demand resolve with
// validation, bases and restore is ResolveCompleteVariant; this one only finds a complete by key
// and, when asked, materializes it with its join rows.
func (this *ProductTemplateDomainServiceImpl) ResolveProductSelection(
	ctx corectx.Context, query itProduct.ResolveProductSelectionQuery,
) (*itProduct.ResolveProductSelectionResult, error) {
	template, err := this.fetchTemplate(ctx, query.TemplateId)
	if err != nil || template == nil {
		return &itProduct.ResolveProductSelectionResult{}, err
	}
	combinationKey := ResolveProductSelection(query.Selections)
	result := &itProduct.ResolveProductSelectionResult{
		Data:    itProduct.ResolveProductSelectionResultData{CombinationKey: combinationKey},
		HasData: true,
	}
	variantEngine, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}
	existing, err := models.FindVariantsByCombination(ctx, variantEngine, query.TemplateId, combinationKey, 1)
	if err != nil {
		return nil, errors.Wrap(err, "ResolveProductSelection")
	}
	if len(existing) > 0 {
		result.Data.VariantId = derefString(models.NewProductVariantFrom(existing[0]).GetId())
		return result, nil
	}
	if !query.MaterializeIfMissing {
		return result, nil
	}
	variantId, cErrs, err := MaterializeVariant(ctx, VariantMaterializeSpec{
		TemplateId: query.TemplateId, OrgId: derefString(template.GetOrgId()),
		Kind: models.VariantKindComplete, Key: combinationKey,
	})
	if err != nil || cErrs.Count() > 0 {
		return &itProduct.ResolveProductSelectionResult{ClientErrors: cErrs}, err
	}
	result.Data.VariantId = variantId
	result.Data.Materialized = variantId != ""
	return result, nil
}

// GenerateVariants reconciles the template's variants with its current configuration, unchanged:
// missing pre-created records are created or restored, invalid ones archived or deleted. It is
// the same reconcile a configuration save runs, in one transaction.
func (this *ProductTemplateDomainServiceImpl) GenerateVariants(
	ctx corectx.Context, query itProduct.GenerateVariantsQuery,
) (*itProduct.GenerateVariantsResult, error) {
	if query.TemplateId == "" {
		return &itProduct.GenerateVariantsResult{}, nil
	}
	var outcome *reconcileOutcome
	refused, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var cErrs ft.ClientErrors
			var err error
			outcome, cErrs, err = this.reconcile(tranxCtx, reconcileRequest{
				templateId: query.TemplateId, deleteAllowed: query.DeleteAllowed,
			})
			return &cErrs, err
		})
	if err != nil {
		return nil, errors.Wrap(err, "GenerateVariants")
	}
	if refused != nil {
		return &itProduct.GenerateVariantsResult{ClientErrors: *refused}, nil
	}
	if outcome == nil {
		return &itProduct.GenerateVariantsResult{}, nil
	}
	return &itProduct.GenerateVariantsResult{Data: generateResultData(outcome.result), HasData: true}, nil
}

func generateResultData(done ReconcileResult) itProduct.GenerateVariantsResultData {
	return itProduct.GenerateVariantsResultData{
		CreatedVariantIds:  done.Created,
		ExtendedVariantIds: done.Extended,
		RestoredVariantIds: done.Restored,
		ArchivedVariantIds: done.Archived,
		DeletedVariantIds:  done.Deleted,
		ObsoleteVariantIds: append(append([]string{}, done.Archived...), done.Deleted...),
		UnchangedCount:     len(done.Kept),
	}
}
