package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The rule-consistency checks (sections 12–14).
//
// Every case here is a row that would otherwise be ACCEPTED and then silently never match anything
// at resolution time. That is the failure worth preventing: a rule that never matches is invisible,
// and whoever wrote it sees a price that did not change with nothing to explain why.

func keyOf(record dmodel.DynamicFields) string {
	vErrs := assertRuleConsistent(record)
	if vErrs == nil || vErrs.Count() == 0 {
		return ""
	}
	return (*vErrs)[0].Key
}

func fixedPriceRow() dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesPricelistItemFieldSalesPricelistId:  "PL1",
		models.SalesPricelistItemFieldAppliesTo:         models.PricelistAppliesToVariant,
		models.SalesPricelistItemFieldProductVariantId:  "VAR1",
		models.SalesPricelistItemFieldUomId:             "UOM1",
		models.SalesPricelistItemFieldPrice:             "100",
		models.SalesPricelistItemFieldCalculationMethod: models.PricelistMethodFixedPrice,
	}
}

func TestAWellFormedFixedPriceRuleIsAccepted(t *testing.T) {
	if key := keyOf(fixedPriceRow()); key != "" {
		t.Fatalf("a complete fixed-price rule must be accepted; got %q", key)
	}
}

func TestVariantRuleWithoutATargetIsRefused(t *testing.T) {
	row := fixedPriceRow()
	delete(row, models.SalesPricelistItemFieldProductVariantId)

	if key := keyOf(row); key != "sales_pricelist_item.target_required" {
		t.Fatalf("a variant rule naming no variant matches nothing; got %q", key)
	}
}

// The surplus direction matters as much as the missing one: a variant rule that also carries a
// category id reads, to anyone browsing the table, as though it priced the category too.
func TestVariantRuleWithASurplusTargetIsRefused(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldProductCategoryId] = "CAT1"

	if key := keyOf(row); key != "sales_pricelist_item.target_not_allowed" {
		t.Fatalf("a variant rule must not also name a category; got %q", key)
	}
}

func TestAllProductsRuleWithATargetIsRefused(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldAppliesTo] = models.PricelistAppliesToAllProducts

	if key := keyOf(row); key != "sales_pricelist_item.target_not_allowed" {
		t.Fatalf("a rule for everything must name nothing; got %q", key)
	}
}

func TestAllProductsRuleWithNoTargetIsAccepted(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldAppliesTo] = models.PricelistAppliesToAllProducts
	delete(row, models.SalesPricelistItemFieldProductVariantId)

	if key := keyOf(row); key != "" {
		t.Fatalf("ALL_PRODUCTS with no target is the correct shape; got %q", key)
	}
}

// A row written before targeting existed carries no applies_to and named a variant. It must keep
// meaning that rather than being rejected or silently widened.
func TestMissingAppliesToIsTreatedAsVariant(t *testing.T) {
	row := fixedPriceRow()
	delete(row, models.SalesPricelistItemFieldAppliesTo)

	if key := keyOf(row); key != "" {
		t.Fatalf("a legacy variant rule must still be valid; got %q", key)
	}
}

// Zero is a real price — a giveaway — so it is the ABSENCE of the field that is refused, never its
// value.
func TestFixedPriceOfZeroIsAccepted(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldPrice] = "0"

	if key := keyOf(row); key != "" {
		t.Fatalf("a giveaway is a legitimate fixed price; got %q", key)
	}
}

func TestFixedPriceWithoutAPriceIsRefused(t *testing.T) {
	row := fixedPriceRow()
	delete(row, models.SalesPricelistItemFieldPrice)

	if key := keyOf(row); key != "sales_pricelist_item.price_required" {
		t.Fatalf("a fixed-price rule with no price would quote nothing; got %q", key)
	}
}

// A fixed price is per a unit, so a rule without one cannot be applied to any line.
func TestFixedPriceWithoutAUnitIsRefused(t *testing.T) {
	row := fixedPriceRow()
	delete(row, models.SalesPricelistItemFieldUomId)

	if key := keyOf(row); key != "sales_pricelist_item.uom_required" {
		t.Fatalf("a fixed price needs the unit it is per; got %q", key)
	}
}

// A discount carries no unit of its own — it adjusts a base already in the line's unit — so the
// unit requirement must NOT apply to it.
func TestDiscountRuleNeedsNoUnit(t *testing.T) {
	row := dmodel.DynamicFields{
		models.SalesPricelistItemFieldAppliesTo:         models.PricelistAppliesToVariant,
		models.SalesPricelistItemFieldProductVariantId:  "VAR1",
		models.SalesPricelistItemFieldCalculationMethod: models.PricelistMethodDiscount,
		models.SalesPricelistItemFieldDiscountPercent:   "10",
	}

	if key := keyOf(row); key != "" {
		t.Fatalf("a percentage discount is unit-agnostic; got %q", key)
	}
}

