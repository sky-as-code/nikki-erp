package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The Stock side of the Product lifecycle contract: what would be stranded if a variant were
// archived now.
//
// It reports and never decides. Whether the numbers block an archive is Product's rule, expressed
// in product_archive_guard.go — the same split as the location guard, where Stock says what is
// there and Warehouse decides what that means.
//
// Archiving must never tidy up on the way through: no unreserving, no cancelling, no zeroing, no
// scrap (CR §14.4, PROD-INT-INV-013). That is why this is a reader with no write path at all.

var _ itStock.StockProductUsageReader = (*StockQuantDomainServiceImpl)(nil)

// GetProductUsage reports what one variant would strand.
func (this *StockQuantDomainServiceImpl) GetProductUsage(
	ctx corectx.Context, query itStock.GetProductUsageQuery,
) (*itStock.GetProductUsageResult, error) {
	if query.VariantId == "" {
		return nil, errors.New("GetProductUsage requires a variant id")
	}

	batch, err := this.GetProductUsageBatch(
		ctx, itStock.GetProductUsageBatchQuery{VariantIds: []string{query.VariantId}})
	if err != nil {
		return nil, err
	}

	return &itStock.GetProductUsageResult{
		HasData: true,
		Data: itStock.GetProductUsageResultData{
			Usage: batch.Data.Usages[query.VariantId],
		},
	}, nil
}

// GetProductUsageBatch reports a whole set of variants together.
//
// Archiving a template has to clear every one of its variants before archiving any of them, so the
// guard needs the whole set in hand before it writes anything. Reading them one at a time inside
// the cascade loop is what leaves half a template archived when the last variant turns out to hold
// stock (CR §14.3, TS-PROD-12).
func (this *StockQuantDomainServiceImpl) GetProductUsageBatch(
	ctx corectx.Context, query itStock.GetProductUsageBatchQuery,
) (*itStock.GetProductUsageBatchResult, error) {
	variantIds := dedupeNonEmpty(query.VariantIds)
	usages := make(map[string]itStock.ProductUsage, len(variantIds))
	for _, id := range variantIds {
		usages[id] = itStock.ProductUsage{}
	}
	if len(variantIds) == 0 {
		return &itStock.GetProductUsageBatchResult{
			HasData: true,
			Data:    itStock.GetProductUsageBatchResultData{Usages: usages},
		}, nil
	}

	if err := this.accumulateUsageQuantities(ctx, variantIds, usages); err != nil {
		return nil, err
	}
	if err := this.accumulateOpenWork(ctx, variantIds, usages); err != nil {
		return nil, err
	}

	return &itStock.GetProductUsageBatchResult{
		HasData: true,
		Data:    itStock.GetProductUsageBatchResultData{Usages: usages},
	}, nil
}

// accumulateUsageQuantities totals on-hand and reserved across every location holding the variants.
func (this *StockQuantDomainServiceImpl) accumulateUsageQuantities(
	ctx corectx.Context, variantIds []string, usages map[string]itStock.ProductUsage,
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

	for page := 0; page < maxSummaryQuantPages; page++ {
		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  summaryScanPageSize,
		})
		if err != nil {
			return errors.Wrap(err, "accumulateUsageQuantities")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}

		for _, row := range found.Data.Items {
			quant := models.NewStockQuantFrom(row)
			variantId := derefId(quant.GetProductVariantId())
			usage, tracked := usages[variantId]
			if !tracked {
				continue
			}
			usage.OnHandQuantity = usage.OnHandQuantity.Add(derefDecimal(quant.GetOnHandQuantity()))
			usage.ReservedQuantity = usage.ReservedQuantity.Add(
				derefDecimal(quant.GetReservedQuantity()))
			usages[variantId] = usage
		}

		if len(found.Data.Items) < summaryScanPageSize {
			break
		}
	}
	return nil
}

// accumulateOpenWork counts the moves and transfers still in flight for each variant.
//
// Done and cancelled are excluded deliberately. A variant referenced only by completed movement
// archives fine — the history keeps resolving it — so counting it would block a safe archive
// forever (CR §14.2, AC-PROD-INT-031, TS-PROD-11).
//
// Transfers are counted through their moves: a transfer has no product of its own, and its moves
// are what name the variant. Distinct transfer ids are collected so a transfer carrying three
// lines of the same product counts once.
func (this *StockQuantDomainServiceImpl) accumulateOpenWork(
	ctx corectx.Context, variantIds []string, usages map[string]itStock.ProductUsage,
) error {
	engine, err := engineFor(models.StockMoveSchemaName)
	if err != nil {
		return err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.StockMoveFieldProductVariantId, dmodel.In, toAnySlice(variantIds)...),
		*dmodel.NewSearchNode().NewCondition(
			models.StockMoveFieldStatus, dmodel.NotIn, closedMoveStatuses()),
	)

	transfersSeen := map[string]map[string]bool{}
	for page := 0; page < maxSummaryQuantPages; page++ {
		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  summaryScanPageSize,
		})
		if err != nil {
			return errors.Wrap(err, "accumulateOpenWork")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}

		for _, row := range found.Data.Items {
			move := models.NewStockMoveFrom(row)
			variantId := derefId(move.GetProductVariantId())
			usage, tracked := usages[variantId]
			if !tracked {
				continue
			}

			usage.OpenMoveCount++
			usages[variantId] = usage

			if transferId := derefId(move.GetTransferId()); transferId != "" {
				if transfersSeen[variantId] == nil {
					transfersSeen[variantId] = map[string]bool{}
				}
				transfersSeen[variantId][transferId] = true
			}
		}

		if len(found.Data.Items) < summaryScanPageSize {
			break
		}
	}

	for variantId, transfers := range transfersSeen {
		usage := usages[variantId]
		usage.OpenTransferCount = len(transfers)
		usages[variantId] = usage
	}
	return nil
}
