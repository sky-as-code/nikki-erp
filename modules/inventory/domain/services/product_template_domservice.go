package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Product Template business rules. Each takes the narrowest dependency it needs — a
// models.ProductSearcher for a read, plain values otherwise — so it runs without building an
// engine.

// MaxCascadeVariants bounds how many variants one archive operation will touch: more is beyond what
// a synchronous request should carry, and archiving only some would leave the cascade half-applied.
const MaxCascadeVariants = 1000

// ShouldSkipCascade decides whether a template's archive change leaves this variant alone: the
// variant is already in the target state, or the template is being unarchived and this variant was
// archived for its own reason. Unarchiving restores only what the template's cascade took down.
func ShouldSkipCascade(variant *models.ProductVariant, archive bool) bool {
	wasArchived := variant.IsArchived() != nil && *variant.IsArchived()
	if wasArchived == archive {
		return true
	}
	if archive {
		return false
	}

	// Only variants the template's own archive took down come back. A nil source counts as "not a
	// cascade", so a variant archived before archive_source existed stays archived: resurrecting a
	// deliberately withdrawn product is worse than leaving one a user can restore by hand.
	source := variant.GetArchiveSource()
	return source == nil || *source != models.ArchiveSourceTemplateCascade
}

// CascadeArchiveFields is the field delta a cascading archive writes onto one variant. Archiving
// stamps archive_source so a later unarchive can tell its own cascade from a deliberate archive;
// unarchiving clears it.
func CascadeArchiveFields(variantId string, archive bool) dmodel.DynamicFields {
	update := dmodel.DynamicFields{
		models.ProductVariantFieldId: variantId,
		basemodel.FieldIsArchived:    archive,
	}
	if archive {
		update[models.ProductVariantFieldArchiveSource] = models.ArchiveSourceTemplateCascade.String()
	} else {
		update[models.ProductVariantFieldArchiveSource] = nil
	}
	return update
}

// AssertTemplateDeletable blocks a hard delete once the template owns any variant: a variant may
// already be referenced by a transaction, so a template with variants is archived, not deleted.
func AssertTemplateDeletable(
	ctx corectx.Context, repo models.ProductSearcher, templateId string, vErrs *ft.ClientErrors,
) error {
	if templateId == "" {
		return nil
	}

	variants, err := models.FindTemplateVariants(ctx, repo, templateId, 1)
	if err != nil {
		return errors.Wrap(err, "AssertTemplateDeletable")
	}
	if len(variants) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductTemplateFieldId,
			"product_template.has_variants",
			"this template still has variants; archive it instead of deleting it"))
	}
	return nil
}
