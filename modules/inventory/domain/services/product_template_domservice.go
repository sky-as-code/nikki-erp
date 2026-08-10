package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Product Template business rules. They live here rather than in dynamicengines, which
// declares engines and wires callbacks but owns no rules of its own.
//
// Each rule takes the narrowest dependency it needs — a models.ProductSearcher for a read, plain
// values otherwise — so it can be exercised without building an engine.

// MaxCascadeVariants bounds how many variants one archive operation will touch. A template with
// more than this many variants is beyond what a synchronous request should carry, and silently
// archiving only some of them would leave the cascade half-applied.
const MaxCascadeVariants = 1000

// ShouldSkipCascade decides whether a template's archive change leaves this variant alone.
//
// Two reasons to skip: the variant is already in the target state, or the template is being
// unarchived and this variant was archived for its own reason. Unarchiving must restore only what
// the template's own archive took down. See BR-PROD-TPL-003 and BR §8.9.
func ShouldSkipCascade(variant *models.ProductVariant, archive bool) bool {
	wasArchived := variant.IsArchived() != nil && *variant.IsArchived()
	if wasArchived == archive {
		return true
	}
	if archive {
		return false
	}

	// Unarchiving. Only variants the template's own archive took down come back; a variant a
	// user archived deliberately stays archived.
	//
	// A nil source is treated as "not a cascade", so a variant archived before archive_source
	// existed stays archived too. Wrongly resurrecting a deliberately withdrawn product is a
	// worse failure than leaving one archived that a user can restore by hand.
	source := variant.GetArchiveSource()
	return source == nil || *source != models.ArchiveSourceTemplateCascade
}

// CascadeArchiveFields is the field delta a cascading archive writes onto one variant.
//
// Archiving stamps archive_source so that a later unarchive can tell its own cascade apart from
// a deliberate archive; unarchiving clears the stamp again. Returning the delta rather than
// performing the write keeps the rule free of the repository.
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

// AssertTemplateDeletable blocks a hard delete once the template owns any variant.
//
// A variant may already be referenced by a transaction, and that reference check belongs to the
// variant's own delete guard, so the safe rule here is that a template with variants is archived
// instead of deleted. See BR-PROD-TPL-005 and AC-PROD-021.
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
