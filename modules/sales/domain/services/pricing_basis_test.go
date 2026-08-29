package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// applyPricingBasis re-reads the product's base price and cost onto the pricing input. Most cases
// here pin the fail-soft decision: an unavailable Inventory leaves the stored price standing rather
// than refusing to reprice, so a till keeps working when another module is slow.

type stubBasis struct {
	bases map[string]itProduct.PricingBasis
	err   error
}

func (this *stubBasis) GetPricingBasis(
	_ corectx.Context, _ itExt.GetPricingBasisQuery,
) (*itExt.GetPricingBasisResult, error) {
	if this.err != nil {
		return nil, this.err
	}
	return &itExt.GetPricingBasisResult{
		Data:    itExt.GetPricingBasisResultData{Bases: this.bases},
		HasData: true,
	}, nil
}

func basisLine() pricing.LineInput {
	return pricing.LineInput{
		Key:                "L1",
		ProductVariantId:   "VAR1",
		CatalogueUnitPrice: mustDecimal("100"),
	}
}

func mustDecimal(text string) decimal.Decimal {
	return decimal.RequireFromString(text)
}

func TestPricingBasisFillsTheLine(t *testing.T) {
	lines := []pricing.LineInput{basisLine()}
	svc := &stubBasis{bases: map[string]itProduct.PricingBasis{
		"VAR1": {
			ProductVariantId:        "VAR1",
			ProductTemplateId:       "TPL1",
			CategoryPath:            []string{"CAT_NEAR", "CAT_FAR"},
			EffectiveBaseSalesPrice: "130",
			Cost:                    "60",
			HasCost:                 true,
		},
	}}

	applyPricingBasis(nil, lines, svc)

	if lines[0].ProductTemplateId != "TPL1" {
		t.Fatalf("the template must be filled for template-targeted rules; got %q",
			lines[0].ProductTemplateId)
	}
	if len(lines[0].CategoryPath) != 2 || lines[0].CategoryPath[0] != "CAT_NEAR" {
		t.Fatalf("the category path must arrive nearest-first; got %v", lines[0].CategoryPath)
	}
	if !lines[0].CatalogueUnitPrice.Equal(mustDecimal("130")) {
		t.Fatalf("the base price must be re-read from the product, not left at the stored 100; got %s",
			lines[0].CatalogueUnitPrice)
	}
	if !lines[0].HasCost || !lines[0].UnitCost.Equal(mustDecimal("60")) {
		t.Fatalf("cost must be filled for a COST formula; got %s (has=%v)",
			lines[0].UnitCost, lines[0].HasCost)
	}
}

// A nil port is a supported deployment, not an error; the line keeps the price it was stored with.
func TestNilPortLeavesTheStoredPrice(t *testing.T) {
	lines := []pricing.LineInput{basisLine()}

	applyPricingBasis(nil, lines, nil)

	if !lines[0].CatalogueUnitPrice.Equal(mustDecimal("100")) {
		t.Fatalf("a nil port must leave the stored price alone; got %s", lines[0].CatalogueUnitPrice)
	}
}

// A failing read must not zero the price, which would give the goods away.
func TestFailedReadLeavesTheStoredPrice(t *testing.T) {
	lines := []pricing.LineInput{basisLine()}

	applyPricingBasis(nil, lines, &stubBasis{err: errors.New("inventory unavailable")})

	if !lines[0].CatalogueUnitPrice.Equal(mustDecimal("100")) {
		t.Fatalf("a failed read must leave the stored price alone; got %s",
			lines[0].CatalogueUnitPrice)
	}
}

// A variant absent from the answer — deleted, or not visible — is left as it was: it is still in the
// basket and the order still has to total.
func TestUnknownVariantIsLeftAlone(t *testing.T) {
	lines := []pricing.LineInput{basisLine()}

	applyPricingBasis(nil, lines, &stubBasis{bases: map[string]itProduct.PricingBasis{}})

	if !lines[0].CatalogueUnitPrice.Equal(mustDecimal("100")) {
		t.Fatalf("an unknown variant must keep its stored price; got %s",
			lines[0].CatalogueUnitPrice)
	}
}

// HasCost false must stay false, so a COST formula declines rather than pricing at zero — zero is a
// legitimate cost for a giveaway.
func TestAbsentCostIsNotReadAsZero(t *testing.T) {
	lines := []pricing.LineInput{basisLine()}
	svc := &stubBasis{bases: map[string]itProduct.PricingBasis{
		"VAR1": {ProductVariantId: "VAR1", EffectiveBaseSalesPrice: "130"},
	}}

	applyPricingBasis(nil, lines, svc)

	if lines[0].HasCost {
		t.Fatal("an unconfigured cost must not be reported as a cost of zero")
	}
}

// A malformed decimal is treated as absent rather than as zero, for the same reason.
func TestMalformedPriceLeavesTheStoredPrice(t *testing.T) {
	lines := []pricing.LineInput{basisLine()}
	svc := &stubBasis{bases: map[string]itProduct.PricingBasis{
		"VAR1": {ProductVariantId: "VAR1", EffectiveBaseSalesPrice: "not-a-number"},
	}}

	applyPricingBasis(nil, lines, svc)

	if !lines[0].CatalogueUnitPrice.Equal(mustDecimal("100")) {
		t.Fatalf("an unparseable price must not become zero; got %s", lines[0].CatalogueUnitPrice)
	}
}
