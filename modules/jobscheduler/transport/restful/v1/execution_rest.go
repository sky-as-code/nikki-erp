package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/execution"
)

type executionRestParams struct {
	dig.In

	ExecutionSvc it.ExecutionAppService
}

func NewExecutionRest(params executionRestParams) *ExecutionRest {
	return &ExecutionRest{ExecutionSvc: params.ExecutionSvc}
}

// ExecutionRest is read-only. There is no create, update or delete handler here, and their
// absence is the design: executions are written by the engine, and history an API could rewrite
// could not be trusted to say what actually ran.
type ExecutionRest struct {
	ExecutionSvc it.ExecutionAppService
}

func (this ExecutionRest) GetExecution(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get execution"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ExecutionSvc.GetExecution,
		httpserver.ItsMeMario,
		NewGetExecutionResponse,
		httpserver.JsonOk,
	)
}

func (this ExecutionRest) SearchExecutions(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchExecutionsRequest, SearchExecutionsResponse, models.Execution](
		"search executions", echoCtx, this.ExecutionSvc.SearchExecutions,
	)
}

// SearchJobExecutions is one job's history, paged.
//
// It stays available after the job is deleted: the execution's job_id is set to NULL on delete,
// so this returns nothing for a deleted job while the rows survive and remain reachable through
// the unscoped listing. That is deliberate - deleting a registration must not destroy the record
// of what it did.
func (this ExecutionRest) SearchJobExecutions(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[
		SearchJobExecutionsRequest, SearchJobExecutionsResponse, models.Execution,
	](
		"search job executions", echoCtx, this.ExecutionSvc.SearchJobExecutions,
	)
}

/* Non-CRUD APIs */

func (this ExecutionRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get execution model schema"); e != nil {
			err = e
		}
	}()
	schema := dmodel.MustGetSchema(models.ExecutionSchemaName)
	return echoCtx.JSON(http.StatusOK, schema.ToSimplized())
}

func (this ExecutionRest) GetAttemptModelSchema(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get attempt model schema"); e != nil {
			err = e
		}
	}()
	schema := dmodel.MustGetSchema(models.AttemptSchemaName)
	return echoCtx.JSON(http.StatusOK, schema.ToSimplized())
}
