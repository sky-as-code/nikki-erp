package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ProductPricingBasisExtService is Sales' port onto the product values a price may be computed
// FROM.
//
// It sits beside ProductVariantExtService rather than inside it, and the separation is deliberate.
// That port answers one question — may this variant be sold? — and its documentation promises no
// way to read a price, because a port that could read Inventory's prices would invite somebody to
// sell at one. This port answers a different question, and merging the two would mean every
// consumer that needed only sellability was handed cost and price as well.
//
// # What this exposes, and why it is not the thing that port refuses
//
// Four values, all named by the pricing change request as pricing INPUTS:
//
//   - the product's template and category ancestry, which a targeted rule is matched against
//     (sections 12 and 18);
//   - the variant's effective base sales price, which is what a rule means by BASE_SALES_PRICE and
//     what pricing falls back to when no rule matches (section 18 step 9, BR-PRICE-VARIANT-003);
//   - the variant's cost, which a FORMULA rule may derive from (section 14).
//
// None of them is a selling price. A resolved selling price still comes only from Sales' own
// pricelists, which is exactly the coupling ProductVariantExtService exists to prevent, and
// nothing here weakens it: the base sales price is the price BEFORE any commercial policy, which
// is why Sales is the module that turns it into one.
//
// # What Sales must never do with cost
//
// Read it. Sales does not calculate a product's cost and does not write it back (PRICE-INV-010,
// AC-PRICE-007): the number belongs to Inventory, and a pricing engine that could change it would
// quietly have become a costing engine. This port is read-only, which is what makes that
// structural rather than a matter of discipline.
type ProductPricingBasisExtService interface {
	// GetPricingBasis reads the pricing inputs for a batch of variants.
	//
	// A BATCH, because an order is priced as a whole: asking per line would cost one round trip
	// per line, and an order is repriced on every edit.
	//
	// A variant that does not exist is absent from the result rather than present and empty, so a
	// caller can tell "no such variant" from "a variant priced at zero" — and zero is a legitimate
	// price for a giveaway.
	GetPricingBasis(
		ctx corectx.Context, query GetPricingBasisQuery,
	) (*GetPricingBasisResult, error)
}

type GetPricingBasisQuery = itProduct.GetPricingBasisQuery
type GetPricingBasisResult = itProduct.GetPricingBasisResult
type GetPricingBasisResultData = itProduct.GetPricingBasisResultData
type PricingBasis = itProduct.PricingBasis
