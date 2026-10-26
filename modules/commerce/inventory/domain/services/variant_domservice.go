package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// NewProductVariantDomainService derives the variant service from the onion's default one, which
// it embeds so built-in actions keep running unchanged.
//
// The template_* fields are declared as related computed fields in product_variant.json, and the
// engine's computed-field layer batches the template read and fills them on every read path.
func NewProductVariantDomainService(
	base composable.CrudDomainService, usageDispatcher *usagecheck.Dispatcher,
) *ProductVariantDomainServiceImpl {
	return &ProductVariantDomainServiceImpl{
		CrudDomainService: base,
		usageDispatcher:   usageDispatcher,
	}
}

// ProductVariantDomainServiceImpl carries the variant's domain behaviors: the archive cascade kept
// in step with the owning template, and the read services.
type ProductVariantDomainServiceImpl struct {
	composable.CrudDomainService

	usageDispatcher *usagecheck.Dispatcher
}

var _ composable.CrudDomainService = (*ProductVariantDomainServiceImpl)(nil)

func (this *ProductVariantDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.BeforeValidation = chainBeforeValidation(opts.BeforeValidation, normalizeVariantIdentifiers())
	opts.BeforeValidation = chainBeforeValidation(opts.BeforeValidation, checkCreateCombination(this.Repository()))
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra,
		validateCreateIdentifiers(this.Repository()))
	return this.CrudDomainService.Create(ctx, params, opts)
}

func (this *ProductVariantDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.UpdateOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	submittedTemplateId := ""
	opts.BeforeValidation = chainBeforeValidation(opts.BeforeValidation, normalizeVariantIdentifiers())
	opts.BeforeValidation = chainBeforeValidation(opts.BeforeValidation, captureSubmittedTemplateId(&submittedTemplateId))
	opts.ValidateExtra = composable.ChainValidateExtra(
		opts.ValidateExtra, validateVariantUpdate(this.Repository(), &submittedTemplateId),
		validateUpdateIdentifiers(this.Repository()))
	return this.CrudDomainService.Update(ctx, params, opts)
}

// Delete refuses to remove a base that still has completes, the last active complete of a
// template with no on-demand line (archive the template instead), and a variant that another
// module still sells, buys or stocks in a kiosk.
//
// The foreign keys on product_variant_id are ON DELETE RESTRICT, so the database would refuse the
// statement anyway — but it would do so as a driver error carrying a constraint name, and the
// answer would say nothing about which module to go and clear. This asks first, so the refusal
// names the modules; the constraint remains underneath as the authority that also settles the
// race between this check and a concurrent insert.
//
// Archiving is the alternative and is guarded separately: SetArchived refuses only on live stock,
// because history does not prevent a variant from being withdrawn from sale.
func (this *ProductVariantDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.DeleteOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	vErrs := ft.NewClientErrors()
	variantId := derefString(models.NewProductVariantFrom(params).GetId())
	if err := assertVariantDeleteKeepsTemplateUsable(ctx, variantId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	if err := assertVariantDeletable(ctx, this.usageDispatcher, params, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, params, options...)
}

// assertVariantDeletable asks every module holding a product_variant_id whether this variant is
// still in use.
func assertVariantDeletable(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher,
	params dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	variantId := derefString(models.NewProductVariantFrom(params).GetId())
	if variantId == "" {
		return nil
	}

	return usagecheck.AssertDeletable(ctx, dispatcher, modconstants.InventoryModuleName,
		models.ProductVariantFieldId, usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceProductVariant,
			Identifier: usagecheck.NewIdentifier(
				variantId, readStringParam(params, basemodel.FieldOrgId)),
		}, vErrs)
}

// SetArchived archives the variant, brings its template in step, and stamps why it was archived.
// The follow-ups must run AFTER the variant row is written, which is why they are not in the
// AfterValidationSuccess hook: that runs before the write, so the "are any variants
// left?" count would still see this variant unarchived and never archive the template.
func (this *ProductVariantDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	// The stock guard runs before the write. Archiving must never make stock disappear: a variant
	// still holding goods, owing them to a reservation, or named by work in flight is refused
	// outright, and nothing about its stock is touched either way.
	if guarded, err := guardVariantStockUsage(ctx, params); err != nil || guarded != nil {
		return guarded, err
	}
	if guarded, err := guardVariantRestore(ctx, params); err != nil || guarded != nil {
		return guarded, err
	}

	result, err := this.CrudDomainService.SetArchived(ctx, params)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 {
		return result, err
	}

	variant := models.NewProductVariantFrom(params)
	archived := variant.IsArchived()
	variantId := derefString(variant.GetId())
	if archived == nil || variantId == "" {
		return result, nil
	}

	templateId, err := this.templateIdOf(ctx, variantId)
	if err != nil {
		return nil, err
	}
	templateArchived, err := this.syncTemplateAvailability(ctx, templateId, *archived)
	if err != nil {
		return nil, err
	}

	// The stamp is a second write, so it supersedes the etag the archive produced. Reporting the
	// archive's stale etag would have the caller's next request rejected as a concurrent modification
	// by a change this same call made.
	stampEtag, err := this.stampArchiveSource(ctx, variantId, archiveStamp(*archived, templateArchived))
	if err != nil {
		return nil, err
	}
	if stampEtag != "" {
		result.Data.Etag = stampEtag
	}
	return result, nil
}

