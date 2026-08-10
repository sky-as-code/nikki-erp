package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// Archiving a product template withdraws its variants from new business along with it, which is
// one operation over two resources. It lives here, on the service, because it writes: a
// dynamicengines callback may adapt and validate, but the writes belong to the service layer.
// See docs/wiki/07. ERP backend module.md §6.7.

// SetArchived archives or restores a template, then brings its variants with it.
//
// It overrides the promoted DynamicResourceService.SetArchived, which is all it takes to change
// what POST /:id/archived does: the built-in set_archived action calls SetArchived on whatever
// service the engine has installed, and Init installs this one. See AC-PROD-019 and BR §8.9.
//
// The template write and the cascade share one transaction, so a failed variant write leaves the
// template unarchived rather than half-cascaded.
func (this *ProductAppServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	archived, hasFlag := readArchivedFlag(params)
	if !hasFlag {
		// is_archived is RequiredAlways on the command schema, so the base call reports the
		// missing flag as a client error. Delegating keeps that message in one place.
		return this.DynamicResourceService.SetArchived(ctx, params)
	}

	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}

	tranx, err := variantEngine.ResourceRepository().BeginTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "SetArchived")
	}
	defer tranx.Rollback()

	// The transaction goes on a scoped copy of the context, never on ctx itself: setting it on
	// the caller's context would leave a committed transaction visible to whatever runs next.
	// CloneRequestContext carries the caller's identity across, which the audit columns need.
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

	templateId := readStringParam(params, models.ProductTemplateFieldId)
	if err := this.cascadeArchiveToVariants(tranxCtx, variantEngine, templateId, archived); err != nil {
		return nil, err
	}
	return result, errors.Wrap(tranx.Commit(), "SetArchived")
}

// cascadeArchiveToVariants applies a template's archive change to each of its variants.
//
// What to do with any one variant is services.ShouldSkipCascade and services.CascadeArchiveFields;
// this walks and writes. The writes go through the variant engine's repository, since the rows
// being updated are variants.
func (this *ProductAppServiceImpl) cascadeArchiveToVariants(
	ctx corectx.Context, variantEngine drif.DynamicResourceEngine, templateId string, archive bool,
) error {
	if templateId == "" {
		return nil
	}

	repo := variantEngine.ResourceRepository()
	variants, err := models.FindTemplateVariants(ctx, repo, templateId, services.MaxCascadeVariants)
	if err != nil {
		return errors.Wrap(err, "cascadeArchiveToVariants")
	}

	for _, item := range variants {
		variant := models.NewProductVariantFrom(item)
		if services.ShouldSkipCascade(variant, archive) {
			continue
		}

		update := services.CascadeArchiveFields(derefString(variant.GetId()), archive)
		if _, err := repo.Update(ctx, update); err != nil {
			return errors.Wrap(err, "cascadeArchiveToVariants")
		}
	}
	return nil
}

// readArchivedFlag reads the archive direction out of the action params, reporting whether it was
// present at all. Absent is not the same as false: it means the caller sent no flag, which the
// command schema rejects, so the two cases must stay distinguishable here.
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
