package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/job"
)

func NewJobDomainServiceImpl(
	repo it.JobRepository,
	executors ActionConfigValidator,
	cfg SchedulerConfig,
	waker EngineWaker,
) it.JobDomainService {
	return &JobDomainServiceImpl{
		repo:      repo,
		executors: executors,
		cfg:       cfg,
		waker:     waker,
	}
}

type JobDomainServiceImpl struct {
	repo      it.JobRepository
	executors ActionConfigValidator
	cfg       SchedulerConfig
	waker     EngineWaker
}

// CreateJob registers a job, idempotently on (module_name, job_key).
//
// The lookup and the insert share one transaction because a module registering its jobs at boot
// races every other instance of itself doing the same. Even so the composite unique index is the
// real guarantee: the lookup narrows the window, it does not close it.
func (this *JobDomainServiceImpl) CreateJob(
	ctx corectx.Context, cmd it.CreateJobCommand,
) (*it.CreateJobResult, error) {
	return corecrud.ExecInTranx(ctx, this.repo, func(tranxCtx corectx.Context) (*it.CreateJobResult, error) {
		existing, err := this.findByModuleAndKey(tranxCtx, cmd.Job)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			// A repeat registration is the normal case on every boot after the first, so it is a
			// success carrying the job that is already there, not a conflict.
			return &it.CreateJobResult{
				Data:    it.CreateJobResultData{Job: *existing, WasCreated: false},
				HasData: true,
			}, nil
		}
		return this.createNew(tranxCtx, cmd)
	})
}

func (this *JobDomainServiceImpl) createNew(
	ctx corectx.Context, cmd it.CreateJobCommand,
) (*it.CreateJobResult, error) {
	result, err := corecrud.Create(ctx, corecrud.CreateParam[models.Job, *models.Job]{
		Action:         "create job",
		BaseRepoGetter: this.repo,
		Data:           cmd,
		ValidateExtra: func(
			ctx corectx.Context, job *models.Job, vErrs *ft.ClientErrors,
		) error {
			ValidateJobRules(job, this.executors, vErrs)
			return nil
		},
		AfterValidationSuccess: func(ctx corectx.Context, job *models.Job) (*models.Job, error) {
			applyNextRunAt(job, this.cfg)
			return job, nil
		},
	})
	if err != nil || result.ClientErrors.Count() > 0 {
		return wrapCreateResult(result, err)
	}
	// The horizon has moved: the new job may be due before anything already scheduled, and
	// without this the engine would not look again until the next reconciliation.
	this.waker.Wake(WakeJobCreated)
	return wrapCreateResult(result, nil)
}

func wrapCreateResult(
	result *dyn.OpResult[models.Job], err error,
) (*it.CreateJobResult, error) {
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &it.CreateJobResult{ClientErrors: result.ClientErrors}, nil
	}
	return &it.CreateJobResult{
		Data:    it.CreateJobResultData{Job: result.Data, WasCreated: true},
		HasData: result.HasData,
	}, nil
}

func (this *JobDomainServiceImpl) UpdateJob(
	ctx corectx.Context, cmd it.UpdateJobCommand,
) (*it.UpdateJobResult, error) {
	// Captured from ValidateExtra, which runs first and is the only step handed the stored
	// row. AfterValidationSuccess only sees the incoming partial update, and next_run_at must
	// be recomputed from the row as it will be after the merge - a PUT sending only
	// is_enabled: false carries no cron_expression of its own, and computing against that
	// alone would leave next_run_at untouched instead of cleared.
	var stored *models.Job

	result, err := corecrud.Update(ctx, corecrud.UpdateParam[models.Job, *models.Job]{
		Action:       "update job",
		DbRepoGetter: this.repo,
		Data:         cmd,
		ValidateExtra: func(
			ctx corectx.Context, job *models.Job, found *models.Job, vErrs *ft.ClientErrors,
		) error {
			stored = found
			ValidateJobUpdateRules(job, found, this.executors, vErrs)
			return nil
		},
		AfterValidationSuccess: func(ctx corectx.Context, job *models.Job) (*models.Job, error) {
			merged := mergeJobForValidation(job, stored)
			applyNextRunAt(merged, this.cfg)
			job.SetNextRunAt(merged.GetNextRunAt())
			return job, nil
		},
	})
	if err == nil && result.ClientErrors.Count() == 0 {
		// An edited cron or effective period changes when this job is next due, in either
		// direction, so the horizon must be recomputed rather than merely extended.
		this.waker.Wake(WakeJobUpdated)
	}
	return result, err
}

// DeleteJob removes the job row. Its executions survive with job_id set to NULL, which the
// foreign key guarantees rather than any code here - application code can forget, a constraint
// cannot.
func (this *JobDomainServiceImpl) DeleteJob(
	ctx corectx.Context, cmd it.DeleteJobCommand,
) (*it.DeleteJobResult, error) {
	result, err := corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete job",
		DbRepoGetter: this.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
	if err == nil && result.ClientErrors.Count() == 0 {
		this.waker.Wake(WakeJobDeleted)
	}
	return result, err
}

