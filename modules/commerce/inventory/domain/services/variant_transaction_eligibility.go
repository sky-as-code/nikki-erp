package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Transaction eligibility has two strengths. Starting new business (a reservation, a transfer, a
// move) needs a variant that is selectable today. Finishing or correcting what already exists
// (executing a drafted scrap, a correction movement, a consumed reservation) must keep working
// for a variant archived or discontinued since, or its stock could never be cleared; it only
// refuses a base, which is never a product and so can never have been transacted.

// EligibilityReason names why a variant cannot be transacted; "" means it can.
type EligibilityReason string

const (
	EligibilityReasonBase                 = EligibilityReason("base_variant")
	EligibilityReasonVariantArchived      = EligibilityReason("variant_archived")
	EligibilityReasonVariantDiscontinued  = EligibilityReason("variant_discontinued")
	EligibilityReasonTemplateArchived     = EligibilityReason("template_archived")
	EligibilityReasonTemplateDiscontinued = EligibilityReason("template_discontinued")
)

// VariantIneligibility decides eligibility for new business when forNewBusiness is true, and for
// completing existing business otherwise. template may be nil when only the kind matters.
func VariantIneligibility(
	variant *models.ProductVariant, template *models.ProductTemplate, forNewBusiness bool,
) EligibilityReason {
	if variantKindOf(variant) == models.VariantKindBase {
		return EligibilityReasonBase
	}
	if !forNewBusiness {
		return ""
	}
	if derefBool(variant.IsArchived()) {
		return EligibilityReasonVariantArchived
	}
	if status := variant.GetStatus(); status != nil && *status == models.ProductVariantStatusDiscontinued {
		return EligibilityReasonVariantDiscontinued
	}
	if template == nil {
		return ""
	}
	if derefBool(template.IsArchived()) {
		return EligibilityReasonTemplateArchived
	}
	if status := template.GetStatus(); status != nil && *status == models.ProductTemplateStatusDiscontinued {
		return EligibilityReasonTemplateDiscontinued
	}
	return ""
}

// assertVariantTransactable refuses starting new business with a variant that is not selectable.
// A missing variant is left to the caller's own not-found handling.
func assertVariantTransactable(ctx corectx.Context, variantId string, field string, vErrs *ft.ClientErrors) error {
	return assertVariantEligible(ctx, variantId, field, true, vErrs)
}

// assertVariantNotBase refuses only a base variant, for paths that finish or correct existing
// business.
func assertVariantNotBase(ctx corectx.Context, variantId string, field string, vErrs *ft.ClientErrors) error {
	return assertVariantEligible(ctx, variantId, field, false, vErrs)
}

func assertVariantEligible(
	ctx corectx.Context, variantId string, field string, forNewBusiness bool, vErrs *ft.ClientErrors,
) error {
	if variantId == "" {
		return nil
	}
	variant, template, err := loadVariantWithTemplate(ctx, variantId)
	if err != nil || variant == nil {
		return err
	}
	appendIneligibility(field, variantId, VariantIneligibility(variant, template, forNewBusiness), vErrs)
	return nil
}

func appendIneligibility(field string, variantId string, reason EligibilityReason, vErrs *ft.ClientErrors) {
	if reason == "" {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(field,
		ft.ErrorKey("product_variant.not_transaction_eligible", "inventory"),
		"product variant '"+variantId+"' cannot be transacted: "+string(reason)))
}