func TestDiscountRuleWithoutAPercentageIsRefused(t *testing.T) {
	row := dmodel.DynamicFields{
		models.SalesPricelistItemFieldAppliesTo:         models.PricelistAppliesToVariant,
		models.SalesPricelistItemFieldProductVariantId:  "VAR1",
		models.SalesPricelistItemFieldCalculationMethod: models.PricelistMethodDiscount,
	}

	if key := keyOf(row); key != "sales_pricelist_item.discount_required" {
		t.Fatalf("a discount rule must state its percentage; got %q", key)
	}
}

func TestFormulaWithoutABaseSourceIsRefused(t *testing.T) {
	row := dmodel.DynamicFields{
		models.SalesPricelistItemFieldAppliesTo:         models.PricelistAppliesToVariant,
		models.SalesPricelistItemFieldProductVariantId:  "VAR1",
		models.SalesPricelistItemFieldCalculationMethod: models.PricelistMethodFormula,
	}

	if key := keyOf(row); key != "sales_pricelist_item.base_source_required" {
		t.Fatalf("a formula must say what it prices from; got %q", key)
	}
}

func TestFormulaOnAnotherPricelistMustNameIt(t *testing.T) {
	row := dmodel.DynamicFields{
		models.SalesPricelistItemFieldAppliesTo:         models.PricelistAppliesToVariant,
		models.SalesPricelistItemFieldProductVariantId:  "VAR1",
		models.SalesPricelistItemFieldCalculationMethod: models.PricelistMethodFormula,
		models.SalesPricelistItemFieldBasePriceSource:   models.PricelistBaseSourceOtherPricelist,
	}

	if key := keyOf(row); key != "sales_pricelist_item.base_pricelist_required" {
		t.Fatalf("OTHER_PRICELIST must name the pricelist; got %q", key)
	}
}

// A COST formula needs no further operand: the cost comes from the product at resolution time.
func TestFormulaOnCostNeedsNothingElse(t *testing.T) {
	row := dmodel.DynamicFields{
		models.SalesPricelistItemFieldAppliesTo:         models.PricelistAppliesToVariant,
		models.SalesPricelistItemFieldProductVariantId:  "VAR1",
		models.SalesPricelistItemFieldCalculationMethod: models.PricelistMethodFormula,
		models.SalesPricelistItemFieldBasePriceSource:   models.PricelistBaseSourceCost,
	}

	if key := keyOf(row); key != "" {
		t.Fatalf("a cost formula is complete on its own; got %q", key)
	}
}

// An inverted window makes the rule match on no date at all — it would never apply, and nothing
// would say so.
func TestInvertedValidityWindowIsRefused(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldValidFrom] = "2026-12-01T00:00:00Z"
	row[models.SalesPricelistItemFieldValidTo] = "2026-01-01T00:00:00Z"

	if key := keyOf(row); key != "sales_pricelist_item.validity_inverted" {
		t.Fatalf("a window that closes before it opens must be refused; got %q", key)
	}
}

func TestOpenEndedValidityIsAccepted(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldValidFrom] = "2026-01-01T00:00:00Z"

	if key := keyOf(row); key != "" {
		t.Fatalf("an open-ended window is the normal case; got %q", key)
	}
}

func TestUnknownMethodIsRefused(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldCalculationMethod] = "MAGIC"

	if key := keyOf(row); key != "sales_pricelist_item.method_unknown" {
		t.Fatalf("an unrecognised method must be refused; got %q", key)
	}
}

func TestUnknownAppliesToIsRefused(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldAppliesTo] = "PRODUCT_COLOUR"

	if key := keyOf(row); key != "sales_pricelist_item.applies_to_unknown" {
		t.Fatalf("an unrecognised target must be refused; got %q", key)
	}
}

// Two surplus targets must be reported in a stable order.
//
// The check iterates a slice rather than a map for exactly this reason: a map would report them in
// whichever order it happened to iterate, so the same bad row would yield different errors on
// different runs — and a client that shows "the first problem" would show a different one each
// time, which makes a bug report impossible to act on.
func TestSurplusTargetsAreReportedDeterministically(t *testing.T) {
	row := fixedPriceRow()
	row[models.SalesPricelistItemFieldAppliesTo] = models.PricelistAppliesToAllProducts
	row[models.SalesPricelistItemFieldProductTemplateId] = "TPL1"
	row[models.SalesPricelistItemFieldProductCategoryId] = "CAT1"

	first := assertRuleConsistent(row)
	if first == nil || first.Count() < 3 {
		t.Fatalf("all three surplus targets must be reported; got %v", first)
	}

	// Repeated because the failure this guards against is intermittent by nature.
	for attempt := 0; attempt < 20; attempt++ {
		again := assertRuleConsistent(row)
		for index := range *first {
			if (*again)[index].Field != (*first)[index].Field {
				t.Fatalf("error order changed between runs: %q then %q",
					(*first)[index].Field, (*again)[index].Field)
			}
		}
	}
}
