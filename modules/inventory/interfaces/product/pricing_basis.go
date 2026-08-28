package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The inputs a pricing engine needs from a product, and nothing else.
//
// This is a purpose-built read rather than a general one, and the narrowness is the point. Every
// other product port here answers "what is this variant" — its name, its stock, whether it may be
// sold. This one answers a different question: "what numbers may a price be computed FROM".
//
// The four values below are exactly what the pricing change request names as pricing inputs
// (sections 14 and 18): the base to fall back to, the cost a FORMULA rule may derive from, and the
// template and category ancestry a targeted rule is matched against. Nothing here can be used to
// manage master data, and none of it is the resolved selling price — that comes only from Sales'
// own pricelists, which is the whole reason those exist.
//
// It is a BATCH read. An order is priced as a whole, so asking per line would cost one round trip
// per line and a basket of twenty would pay for it on every reprice.

type GetPricingBasisQuery struct {
	ProductVariantIds []string
}

type GetPricingBasisResultData struct {
	// Bases is keyed by variant id. A variant that does not exist is simply absent rather than
	// present-and-empty, so a caller can tell "no such variant" from "a variant priced at zero" —
	// which matters because zero is a legitimate price and a legitimate cost.
	Bases map[string]PricingBasis
}

type GetPricingBasisResult = dyn.OpResult[GetPricingBasisResultData]

// PricingBasis is one variant's pricing inputs.
type PricingBasis struct {
	ProductVariantId  string
	ProductTemplateId string

	// CategoryPath runs from the variant's OWN category outward to the root.
	//
	// The order IS the precedence: a pricing rule on the nearest category wins over one on an
	// ancestor. Resolved here rather than by the consumer because walking parent_category_id is a
	// read per level, and Inventory can do it once for a whole batch while a consumer would pay
	// per line — and would need Inventory's category table to do it at all.
	CategoryPath []string

	// EffectiveBaseSalesPrice is the template's base price plus this variant's attribute price
	// extras — the value BR-PRICE-VARIANT-003 requires a rule to mean by BASE_SALES_PRICE. Not the
	// template's raw base price, which would ignore what the variant adds.
	EffectiveBaseSalesPrice string

	// Cost is the variant's current cost, for a FORMULA rule based on COST. A consumer READS it
	// and never writes it back (PRICE-INV-010).
	//
	// A string, like the price above, because both are decimals: carrying them as float64 would
	// lose precision on exactly the values that must not lose it. HasCost distinguishes an
	// unconfigured cost from a cost of zero, which the number alone cannot.
	Cost    string
	HasCost bool
}

// ProductPricingBasisService publishes the pricing inputs.
//
// Separate from ProductVariantDomainService deliberately. That port is a general variant reader and
// a consumer holding it can already see most of a product; this one exists so a pricing consumer
// can be given the four values it needs WITHOUT being given the general reader. The narrower grant
// is the reason for the extra interface.
type ProductPricingBasisService interface {
	GetPricingBasis(ctx corectx.Context, query GetPricingBasisQuery) (*GetPricingBasisResult, error)
}