func (this *JobDomainServiceImpl) GetJob(
	ctx corectx.Context, query it.GetJobQuery,
) (*it.GetJobResult, error) {
	return corecrud.GetOne[models.Job, *models.Job](ctx, corecrud.GetOneParam{
		Action:       "get job",
		DbRepoGetter: this.repo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *JobDomainServiceImpl) JobExists(
	ctx corectx.Context, query it.JobExistsQuery,
) (*it.JobExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if job exists",
		DbRepoGetter: this.repo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *JobDomainServiceImpl) SearchJobs(
	ctx corectx.Context, query it.SearchJobsQuery, opts corecrud.ServiceSearchOptions,
) (*it.SearchJobsResult, error) {
	return corecrud.Search[models.Job, *models.Job](ctx, corecrud.SearchParam{
		Action:                 "search jobs",
		DbRepoGetter:           this.repo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

// findByModuleAndKey answers the idempotency question. It searches rather than reading by id
// because (module_name, job_key) is the registration identity a caller knows; the row id is
// this module's business.
func (this *JobDomainServiceImpl) findByModuleAndKey(
	ctx corectx.Context, job models.Job,
) (*models.Job, error) {
	moduleName, jobKey := job.GetModuleName(), job.GetJobKey()
	if moduleName == nil || jobKey == nil {
		// Both are required by the schema. Letting the schema report the omission keeps one
		// error message for one mistake instead of two competing ones.
		return nil, nil
	}

	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchGraph().
			NewCondition(models.JobFieldModuleName, dmodel.Equals, *moduleName).ToSearchNode(),
		*dmodel.NewSearchGraph().
			NewCondition(models.JobFieldJobKey, dmodel.Equals, *jobKey).ToSearchNode(),
	)

	found, err := corecrud.Search[models.Job, *models.Job](ctx, corecrud.SearchParam{
		Action:       "find job by module and key",
		DbRepoGetter: this.repo,
		// Page is zero-indexed, the same as every other search in this codebase (see
		// model.MODEL_RULE_PAGE_INDEX_START). Page: 1 here silently OFFSETs past the only
		// matching row, so the lookup always finds nothing and registration was never
		// actually idempotent - every repeat POST fell through to the insert and hit the
		// unique constraint. Caught by a live run against Postgres, not by any unit test:
		// the pure test doubles never had a second matching row to page past.
		Query: dyn.SearchQuery{Graph: graph, Page: 0, Size: 1},
	})
	if err != nil {
		return nil, err
	}
	if !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return &found.Data.Items[0], nil
}

// findTechnicalJobIdsOfModule lists what a delete-by-module would remove.
//
// The job_type predicate is not decoration: user-managed jobs are somebody's saved configuration
// rather than a module's registration, and a module uninstalling itself has no business deleting
// them. It costs one clause here and would be unrecoverable if omitted.
func (this *JobDomainServiceImpl) findTechnicalJobIdsOfModule(
	ctx corectx.Context, moduleName string,
) ([]model.Id, error) {
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchGraph().
			NewCondition(models.JobFieldModuleName, dmodel.Equals, moduleName).ToSearchNode(),
		*dmodel.NewSearchGraph().
			NewCondition(models.JobFieldJobType, dmodel.Equals, models.JobTypeTechnical).ToSearchNode(),
	)

	found, err := corecrud.SearchAll(func(page int, size int) (
		*dyn.OpResult[dyn.PagedResultData[models.Job]], error,
	) {
		return corecrud.Search[models.Job, *models.Job](ctx, corecrud.SearchParam{
			Action:       "find technical jobs of module",
			DbRepoGetter: this.repo,
			Query:        dyn.SearchQuery{Graph: graph, Page: page, Size: size},
		})
	})
	if err != nil {
		return nil, err
	}
	if !found.HasData {
		return nil, nil
	}

	ids := make([]model.Id, 0, len(found.Data))
	for _, job := range found.Data {
		if id := job.GetId(); id != nil {
			ids = append(ids, *id)
		}
	}
	return ids, nil
}

// DeleteJobsByModule withdraws every technical job of one module.
//
// Deleting nothing is a success, not a 404: a module uninstalling itself should not have to know
// whether it ever registered anything, and making the caller distinguish those two cases would
// only produce callers that ignore the difference.
func (this *JobDomainServiceImpl) DeleteJobsByModule(
	ctx corectx.Context, cmd it.DeleteJobsByModuleCommand,
) (*it.DeleteJobsByModuleResult, error) {
	if vErrs := validateModuleName(cmd.ModuleName); vErrs != nil {
		return &it.DeleteJobsByModuleResult{ClientErrors: *vErrs}, nil
	}

	return corecrud.ExecInTranx(ctx, this.repo, func(
		tranxCtx corectx.Context,
	) (*it.DeleteJobsByModuleResult, error) {
		ids, err := this.findTechnicalJobIdsOfModule(tranxCtx, cmd.ModuleName)
		if err != nil {
			return nil, err
		}
		deleted := 0
		for _, id := range ids {
			result, err := corecrud.DeleteOne(tranxCtx, corecrud.DeleteOneParam{
				Action:       "delete job by module",
				DbRepoGetter: this.repo,
				Cmd:          dyn.DeleteOneCommand{Id: id},
			})
			if err != nil {
				return nil, err
			}
			if result.HasData {
				deleted += result.Data.AffectedCount
			}
		}
		if deleted > 0 {
			this.waker.Wake(WakeJobDeleted)
		}
		return &it.DeleteJobsByModuleResult{
			Data: dyn.MutateResultData{
				AffectedCount: deleted,
				AffectedAt:    *util.ToPtr(nowModelDateTime()),
			},
			HasData: true,
		}, nil
	})
}
