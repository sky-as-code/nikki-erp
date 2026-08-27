package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	c "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

// TaxCalculationApplicationServiceImpl authorizes a calculation and delegates it.
//
// It adds exactly two things to the domain service: the permission check, and the request
// validation that has to happen before any configuration is read. Everything else is the domain's,
// which is what lets the engine be tested exhaustively without a caller identity.
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
	// An override substitutes the tax on a live transaction, which is a materially different power
	// from being allowed to price one. It is checked separately and only when actually used, so
	// that the ordinary path does not require the elevated entitlement (BR-TAX-ESS-053).
	if cErrs := this.assertOverrideAllowed(ctx, request); cErrs != nil {
		return &it.CalculateResult{ClientErrors: *cErrs}, nil
	}

	return this.calculationSvc.Calculate(ctx, request)
}

func (this *TaxCalculationApplicationServiceImpl) Simulate(
	ctx corectx.Context, request it.CalculationRequest,
) (*it.SimulateResult, error) {
	// The simulator has its own entitlement rather than reusing read: it explains the whole rule
	// base, including rules the caller's own transactions never hit, and that is a broader
	// disclosure than pricing one order.
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

// assertOverrideAllowed gates a manual tax substitution.
//
// Two conditions, both required by BR-TAX-ESS-SUP-023: the dedicated entitlement, and a written
// reason. The reason is mandatory because an override that nobody has to justify is indistinguishable
// from a mistake when an auditor reads it three years later.
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

// validateCalculationRequest refuses a request the engine could not answer meaningfully.
//
// These are the checks that must happen before any configuration is read, because each of them
// would otherwise produce a confidently wrong answer rather than an error.
func validateCalculationRequest(request it.CalculationRequest) *ft.ClientErrors {
	// The tax date is never defaulted from the server clock. BR-TAX-ESS-SUP-020 deleted that
	// fallback: a request that forgot the date would otherwise be priced against whatever was
	// effective at the moment it happened to be processed, rather than the date the sale legally
	// occurred — a difference that only shows up around a rate change, which is exactly when it
	// matters most.
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

	// V1 defines sale semantics only. The purchase values exist in the enum as a reserved contract
	// and are refused here rather than silently treated as sales, which would apply output-VAT
	// rules to an input-VAT transaction.
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
		// Results are keyed by line reference, and document-scoped rounding is allocated against
		// that key. A duplicate would silently overwrite another line's rounding.
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
