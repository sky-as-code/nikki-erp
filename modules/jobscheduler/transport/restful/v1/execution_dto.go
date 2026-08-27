package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/execution"
)

type (
	GetExecutionRequest = it.GetExecutionQuery

	SearchExecutionsRequest  = it.SearchExecutionsQuery
	SearchExecutionsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

	SearchJobExecutionsRequest  = it.SearchJobExecutionsQuery
	SearchJobExecutionsResponse = SearchExecutionsResponse
)

// GetExecutionResponse is the execution detail: the row, its attempts, and the field metadata.
//
// It is hand-written rather than aliased to RestGetOneResponse because that envelope carries one
// item, and the question this endpoint answers - why did this execution end the way it did -
// cannot be answered without the attempts beside it.
type GetExecutionResponse struct {
	Item     dmodel.DynamicFields   `json:"item"`
	Attempts []dmodel.DynamicFields `json:"attempts"`
	Meta     dyn.SingleMetaData     `json:"meta"`
}

func NewGetExecutionResponse(data it.GetExecutionResultData) GetExecutionResponse {
	// An execution with no attempts is ordinary - it may be queued and not yet run - so the list
	// is emptied rather than left nil, sparing every client a null check for a normal state.
	attempts := make([]dmodel.DynamicFields, 0, len(data.Attempts))
	for _, attempt := range data.Attempts {
		attempts = append(attempts, attempt.GetFieldData())
	}
	return GetExecutionResponse{
		Item:     data.Execution.GetFieldData(),
		Attempts: attempts,
		Meta:     data.Meta,
	}
}
