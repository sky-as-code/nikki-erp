package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// "Could any of these places supply any of these items" — the question a caller offering a customer
// a shortlist of kiosks is really asking, answered in one scan rather than one per variant.
//
// It is ADVISORY and takes no lock, which is the whole reason it can be cheap. The numbers are true
// of the instant they were read and of no instant after it: another sale may claim the last unit
// before the customer finishes choosing. Only a reservation secures stock, so nothing here returns a
// reference a caller could mistake for a hold, and a caller must never treat a positive answer as a
// guarantee that a fulfillment will succeed.

// AvailableByLocations totals free stock for each (location, variant) pair that has any, keyed
// location first. A pair with no quant row is simply absent, and callers read a missing entry as
// zero rather than as "unknown": a location holding none of something has nothing to hold.
func (this *StockQuantDomainServiceImpl) AvailableByLocations(
	ctx corectx.Context, query itStock.LocationAvailabilityQuery,
) (map[string]map[string]decimal.Decimal, error) {
	if len(query.LocationIds) == 0 || len(query.VariantIds) == 0 {
		return map[string]map[string]decimal.Decimal{}, nil
	}

	engine, err := repoFor(models.StockQuantSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.StockQuantFieldLocationId, dmodel.In, anyValuesOf(query.LocationIds)...),
		*dmodel.NewSearchNode().NewCondition(
			models.StockQuantFieldProductVariantId, dmodel.In, anyValuesOf(query.VariantIds)...),
	)

	available := map[string]map[string]decimal.Decimal{}
	for page := 0; page < maxSummaryQuantPages; page++ {
		found, err := engine.Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  summaryScanPageSize,
		})
		if err != nil {
			return nil, errors.Wrap(err, "AvailableByLocations")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}

		for _, row := range found.Data.Items {
			quant := models.NewStockQuantFrom(row)
			locationId := derefId(quant.GetLocationId())
			variantId := derefId(quant.GetProductVariantId())
			if locationId == "" || variantId == "" {
				continue
			}

			// Free stock, not on-hand: goods already promised to another sale are not available to
			// this one, and reporting them would offer a customer a kiosk that cannot serve them.
			free := derefDecimal(quant.GetOnHandQuantity()).
				Sub(derefDecimal(quant.GetReservedQuantity()))

			byVariant, seen := available[locationId]
			if !seen {
				byVariant = map[string]decimal.Decimal{}
				available[locationId] = byVariant
			}
			// Summed rather than assigned: one location commonly holds several quants of the same
			// variant, one per lot, package and owner combination.
			byVariant[variantId] = byVariant[variantId].Add(free)
		}

		if len(found.Data.Items) < summaryScanPageSize {
			break
		}
	}

	// An over-reserved quant reads as negative free stock, which is a data problem rather than a
	// debt the next caller should inherit; clamping keeps it from cancelling out a sibling quant
	// that genuinely has goods.
	for _, byVariant := range available {
		for variantId, free := range byVariant {
			if free.IsNegative() {
				byVariant[variantId] = decimal.Zero
			}
		}
	}
	return available, nil
}

func anyValuesOf(values []string) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}
