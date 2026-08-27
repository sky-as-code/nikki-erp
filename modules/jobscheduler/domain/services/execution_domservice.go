package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	itexec "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/execution"
)

func NewExecutionDomainServiceImpl(
	executionRepo itexec.ExecutionRepository,
	attemptRepo itexec.AttemptRepository,
) itexec.ExecutionDomainService {
	return &ExecutionDomainServiceImpl{
		executionRepo: executionRepo,
		attemptRepo:   attemptRepo,
	}
}

type ExecutionDomainServiceImpl struct {
	executionRepo itexec.ExecutionRepository
	attemptRepo   itexec.AttemptRepository
}

func (this *ExecutionDomainServiceImpl) GetExecution(
	ctx corectx.Context, query itexec.GetExecutionQuery,
) (*itexec.GetExecutionRawResult, error) {
	return corecrud.GetOne[models.Execution, *models.Execution](ctx, corecrud.GetOneParam{
		Action:       "get execution",
		DbRepoGetter: this.executionRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ExecutionDomainServiceImpl) SearchExecutions(
	ctx corectx.Context, query itexec.SearchExecutionsQuery, opts corecrud.ServiceSearchOptions,
) (*itexec.SearchExecutionsResult, error) {
	return corecrud.Search[models.Execution, *models.Execution](ctx, corecrud.SearchParam{
		Action:                 "search executions",
		DbRepoGetter:           this.executionRepo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

// SearchJobExecutions is one job's history.
//
// The job predicate is ANDed on top of whatever graph the client sent rather than replacing it,
// so a client cannot widen the search past the job named in the path by supplying a graph of its
// own. Scoping that can be overridden by the request it scopes is not scoping.
func (this *ExecutionDomainServiceImpl) SearchJobExecutions(
	ctx corectx.Context, query itexec.SearchJobExecutionsQuery, opts corecrud.ServiceSearchOptions,
) (*itexec.SearchJobExecutionsResult, error) {
	inner := query.SearchQuery
	inner.Graph = scopeGraphToJob(inner.Graph, query.JobId)

	return corecrud.Search[models.Execution, *models.Execution](ctx, corecrud.SearchParam{
		Action:                 "search job executions",
		DbRepoGetter:           this.executionRepo,
		Query:                  inner,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func scopeGraphToJob(graph *dmodel.SearchGraph, jobId model.Id) *dmodel.SearchGraph {
	jobNode := *dmodel.NewSearchGraph().
		NewCondition(models.ExecutionFieldJobId, dmodel.Equals, jobId).ToSearchNode()

	if graph == nil {
		return dmodel.NewSearchGraph().And(jobNode)
	}
	scoped := dmodel.NewSearchGraph().And(jobNode, *graph.ToSearchNode())
	// The client's ordering is theirs to choose; only the predicate is constrained.
	if order := graph.GetOrder(); order != nil {
		scoped = scoped.Order(order)
	}
	return scoped
}

// LoadAttemptsOf reads one execution's attempts in the order they ran.
//
// Ordering by attempt_number rather than by started_at is deliberate: a reaped attempt whose
// worker died has no reliable finish time, and two attempts of one execution are strictly
// ordered by their number whatever their clocks said.
func (this *ExecutionDomainServiceImpl) LoadAttemptsOf(
	ctx corectx.Context, executionId model.Id, limit int,
) ([]models.Attempt, error) {
	graph := dmodel.NewSearchGraph().
		NewCondition(models.AttemptFieldExecutionId, dmodel.Equals, executionId).
		OrderBy(models.AttemptFieldAttemptNumber, dmodel.Asc)

	found, err := corecrud.Search[models.Attempt, *models.Attempt](ctx, corecrud.SearchParam{
		Action:       "load attempts of execution",
		DbRepoGetter: this.attemptRepo,
		// Page 0 is the first page (model.MODEL_RULE_PAGE_INDEX_START). Page: 1 here would
		// OFFSET past every attempt below the limit, leaving the detail view's attempt list
		// empty for any execution that has not run limit+1 times.
		Query: dyn.SearchQuery{Graph: graph, Page: 0, Size: limit},
	})
	if err != nil {
		return nil, err
	}
	if !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}
