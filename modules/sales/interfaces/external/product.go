package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ProductVariantExtService answers one question — may this variant be sold? — and nothing else. It
// does not hand back the variant for Sales to judge, because selectability is not derivable from any
// single field: inventory combines archival with lifecycle status, and re-deriving the rule here
// would mean Sales keeps selling something inventory has withdrawn the day a third condition is
// added. Selling a discontinued or archived variant otherwise fails quietly, surfacing at fulfilment
// after the money is taken.
//
// There is no search, no listing and no way to read a price: a price comes from Sales' own
// pricelists, and a port that could read inventory's would invite somebody to sell at it.
type ProductVariantExtService interface {
	// AssertSellable answers whether each of the given variants may be sold. A batch, because an
	// order is validated as a whole and repriced often. The result names the variants that failed
	// rather than returning a bare false, since "this order is invalid" is not actionable.
	AssertSellable(
		ctx corectx.Context, query AssertSellableQuery,
	) (*AssertSellableResult, error)
}

// AssertSellableQuery names the variants to check.
type AssertSellableQuery struct {
	ProductVariantIds []string
}

// AssertSellableResult reports what may not be sold.
type AssertSellableResult struct {
	// NotSellable maps a variant id to why it was refused; empty means every variant may be sold.
	NotSellable map[string]string
}

// The reasons a variant may not be sold. Re-exported as constants so Sales names them without
// importing inventory outside infra/external.
const (
	// ReasonVariantNotFound: the id names nothing. Distinct from "withdrawn" because it usually
	// means a bad reference rather than a business decision.
	ReasonVariantNotFound = "sales_order_line.variant_not_found"

	// ReasonVariantNotSellable: the variant exists and inventory has withdrawn it from new business,
	// by archiving or by lifecycle status. Sales does not distinguish which; the remedy is the same.
	ReasonVariantNotSellable = "sales_order_line.variant_not_sellable"
)
