package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/job"
)

type jobRestParams struct {
	dig.In

	JobSvc it.JobAppService
}

func NewJobRest(params jobRestParams) *JobRest {
	return &JobRest{JobSvc: params.JobSvc}
}

type JobRest struct {
	JobSvc it.JobAppService
}

// CreateJob registers a job, answering 201 the first time and 200 for a repeat registration.
//
// It cannot use ServeCreate, which hardcodes 201: a module registers its jobs on every boot, and
// answering 201 to the tenth boot would claim a job was created that has existed for weeks.
// Distinguishing the two is the whole point of making registration idempotent.
func (this JobRest) CreateJob(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create job"); e != nil {
			err = e
		}
	}()

	// Captured from the result rather than passed through the response, because the status is
	// the only place the distinction belongs: the body is the job either way.
	wasCreated := false

	return httpserver.ServeRequestDynamic(
		echoCtx,
		func(ctx corectx.Context, cmd it.CreateJobCommand) (*it.CreateJobResult, error) {
			result, err := this.JobSvc.CreateJob(ctx, cmd)
			if err == nil && result != nil && result.HasData {
				wasCreated = result.Data.WasCreated
			}
			return result, err
		},
		func(requestFields dmodel.DynamicFields) it.CreateJobCommand {
			cmd := it.CreateJobCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data it.CreateJobResultData) CreateJobResponse {
			return *httpserver.NewRestCreateResponseDyn(data.Job.GetFieldData())
		},
		func(c *echo.Context, body any) error {
			if wasCreated {
				return httpserver.JsonCreated(c, body)
			}
			return httpserver.JsonOk(c, body)
		},
	)
}

func (this JobRest) UpdateJob(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate[UpdateJobRequest, UpdateJobResponse](
		"update job", echoCtx, &it.UpdateJobCommand{}, this.JobSvc.UpdateJob,
	)
}

func (this JobRest) GetJob(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetJobRequest, GetJobResponse, models.Job](
		"get job", echoCtx, this.JobSvc.GetJob,
	)
}

func (this JobRest) SearchJobs(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchJobsRequest, SearchJobsResponse, models.Job](
		"search jobs", echoCtx, this.JobSvc.SearchJobs,
	)
}

func (this JobRest) JobExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists[JobExistsRequest, JobExistsResponse](
		"check if job exists", echoCtx, this.JobSvc.JobExists,
	)
}

func (this JobRest) DeleteJob(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[DeleteJobRequest, DeleteJobResponse](
		"delete job", echoCtx, this.JobSvc.DeleteJob,
	)
}

// DeleteJobsByModule withdraws every technical job of one module.
//
// It answers 200 with a count of zero when the module had none, rather than 404, so a module
// uninstalling itself need not know whether it ever registered anything. skipNotFoundError is
// what makes that so: without it the helper would turn "nothing matched" into an error.
func (this JobRest) DeleteJobsByModule(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete jobs by module"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.JobSvc.DeleteJobsByModule,
		httpserver.ItsMeMario,
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
		true,
	)
}

/* Non-CRUD APIs */

// GetModelSchema serves the job schema so a client can render a form without hard-coding the
// field list. It is routed before /jobs/:id, or the parameter route would swallow "meta".
func (this JobRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get job model schema"); e != nil {
			err = e
		}
	}()
	schema := dmodel.MustGetSchema(models.JobSchemaName)
	return echoCtx.JSON(http.StatusOK, schema.ToSimplized())
}
