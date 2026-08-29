package external

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"

	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// An adapter rather than a direct hand-over: inventory publishes a general reader, and handing it
// through would let a Sales caller read a variant's price — a price somebody eventually sells at,
// which Sales' own pricelists exist to prevent. The selectability rule stays inventory's, so this
// calls IsSelectable() rather than reading `is_archived` and `status` itself.

type productVariantAdapter struct {
	variants itProduct.ProductVariantDomainService
}

// AssertSellable answers whether each variant may be sold. It reads one variant per id — N round
// trips for an N-line order — because a batched search cannot distinguish "not found" from "found
// but withdrawn", and those need different remedies. If this ever hurts, the fix is a batched
// exists-and-selectable query upstream, not a Sales-side guess.
func (this *productVariantAdapter) AssertSellable(
	ctx corectx.Context, query itExt.AssertSellableQuery,
) (*itExt.AssertSellableResult, error) {
	result := &itExt.AssertSellableResult{NotSellable: map[string]string{}}

	// Deduplicated: an order may name the same variant on several lines, and a repeat check costs
	// a round trip for an answer already known.
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
