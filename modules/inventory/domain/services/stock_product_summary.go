package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The Stock side of the Product integration contract, implemented on the quant service because
// that is where the balances already live — the same placement, and for the same reason, as
// GetLocationUsage in location_usage_read_service.go.
//
// Everything here reads. Product displays these numbers and stores none of them, so no method
// writes and none returns anything a caller could use to change a quantity (CR §4.1, §6.2).
//
// The rollups are done in Go, in one pass over one bounded search. That is not a preference: the
// dynamic-model layer has no aggregation — SearchGraph filters and pages, and there is no
// group-by — so a database-side GROUP BY is not available to us. Batching is what keeps this from
// becoming the N+1 the requirement forbids (CR §8.4, AC-PROD-INT-035): one search covers a whole
// page of variants, and the grouping happens in memory afterwards.

var _ itStock.StockProductSummaryReader = (*StockQuantDomainServiceImpl)(nil)

// summaryScanPageSize is how many quants are read at a time when summarising variants.
//
// It matches usageScanPageSize's reasoning: page to the end rather than stopping at the first
// page, because a partial total reported as a whole one is a wrong number presented as right.
const summaryScanPageSize = 200

// maxSummaryQuantPages bounds the scan so a pathological variant cannot turn one page render into
// an unbounded read. Hitting it sets Truncated rather than failing: a flagged partial is more
// useful than an error, and the caller is told not to trust the total.
const maxSummaryQuantPages = 50

// GetVariantSummaries resolves a batch of variants in a single pass.
//
// Every requested id appears in the result, zero-valued when the variant holds no stock, so a
// caller never has to tell "no stock" apart from "not returned".
func (this *StockQuantDomainServiceImpl) GetVariantSummaries(
	ctx corectx.Context, query itStock.GetVariantSummariesQuery,
) (*itStock.GetVariantSummariesResult, error) {
	variantIds := dedupeNonEmpty(query.VariantIds)
	if len(variantIds) > itStock.MaxSummaryVariants {
		variantIds = variantIds[:itStock.MaxSummaryVariants]
	}

	summaries := make(map[string]itStock.VariantStockSummary, len(variantIds))
	for _, id := range variantIds {
		summaries[id] = itStock.VariantStockSummary{}
	}
	if len(variantIds) == 0 {
		return &itStock.GetVariantSummariesResult{
			HasData: true,
			Data:    itStock.GetVariantSummariesResultData{Summaries: summaries},
		}, nil
	}

	if err := this.accumulateQuants(ctx, variantIds, summaries); err != nil {
		return nil, err
	}
	if err := this.accumulateMoves(ctx, variantIds, summaries); err != nil {
		return nil, err
	}

	for id, summary := range summaries {
		summary.Available = summary.OnHand.Sub(summary.Reserved)
		summary.Forecasted = summary.OnHand.Add(summary.Forecasted)
		summaries[id] = summary
	}

	return &itStock.GetVariantSummariesResult{
		HasData: true,
		Data:    itStock.GetVariantSummariesResultData{Summaries: summaries},
	}, nil
}

// accumulateQuants folds every quant of the requested variants into their summaries.
//
// One search for the whole batch, grouped in memory. The location and warehouse counts are built
// from distinct sets rather than a row count, since one variant commonly has several quants in the
// same location — different lots, packages or owners — and each must count once.
func (this *StockQuantDomainServiceImpl) accumulateQuants(
	ctx corectx.Context, variantIds []string, summaries map[string]itStock.VariantStockSummary,
) error {
	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.StockQuantFieldProductVariantId, dmodel.In, toAnySlice(variantIds)...),
	)

	locationsSeen := map[string]map[string]bool{}
	transitLocations, err := this.transitLocationIds(ctx)
	if err != nil {
		return err
	}

	truncated := false
	for page := 0; ; page++ {
		if page >= maxSummaryQuantPages {
			truncated = true
			break
		}

		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  summaryScanPageSize,
		})
		if err != nil {
			return errors.Wrap(err, "accumulateQuants")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}

		for _, row := range found.Data.Items {
			quant := models.NewStockQuantFrom(row)
			variantId := derefId(quant.GetProductVariantId())
			summary, tracked := summaries[variantId]
			if !tracked {
				continue
			}

			onHand := derefDecimal(quant.GetOnHandQuantity())
			summary.OnHand = summary.OnHand.Add(onHand)
			summary.Reserved = summary.Reserved.Add(derefDecimal(quant.GetReservedQuantity()))

			locationId := derefId(quant.GetLocationId())
			if transitLocations[locationId] {
				summary.InTransit = summary.InTransit.Add(onHand)
			}

			// A location counts as holding the variant only when something is actually
			// there. A zeroed quant is a place it used to be, which is not what
			// "Number of Locations" means to someone reading the product page.
			if !onHand.IsZero() && locationId != "" {
				if locationsSeen[variantId] == nil {
					locationsSeen[variantId] = map[string]bool{}
				}
				locationsSeen[variantId][locationId] = true
			}

			// Read off the field data: base_uom_id is declared on the quant schema but has
			// no generated accessor, since nothing inside Stock had needed it until now.
			if summary.BaseUomId == nil {
				summary.BaseUomId = quant.GetFieldData().GetModelId(models.StockQuantFieldBaseUomId)
			}
			summaries[variantId] = summary
		}

		if len(found.Data.Items) < summaryScanPageSize {
			break
		}
	}

	return this.fillLocationCounts(ctx, summaries, locationsSeen, truncated)
}

