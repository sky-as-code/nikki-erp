package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/execution"
)

// maxAttemptsPerPage bounds the attempt list attached to one execution.
//
// It is far above any real retry chain, which the max_attempts quota caps in the low tens. It
// exists so that a corrupt row - an execution whose attempts were written in a loop that did not
// terminate - cannot turn one detail request into an unbounded read.
const maxAttemptsPerPage = 500

func NewExecutionApplicationServiceImpl(
	executionSvc it.ExecutionDomainService,
	executionRepo it.ExecutionRepository,
) it.ExecutionAppService {
	return &ExecutionApplicationServiceImpl{
		executionSvc:  executionSvc,
		executionRepo: executionRepo,
	}
}

type ExecutionApplicationServiceImpl struct {
	executionSvc  it.ExecutionDomainService
	executionRepo it.ExecutionRepository
}

// GetExecution answers the execution together with its attempts.
//
// Reading the attempts requires the attempt resource's own read permission, not the execution's.
// They are seeded as separate resources, and an installation that grants sight of what ran
// without granting sight of the error messages inside it is a coherent thing to want.
func (this *ExecutionApplicationServiceImpl) GetExecution(
	ctx corectx.Context, query it.GetExecutionQuery,
) (*it.GetExecutionResult, error) {
	if cErr := assertPermission(ctx, actionRead, executionResource); cErr != nil {
		return &it.GetExecutionResult{ClientErrors: *cErr}, nil
	}

	found, err := corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[models.Execution, *models.Execution]{
		Action: "get execution",
		Schema: this.executionRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[models.Execution], error) {
			return this.executionSvc.GetExecution(ctx, query)
		},
	})
	if err != nil {
		return nil, err
	}
	if found.ClientErrors.Count() > 0 {
		return &it.GetExecutionResult{ClientErrors: found.ClientErrors}, nil
	}
	if !found.HasData {
		return &it.GetExecutionResult{HasData: false}, nil
	}

	attempts, err := this.loadAttempts(ctx, found.Data.Item)
	if err != nil {
		return nil, err
	}

	return &it.GetExecutionResult{
		Data: it.GetExecutionResultData{
			Execution: found.Data.Item,
			Attempts:  attempts,
			Meta:      found.Data.Meta,
		},
		HasData: true,
	}, nil
}

// loadAttempts returns the attempts, or an empty list when the caller may not read them.
//
// A missing permission empties the list rather than failing the whole request: the execution the
// caller did ask for, and may read, should still come back.
func (this *ExecutionApplicationServiceImpl) loadAttempts(
	ctx corectx.Context, execution models.Execution,
) ([]models.Attempt, error) {
	if cErr := assertPermission(ctx, actionRead, attemptResource); cErr != nil {
		return nil, nil
	}
	id := execution.GetId()
	if id == nil {
		return nil, nil
	}
	return this.executionSvc.LoadAttemptsOf(ctx, *id, maxAttemptsPerPage)
}

func (this *ExecutionApplicationServiceImpl) SearchExecutions(
	ctx corectx.Context, query it.SearchExecutionsQuery,
) (*it.SearchExecutionsResult, error) {
	if cErr := assertPermission(ctx, actionRead, executionResource); cErr != nil {
		return &it.SearchExecutionsResult{ClientErrors: *cErr}, nil
	}
	return this.uiSearch(ctx, "search executions", func(
		fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery],
	) (*dyn.OpResult[dyn.PagedResultData[models.Execution]], error) {
		return this.executionSvc.SearchExecutions(ctx, query, corecrud.ServiceSearchOptions{
			AfterValidationSuccess: fn,
		})
	})
}

func (this *ExecutionApplicationServiceImpl) SearchJobExecutions(
	ctx corectx.Context, query it.SearchJobExecutionsQuery,
) (*it.SearchJobExecutionsResult, error) {
	if cErr := assertPermission(ctx, actionRead, executionResource); cErr != nil {
		return &it.SearchJobExecutionsResult{ClientErrors: *cErr}, nil
	}
	return this.uiSearch(ctx, "search job executions", func(
		fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery],
	) (*dyn.OpResult[dyn.PagedResultData[models.Execution]], error) {
		return this.executionSvc.SearchJobExecutions(ctx, query, corecrud.ServiceSearchOptions{
			AfterValidationSuccess: fn,
		})
	})
}

// uiSearch is shared by both listings so that the two views of the same history cannot drift
// into showing different columns.
func (this *ExecutionApplicationServiceImpl) uiSearch(
	ctx corectx.Context, action string, searchFn corecrud.SearchFn[models.Execution],
) (*it.SearchExecutionsResult, error) {
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[models.Execution, *models.Execution]{
		Action: action,
		Schema: this.executionRepo.GetBaseRepo().Schema(),
		// job_snapshot is omitted: it is a whole frozen job per row, which would dominate a page
		// of results, and the detail view is where it belongs.
		DefaultFields: []string{
			models.ExecutionFieldJobId,
			models.ExecutionFieldExecutionKey,
			models.ExecutionFieldScheduledFor,
			models.ExecutionFieldStatus,
			models.ExecutionFieldAvailableAt,
			models.ExecutionFieldStartedAt,
			models.ExecutionFieldFinishedAt,
			models.ExecutionFieldAttemptCount,
			models.ExecutionFieldFailureCode,
		},
		SearchFn: searchFn,
	})
}
