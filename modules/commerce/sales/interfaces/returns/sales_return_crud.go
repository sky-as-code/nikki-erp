// Package returns declares the layers of the return and refund resources: the customer's request
// to send goods back, the lines it covers, and the money paid back against it.
package returns

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SalesReturnRepository reads and writes returns.
type SalesReturnRepository interface {
	dynamicresource.CrudRepository
}

// SalesReturnDomainService is the CRUD of the resource. The state machine lives in the operation
// services, each a multi-row transaction under a distributed lock.
type SalesReturnDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesReturnApplicationService is the authorized surface the REST handler serves.
type SalesReturnApplicationService interface {
	dynamicresource.CrudApplicationService

	// CreateReturn is collection-level: it records the request, and deliberately is not the power
	// to refund.
	CreateReturn(ctx corectx.Context, cmd ReturnActionCommand) (*dyn.OpResult[any], error)

	// Confirm approves a draft refund request, records the note, and dispatches it.
	Confirm(ctx corectx.Context, cmd ReturnActionCommand) (*dyn.OpResult[any], error)

	// Process carries its own permission rather than update, because it moves money out of the
	// business.
	Process(ctx corectx.Context, cmd ReturnActionCommand) (*dyn.OpResult[any], error)

	// Cancel reuses update: cancelling before anything irreversible is an ordinary correction, and
	// the state machine already refuses it afterwards.
	Cancel(ctx corectx.Context, cmd ReturnActionCommand) (*dyn.OpResult[any], error)
}

// ReturnActionCommand is the bound request of a custom return action.
type ReturnActionCommand = dynamicresource.UpdateCommand

// The read-only return resources: a line is written by the return that covers it, a refund leg by
// the processing that paid it back.

type SalesReturnLineRepository interface {
	dynamicresource.CrudRepository
}

type SalesReturnLineApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesRefundPaymentRepository interface {
	dynamicresource.CrudRepository
}

type SalesRefundPaymentApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateSalesReturnCommand      = dynamicresource.CreateCommand
	UpdateSalesReturnCommand      = dynamicresource.UpdateCommand
	DeleteSalesReturnCommand      = dynamicresource.DeleteCommand
	SetSalesReturnArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesReturnByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesReturnsQuery       = dynamicresource.SearchQuery
	SalesReturnExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesReturnResult      = dynamicresource.CreateResult
	UpdateSalesReturnResult      = dynamicresource.MutateResult
	DeleteSalesReturnResult      = dynamicresource.MutateResult
	SetSalesReturnArchivedResult = dynamicresource.MutateResult
	GetSalesReturnByIdResult     = dynamicresource.GetOneResult
	SearchSalesReturnsResult     = dynamicresource.SearchResult
	SalesReturnExistsResult      = dynamicresource.ExistsResult
)
