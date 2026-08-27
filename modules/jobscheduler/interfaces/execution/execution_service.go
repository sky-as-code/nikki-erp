package execution

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// ExecutionDomainService reads execution history.
//
// There is no create, update or delete here on purpose. Executions and attempts are written by
// the engine through the repositories, and history that an API could rewrite would not be able
// to answer the one question it exists for: what actually ran.
type ExecutionDomainService interface {
	GetExecution(ctx corectx.Context, query GetExecutionQuery) (*GetExecutionRawResult, error)
	SearchExecutions(
		ctx corectx.Context, query SearchExecutionsQuery, opts corecrud.ServiceSearchOptions,
	) (*SearchExecutionsResult, error)
	SearchJobExecutions(
		ctx corectx.Context, query SearchJobExecutionsQuery, opts corecrud.ServiceSearchOptions,
	) (*SearchJobExecutionsResult, error)

	// LoadAttemptsOf reads one execution's attempts in the order they ran, up to limit.
	//
	// It is on the interface rather than on the implementation because the application service
	// composes the detail response from it, and reaching for it through a type assertion would
	// make the dependency invisible to the compiler and to anyone substituting a fake.
	LoadAttemptsOf(
		ctx corectx.Context, executionId model.Id, limit int,
	) ([]models.Attempt, error)
}

type ExecutionAppService interface {
	GetExecution(ctx corectx.Context, query GetExecutionQuery) (*GetExecutionResult, error)
	SearchExecutions(
		ctx corectx.Context, query SearchExecutionsQuery,
	) (*SearchExecutionsResult, error)
	SearchJobExecutions(
		ctx corectx.Context, query SearchJobExecutionsQuery,
	) (*SearchJobExecutionsResult, error)
}
