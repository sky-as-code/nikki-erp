package services

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	taxsvc "github.com/sky-as-code/nikki-erp/modules/accounting/domain/services/tax"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

func TestContextCarriesTheWhitelistedFacts(t *testing.T) {
	request := it.CalculationRequest{
		OperationType:          models.OperationSale,
		TaxDate:                "2026-08-25",
		CurrencyCode:           "VND",
		ShipFromJurisdictionId: "jur-from",
		ShipToJurisdictionId:   "jur-to",
		BusinessChannelCode:    "pos",
		Seller: it.TaxPartyContext{
			PrimaryJurisdictionId: "jur-seller",
			TaxRegistrations:      []it.TaxRegistration{{IsRegistered: true}},
		},
		Buyer: it.TaxPartyContext{
			PartyTaxClassification: "business",
			PrimaryJurisdictionId:  "jur-buyer",
		},
	}
	line := it.CalculationLine{
		LineReference:            "L1",
		ProductReference:         "prod-1",
		ProductTaxClassification: "standard",
		CommercialBaseAmount:     decimal.NewFromInt(1000),
		CandidateTaxIds:          []string{"tax-1", "tax-2"},
	}

	context := buildContext(request, line)

	expected := map[string]string{
		models.CtxOperationType:            "sale",
		models.CtxTaxDate:                  "2026-08-25",
		models.CtxCurrencyCode:             "VND",
		models.CtxProductTaxClassification: "standard",
		models.CtxPartyTaxClassification:   "business",
		models.CtxSellerJurisdictionId:     "jur-seller",
		models.CtxBuyerJurisdictionId:      "jur-buyer",
		models.CtxShipFromJurisdictionId:   "jur-from",
		models.CtxShipToJurisdictionId:     "jur-to",
		models.CtxSellerIsTaxRegistered:    "true",
		models.CtxBuyerIsTaxRegistered:     "false",
		models.CtxCommercialBaseAmount:     "1000",
		models.CtxBusinessChannelCode:      "pos",
		models.CtxProductReference:         "prod-1",
		models.CtxCandidateTaxId:           "tax-1",
	}
	for key, want := range expected {
		if got := context[key]; got != want {
			t.Errorf("context[%q] = %q, want %q", key, got, want)
		}
	}
}

// The whitelist is closed by BR-TAX-ESS-SUP-022: a rule may test these fields and nothing else, so
// a context key outside the declared set would be a rule depending on something no one guaranteed.
func TestContextCarriesNothingOutsideTheWhitelist(t *testing.T) {
	context := buildContext(
		it.CalculationRequest{OperationType: models.OperationSale, TaxDate: "2026-08-25"},
		it.CalculationLine{LineReference: "L1"},
	)

	for key := range context {
		if !models.IsKnownContextField(key) {
			t.Errorf("context carries %q, which is not a whitelisted testable field", key)
		}
	}
}

func TestRegistrationAnywhereCountsAsRegistered(t *testing.T) {
	party := it.TaxPartyContext{TaxRegistrations: []it.TaxRegistration{
		{JurisdictionId: "a", IsRegistered: false},
		{JurisdictionId: "b", IsRegistered: true},
	}}
	if !isRegisteredAnywhere(party) {
		t.Fatal("expected a party registered in any jurisdiction to count as registered")
	}

	none := it.TaxPartyContext{TaxRegistrations: []it.TaxRegistration{{IsRegistered: false}}}
	if isRegisteredAnywhere(none) {
		t.Fatal("expected a party with no active registration to count as unregistered")
	}
	if isRegisteredAnywhere(it.TaxPartyContext{}) {
		t.Fatal("expected a party with no registrations at all to count as unregistered")
	}
}

// Two components of different taxes must never pool for rounding, because the per-tax document
// total is exactly what a VAT return reports.
func TestGroupKeySeparatesDistinctTaxes(t *testing.T) {
	vat10 := ResolvedTax{TaxId: "vat10", RateVersionId: "rv1", Treatment: models.TaxTreatmentTaxable, JurisdictionId: "vn"}
	vat8 := ResolvedTax{TaxId: "vat8", RateVersionId: "rv2", Treatment: models.TaxTreatmentTaxable, JurisdictionId: "vn"}

	if groupKeyOf(vat10) == groupKeyOf(vat8) {
		t.Fatal("expected different taxes to fall in different rounding groups")
	}
	if groupKeyOf(vat10) != groupKeyOf(vat10) {
		t.Fatal("expected the group key to be stable")
	}
}

