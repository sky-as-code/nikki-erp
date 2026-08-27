package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/job"
)

func NewJobApplicationServiceImpl(
	jobSvc it.JobDomainService, jobRepo it.JobRepository,
) it.JobAppService {
	return &JobApplicationServiceImpl{jobSvc: jobSvc, jobRepo: jobRepo}
}

type JobApplicationServiceImpl struct {
	jobSvc it.JobDomainService

	// jobRepo is held only for its schema, which the Ui helpers need in order to stamp the
	// schema etag onto the response. No query goes through it here.
	jobRepo it.JobRepository
}

func (this *JobApplicationServiceImpl) CreateJob(
	ctx corectx.Context, cmd it.CreateJobCommand,
) (*it.CreateJobResult, error) {
	if cErr := assertPermission(ctx, actionCreate, jobResource); cErr != nil {
		return &it.CreateJobResult{ClientErrors: *cErr}, nil
	}
	return this.jobSvc.CreateJob(ctx, cmd)
}

func (this *JobApplicationServiceImpl) UpdateJob(
	ctx corectx.Context, cmd it.UpdateJobCommand,
) (*it.UpdateJobResult, error) {
	if cErr := assertPermission(ctx, actionUpdate, jobResource); cErr != nil {
		return &it.UpdateJobResult{ClientErrors: *cErr}, nil
	}
	return this.jobSvc.UpdateJob(ctx, cmd)
}

func (this *JobApplicationServiceImpl) DeleteJob(
	ctx corectx.Context, cmd it.DeleteJobCommand,
) (*it.DeleteJobResult, error) {
	if cErr := assertPermission(ctx, actionDelete, jobResource); cErr != nil {
		return &it.DeleteJobResult{ClientErrors: *cErr}, nil
	}
	return this.jobSvc.DeleteJob(ctx, cmd)
}

func (this *JobApplicationServiceImpl) DeleteJobsByModule(
	ctx corectx.Context, cmd it.DeleteJobsByModuleCommand,
) (*it.DeleteJobsByModuleResult, error) {
	if cErr := assertPermission(ctx, actionDelete, jobResource); cErr != nil {
		return &it.DeleteJobsByModuleResult{ClientErrors: *cErr}, nil
	}
	return this.jobSvc.DeleteJobsByModule(ctx, cmd)
}

func (this *JobApplicationServiceImpl) GetJob(
	ctx corectx.Context, query it.GetJobQuery,
) (*it.GetJobUiResult, error) {
	if cErr := assertPermission(ctx, actionRead, jobResource); cErr != nil {
		return &it.GetJobUiResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[models.Job, *models.Job]{
		Action: "get job",
		Schema: this.jobRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[models.Job], error) {
			return this.jobSvc.GetJob(ctx, query)
		},
	})
}

func (this *JobApplicationServiceImpl) JobExists(
	ctx corectx.Context, query it.JobExistsQuery,
) (*it.JobExistsResult, error) {
	if cErr := assertPermission(ctx, actionRead, jobResource); cErr != nil {
		return &it.JobExistsResult{ClientErrors: *cErr}, nil
	}
	return this.jobSvc.JobExists(ctx, query)
}

func (this *JobApplicationServiceImpl) SearchJobs(
	ctx corectx.Context, query it.SearchJobsQuery,
) (*it.SearchJobsResult, error) {
	if cErr := assertPermission(ctx, actionRead, jobResource); cErr != nil {
		return &it.SearchJobsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[models.Job, *models.Job]{
		Action: "search jobs",
		Schema: this.jobRepo.GetBaseRepo().Schema(),
		// The list view answers "what is registered, and is it running": identity, schedule and
		// whether it is on. action_config is deliberately absent - it can hold a URL with
		// credentials in it, and a list is the wrong place to spray that across a screen.
		DefaultFields: []string{
			models.JobFieldName,
			models.JobFieldModuleName,
			models.JobFieldJobKey,
			models.JobFieldJobType,
			models.JobFieldActionType,
			models.JobFieldCronExpression,
			models.JobFieldIsEnabled,
			models.JobFieldNextRunAt,
			models.JobFieldEffectiveFrom,
			models.JobFieldEffectiveUntil,
		},
		SearchFn: func(
			fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery],
		) (*dyn.OpResult[dyn.PagedResultData[models.Job]], error) {
			return this.jobSvc.SearchJobs(ctx, query, corecrud.ServiceSearchOptions{
				AfterValidationSuccess: fn,
			})
		},
	})
}
