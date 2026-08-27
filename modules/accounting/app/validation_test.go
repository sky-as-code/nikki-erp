package app

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

func validRequest() it.CalculationRequest {
	return it.CalculationRequest{
		OperationType: models.OperationSale,
		TaxDate:       "2026-08-25",
		CurrencyCode:  "VND",
		Lines: []it.CalculationLine{{
			LineReference:        "L1",
			CommercialBaseAmount: decimal.NewFromInt(100),
		}},
	}
}

func TestValidRequestPasses(t *testing.T) {
	if cErrs := validateCalculationRequest(validRequest()); cErrs != nil {
		t.Fatalf("expected a valid request to pass, got %v", cErrs.ToError())
	}
}

// BR-TAX-ESS-SUP-020 deleted the server-clock fallback. A request without a date must be refused
// rather than priced against whatever is effective at the moment it happens to be processed.
func TestTaxDateIsMandatory(t *testing.T) {
	request := validRequest()
	request.TaxDate = ""

	cErrs := validateCalculationRequest(request)
	if cErrs == nil {
		t.Fatal("expected a request without a tax date to be refused")
	}
}

func TestTaxDateMustBeWellFormed(t *testing.T) {
	// Deliberately includes forms time.Parse would accept but that do not sort correctly against
	// the stored YYYY-MM-DD bounds, which is what the effective-period lookup compares against.
	for _, malformed := range []string{"2026-8-25", "25/08/2026", "2026-08-25T00:00:00Z", "not-a-date", "2026-13-01"} {
		request := validRequest()
		request.TaxDate = malformed

		if cErrs := validateCalculationRequest(request); cErrs == nil {
			t.Errorf("expected tax date %q to be refused", malformed)
		}
	}
}

func TestWellFormedDatesAreAccepted(t *testing.T) {
	for _, wellFormed := range []string{"2026-08-25", "2025-01-01", "2026-12-31"} {
		if !models.IsWellFormedDate(wellFormed) {
			t.Errorf("expected %q to be accepted", wellFormed)
		}
	}
}

func TestCurrencyIsMandatory(t *testing.T) {
	request := validRequest()
	request.CurrencyCode = ""

	if cErrs := validateCalculationRequest(request); cErrs == nil {
		t.Fatal("expected a request without a currency to be refused")
	}
}

func TestAtLeastOneLineIsRequired(t *testing.T) {
	request := validRequest()
	request.Lines = nil

	if cErrs := validateCalculationRequest(request); cErrs == nil {
		t.Fatal("expected a request with no lines to be refused")
	}
}

// V1 defines sale semantics only. A purchase must be refused rather than silently treated as a
// sale, which would apply output-VAT rules to an input-VAT transaction.
func TestPurchaseOperationsAreRefused(t *testing.T) {
	for _, operation := range []models.OperationType{models.OperationPurchase, models.OperationPurchaseRefund} {
		request := validRequest()
		request.OperationType = operation

		if cErrs := validateCalculationRequest(request); cErrs == nil {
			t.Errorf("expected operation %q to be refused in V1", operation)
		}
	}
}

func TestSaleOperationsAreAccepted(t *testing.T) {
	for _, operation := range []models.OperationType{models.OperationSale, models.OperationSaleRefund} {
		request := validRequest()
		request.OperationType = operation

		if cErrs := validateCalculationRequest(request); cErrs != nil {
			t.Errorf("expected operation %q to be accepted, got %v", operation, cErrs.ToError())
		}
	}
}

func TestUnknownOperationIsRefused(t *testing.T) {
	request := validRequest()
	request.OperationType = models.OperationType("barter")

	if cErrs := validateCalculationRequest(request); cErrs == nil {
		t.Fatal("expected an unknown operation type to be refused")
	}
}

func TestLineReferenceIsMandatory(t *testing.T) {
	request := validRequest()
	request.Lines[0].LineReference = ""

	if cErrs := validateCalculationRequest(request); cErrs == nil {
		t.Fatal("expected a line without a reference to be refused")
	}
}

// Document-scoped rounding is allocated by line reference, so a duplicate would silently overwrite
// another line's rounding rather than merely being confusing.
func TestDuplicateLineReferencesAreRefused(t *testing.T) {
	request := validRequest()
	request.Lines = append(request.Lines, it.CalculationLine{
		LineReference:        "L1",
		CommercialBaseAmount: decimal.NewFromInt(50),
	})

	if cErrs := validateCalculationRequest(request); cErrs == nil {
		t.Fatal("expected a duplicated line reference to be refused")
	}
}

// The reason is mandatory whenever an override is used, and is checked before the entitlement so
// that a caller who holds the permission still cannot override without justifying it.
func TestOverrideWithoutReasonIsRefused(t *testing.T) {
	request := validRequest()
	request.Lines[0].OverrideTaxIds = []string{"tax-1"}

	svc := &TaxCalculationApplicationServiceImpl{}
	if cErrs := svc.assertOverrideAllowed(nil, request); cErrs == nil {
		t.Fatal("expected an override without a reason to be refused")
	}
}

// A request that overrides nothing must not require the elevated entitlement, or every ordinary
// calculation would demand a permission only a tax administrator holds.
func TestNoOverrideSkipsTheEntitlementCheck(t *testing.T) {
	svc := &TaxCalculationApplicationServiceImpl{}
	// Passing a nil context is safe precisely because the entitlement check must not be reached.
	// If this ever panics, the short-circuit has been lost.
	if cErrs := svc.assertOverrideAllowed(nil, validRequest()); cErrs != nil {
		t.Fatalf("expected no override to skip the check, got %v", cErrs.ToError())
	}
}

func TestReversalRequiresDateAndComponents(t *testing.T) {
	if cErrs := validateReversal("", 1); cErrs == nil {
		t.Error("expected a reversal without a tax date to be refused")
	}
	if cErrs := validateReversal("2026-8-25", 1); cErrs == nil {
		t.Error("expected a malformed reversal tax date to be refused")
	}
	if cErrs := validateReversal("2026-08-25", 0); cErrs == nil {
		t.Error("expected a reversal with no components to be refused")
	}
	if cErrs := validateReversal("2026-08-25", 1); cErrs != nil {
		t.Errorf("expected a well-formed reversal to pass, got %v", cErrs.ToError())
	}
}
