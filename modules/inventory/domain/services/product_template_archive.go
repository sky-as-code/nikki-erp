package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Archiving a product template withdraws its variants from new business along with it: one
// operation over two resources. It lives on the service because it writes; a dynamicengines
// callback may only adapt and validate.

// SetArchived archives or restores a template, then brings its variants with it. It overrides the
// promoted DynamicResourceService.SetArchived, which is what changes POST /:id/archived.
//
// The template write and the cascade share one transaction, so a failed variant write leaves the
// template unarchived rather than half-cascaded.
func (this *ProductTemplateDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	archived, hasFlag := readArchivedFlag(params)
	if !hasFlag {
		// is_archived is RequiredAlways on the command schema, so the base call reports the missing
		// flag as a client error.
		return this.DynamicResourceService.SetArchived(ctx, params)
	}

	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}

	// The stock guard runs first, over every variant, and outside the transaction that follows: a
	// check made while the cascade walks would leave the variants it had already reached archived
	// when a later one turned out to hold stock. One variant with stock blocks the whole line.
	templateId := readStringParam(params, models.ProductTemplateFieldId)
	if guarded, err := this.guardStockUsage(ctx, templateId, archived); err != nil || guarded != nil {
		return guarded, err
	}

	tranx, err := variantEngine.ResourceRepository().BeginTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "SetArchived")
	}
	defer tranx.Rollback()

	// The transaction goes on a cloned context, never ctx itself, or a committed transaction stays
	// visible to whatever runs next. CloneRequestContext carries the identity the audit columns need.
	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	result, err := this.DynamicResourceService.SetArchived(tranxCtx, params)
	if err != nil {
		return nil, errors.Wrap(err, "SetArchived")
	}
	// A client error means the template was not archived, so there is nothing to cascade.
	if result.ClientErrors.Count() > 0 {
		return result, nil
	}

	if err := this.cascadeArchiveToVariants(tranxCtx, variantEngine, templateId, archived); err != nil {
		return nil, err
	}
	return result, errors.Wrap(tranx.Commit(), "SetArchived")
}

// guardStockUsage refuses a template archive that would strand stock. A non-nil result is the
// refusal, which the caller returns as-is; nil means proceed. Only archiving is guarded: guarding
// unarchive would make a mistakenly archived line unrecoverable.
func (this *ProductTemplateDomainServiceImpl) guardStockUsage(
	ctx corectx.Context, templateId string, archiving bool,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if !archiving {
		return nil, nil
	}

	reader, err := stockUsageReader()
	if err != nil {
		// No reader means the stock side is not wired in this deployment; the archive proceeds under
		// the module's own rules rather than Product becoming unusable.
		return nil, nil
	}

	vErrs := &ft.ClientErrors{}
	if err := GuardTemplateArchive(ctx, reader, templateId, archiving, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() == 0 {
		return nil, nil
	}
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
}

// cascadeArchiveToVariants applies a template's archive change to each of its variants, deciding
// per variant with ShouldSkipCascade and CascadeArchiveFields.
func (this *ProductTemplateDomainServiceImpl) cascadeArchiveToVariants(
	ctx corectx.Context, variantEngine drif.DynamicResourceEngine, templateId string, archive bool,
) error {
	if templateId == "" {
		return nil
	}

	repo := variantEngine.ResourceRepository()
	variants, err := models.FindTemplateVariants(ctx, repo, templateId, MaxCascadeVariants)
	if err != nil {
		return errors.Wrap(err, "cascadeArchiveToVariants")
	}

	for _, item := range variants {
		variant := models.NewProductVariantFrom(item)
		if ShouldSkipCascade(variant, archive) {
			continue
		}

		update := CascadeArchiveFields(derefString(variant.GetId()), archive)
		if _, err := repo.Update(ctx, update); err != nil {
			return errors.Wrap(err, "cascadeArchiveToVariants")
		}
	}
	return nil
}

// readArchivedFlag reads the archive direction out of the action params, reporting whether it was
// present. Absent is not the same as false — it means no flag was sent, which the command schema
// rejects — so the two must stay distinguishable.
func readArchivedFlag(params dmodel.DynamicFields) (bool, bool) {
	val, ok := params[basemodel.FieldIsArchived]
	if !ok || val == nil {
		return false, false
	}

	switch typed := val.(type) {
	case bool:
		return typed, true
	case *bool:
		if typed == nil {
			return false, false
		}
		return *typed, true
	default:
		return false, false
	}
}

func readStringParam(params dmodel.DynamicFields, field string) string {
	val, ok := params[field]
	if !ok || val == nil {
		return ""
	}
	// model.Id is a string type, so the string case covers it too.
	if typed, ok := val.(string); ok {
		return typed
	}
	return ""
}
