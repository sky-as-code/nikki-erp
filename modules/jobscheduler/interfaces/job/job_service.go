package job

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

// JobDomainService holds the scheduling rules: cron validation, the effective period, the
// action-config check, and idempotent registration. It asserts no permission - that is the
// application service's only job, and keeping the two apart is what stops an internal caller
// (the engine, a CQRS handler) from needing a user's entitlements to read a job it must run.
type JobDomainService interface {
	CreateJob(ctx corectx.Context, cmd CreateJobCommand) (*CreateJobResult, error)
	UpdateJob(ctx corectx.Context, cmd UpdateJobCommand) (*UpdateJobResult, error)
	DeleteJob(ctx corectx.Context, cmd DeleteJobCommand) (*DeleteJobResult, error)
	DeleteJobsByModule(
		ctx corectx.Context, cmd DeleteJobsByModuleCommand,
	) (*DeleteJobsByModuleResult, error)
	GetJob(ctx corectx.Context, query GetJobQuery) (*GetJobResult, error)
	JobExists(ctx corectx.Context, query JobExistsQuery) (*JobExistsResult, error)
	SearchJobs(
		ctx corectx.Context, query SearchJobsQuery, opts corecrud.ServiceSearchOptions,
	) (*SearchJobsResult, error)
}

// JobAppService is the surface the transport calls. Every method asserts a permission first and
// then delegates.
//
// GetJob and SearchJobs return the Ui-decorated shapes rather than the domain ones, because field
// masking is a presentation concern that the engine, which also reads jobs, must not pay for.
type JobAppService interface {
	CreateJob(ctx corectx.Context, cmd CreateJobCommand) (*CreateJobResult, error)
	UpdateJob(ctx corectx.Context, cmd UpdateJobCommand) (*UpdateJobResult, error)
	DeleteJob(ctx corectx.Context, cmd DeleteJobCommand) (*DeleteJobResult, error)
	DeleteJobsByModule(
		ctx corectx.Context, cmd DeleteJobsByModuleCommand,
	) (*DeleteJobsByModuleResult, error)
	GetJob(ctx corectx.Context, query GetJobQuery) (*GetJobUiResult, error)
	JobExists(ctx corectx.Context, query JobExistsQuery) (*JobExistsResult, error)
	SearchJobs(ctx corectx.Context, query SearchJobsQuery) (*SearchJobsResult, error)
}
