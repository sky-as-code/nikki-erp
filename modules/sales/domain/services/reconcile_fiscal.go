package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// Resolving fiscal requests whose answer never came back.
//
// A request left `pending` is the one state nobody can act on: Sales asked for a document and does
// not know whether it exists. Retrying blind would risk a second document for one sale; giving up
// would abandon a document that may have been issued. Only the provider knows, so this asks.
//
// It is the recovery path the port's own GetStatus comment describes, and the reason issuing carries
// an idempotency key at all.

// ReconcileFiscalResult reports what one pass did.
type ReconcileFiscalResult struct {
	Examined int

	// Resolved counts requests that turned out to have a document after all.
	Resolved int

	// Unresolved counts those the provider still has nothing for. They stay `pending` — a request
	// the provider has not answered is not a request that failed, and marking it failed would
	// invite a retry that mints a duplicate.
	Unresolved int
}

// ReconcileStaleFiscalRequests resolves requests whose answer was lost.
//
// olderThan keeps the sweep off requests that were only just sent: the provider may still be
// working, and asking about every fresh request would be a round trip for an answer nobody has yet.
func ReconcileStaleFiscalRequests(
	ctx corectx.Context,
	provider itInvoicing.InvoicingExtService,
	now time.Time,
	olderThan time.Duration,
	limit int,
) (*ReconcileFiscalResult, error) {
	result := &ReconcileFiscalResult{}
	if provider == nil {
		// No provider bound, so nothing can be asked. Not an error: requests stay pending, which is
		// exactly the state they are meant to have while nothing can issue them.
		return result, nil
	}

	stale, err := pendingFiscalRequestsOlderThan(ctx, now.Add(-olderThan), limit)
	if err != nil {
		return nil, err
	}

	for _, request := range stale {
		result.Examined++

		requestId := stringOf(request, models.SalesFiscalRequestFieldId)
		idempotencyKey := stringOf(request, models.SalesFiscalRequestFieldIdempotencyKey)
		if idempotencyKey == "" {
			// Without a key there is nothing to ask about. Left alone rather than guessed at.
			result.Unresolved++
			continue
		}

		status, err := provider.GetStatus(ctx, idempotencyKey)
		if err != nil {
			// One unreachable provider call must not end the pass; the request comes round again.
			continue
		}
		if status == nil || !status.Issued {
			result.Unresolved++
			continue
		}

		bill, err := loadRecord(ctx, models.SalesBillSchemaName, models.SalesBillFieldId,
			stringOf(request, models.SalesFiscalRequestFieldSalesBillId))
		if err != nil {
			continue
		}
		if bill == nil {
			continue
		}

		// The same write the original call would have made, so a recovered request is
		// indistinguishable from one that answered in time.
		if err := recordFiscalOutcome(ctx, bill, requestId, status); err != nil {
			continue
		}
		result.Resolved++
	}
	return result, nil
}

// pendingFiscalRequestsOlderThan finds the requests still waiting on an answer.
func pendingFiscalRequestsOlderThan(
	ctx corectx.Context, cutoff time.Time, limit int,
) ([]dmodel.DynamicFields, error) {
	engine, err := engineFor(models.SalesFiscalRequestSchemaName)
	if err != nil {
		return nil, err
	}

	size := limit
	if size <= 0 {
		size = model.MODEL_RULE_PAGE_MAX_SIZE
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.SalesFiscalRequestFieldStatus, dmodel.Equals,
			string(models.SalesFiscalStatusPending)),
		*dmodel.NewSearchNode().NewCondition(
			basemodel.FieldCreatedAt, dmodel.LessThan, model.ModelDateTime(cutoff)),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  size,
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}
