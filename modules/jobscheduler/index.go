// Package jobscheduler runs recurring jobs registered by other modules.
//
// It exists because the platform needs scheduling that survives a restart, spreads across
// instances, retries with a bounded policy, and keeps a history of what ran - none of which
// an in-process cron ticker provides. The scheduling engine, the cron parser and the
// distributed claim are all implemented here rather than taken from a library.
//
// The whole module works in UTC. There is no per-job timezone, cron expressions are evaluated
// in UTC, and a datetime submitted with any other offset is rejected.
package jobscheduler

import (
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/app"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/constants"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/services"
	external "github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/external"
	repo "github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
//
// It is typed as DynamicModule rather than InCodeModule so that dropping RegisterModels
// breaks the build instead of silently leaving three tables unregistered.
var ModuleSingleton modules.DynamicModule = &JobSchedulerModule{}

type JobSchedulerModule struct {
}

// LabelKey implements NikkiModule.
func (*JobSchedulerModule) LabelKey() string {
	return "jobscheduler.moduleLabel"
}

// Name implements NikkiModule.
func (*JobSchedulerModule) Name() string {
	return constants.JobSchedulerModuleName
}

// Deps implements NikkiModule.
//
// core and apptrait are implicit and must not be listed. The scheduler deliberately depends
// on no feature module: a technical job names its owning module as a free-form string and is
// dispatched through the command bus or HTTP, so the scheduler never links against the
// modules whose work it runs.
func (*JobSchedulerModule) Deps() []string {
	return []string{
		"dynamicresource",
	}
}

// IsInternal implements InCodeModule.
func (*JobSchedulerModule) IsInternal() bool {
	return false
}

// Version implements NikkiModule.
func (*JobSchedulerModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*JobSchedulerModule) Init() error {
	return errors.Join(
		external.InitExternalServices(),
		repo.InitRepositories(),
		services.InitDomainServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)
}

// RegisterModels implements DynamicModule.
//
// The job is registered first because the execution's edge points at it, and the execution
// before the attempt for the same reason: a schema may only reference one already registered.
func (*JobSchedulerModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(models.JobSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ExecutionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.AttemptSchemaBuilder()),
	)
}