// archiveStamp is the archive_source a set_archived call leaves on the variant: nil on unarchive,
// user for a deliberate archive, and template_cascade when that archive took the template's last
// selectable variant and so archived the template with it. The template's unarchive restores
// exactly its cascade, so the variant that caused it comes back with it rather than leaving the
// restored template with nothing to sell.
func archiveStamp(archived bool, templateArchived bool) *models.ArchiveSource {
	switch {
	case !archived:
		return nil
	case templateArchived:
		return util.ToPtr(models.ArchiveSourceTemplateCascade)
	default:
		return util.ToPtr(models.ArchiveSourceUser)
	}
}

// stampArchiveSource records why the variant was archived (nil clears it), so a later template
// unarchive restores only the variants its cascade took down. It returns the etag the stamp
// produced, which becomes the row's current one.
func (this *ProductVariantDomainServiceImpl) stampArchiveSource(
	ctx corectx.Context, variantId string, source *models.ArchiveSource,
) (string, error) {
	update := dmodel.DynamicFields{models.ProductVariantFieldId: variantId}
	if source != nil {
		update[models.ProductVariantFieldArchiveSource] = source.String()
	} else {
		update[models.ProductVariantFieldArchiveSource] = nil
	}

	engine, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return "", err
	}
	result, err := engine.Update(ctx, update)
	if err != nil {
		return "", errors.Wrap(err, "stampArchiveSource")
	}
	if result == nil || !result.HasData {
		return "", nil
	}
	return string(result.Data.Etag), nil
}

// syncTemplateAvailability archives a template once its last selectable variant is archived: a
// template with nothing transactable must not advertise itself as available. Only completes are
// selectable, and a template with a dynamic or never line is left alone, because it resolves its
// products on demand and having no active variant is normal for it.
//
// It reports whether it archived the template, which makes this variant's archive part of the
// template's cascade (see archiveStamp).
//
// Unarchiving never runs it: a manual unarchive is refused while the template is archived, so the
// template is already active by then.
func (this *ProductVariantDomainServiceImpl) syncTemplateAvailability(
	ctx corectx.Context, templateId string, archived bool,
) (bool, error) {
	if !archived || templateId == "" {
		return false, nil
	}
	config, err := LoadTemplateConfiguration(ctx, templateId)
	if err != nil || config.HasOnDemandLine() {
		return false, err
	}
	remaining, err := findActiveCompleteVariants(ctx, templateId, 1)
	if err != nil || len(remaining) > 0 {
		return false, errors.Wrap(err, "syncTemplateAvailability")
	}

	templateEngine, err := repoFor(models.ProductTemplateSchemaName)
	if err != nil {
		return false, err
	}
	// Written through the repository, not the template's set_archived action: re-entering that action
	// would run the template's cascade back over the variants that triggered it.
	_, err = templateEngine.Update(ctx, dmodel.DynamicFields{
		models.ProductTemplateFieldId: templateId,
		basemodel.FieldIsArchived:     true,
	})
	if err != nil {
		return false, errors.Wrap(err, "syncTemplateAvailability")
	}
	return true, nil
}

// guardVariantRestore runs the configuration check of a manual unarchive. A non-nil result is the
// refusal; archiving is never refused here.
func guardVariantRestore(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	archived, hasFlag := readArchivedFlag(params)
	variantId := derefString(models.NewProductVariantFrom(params).GetId())
	if !hasFlag || archived || variantId == "" {
		return nil, nil
	}
	vErrs := ft.NewClientErrors()
	if err := assertVariantRestorable(ctx, variantId, vErrs); err != nil || vErrs.Count() == 0 {
		return nil, err
	}
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
}

// templateIdOf reads the owning template from the stored row: a set_archived payload carries only
// the id and the flag.
func (this *ProductVariantDomainServiceImpl) templateIdOf(
	ctx corectx.Context, variantId string,
) (string, error) {
	engine, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return "", err
	}

	found, err := engine.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.ProductVariantFieldId: variantId},
		Fields: []string{models.ProductVariantFieldId, models.ProductVariantFieldProductTemplateId},
	})
	if err != nil {
		return "", errors.Wrap(err, "templateIdOf")
	}
	if !found.HasData {
		return "", nil
	}
	return derefString(models.NewProductVariantFrom(found.Data).GetProductTemplateId()), nil
}

// paramFieldNames is the request key carrying the field projection, shared by the read services
// in this package.
const paramFieldNames = "fields"
