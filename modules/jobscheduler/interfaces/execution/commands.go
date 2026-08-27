package execution

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*GetExecutionQuery)(nil)
	req = (*SearchExecutionsQuery)(nil)
	req = (*SearchJobExecutionsQuery)(nil)
	util.Unused(req)
}

var getExecutionQueryType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "execution", Action: "getExecution",
}

type GetExecutionQuery dyn.GetOneQuery

func (GetExecutionQuery) CqrsRequestType() cqrs.RequestType {
	return getExecutionQueryType
}

// GetExecutionResultData is the execution together with its attempts, ordered by attempt number.
//
// The two are returned as one because an execution on its own does not answer the question the
// detail view exists for - why it ended the way it did - and that answer lives entirely in the
// attempts. The engine's GetOne will not join, so the composition happens in the app service.
type GetExecutionResultData struct {
	Execution models.Execution   `json:"execution"`
	Attempts  []models.Attempt   `json:"attempts"`
	Meta      dyn.SingleMetaData `json:"meta"`
}

type GetExecutionResult = dyn.OpResult[GetExecutionResultData]

// GetExecutionRawResult is what the domain service returns before attempts are attached.
type GetExecutionRawResult = dyn.OpResult[models.Execution]

var searchExecutionsQueryType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "execution", Action: "searchExecutions",
}

type SearchExecutionsQuery dyn.SearchQuery

func (SearchExecutionsQuery) CqrsRequestType() cqrs.RequestType {
	return searchExecutionsQueryType
}

type SearchExecutionsResultData = dyn.PagedResultData[models.Execution]

type SearchExecutionsResult = dyn.OpResult[SearchExecutionsResultData]

var searchJobExecutionsQueryType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "execution", Action: "searchJobExecutions",
}

// SearchJobExecutionsQuery is the history of one job, paged.
//
// JobId comes from the path rather than from the search graph so that the scoping cannot be
// dropped by a client that supplies its own graph: the job predicate is added on top of whatever
// was sent, not merged into it.
type SearchJobExecutionsQuery struct {
	dyn.SearchQuery
	JobId model.Id `json:"job_id" param:"job_id"`
}

func (SearchJobExecutionsQuery) CqrsRequestType() cqrs.RequestType {
	return searchJobExecutionsQueryType
}

type SearchJobExecutionsResult = SearchExecutionsResult
