package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/transport/restful/v1"
)

func InitRestfulHandlers() error {
	return stdErr.Join(
		deps.Register(v1.NewJobRest, v1.NewExecutionRest),
		initJobSchedulerV1(),
	)
}

// initJobSchedulerV1 mounts the scheduler's routes.
//
// Every route carries UserSmokeAuthz. It is what populates the permissions the application services
// read, so a route without it does not merely skip a check - it makes every check inside that
// request deny, which looks like a broken permission system rather than a missing middleware.
func initJobSchedulerV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		jobRest *v1.JobRest,
		executionRest *v1.ExecutionRest,
	) error {
		routeV1 := route.Group("/v1/jobscheduler")

		// The schema routes come first. Registered after "/jobs/:id" the parameter route would
		// match "meta" and answer 404 for a job that does not exist.
		routeV1.GET("/jobs/meta/schema", jobRest.GetModelSchema, m.UserSmokeAuthz())
		routeV1.GET("/executions/meta/schema", executionRest.GetModelSchema, m.UserSmokeAuthz())
		routeV1.GET("/attempts/meta/schema", executionRest.GetAttemptModelSchema, m.UserSmokeAuthz())

		// DELETE "/jobs" and DELETE "/jobs/:id" are distinct routes, not one with an optional
		// segment: the module-wide delete must be impossible to reach by omitting an id.
		routeV1.DELETE("/jobs", jobRest.DeleteJobsByModule, m.UserSmokeAuthz())
		routeV1.DELETE("/jobs/:id", jobRest.DeleteJob, m.UserSmokeAuthz())
		routeV1.GET("/jobs", jobRest.SearchJobs, m.UserSmokeAuthz())
		routeV1.GET("/jobs/:id", jobRest.GetJob, m.UserSmokeAuthz())
		routeV1.POST("/jobs", jobRest.CreateJob, m.UserSmokeAuthz())
		routeV1.POST("/jobs/exists", jobRest.JobExists, m.UserSmokeAuthz())
		routeV1.PUT("/jobs/:id", jobRest.UpdateJob, m.UserSmokeAuthz())

		routeV1.GET("/jobs/:job_id/executions", executionRest.SearchJobExecutions, m.UserSmokeAuthz())
		routeV1.GET("/executions", executionRest.SearchExecutions, m.UserSmokeAuthz())
		routeV1.GET("/executions/:id", executionRest.GetExecution, m.UserSmokeAuthz())

		return nil
	})
}
