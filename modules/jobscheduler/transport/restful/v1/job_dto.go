package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/job"
)

// The request types are aliases of the commands rather than separate structs, so that adding a
// field to a command cannot silently fail to reach the API.
type (
	CreateJobRequest  = it.CreateJobCommand
	CreateJobResponse = httpserver.RestCreateResponse

	UpdateJobRequest  = it.UpdateJobCommand
	UpdateJobResponse = httpserver.RestMutateResponse

	DeleteJobRequest  = it.DeleteJobCommand
	DeleteJobResponse = httpserver.RestMutateResponse

	DeleteJobsByModuleRequest  = it.DeleteJobsByModuleCommand
	DeleteJobsByModuleResponse = httpserver.RestMutateResponse

	GetJobRequest  = it.GetJobQuery
	GetJobResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

	JobExistsRequest  = it.JobExistsQuery
	JobExistsResponse = dyn.ExistsResultData

	SearchJobsRequest  = it.SearchJobsQuery
	SearchJobsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
)