func TestGroupKeySeparatesTreatmentAndJurisdiction(t *testing.T) {
	base := ResolvedTax{TaxId: "t", RateVersionId: "rv", Treatment: models.TaxTreatmentTaxable, JurisdictionId: "vn"}

	differentTreatment := base
	differentTreatment.Treatment = models.TaxTreatmentZeroRated
	if groupKeyOf(base) == groupKeyOf(differentTreatment) {
		t.Error("expected a different treatment to separate the rounding group")
	}

	differentJurisdiction := base
	differentJurisdiction.JurisdictionId = "sg"
	if groupKeyOf(base) == groupKeyOf(differentJurisdiction) {
		t.Error("expected a different jurisdiction to separate the rounding group")
	}
}

// A group tax is a container for display and reporting, not a rate: calculating it as well as its
// parts would tax the line twice (BR-TAX-ESS-018).
func TestGroupTaxFlattensToItsComponentsOnly(t *testing.T) {
	group := ResolvedTax{
		TaxId:           "group",
		CalculationType: models.CalculationGroup,
		Components: []ResolvedTax{
			{TaxId: "child-a", CalculationType: models.CalculationPercentage, Sequence: 1},
			{TaxId: "child-b", CalculationType: models.CalculationPercentage, Sequence: 2},
		},
	}

	specs := ToComponentSpecs(group, nil)
	if len(specs) != 2 {
		t.Fatalf("expected the group to contribute its 2 children, got %d specs", len(specs))
	}
	for _, spec := range specs {
		if spec.TaxId == "group" {
			t.Fatal("expected the group itself to contribute no component of its own")
		}
	}

	leaves := FlattenResolved(group)
	if len(leaves) != 2 {
		t.Fatalf("expected 2 leaves, got %d", len(leaves))
	}
}

func TestNonGroupTaxIsItsOwnLeaf(t *testing.T) {
	tax := ResolvedTax{TaxId: "vat", CalculationType: models.CalculationPercentage}

	if specs := ToComponentSpecs(tax, nil); len(specs) != 1 || specs[0].TaxId != "vat" {
		t.Fatalf("expected a plain tax to yield exactly itself, got %+v", specs)
	}
	if leaves := FlattenResolved(tax); len(leaves) != 1 || leaves[0].TaxId != "vat" {
		t.Fatalf("expected a plain tax to be its own leaf, got %+v", leaves)
	}
}

func TestFixedTaxCarriesItsConvertedQuantity(t *testing.T) {
	tax := ResolvedTax{
		TaxId:           "excise",
		CalculationType: models.CalculationFixed,
		FixedAmount:     decimal.NewFromInt(4000),
	}
	quantities := map[string]decimal.Decimal{"excise": decimal.NewFromInt(3)}

	specs := ToComponentSpecs(tax, quantities)
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if !specs[0].Quantity.Equal(decimal.NewFromInt(3)) {
		t.Fatalf("expected the converted quantity to reach the calculator, got %s", specs[0].Quantity)
	}
}

// A document is only as resolved as its least resolved line. Reporting "resolved" while a line
// silently contributed no tax is how a caller stores a total that is quietly wrong.
func TestOneUnresolvedLineDowngradesTheDocument(t *testing.T) {
	result := it.CalculationResult{
		Status: models.DeterminationResolved,
		Lines: []it.LineResult{
			{LineReference: "L1", Status: models.DeterminationResolved,
				TotalExcluded: decimal.NewFromInt(100), TotalTax: decimal.NewFromInt(10),
				TotalIncluded: decimal.NewFromInt(110)},
			{LineReference: "L2", Status: models.DeterminationUnresolved,
				TotalExcluded: decimal.NewFromInt(50), TotalIncluded: decimal.NewFromInt(50)},
		},
	}

	totalDocument(&result)

	if result.Status != models.DeterminationUnresolved {
		t.Fatalf("expected the document to be unresolved, got %q", result.Status)
	}
	if !result.TotalTax.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected totals to still sum, got %s", result.TotalTax)
	}
}

// A lawful zero is not a failure: a no_tax_applicable line leaves the document resolved.
func TestNoTaxApplicableDoesNotDowngradeTheDocument(t *testing.T) {
	result := it.CalculationResult{
		Status: models.DeterminationResolved,
		Lines: []it.LineResult{{
			LineReference: "L1",
			Status:        models.DeterminationNoTaxApplicable,
			TotalExcluded: decimal.NewFromInt(100),
			TotalIncluded: decimal.NewFromInt(100),
		}},
	}

	totalDocument(&result)

	if result.Status != models.DeterminationResolved {
		t.Fatalf("expected an exempt line to leave the document resolved, got %q", result.Status)
	}
}

