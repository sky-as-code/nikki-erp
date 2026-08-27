package external

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"

	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The product-variant adapter (BR 69, SALES-048).
//
// An ADAPTER rather than a direct hand-over, unlike the tax and payment-method bindings. Inventory
// publishes a general reader — get a variant, search variants — and Sales needs one narrow judgement:
// may this be sold? Handing the reader straight through would let any Sales caller pull a variant's
// whole record, including the price, and a price read from Inventory is a price somebody will
// eventually sell at, which is exactly what Sales' own pricelists exist to prevent (BR 16).
//
// The selectability rule is INVENTORY'S, and stays there. This adapter calls IsSelectable() rather
// than reading `is_archived` and `status` itself, so that the day Inventory adds a third condition
// Sales inherits it instead of quietly disagreeing.

type productVariantAdapter struct {
	variants itProduct.ProductVariantDomainService
}

// AssertSellable answers whether each variant may be sold.
//
// Reads one variant per id rather than a single batched search. That is a real cost — N round trips
// for an N-line order — and the reason is that the batch alternative cannot distinguish "not found"
// from "found but withdrawn": a search returns what matched, and an id missing from the results
// could be either. Those two need different remedies, so the loop buys an answer an operator can act
// on. If order sizes ever make this hurt, the fix belongs upstream as a batched exists-and-selectable
// query, not as a Sales-side guess.
func (this *productVariantAdapter) AssertSellable(
	ctx corectx.Context, query itExt.AssertSellableQuery,
) (*itExt.AssertSellableResult, error) {
	result := &itExt.AssertSellableResult{NotSellable: map[string]string{}}

	// Deduplicated, because an order may name the same variant on several lines — two of the same
	// product bought separately — and checking it twice would cost a round trip to learn what the
	// first one already answered.
	seen := map[string]bool{}
	for _, variantId := range query.ProductVariantIds {
		if variantId == "" || seen[variantId] {
			continue
		}
		seen[variantId] = true

		found, err := this.variants.GetProductVariant(ctx, itProduct.GetProductVariantQuery{
			Id: model.Id(variantId),
		})
		if err != nil {
			return nil, errors.Wrapf(err, "reading product variant '%s'", variantId)
		}

		if found == nil || !found.HasData {
			result.NotSellable[variantId] = itExt.ReasonVariantNotFound
			continue
		}
		if !found.Data.IsSelectable() {
			result.NotSellable[variantId] = itExt.ReasonVariantNotSellable
		}
	}
	return result, nil
}