// fillLocationCounts turns the per-variant location sets into counts, and derives how many
// distinct warehouses those locations belong to.
//
// The warehouse count needs the locations' warehouse ids, which the quant does not carry, so the
// locations are read once for the whole batch rather than once per variant.
func (this *StockQuantDomainServiceImpl) fillLocationCounts(
	ctx corectx.Context,
	summaries map[string]itStock.VariantStockSummary,
	locationsSeen map[string]map[string]bool,
	truncated bool,
) error {
	allLocationIds := map[string]bool{}
	for _, locations := range locationsSeen {
		for locationId := range locations {
			allLocationIds[locationId] = true
		}
	}

	warehouseOf, err := this.warehouseIdsOfLocations(ctx, mapKeys(allLocationIds))
	if err != nil {
		return err
	}

	for variantId, locations := range locationsSeen {
		summary := summaries[variantId]
		summary.LocationCount = len(locations)

		warehouses := map[string]bool{}
		for locationId := range locations {
			// A location outside any warehouse — vendor, customer, transit, loss — has no
			// warehouse to count, and must not be folded into a single empty-string bucket.
			if warehouseId := warehouseOf[locationId]; warehouseId != "" {
				warehouses[warehouseId] = true
			}
		}
		summary.WarehouseCount = len(warehouses)
		summaries[variantId] = summary
	}

	if truncated {
		for id, summary := range summaries {
			summary.Truncated = true
			summaries[id] = summary
		}
	}
	return nil
}

// accumulateMoves folds movement into the forecast, and records the last time stock for each
// variant actually moved.
//
// Forecast is current on-hand plus confirmed incoming minus confirmed outgoing (BR §4.2.13.3).
// Draft moves are excluded: they are not commitments, and counting them would forecast stock
// nobody has agreed to move. The on-hand part is added by the caller, once, after this returns.
//
// Done moves are read in the same pass rather than a separate one. They contribute nothing to the
// forecast — they are already in on-hand — but they are what "Last Movement" means, and a second
// query per variant to find each one would be the N+1 this whole reader exists to avoid. The
// repository offers no sorting, so the latest is found by comparing as the rows go past.
func (this *StockQuantDomainServiceImpl) accumulateMoves(
	ctx corectx.Context, variantIds []string, summaries map[string]itStock.VariantStockSummary,
) error {
	engine, err := engineFor(models.StockMoveSchemaName)
	if err != nil {
		return err
	}

	internalLocations, err := this.internalLocationIds(ctx)
	if err != nil {
		return err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.StockMoveFieldProductVariantId, dmodel.In, toAnySlice(variantIds)...),
		*dmodel.NewSearchNode().NewCondition(
			models.StockMoveFieldStatus, dmodel.NotIn, forecastIgnoredMoveStatuses()),
	)

	for page := 0; page < maxSummaryQuantPages; page++ {
		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  summaryScanPageSize,
		})
		if err != nil {
			return errors.Wrap(err, "accumulateMoves")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}

		for _, row := range found.Data.Items {
			move := models.NewStockMoveFrom(row)
			variantId := derefId(move.GetProductVariantId())
			summary, tracked := summaries[variantId]
			if !tracked {
				continue
			}

			if derefString(move.GetStatus()) == models.StockMoveStatusDone {
				summary.LastMovementAt = laterOf(summary.LastMovementAt, move.GetUpdatedAt())
				summaries[variantId] = summary
				continue
			}

			quantity := derefDecimal(move.GetBaseDemandQuantity())
			source := derefId(move.GetSourceLocationId())
			destination := derefId(move.GetDestinationLocationId())

			// Direction is judged by which end is stock we hold. A move from a vendor into
			// our own location is incoming; the reverse is outgoing. A move between two of
			// our own locations changes no total and is deliberately neither.
			inbound := internalLocations[destination] && !internalLocations[source]
			outbound := internalLocations[source] && !internalLocations[destination]
			switch {
			case inbound:
				summary.Forecasted = summary.Forecasted.Add(quantity)
			case outbound:
				summary.Forecasted = summary.Forecasted.Sub(quantity)
			}

			summaries[variantId] = summary
		}

		if len(found.Data.Items) < summaryScanPageSize {
			break
		}
	}
	return nil
}

