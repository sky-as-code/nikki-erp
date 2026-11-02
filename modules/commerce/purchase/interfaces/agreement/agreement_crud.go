// Package agreement declares the CRUD layers of the purchase agreement and its lines.
package agreement

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ActionCommand is the bound request of a custom action: path, body and the resolved org.
type ActionCommand = dynamicresource.UpdateCommand

// ActionResult answers a custom action with the body the legacy engine action returned: a
// mutation summary, the created record, or the action's own report.
type ActionResult = dyn.OpResult[any]

// AgreementCrudRepository reads and writes purchase agreements.
type AgreementCrudRepository interface {
	dynamicresource.CrudRepository
}

// AgreementCrudDomainService is the CRUD of the resource. Its Create stamps code and status, its Delete
// refuses an agreement that is neither a draft nor cancelled.
type AgreementCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// AgreementCrudApplicationService is the authorized surface the REST handler serves.
type AgreementCrudApplicationService interface {
	dynamicresource.CrudApplicationService

	ConfirmAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	CloseAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	CancelAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)

	// CreateRfqAction raises a purchase order from the agreement, so it carries "create".
	CreateRfqAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
}

type (
	CreateAgreementCommand      = dynamicresource.CreateCommand
	UpdateAgreementCommand      = dynamicresource.UpdateCommand
	DeleteAgreementCommand      = dynamicresource.DeleteCommand
	SetAgreementArchivedCommand = dynamicresource.SetArchivedCommand
	GetAgreementByIdQuery       = dynamicresource.GetByIdQuery
	SearchAgreementsQuery       = dynamicresource.SearchQuery
	AgreementExistsQuery        = dynamicresource.ExistsQuery
)

// AgreementLineCrudRepository reads and writes purchase agreement lines.
type AgreementLineCrudRepository interface {
	dynamicresource.CrudRepository
}

// AgreementLineCrudDomainService is the CRUD of the resource.
type AgreementLineCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// AgreementLineCrudApplicationService is the authorized surface the REST handler serves.
type AgreementLineCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateAgreementLineCommand      = dynamicresource.CreateCommand
	UpdateAgreementLineCommand      = dynamicresource.UpdateCommand
	DeleteAgreementLineCommand      = dynamicresource.DeleteCommand
	SetAgreementLineArchivedCommand = dynamicresource.SetArchivedCommand
	GetAgreementLineByIdQuery       = dynamicresource.GetByIdQuery
	SearchAgreementLinesQuery       = dynamicresource.SearchQuery
	AgreementLineExistsQuery        = dynamicresource.ExistsQuery
)
