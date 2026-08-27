package execution

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// ExecutionRepository is the persistence contract for job executions.
type ExecutionRepository interface {
	GetBaseRepo() dyn.BaseDynamicRepository
	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)

	Insert(ctx corectx.Context, execution models.Execution) (*dyn.OpResult[int], error)
	Update(ctx corectx.Context, execution models.Execution) (*dyn.OpResult[dyn.MutateResultData], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Execution], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Execution]], error)
}

// AttemptRepository is the persistence contract for the attempts within an execution.
//
// There is no Update for a finished attempt beyond recording its outcome, and no Delete at all:
// attempt history is the record of what actually ran, and a scheduler that could rewrite it would
// be unable to answer the one question the history exists for.
type AttemptRepository interface {
	GetBaseRepo() dyn.BaseDynamicRepository
	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)

	Insert(ctx corectx.Context, attempt models.Attempt) (*dyn.OpResult[int], error)
	Update(ctx corectx.Context, attempt models.Attempt) (*dyn.OpResult[dyn.MutateResultData], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Attempt], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Attempt]], error)
}
