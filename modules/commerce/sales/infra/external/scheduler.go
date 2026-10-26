package external

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	jobmodels "github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	itJob "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/job"

	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// An adapter rather than a hand-over, because the upstream command is a dynamic model: it embeds
// models.Job, so a caller must know the scheduler's own field names and enum values to build one.
// Passing that through would put the scheduler's schema into Sales' domain, where a renamed column
// upstream would break code that has nothing to do with scheduling.
//
// The three policy choices below are made here, once, rather than at each call site — they are
// properties of how Sales' jobs must behave, not of any one job.

type schedulerAdapter struct {
	jobs itJob.JobDomainService
}

func (this *schedulerAdapter) EnsureJob(
	ctx corectx.Context, cmd itExt.EnsureJobCommand,
) (*itExt.EnsureJobResult, error) {
	job := jobmodels.NewJob()
	job.SetFieldData(dmodel.DynamicFields{
		jobmodels.JobFieldName:       cmd.Name,
		jobmodels.JobFieldModuleName: cmd.ModuleName,
		jobmodels.JobFieldJobKey:     cmd.JobKey,

		// Technical, not user: nobody creates this from a screen, and it must not appear as
		// something an operator may freely rewrite.
		jobmodels.JobFieldJobType: jobmodels.JobTypeTechnical,

		// The scheduler dispatches over the command bus rather than calling an HTTP endpoint. The
		// handler is in this same process, and a loopback request would need a routable URL and an
		// authenticated caller for work that has neither.
		jobmodels.JobFieldActionType: jobmodels.ActionTypeCommandBus,
		jobmodels.JobFieldActionConfig: map[string]any{
			"command_name": cmd.CommandName,
		},

		jobmodels.JobFieldCronExpression: cmd.CronExpression,
		jobmodels.JobFieldIsEnabled:      true,

		// FORBID OVERLAP. A second run starting while the first still holds instructions in
		// `processing` would find nothing to do at best, and at worst race the first for the same
		// claim. The claim itself is what guarantees correctness; this keeps the two from competing
		// for it in the first place.
		jobmodels.JobFieldConcurrencyPolicy: jobmodels.ConcurrencyForbidOverlap,

		// SKIP a missed run rather than firing it late. These jobs sweep whatever is due whenever
		// they run, so a missed tick is picked up by the next one — replaying it would do the same
		// work twice for no gain.
		jobmodels.JobFieldMisfirePolicy: jobmodels.MisfireSkip,

		jobmodels.JobFieldMaxAttempts:          cmd.MaxAttempts,
		jobmodels.JobFieldRetryIntervalSeconds: cmd.RetryIntervalSeconds,
	})

	created, err := this.jobs.CreateJob(ctx, itJob.CreateJobCommand{Job: *job})
	if err != nil {
		return nil, errors.Wrapf(err, "registering the '%s' job", cmd.JobKey)
	}
	if created == nil || !created.HasData {
		return nil, errors.Errorf("registering the '%s' job returned nothing", cmd.JobKey)
	}

	result := &itExt.EnsureJobResult{WasCreated: created.Data.WasCreated}
	if id := created.Data.Job.GetId(); id != nil {
		result.JobId = string(*id)
	}
	return result, nil
}
