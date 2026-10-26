package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The inputs a pricing engine needs from a product, and nothing else: nothing here can manage
// master data, and none of it is the resolved selling price, which comes only from Sales'
// pricelists. Batched deliberately — a per-line read would cost one round trip per order line.

type GetPricingBasisQuery struct {
	ProductVariantIds []string
}

type GetPricingBasisResultData struct {
	// Bases is keyed by variant id. A missing variant is absent rather than present-and-empty, so
	// a caller can tell "no such variant" from "priced at zero" — zero is a legitimate price.
	Bases map[string]PricingBasis
}

type GetPricingBasisResult = dyn.OpResult[GetPricingBasisResultData]

// PricingBasis is one variant's pricing inputs.
type PricingBasis struct {
	ProductVariantId  string
	ProductTemplateId string

	// CategoryPath runs from the variant's OWN category outward to the root. The order IS the
	// precedence: a rule on the nearest category wins over one on an ancestor. Resolved here
	// because walking parent_category_id costs a read per level and Inventory can batch it.
	CategoryPath []string

	// EffectiveBaseSalesPrice is the template's base price plus this variant's attribute price
	// extras — what a rule means by BASE_SALES_PRICE, not the template's raw base price.
	EffectiveBaseSalesPrice string

	// Cost is the variant's current cost, for a FORMULA rule based on COST. Consumers read it and
	// never write it back. A decimal string, like the price above, because float64 would lose
	// precision. HasCost distinguishes an unconfigured cost from a cost of zero.
	Cost    string
	HasCost bool
}

// ProductPricingBasisService publishes the pricing inputs. Kept separate from
// ProductVariantDomainService so a pricing consumer gets these values without the general
// variant reader; the narrower grant is the reason for the extra interface.
type ProductPricingBasisService interface {
	GetPricingBasis(ctx corectx.Context, query GetPricingBasisQuery) (*GetPricingBasisResult, error)
}
