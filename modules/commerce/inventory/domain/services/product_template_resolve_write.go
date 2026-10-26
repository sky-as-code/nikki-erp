package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// createResolvedComplete creates the complete and its value rows in one transaction. Two
// concurrent requests for the same key race on the (template, kind, key) unique constraint; the
// loser's insert is refused, its transaction rolls back, and it answers with the winner instead.
func (this *ProductTemplateDomainServiceImpl) createResolvedComplete(
	ctx corectx.Context, template *models.ProductTemplate, data itProduct.ResolveCompleteVariantResultData,
	refs TemplateValueRefs,
) (*itProduct.ResolveCompleteVariantResult, error) {
	repo, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}
	createdId := ""
	refused, err := composable.InRefusableTransaction(ctx, repo,
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			id, cErrs, err := MaterializeVariant(tranxCtx, VariantMaterializeSpec{
				TemplateId:    derefString(template.GetId()),
				OrgId:         derefString(template.GetOrgId()),
				Kind:          models.VariantKindComplete,
				Key:           data.CombinationKey,
				BaseVariantId: data.BaseVariantId,
				ValueRefs:     refs,
			})
			createdId = id
			return &cErrs, err
		})
	if err != nil {
		return nil, err
	}
	if refused != nil && refused.Count() > 0 {
		return resolveRaceWinner(ctx, derefString(template.GetId()), data, *refused)
	}
	data.VariantId, data.Outcome = createdId, itProduct.ResolveOutcomeCreated
	return &itProduct.ResolveCompleteVariantResult{Data: data, HasData: true}, nil
}

// resolveRaceWinner re-reads the key after a refused create. An active winner is the answer; no
// winner means the refusal had another cause and is reported as is. Inside a caller's transaction
// the re-read is not attempted: a database refusal has aborted that transaction, so the caller
// must retry as a whole.
func resolveRaceWinner(
	ctx corectx.Context, templateId string, data itProduct.ResolveCompleteVariantResultData,
	refused ft.ClientErrors,
) (*itProduct.ResolveCompleteVariantResult, error) {
	if ctx.GetDbTranx() != nil {
		return &itProduct.ResolveCompleteVariantResult{ClientErrors: refused}, nil
	}
	winner, err := findKeyedVariant(ctx, templateId, models.VariantKindComplete, data.CombinationKey)
	if err != nil {
		return nil, err
	}
	if winner == nil || EffectiveArchiveSource(winner) != "" {
		return &itProduct.ResolveCompleteVariantResult{ClientErrors: refused}, nil
	}
	data.VariantId, data.Outcome = derefString(winner.GetId()), itProduct.ResolveOutcomeExisting
	return &itProduct.ResolveCompleteVariantResult{Data: data, HasData: true}, nil
}
