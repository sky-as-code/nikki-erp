package job

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// The build fails here if a command forgets CqrsRequestType, which is the only point at which
// that omission is cheap to notice: a command reaching the bus without one is a runtime panic.
func init() {
	var req cqrs.Request
	req = (*CreateJobCommand)(nil)
	req = (*UpdateJobCommand)(nil)
	req = (*DeleteJobCommand)(nil)
	req = (*DeleteJobsByModuleCommand)(nil)
	req = (*GetJobQuery)(nil)
	req = (*JobExistsQuery)(nil)
	req = (*SearchJobsQuery)(nil)
	util.Unused(req)
}

var createJobCommandType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "job", Action: "createJob",
}

// CreateJobCommand registers a job.
//
// Registration is idempotent on (module_name, job_key): re-submitting an identical registration
// returns the existing job rather than a conflict. A module registers its jobs on every boot, so
// a second boot must not be an error, and the composite unique index is the backstop for two
// instances booting at once.
type CreateJobCommand struct {
	models.Job
}

func (CreateJobCommand) CqrsRequestType() cqrs.RequestType {
	return createJobCommandType
}

func (CreateJobCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.JobSchemaName)
}

// CreateJobResultData carries the job together with whether this call is what created it.
//
// WasCreated is what lets the transport answer 201 for a first registration and 200 for a repeat.
// Without it the two are indistinguishable from the result, and a caller could not tell a fresh
// registration from a no-op.
type CreateJobResultData struct {
	Job        models.Job
	WasCreated bool
}

type CreateJobResult = dyn.OpResult[CreateJobResultData]

var updateJobCommandType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "job", Action: "updateJob",
}

type UpdateJobCommand struct {
	models.Job
}

func (UpdateJobCommand) CqrsRequestType() cqrs.RequestType {
	return updateJobCommandType
}

func (UpdateJobCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.JobSchemaName)
}

type UpdateJobResult = dyn.OpResult[dyn.MutateResultData]

var deleteJobCommandType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "job", Action: "deleteJob",
}

// DeleteJobCommand removes one job permanently. Its execution history survives, because the
// execution's edge to the job is ON DELETE SET NULL and each execution carries a job snapshot.
type DeleteJobCommand dyn.DeleteOneCommand

func (DeleteJobCommand) CqrsRequestType() cqrs.RequestType {
	return deleteJobCommandType
}

type DeleteJobResult = dyn.OpResult[dyn.MutateResultData]

var deleteJobsByModuleCommandType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "job", Action: "deleteJobsByModule",
}

// DeleteJobsByModuleCommand removes every technical job belonging to one module, which is how a
// module that is being uninstalled withdraws its registrations.
//
// ModuleName is mandatory and matched exactly. There is deliberately no wildcard and no
// "delete all" form: the difference between a typo and a catastrophe should not be one character.
type DeleteJobsByModuleCommand struct {
	ModuleName string `json:"module_name" query:"module_name"`
}

func (DeleteJobsByModuleCommand) CqrsRequestType() cqrs.RequestType {
	return deleteJobsByModuleCommandType
}

type DeleteJobsByModuleResult = dyn.OpResult[dyn.MutateResultData]

var getJobQueryType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "job", Action: "getJob",
}

type GetJobQuery dyn.GetOneQuery

func (GetJobQuery) CqrsRequestType() cqrs.RequestType {
	return getJobQueryType
}

// GetJobResult is what the domain service returns. The application service wraps it into
// SingleResultData so the transport can emit the {item, meta} envelope.
type GetJobResult = dyn.OpResult[models.Job]

type GetJobUiResult = dyn.OpResult[dyn.SingleResultData[models.Job]]

var jobExistsQueryType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "job", Action: "jobExists",
}

type JobExistsQuery dyn.ExistsQuery

func (JobExistsQuery) CqrsRequestType() cqrs.RequestType {
	return jobExistsQueryType
}

type JobExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchJobsQueryType = cqrs.RequestType{
	Module: "jobscheduler", Submodule: "job", Action: "searchJobs",
}

type SearchJobsQuery dyn.SearchQuery

func (SearchJobsQuery) CqrsRequestType() cqrs.RequestType {
	return searchJobsQueryType
}

type SearchJobsResultData = dyn.PagedResultData[models.Job]

type SearchJobsResult = dyn.OpResult[SearchJobsResultData]
