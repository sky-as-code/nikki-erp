package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ProductVariantExtService is Sales' port onto Inventory's product master (BR 69, BR 3.2).
//
// It answers ONE question — may this variant be sold? — and deliberately nothing else. The narrower
// alternative was tempting: hand back the variant and let Sales read its fields. That is wrong for
// the same reason the payment-method port refuses it (D-05a): selectability is not derivable from
// any single field. Inventory combines archival with lifecycle status, and either alone withdraws a
// variant from new business. Re-deriving that rule here would mean two implementations of it, and
// the day Inventory adds a third condition, Sales would keep selling something it should not.
//
// # Why Sales must ask at all
//
// BR 69 forbids selling a discontinued or archived variant, and until now nothing enforced it: an
// order could name any variant id and the line would be written. The failure is quiet and expensive
// — a customer is charged for something the business has withdrawn, and it surfaces at fulfilment,
// after the money has been taken.
//
// # What is NOT here
//
// No search, no listing, no way to read a price. A price comes from Sales' own pricelists (BR 16),
// and a port that could read Inventory's would invite somebody to sell at it — which is precisely
// the coupling the pricelist exists to avoid.
type ProductVariantExtService interface {
	// AssertSellable answers whether each of the given variants may be sold.
	//
	// A BATCH, because an order is validated as a whole: asking per line would cost one round trip
	// per line, and a basket of twenty would pay for it on every reprice. The result names the
	// variants that failed rather than returning a bare false, since an operator told only "this
	// order is invalid" cannot fix it.
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
	// NotSellable maps a variant id to why it was refused. EMPTY means every variant may be sold,
	// which is the common case — an empty map rather than a boolean because the caller needs the
	// reasons, and a separate `AllSellable` flag would be a second thing to keep in step with this.
	NotSellable map[string]string
}

// The reasons a variant may not be sold. Re-exported as constants so Sales names them without
// importing inventory outside infra/external.
const (
	// ReasonVariantNotFound: the id names nothing. Distinct from "withdrawn" because it usually
	// means a bad reference rather than a business decision, and an operator chases the two
	// differently.
	ReasonVariantNotFound = "sales_order_line.variant_not_found"

	// ReasonVariantNotSellable: the variant exists and Inventory has withdrawn it from new
	// business, by archiving it or by its lifecycle status. Sales does not distinguish which,
	// because the remedy is the same and the distinction is Inventory's to explain.
	ReasonVariantNotSellable = "sales_order_line.variant_not_sellable"
)
