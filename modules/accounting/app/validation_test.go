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

// There is no server-clock fallback: a request without a date must be refused rather than priced
// against whatever is effective at processing time.
func TestTaxDateIsMandatory(t *testing.T) {
	request := validRequest()
	request.TaxDate = ""

	cErrs := validateCalculationRequest(request)
	if cErrs == nil {
		t.Fatal("expected a request without a tax date to be refused")
	}
}

func TestTaxDateMustBeWellFormed(t *testing.T) {
	// Includes forms time.Parse accepts but that do not sort correctly against the stored
	// YYYY-MM-DD bounds the effective-period lookup compares.
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

// V1 is sale-only: treating a purchase as a sale would apply output-VAT rules to an input-VAT
// transaction.
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

// Document-scoped rounding is allocated by line reference, so a duplicate silently overwrites
// another line's rounding.
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

// The reason is checked before the entitlement, so holding the permission does not excuse it.
func TestOverrideWithoutReasonIsRefused(t *testing.T) {
	request := validRequest()
	request.Lines[0].OverrideTaxIds = []string{"tax-1"}

	svc := &TaxCalculationApplicationServiceImpl{}
	if cErrs := svc.assertOverrideAllowed(nil, request); cErrs == nil {
		t.Fatal("expected an override without a reason to be refused")
	}
}

// A request that overrides nothing must not require the elevated entitlement.
func TestNoOverrideSkipsTheEntitlementCheck(t *testing.T) {
	svc := &TaxCalculationApplicationServiceImpl{}
	// The nil context is safe only because the entitlement check must not be reached; a panic here
	// means the short-circuit has been lost.
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
