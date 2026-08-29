package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ProductPricingBasisExtService is Sales' port onto the product values a price may be computed FROM:
// template and category ancestry a targeted rule matches against, the variant's effective base sales
// price, and its cost, which a FORMULA rule may derive from. None is a selling price — a resolved
// selling price still comes only from Sales' own pricelists.
//
// It is separate from ProductVariantExtService so consumers needing only sellability are not handed
// cost and price too. The port is read-only: cost belongs to inventory, and a pricing engine that
// could write it back would have become a costing engine.
type ProductPricingBasisExtService interface {
	// GetPricingBasis reads the pricing inputs for a batch of variants — a batch because an order is
	// priced as a whole and repriced on every edit. A variant that does not exist is absent from the
	// result rather than present and empty, so a caller can tell "no such variant" from "priced at
	// zero", which is legitimate for a giveaway.
	GetPricingBasis(
		ctx corectx.Context, query GetPricingBasisQuery,
	) (*GetPricingBasisResult, error)
}

type GetPricingBasisQuery = itProduct.GetPricingBasisQuery
type GetPricingBasisResult = itProduct.GetPricingBasisResult
type GetPricingBasisResultData = itProduct.GetPricingBasisResultData
type PricingBasis = itProduct.PricingBasis
