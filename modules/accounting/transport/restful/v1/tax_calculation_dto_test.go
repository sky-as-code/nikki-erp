package v1

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

// Money must survive the wire exactly. A JSON number would be parsed as a float64 by most clients,
// and a float64 cannot hold a decimal fraction — which on a tax figure means an invoice total that
// disagrees with the sum of its lines.
func TestAmountsBindFromJsonStringsExactly(t *testing.T) {
	body := `{
		"operation_type": "sale",
		"tax_date": "2026-08-25",
		"currency_code": "VND",
		"lines": [{
			"line_reference": "L1",
			"quantity": "3",
			"unit_price": "19.99",
			"commercial_base_amount": "59.97"
		}]
	}`

	var request CalculateTaxRequest
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatalf("failed to bind the request: %v", err)
	}

	command := request.ToCommand()
	if len(command.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(command.Lines))
	}
	if !command.Lines[0].CommercialBaseAmount.Equal(decimal.RequireFromString("59.97")) {
		t.Errorf("base = %s, want exactly 59.97", command.Lines[0].CommercialBaseAmount)
	}
	if !command.Lines[0].UnitPrice.Equal(decimal.RequireFromString("19.99")) {
		t.Errorf("unit price = %s, want exactly 19.99", command.Lines[0].UnitPrice)
	}
	if command.OperationType != models.OperationSale {
		t.Errorf("operation = %q, want sale", command.OperationType)
	}
	if command.TaxDate != "2026-08-25" {
		t.Errorf("tax date = %q, want 2026-08-25", command.TaxDate)
	}
}

// A value with more precision than float64 can represent must round-trip unchanged. If this ever
// fails, a float has crept into the path.
func TestHighPrecisionAmountSurvivesBinding(t *testing.T) {
	body := `{"lines":[{"line_reference":"L1","commercial_base_amount":"12345678901234.123456789"}]}`

	var request CalculateTaxRequest
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatalf("failed to bind: %v", err)
	}

	got := request.ToCommand().Lines[0].CommercialBaseAmount.String()
	if got != "12345678901234.123456789" {
		t.Fatalf("precision was lost in binding: got %s", got)
	}
}

func TestResponseEmitsAmountsAsStrings(t *testing.T) {
	result := it.CalculationResult{
		Status:        models.DeterminationResolved,
		TotalExcluded: decimal.RequireFromString("100.50"),
		TotalTax:      decimal.RequireFromString("10.05"),
		TotalIncluded: decimal.RequireFromString("110.55"),
		Lines: []it.LineResult{{
			LineReference: "L1",
			Status:        models.DeterminationResolved,
			TotalTax:      decimal.RequireFromString("10.05"),
			Components: []it.ComponentResult{{
				TaxId:     "vat10",
				TaxCode:   "VN_VAT_10",
				Rate:      decimal.RequireFromString("10"),
				TaxAmount: decimal.RequireFromString("10.05"),
			}},
		}},
	}

	encoded, err := json.Marshal(NewCalculateTaxResponse(result))
	if err != nil {
		t.Fatalf("failed to encode: %v", err)
	}
	rendered := string(encoded)

	// Quoted, so a client parsing this into a decimal type gets the exact value.
	if !strings.Contains(rendered, `"total_tax":"10.05"`) {
		t.Errorf("expected total_tax as a quoted string, got %s", rendered)
	}
	if !strings.Contains(rendered, `"total_included":"110.55"`) {
		t.Errorf("expected total_included as a quoted string, got %s", rendered)
	}
	// A bare number anywhere in a money field would mean a float crept in.
	if strings.Contains(rendered, `"total_tax":10.05`) {
		t.Error("total_tax was emitted as a JSON number, which loses precision")
	}
}

