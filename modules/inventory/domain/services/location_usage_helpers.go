package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// usageScanPageSize is how many quants are read at a time. The scan pages to the end rather than
// stopping at the first page: a partial total would report an empty location that is not, and the
// caller decides whether archiving is safe from that answer.
const usageScanPageSize = 200

// closedMoveStatuses are the move states that are history rather than work in flight. A location
// referenced only by these can still be archived, so they are excluded from the open count.
func closedMoveStatuses() []string {
	return []string{models.StockMoveStatusDone, models.StockMoveStatusCancelled}
}

// closedTransferStatuses are the transfer states that are history rather than work in flight.
func closedTransferStatuses() []string {
	return []string{models.StockTransferStatusDone, models.StockTransferStatusCancelled}
}

// countMatching returns how many rows match, without fetching them: the repository reports Total
// alongside the page, so one row is enough to learn the count of a large set.
func countMatching(
	ctx corectx.Context, engine drif.DynamicResourceEngine, graph *dmodel.SearchGraph,
) (int, error) {
	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return 0, errors.Wrap(err, "countMatching")
	}
	if found == nil || !found.HasData {
		return 0, nil
	}
	return found.Data.Total, nil
}

func derefDecimal(v *decimal.Decimal) decimal.Decimal {
	if v == nil {
		return decimal.Zero
	}
	return *v
}