// GetTemplateSummary aggregates a template's variants into one summary, and reports the rows it
// was built from.
//
// The total is the sum of the variants and nothing else. A template has no quants of its own and
// must never acquire any: the aggregate exists for display, search and reporting, and may not be
// used to execute a stock operation (CR §5.2, §7, PROD-INT-INV-002/003, TS-PROD-02).
func (this *StockQuantDomainServiceImpl) GetTemplateSummary(
	ctx corectx.Context, query itStock.GetTemplateSummaryQuery,
) (*itStock.GetTemplateSummaryResult, error) {
	if query.TemplateId == "" {
		return nil, errors.New("GetTemplateSummary requires a template id")
	}

	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}

	rows, err := models.FindTemplateVariants(
		ctx, variantEngine.ResourceRepository(), query.TemplateId, itStock.MaxSummaryVariants)
	if err != nil {
		return nil, errors.Wrap(err, "GetTemplateSummary")
	}

	variantIds := make([]string, 0, len(rows))
	variantsById := make(map[string]models.ProductVariant, len(rows))
	for _, row := range rows {
		variant := models.NewProductVariantFrom(row)
		id := derefId(variant.GetId())
		if id == "" {
			continue
		}
		variantIds = append(variantIds, id)
		variantsById[id] = *variant
	}

	summaries, err := this.GetVariantSummaries(
		ctx, itStock.GetVariantSummariesQuery{VariantIds: variantIds})
	if err != nil {
		return nil, err
	}

	total := itStock.VariantStockSummary{}
	breakdown := make([]itStock.TemplateVariantStockRow, 0, len(variantIds))
	for _, id := range variantIds {
		summary := summaries.Data.Summaries[id]
		variant := variantsById[id]

		total.OnHand = total.OnHand.Add(summary.OnHand)
		total.Reserved = total.Reserved.Add(summary.Reserved)
		total.Available = total.Available.Add(summary.Available)
		total.Forecasted = total.Forecasted.Add(summary.Forecasted)
		total.InTransit = total.InTransit.Add(summary.InTransit)
		if summary.Truncated {
			total.Truncated = true
		}
		if total.BaseUomId == nil {
			total.BaseUomId = summary.BaseUomId
		}
		if summary.LastMovementAt != nil &&
			(total.LastMovementAt == nil || summary.LastMovementAt.After(*total.LastMovementAt)) {
			total.LastMovementAt = summary.LastMovementAt
		}

		breakdown = append(breakdown, itStock.TemplateVariantStockRow{
			VariantId:      derefId(variant.GetId()),
			Sku:            derefString(variant.GetSku()),
			CombinationKey: derefString(variant.GetCombinationKey()),
			Summary:        summary,
		})
	}

	// Location and warehouse counts are deliberately not summed: two variants in the same
	// aisle would count that aisle twice. They are left zero on the aggregate, where they have
	// no meaning, and remain accurate on each variant row.

	return &itStock.GetTemplateSummaryResult{
		HasData: true,
		Data: itStock.GetTemplateSummaryResultData{
			Summary:  total,
			Variants: breakdown,
		},
	}, nil
}

// laterOf keeps whichever of the two timestamps is more recent, treating a nil candidate as no
// information rather than as the zero time.
func laterOf(current *time.Time, candidate *model.ModelDateTime) *time.Time {
	if candidate == nil {
		return current
	}
	at := candidate.GoTime()
	if current == nil || at.After(*current) {
		return &at
	}
	return current
}
