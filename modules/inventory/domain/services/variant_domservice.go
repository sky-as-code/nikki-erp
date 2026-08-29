package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// engineFor resolves another resource's engine from the registry. It is a variable so a test can
// substitute its own engines: the registry is a package singleton populated during Init, which a
// unit test cannot build.
var engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for '%s'", schemaName)
	}
	return engine, nil
}

// EngineFor exposes the registry lookup to the dynamicengines package, whose action callbacks
// receive only their own engine but need another resource's repository. It delegates to engineFor
// so a test's substitution is honoured here too.
func EngineFor(schemaName string) (drif.DynamicResourceEngine, error) {
	return engineFor(schemaName)
}

// NewProductVariantDomainService derives the variant service from the engine's default one, which
// it embeds so built-in actions keep running unchanged. Installed with Engine.SetResourceService.
//
// The template_* fields are declared as related computed fields in product_variant.json, and the
// engine's computed-field layer batches the template read and fills them on every read path.
func NewProductVariantDomainService(base drif.DynamicResourceService) *ProductVariantDomainServiceImpl {
	return &ProductVariantDomainServiceImpl{DynamicResourceService: base}
}

// ProductVariantDomainServiceImpl carries the variant's domain behaviors: the archive cascade kept
// in step with the owning template, and the read services.
type ProductVariantDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*ProductVariantDomainServiceImpl)(nil)

// SetArchived archives the variant, stamps why it was archived, and brings its template in step.
// Both follow-ups must run AFTER the variant row is written, which is why they are not in the
// engine's AfterValidationSuccess hook: that runs before MainProcess, so the "are any variants
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

	result, err := this.DynamicResourceService.SetArchived(ctx, params)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 {
		return result, err
	}

	variant := models.NewProductVariantFrom(params)
	archived := variant.IsArchived()
	variantId := derefString(variant.GetId())
	if archived == nil || variantId == "" {
		return result, nil
	}

	// The stamp is a second write, so it supersedes the etag the archive produced. Reporting the
	// archive's stale etag would have the caller's next request rejected as a concurrent modification
	// by a change this same call made.
	stampEtag, err := this.stampArchiveSource(ctx, variantId, *archived)
	if err != nil {
		return nil, err
	}
	if stampEtag != "" {
		result.Data.Etag = stampEtag
	}

	templateId, err := this.templateIdOf(ctx, variantId)
	if err != nil || templateId == "" {
		return result, err
	}
	if err := this.syncTemplateAvailability(ctx, templateId, *archived); err != nil {
		return nil, err
	}
	return result, nil
}

// stampArchiveSource records that this archive was the user's own doing, so a later template
// unarchive restores only the variants its cascade took down. It returns the etag the stamp
// produced, which becomes the row's current one.
func (this *ProductVariantDomainServiceImpl) stampArchiveSource(
	ctx corectx.Context, variantId string, archived bool,
) (string, error) {
	update := dmodel.DynamicFields{models.ProductVariantFieldId: variantId}
	if archived {
		update[models.ProductVariantFieldArchiveSource] = models.ArchiveSourceUser.String()
	} else {
		update[models.ProductVariantFieldArchiveSource] = nil
	}

	engine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return "", err
	}
	result, err := engine.ResourceRepository().Update(ctx, update)
	if err != nil {
		return "", errors.Wrap(err, "stampArchiveSource")
	}
	if result == nil || !result.HasData {
		return "", nil
	}
	return string(result.Data.Etag), nil
}

// syncTemplateAvailability archives a template once its last selectable variant is gone, and brings
// it back when a variant returns: a template with nothing transactable must not advertise itself as
// available.
func (this *ProductVariantDomainServiceImpl) syncTemplateAvailability(
	ctx corectx.Context, templateId string, archived bool,
) error {
	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return err
	}

	if archived {
		remaining, err := models.FindActiveTemplateVariants(
			ctx, variantEngine.ResourceRepository(), templateId, 1)
		if err != nil {
			return errors.Wrap(err, "syncTemplateAvailability")
		}
		if len(remaining) > 0 {
			return nil
		}
	}

	templateEngine, err := engineFor(models.ProductTemplateSchemaName)
	if err != nil {
		return err
	}
	// Written through the repository, not the template's set_archived action: re-entering that action
	// would run the template's cascade back over the variants that triggered it.
	_, err = templateEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.ProductTemplateFieldId: templateId,
		basemodel.FieldIsArchived:     archived,
	})
	return errors.Wrap(err, "syncTemplateAvailability")
}

// templateIdOf reads the owning template from the stored row: a set_archived payload carries only
// the id and the flag.
func (this *ProductVariantDomainServiceImpl) templateIdOf(
	ctx corectx.Context, variantId string,
) (string, error) {
	engine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return "", err
	}

	found, err := engine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
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