func TestComponentResponseCarriesTheAuditTrail(t *testing.T) {
	component := it.ComponentResult{
		TaxId:                  "vat10",
		TaxCode:                "VN_VAT_10",
		TaxName:                "VAT 10%",
		TaxDefinitionVersionId: "dv-1",
		TaxDefinitionVersionNo: 2,
		TaxRateVersionId:       "rv-1",
		TaxRateVersionNo:       3,
		Rate:                   decimal.RequireFromString("10"),
		TaxableBase:            decimal.RequireFromString("100"),
		UnroundedTaxAmount:     decimal.RequireFromString("10.004"),
		TaxAmount:              decimal.RequireFromString("10"),
		RoundingAdjustment:     decimal.RequireFromString("-0.004"),
		LegalReference:         "Luat Thue GTGT",
	}

	response := newTaxComponentResponse(component)

	// "tax = 10%" is not enough to issue a VAT invoice or answer an auditor: the version that
	// produced the number and the base it was computed on both have to travel with it.
	if response.TaxDefinitionVersionId != "dv-1" || response.TaxDefinitionVersionNo != 2 {
		t.Error("expected the definition version in the response")
	}
	if response.TaxRateVersionId != "rv-1" || response.TaxRateVersionNo != 3 {
		t.Error("expected the rate version in the response")
	}
	if response.TaxableBase != "100" {
		t.Errorf("taxable base = %q, want 100", response.TaxableBase)
	}
	// Both figures are carried: a refund reverses the rounded amount actually charged, while the
	// difference between them is what the rounding adjustment accounts for.
	if response.UnroundedTaxAmount != "10.004" || response.TaxAmount != "10" {
		t.Errorf("expected both the unrounded and rounded amounts, got %q and %q",
			response.UnroundedTaxAmount, response.TaxAmount)
	}
	if response.LegalReference != "Luat Thue GTGT" {
		t.Error("expected the legal reference in the response")
	}
}

func TestReversalRequestBindsAmountsAsStrings(t *testing.T) {
	body := `{
		"tax_date": "2026-09-01",
		"rounding_policy_code": "VN_DEFAULT",
		"original_snapshot": {"currency_code": "VND", "rounding_method": "half_up", "rounding_scope": "line"},
		"components": [{
			"original_component_reference": "C1",
			"original_reversible_basis": "100",
			"original_tax_amount": "10",
			"requested_reversal_basis": "40",
			"is_final_reversal": false
		}]
	}`

	var request ReversePartialRequest
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatalf("failed to bind: %v", err)
	}

	command := request.ToCommand()
	if command.RoundingPolicyCode != "VN_DEFAULT" {
		t.Errorf("policy code = %q", command.RoundingPolicyCode)
	}
	if len(command.Components) != 1 || command.Components[0].RequestedReversalBasis != "40" {
		t.Fatalf("expected the reversal basis through as a string, got %+v", command.Components)
	}
	if command.OriginalSnapshot.RoundingMethod != models.RoundingHalfUp {
		t.Errorf("rounding method = %q, want half_up", command.OriginalSnapshot.RoundingMethod)
	}
	if command.OriginalSnapshot.RoundingScope != models.RoundingScopeLine {
		t.Errorf("rounding scope = %q, want line", command.OriginalSnapshot.RoundingScope)
	}
}

// A snapshot comes back from the caller's own storage, so its enum strings may have been written by
// an older version or corrupted. An unrecognized value must not travel on as if it were meaningful.
func TestUnknownSnapshotEnumsCoerceToEmpty(t *testing.T) {
	snapshot := TaxSnapshotRequest{
		RoundingMethod: "banker's rounding, probably",
		RoundingScope:  "whatever",
	}

	converted := snapshot.toSnapshot()

	if converted.RoundingMethod != "" {
		t.Errorf("expected an unknown rounding method to coerce to empty, got %q", converted.RoundingMethod)
	}
	if converted.RoundingScope != "" {
		t.Errorf("expected an unknown rounding scope to coerce to empty, got %q", converted.RoundingScope)
	}
}

func TestKnownSnapshotEnumsSurvive(t *testing.T) {
	for _, method := range []models.RoundingMethod{
		models.RoundingHalfUp, models.RoundingHalfEven, models.RoundingUp, models.RoundingDown,
	} {
		if got := roundingMethodOf(string(method)); got != method {
			t.Errorf("roundingMethodOf(%q) = %q", method, got)
		}
	}
	for _, scope := range []models.RoundingScope{models.RoundingScopeLine, models.RoundingScopeDocument} {
		if got := roundingScopeOf(string(scope)); got != scope {
			t.Errorf("roundingScopeOf(%q) = %q", scope, got)
		}
	}
}

func TestSimulateResponseCarriesTheTrace(t *testing.T) {
	result := it.SimulationResult{
		Calculation: it.CalculationResult{Status: models.DeterminationResolved},
		Trace: []it.TraceStep{
			{Stage: "determination", Detail: "matched rule R1", RuleIds: []string{"R1"}, TaxIds: []string{"vat10"}},
			{Stage: "rounding", Detail: "line scope"},
		},
	}

	response := NewSimulateTaxResponse(result)

	if len(response.Trace) != 2 {
		t.Fatalf("expected 2 trace steps, got %d", len(response.Trace))
	}
	if response.Trace[0].Stage != "determination" || len(response.Trace[0].RuleIds) != 1 {
		t.Errorf("expected the determination step with its rule ids, got %+v", response.Trace[0])
	}
}
