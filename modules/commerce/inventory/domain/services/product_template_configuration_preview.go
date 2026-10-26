package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type previewResult = dyn.OpResult[itProduct.TemplateConfigurationPreview]

// PreviewTemplateConfiguration runs the apply path as a dry run: the same validation, planner and
// delete-or-archive decisions (usage checks included), and no write. The etag binds the preview
// to the template version it was computed for; the apply re-checks everything.
func (this *ProductTemplateDomainServiceImpl) PreviewTemplateConfiguration(
	ctx corectx.Context, query itProduct.PreviewTemplateConfigurationQuery,
) (*previewResult, error) {
	template, err := this.fetchTemplate(ctx, query.TemplateId)
	if err != nil || template == nil {
		return &previewResult{}, err
	}
	if cErrs := assertPreviewEtag(template, query.Etag); cErrs.Count() > 0 {
		return &previewResult{ClientErrors: cErrs}, nil
	}
	if cErrs := validateConfigurationShape(query.Input); cErrs.Count() > 0 {
		return &previewResult{ClientErrors: cErrs}, nil
	}
	input := query.Input
	outcome, cErrs, err := this.reconcile(ctx, reconcileRequest{
		templateId: query.TemplateId, input: &input, deleteAllowed: query.DeleteAllowed, dryRun: true,
	})
	if err != nil || cErrs.Count() > 0 || outcome == nil {
		return &previewResult{ClientErrors: cErrs}, err
	}
	preview, err := buildPreview(ctx, outcome)
	if err != nil {
		return nil, err
	}
	preview.Etag = query.Etag
	return &previewResult{Data: *preview, HasData: true}, nil
}

func assertPreviewEtag(template *models.ProductTemplate, etag string) ft.ClientErrors {
	if etag == "" {
		return ft.ClientErrors{*ft.NewValidationError("etag", ft.ErrorKey("err_missing_required_field"), "etag is required")}
	}
	stored := template.GetEtag()
	if stored == nil || *stored != etag {
		return ft.ClientErrors{*ft.NewEtagMismatchedError()}
	}
	return nil
}

func buildPreview(ctx corectx.Context, outcome *reconcileOutcome) (*itProduct.TemplateConfigurationPreview, error) {
	plan := outcome.plan
	preview := &itProduct.TemplateConfigurationPreview{
		PreCreateKind: plan.PreCreateKind.String(), PreCreateCount: plan.PreCreateCount,
		PossibleCompleteCount: plan.PossibleCompleteCount, PossibleOverThreshold: plan.PossibleOverThreshold,
		InactiveExclusions: []itProduct.ExclusionInput{},
	}
	for _, exclusion := range plan.InactiveExclusions {
		preview.InactiveExclusions = append(preview.InactiveExclusions,
			itProduct.ExclusionInput{LeftValueId: exclusion.LeftValueId, RightValueId: exclusion.RightValueId})
	}
	labeler, err := newPreviewLabeler(ctx, outcome)
	if err != nil {
		return nil, err
	}
	preview.Keep = previewItems(outcome, labeler, plan.Keep)
	preview.Extend = previewItems(outcome, labeler, plan.Extend)
	preview.Restore = previewItems(outcome, labeler, plan.Restore)
	preview.Create = previewItems(outcome, labeler, plan.Create)
	preview.Untouched = previewItems(outcome, labeler,
		append(append([]PlannedVariant{}, plan.Untouched...), plan.Conflict...))
	preview.Archive, preview.Delete = splitInvalid(outcome, labeler)
	if err := fillStockFigures(ctx, preview.Archive, preview.Delete); err != nil {
		return nil, err
	}
	preview.Counts = itProduct.PreviewCounts{
		Keep: len(preview.Keep), Extend: len(preview.Extend), Restore: len(preview.Restore),
		Create: len(preview.Create), Archive: len(preview.Archive), Delete: len(preview.Delete),
		Untouched: len(preview.Untouched),
	}
	preview.ChangesIdentity = preview.Counts.Extend+preview.Counts.Restore+preview.Counts.Archive+preview.Counts.Delete > 0
	return preview, nil
}

func previewItems(
	outcome *reconcileOutcome, labeler *previewLabeler, items []PlannedVariant,
) []itProduct.PreviewVariant {
	out := make([]itProduct.PreviewVariant, 0, len(items))
	for _, item := range items {
		out = append(out, previewItem(outcome, labeler, item))
	}
	return out
}

func previewItem(outcome *reconcileOutcome, labeler *previewLabeler, item PlannedVariant) itProduct.PreviewVariant {
	key := item.Key
	if item.NewKey != "" {
		key = item.NewKey
	}
	view := itProduct.PreviewVariant{
		VariantId: item.VariantId, VariantKind: item.Kind.String(), CombinationKey: item.Key,
		NewCombinationKey: item.NewKey, BaseCombinationKey: item.BaseKey, ValueIds: keyValueIds(key),
		IsArchived: item.IsArchived, ArchiveSource: item.ArchiveSource.String(),
	}
	variant := outcome.variants.byId[item.VariantId]
	if variant == nil {
		view.DisplayName = labeler.desiredLabel(item.Key)
		return view
	}
	view.Sku = variant.GetSku()
	view.PrimaryBarcode = variant.GetPrimaryBarcode()
	view.DisplayName = labeler.storedLabel(item.Key, derefString(view.Sku))
	if item.NewKey != "" {
		view.NewDisplayName = labeler.desiredLabel(item.NewKey)
	}
	return view
}

func keyValueIds(key string) []string {
	pairs := ParseCombinationKey(key)
	out := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		out = append(out, pair.ValueId)
	}
	return out
}

// splitInvalid separates the invalid records by the executor's verdict.
func splitInvalid(
	outcome *reconcileOutcome, labeler *previewLabeler,
) (archive []itProduct.PreviewVariant, remove []itProduct.PreviewVariant) {
	archive, remove = []itProduct.PreviewVariant{}, []itProduct.PreviewVariant{}
	for _, item := range outcome.plan.Invalid {
		view := previewItem(outcome, labeler, item)
		decision := outcome.decisions[item.VariantId]
		if decision.delete {
			remove = append(remove, view)
			continue
		}
		view.ArchiveReason = decision.reason
		view.UsedBy = decision.usedBy
		archive = append(archive, view)
	}
	return archive, remove
}

// fillStockFigures adds on-hand and reserved quantities to the records leaving use, so the preview
// warns that archiving keeps stock on a withdrawn id. No reader wired means no figures.
func fillStockFigures(ctx corectx.Context, lists ...[]itProduct.PreviewVariant) error {
	ids := []string{}
	for _, list := range lists {
		for _, item := range list {
			ids = append(ids, item.VariantId)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	reader, err := stockUsageReader()
	if err != nil {
		return nil
	}
	usages, err := reader.GetProductUsageBatch(ctx, itStock.GetProductUsageBatchQuery{VariantIds: ids})
	if err != nil || usages == nil {
		return err
	}
	for _, list := range lists {
		for i := range list {
			usage := usages.Data.Usages[list[i].VariantId]
			list[i].OnHandQuantity = usage.OnHandQuantity.String()
			list[i].ReservedQuantity = usage.ReservedQuantity.String()
		}
	}
	return nil
}
