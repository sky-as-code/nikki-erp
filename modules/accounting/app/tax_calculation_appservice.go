package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	c "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

// TaxCalculationApplicationServiceImpl adds two things to the domain service: the permission check,
// and the request validation that must happen before any configuration is read.
type TaxCalculationApplicationServiceImpl struct {
	calculationSvc it.TaxCalculationDomainService
}

func NewTaxCalculationApplicationServiceImpl(
	calculationSvc it.TaxCalculationDomainService,
) it.TaxCalculationAppService {
	return &TaxCalculationApplicationServiceImpl{calculationSvc: calculationSvc}
}

func (this *TaxCalculationApplicationServiceImpl) Calculate(
	ctx corectx.Context, request it.CalculationRequest,
) (*it.CalculateResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionRead,
		models.TaxSchemaName, c.ResourceScopeOrg); cErrs != nil {
		return &it.CalculateResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := validateCalculationRequest(request); cErrs != nil {
		return &it.CalculateResult{ClientErrors: *cErrs}, nil
	}
	// An override is a different power from pricing a transaction, so it is checked separately and
	// only when used, keeping the ordinary path free of the elevated entitlement.
	if cErrs := this.assertOverrideAllowed(ctx, request); cErrs != nil {
		return &it.CalculateResult{ClientErrors: *cErrs}, nil
	}

	return this.calculationSvc.Calculate(ctx, request)
}

func (this *TaxCalculationApplicationServiceImpl) Simulate(
	ctx corectx.Context, request it.CalculationRequest,
) (*it.SimulateResult, error) {
	// The simulator has its own entitlement rather than reusing read: it discloses the whole rule
	// base, not just the rules one order hits.
	if cErrs := assertPermission(ctx, c.ActionSimulate,
		models.TaxSchemaName, c.ResourceScopeOrg); cErrs != nil {
		return &it.SimulateResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := validateCalculationRequest(request); cErrs != nil {
		return &it.SimulateResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := this.assertOverrideAllowed(ctx, request); cErrs != nil {
		return &it.SimulateResult{ClientErrors: *cErrs}, nil
	}

	return this.calculationSvc.Simulate(ctx, request)
}

func (this *TaxCalculationApplicationServiceImpl) ReverseFull(
	ctx corectx.Context, request it.FullReversalRequest,
) (*it.ReverseResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionRead,
		models.TaxSchemaName, c.ResourceScopeOrg); cErrs != nil {
		return &it.ReverseResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := validateReversal(request.TaxDate, len(request.Components)); cErrs != nil {
		return &it.ReverseResult{ClientErrors: *cErrs}, nil
	}

	return this.calculationSvc.ReverseFull(ctx, request)
}

func (this *TaxCalculationApplicationServiceImpl) ReversePartial(
	ctx corectx.Context, request it.PartialReversalRequest,
) (*it.ReverseResult, error) {
	if cErrs := assertPermission(ctx, drif.PermissionRead,
		models.TaxSchemaName, c.ResourceScopeOrg); cErrs != nil {
		return &it.ReverseResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := validateReversal(request.TaxDate, len(request.Components)); cErrs != nil {
		return &it.ReverseResult{ClientErrors: *cErrs}, nil
	}

	return this.calculationSvc.ReversePartial(ctx, request)
}

// assertOverrideAllowed gates a manual tax substitution on two conditions: the dedicated
// entitlement and a written reason, so an audit can tell an override from a mistake.
func (this *TaxCalculationApplicationServiceImpl) assertOverrideAllowed(
	ctx corectx.Context, request it.CalculationRequest,
) *ft.ClientErrors {
	overridden := false
	for _, line := range request.Lines {
		if len(line.OverrideTaxIds) > 0 {
			overridden = true
			if line.OverrideReason == "" {
				return rejection(models.TaxSchemaName, "tax.override_reason_required",
					"overriding the determined tax on line '"+line.LineReference+"' requires a reason")
			}
		}
	}
	if !overridden {
		return nil
	}
	return assertPermission(ctx, c.ActionOverride, models.TaxSchemaName, c.ResourceScopeOrg)
}

// validateCalculationRequest refuses a request the engine could not answer meaningfully. These
// checks must run before any configuration is read, or each yields a confidently wrong answer.
func validateCalculationRequest(request it.CalculationRequest) *ft.ClientErrors {
	// The tax date is never defaulted from the server clock: a missing date would otherwise price
	// against whatever is effective at processing time rather than the date of the sale, which
	// diverges exactly around a rate change.
	if request.TaxDate == "" {
		return rejection(models.TaxSchemaName, "tax.tax_date_required",
			"a tax calculation requires the tax date; it is never taken from the server clock")
	}
	if !models.IsWellFormedDate(request.TaxDate) {
		return rejection(models.TaxSchemaName, "tax.tax_date_malformed",
			"the tax date must be an ISO date of the form YYYY-MM-DD")
	}
	if request.CurrencyCode == "" {
		return rejection(models.TaxSchemaName, "tax.currency_required",
			"a tax calculation requires the document currency")
	}
	if len(request.Lines) == 0 {
		return rejection(models.TaxSchemaName, "tax.lines_required",
			"a tax calculation requires at least one line")
	}

	// V1 defines sale semantics only. The purchase enum values are a reserved contract and are
	// refused here; treating them as sales would apply output-VAT rules to an input-VAT
	// transaction.
	if !models.IsImplementedOperation(request.OperationType) {
		return rejection(models.TaxSchemaName, "tax.operation_type_unsupported",
			"only 'sale' and 'sale_refund' are supported in this version")
	}

	seen := map[string]bool{}
	for _, line := range request.Lines {
		if line.LineReference == "" {
			return rejection(models.TaxSchemaName, "tax.line_reference_required",
				"every line requires a reference so its result can be matched back")
		}
		// Document-scoped rounding is allocated by line reference, so a duplicate would silently
		// overwrite another line's rounding.
		if seen[line.LineReference] {
			return rejection(models.TaxSchemaName, "tax.line_reference_duplicated",
				"line reference '"+line.LineReference+"' appears more than once")
		}
		seen[line.LineReference] = true
	}
	return nil
}

func validateReversal(taxDate string, componentCount int) *ft.ClientErrors {
	if taxDate == "" {
		return rejection(models.TaxSchemaName, "tax.tax_date_required",
			"a reversal requires the refund's own tax date")
	}
	if !models.IsWellFormedDate(taxDate) {
		return rejection(models.TaxSchemaName, "tax.tax_date_malformed",
			"the tax date must be an ISO date of the form YYYY-MM-DD")
	}
	if componentCount == 0 {
		return rejection(models.TaxSchemaName, "tax.reversal_components_required",
			"a reversal requires at least one original component to reverse")
	}
	return nil
}

func rejection(field string, key string, message string) *ft.ClientErrors {
	cErrs := ft.NewClientErrors()
	cErrs.Append(*ft.NewBusinessViolation(field, key, message))
	return cErrs
}