// The line total is the sum of the rounded components, not the rounding of the summed ones: an
// invoice shows the components, and a total that does not add up to what is printed beside it is
// the defect this ordering avoids.
func TestLineTotalIsTheSumOfRoundedComponents(t *testing.T) {
	line := it.LineResult{
		TotalExcluded: decimal.NewFromInt(100),
		Components: []it.ComponentResult{
			{TaxAmount: decimal.RequireFromString("5.01")},
			{TaxAmount: decimal.RequireFromString("4.99")},
		},
	}

	retotalLine(&line)

	if !line.TotalTax.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected the tax total to be 10, got %s", line.TotalTax)
	}
	if !line.TotalIncluded.Equal(decimal.NewFromInt(110)) {
		t.Fatalf("expected the included total to be 110, got %s", line.TotalIncluded)
	}
}

func TestUnresolvedDocumentEchoesEveryLine(t *testing.T) {
	request := it.CalculationRequest{Lines: []it.CalculationLine{
		{LineReference: "L1", CommercialBaseAmount: decimal.NewFromInt(100)},
		{LineReference: "L2", CommercialBaseAmount: decimal.NewFromInt(250)},
	}}

	result := unresolvedDocument(request, models.ErrCodeRoundingPolicyMissing)

	if result.Status != models.DeterminationUnresolved {
		t.Fatalf("expected an unresolved document, got %q", result.Status)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("expected every line echoed back, got %d", len(result.Lines))
	}
	for _, line := range result.Lines {
		if line.ErrorCode != models.ErrCodeRoundingPolicyMissing {
			t.Errorf("expected line %q to carry the error code, got %q", line.LineReference, line.ErrorCode)
		}
	}
	// No tax was determined, so the excluded total is the untaxed base and nothing was added.
	if !result.TotalExcluded.Equal(decimal.NewFromInt(350)) {
		t.Errorf("expected the untaxed base to total 350, got %s", result.TotalExcluded)
	}
	if !result.TotalTax.IsZero() {
		t.Errorf("expected no tax on an unresolved document, got %s", result.TotalTax)
	}
}

func TestAllocationKeyIsUniquePerLineAndSequence(t *testing.T) {
	if allocationKey("L1", 1) == allocationKey("L1", 2) {
		t.Error("expected different sequences to key differently")
	}
	if allocationKey("L1", 1) == allocationKey("L2", 1) {
		t.Error("expected different lines to key differently")
	}
	if allocationKey("L1", 1) != allocationKey("L1", 1) {
		t.Error("expected the allocation key to be stable")
	}
}

// Document-scoped rounding writes back to the component the allocation names, so a mismatched key
// would silently leave a component unrounded.
func TestDocumentRoundingWritesBackToItsComponent(t *testing.T) {
	result := it.CalculationResult{Lines: []it.LineResult{{
		LineReference: "L1",
		TotalExcluded: decimal.NewFromInt(100),
		Components: []it.ComponentResult{{
			Sequence:           1,
			TaxAmount:          decimal.RequireFromString("10.004"),
			UnroundedTaxAmount: decimal.RequireFromString("10.004"),
		}},
	}}}
	allocations := []taxsvc.AllocationInput{{
		LineReference:     "L1",
		ComponentSequence: 1,
		GroupKey:          "vat",
		Unrounded:         decimal.RequireFromString("10.004"),
	}}
	policy := taxsvc.RoundingPolicy{
		Scope:     models.RoundingScopeDocument,
		Method:    models.RoundingHalfUp,
		Increment: decimal.RequireFromString("0.01"),
	}

	applyDocumentRounding(&result, allocations, policy)

	component := result.Lines[0].Components[0]
	if !component.TaxAmount.Equal(decimal.RequireFromString("10")) {
		t.Fatalf("expected the component rounded to 10, got %s", component.TaxAmount)
	}
	if !result.Lines[0].TotalTax.Equal(component.TaxAmount) {
		t.Fatalf("expected the line to be retotalled from its rounded component, got %s",
			result.Lines[0].TotalTax)
	}
}

func TestSortedKeysIsDeterministic(t *testing.T) {
	set := map[string]bool{"c": true, "a": true, "b": true}

	keys := sortedKeys(set)

	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("expected sorted keys, got %v", keys)
	}
	if sortedKeys(map[string]bool{}) != nil {
		t.Fatal("expected an empty set to yield nil rather than an empty slice")
	}
}
