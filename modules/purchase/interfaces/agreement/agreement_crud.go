// Package agreement declares the CRUD layers of the purchase agreement and its lines.
package agreement

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ActionCommand is the bound request of a custom action: path, body and the resolved org.
type ActionCommand = composable.UpdateCommand

// ActionResult answers a custom action with the body the legacy engine action returned: a
// mutation summary, the created record, or the action's own report.
type ActionResult = dyn.OpResult[any]

// AgreementCrudRepository reads and writes purchase agreements.
type AgreementCrudRepository interface {
	composable.CrudRepository
}

// AgreementCrudDomainService is the CRUD of the resource. Its Create stamps code and status, its Delete
// refuses an agreement that is neither a draft nor cancelled.
type AgreementCrudDomainService interface {
	composable.CrudDomainService
}

// AgreementCrudApplicationService is the authorized surface the REST handler serves.
type AgreementCrudApplicationService interface {
	composable.CrudApplicationService

	ConfirmAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	CloseAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	CancelAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)

	// CreateRfqAction raises a purchase order from the agreement, so it carries "create".
	CreateRfqAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
}

type (
	CreateAgreementCommand      = composable.CreateCommand
	UpdateAgreementCommand      = composable.UpdateCommand
	DeleteAgreementCommand      = composable.DeleteCommand
	SetAgreementArchivedCommand = composable.SetArchivedCommand
	GetAgreementByIdQuery       = composable.GetByIdQuery
	SearchAgreementsQuery       = composable.SearchQuery
	AgreementExistsQuery        = composable.ExistsQuery
)

// AgreementLineCrudRepository reads and writes purchase agreement lines.
type AgreementLineCrudRepository interface {
	composable.CrudRepository
}

// AgreementLineCrudDomainService is the CRUD of the resource.
type AgreementLineCrudDomainService interface {
	composable.CrudDomainService
}

// AgreementLineCrudApplicationService is the authorized surface the REST handler serves.
type AgreementLineCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateAgreementLineCommand      = composable.CreateCommand
	UpdateAgreementLineCommand      = composable.UpdateCommand
	DeleteAgreementLineCommand      = composable.DeleteCommand
	SetAgreementLineArchivedCommand = composable.SetArchivedCommand
	GetAgreementLineByIdQuery       = composable.GetByIdQuery
	SearchAgreementLinesQuery       = composable.SearchQuery
	AgreementLineExistsQuery        = composable.ExistsQuery
)
