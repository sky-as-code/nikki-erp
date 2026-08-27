package job

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// JobRepository is the persistence contract for scheduled jobs.
//
// The scheduler-specific reads - due jobs, and the earliest upcoming instant - are not on this
// interface but on SchedulerClaimRepository below, because they are raw SQL rather than engine
// calls and it is worth having the boundary between the two be visible in the type.
type JobRepository interface {
	GetBaseRepo() dyn.BaseDynamicRepository
	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)

	Insert(ctx corectx.Context, job models.Job) (*dyn.OpResult[int], error)
	Update(ctx corectx.Context, job models.Job) (*dyn.OpResult[dyn.MutateResultData], error)
	DeleteOne(ctx corectx.Context, keys models.Job) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.Job) (*dyn.OpResult[dyn.RepoExistsResult], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Job], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Job]], error)
}
