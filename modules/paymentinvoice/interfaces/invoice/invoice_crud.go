package invoice

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// InvoiceCrudRepository reads and writes invoices.
type InvoiceCrudRepository interface {
	composable.CrudRepository
}

// InvoiceCrudDomainService is the CRUD of the resource. Issuing and raising an invoice from a
// source document live on the InvoiceDomainService port.
type InvoiceCrudDomainService interface {
	composable.CrudDomainService
}

// InvoiceCrudApplicationService is the authorized surface the REST handler serves.
type InvoiceCrudApplicationService interface {
	composable.CrudApplicationService

	// IssueAction closes a draft. It carries its own permission: closing a draft and minting its
	// number is not the same authority as correcting its note.
	IssueAction(ctx corectx.Context, cmd InvoiceActionCommand) (*InvoiceActionResult, error)
}

// InvoiceActionCommand is the bound request of a custom invoice action.
type InvoiceActionCommand = composable.UpdateCommand

// InvoiceActionResult answers a custom invoice action with the flat body the legacy action returned.
type InvoiceActionResult = dyn.OpResult[map[string]any]

// InvoiceLineCrudRepository reads and writes invoice lines.
type InvoiceLineCrudRepository interface {
	composable.CrudRepository
}

// InvoiceLineCrudDomainService is the CRUD of the resource.
type InvoiceLineCrudDomainService interface {
	composable.CrudDomainService
}

// InvoiceLineCrudApplicationService is the authorized surface the REST handler serves.
type InvoiceLineCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateInvoiceCommand      = composable.CreateCommand
	UpdateInvoiceCommand      = composable.UpdateCommand
	DeleteInvoiceCommand      = composable.DeleteCommand
	SetInvoiceArchivedCommand = composable.SetArchivedCommand
	GetInvoiceByIdQuery       = composable.GetByIdQuery
	SearchInvoicesQuery       = composable.SearchQuery
	InvoiceExistsQuery        = composable.ExistsQuery
)

type (
	CreateInvoiceResult      = composable.CreateResult
	UpdateInvoiceResult      = composable.MutateResult
	DeleteInvoiceResult      = composable.MutateResult
	SetInvoiceArchivedResult = composable.MutateResult
	GetInvoiceByIdResult     = composable.GetOneResult
	SearchInvoicesResult     = composable.SearchResult
	InvoiceExistsResult      = composable.ExistsResult
)

type (
	CreateInvoiceLineCommand      = composable.CreateCommand
	UpdateInvoiceLineCommand      = composable.UpdateCommand
	DeleteInvoiceLineCommand      = composable.DeleteCommand
	SetInvoiceLineArchivedCommand = composable.SetArchivedCommand
	GetInvoiceLineByIdQuery       = composable.GetByIdQuery
	SearchInvoiceLinesQuery       = composable.SearchQuery
	InvoiceLineExistsQuery        = composable.ExistsQuery
)

type (
	CreateInvoiceLineResult      = composable.CreateResult
	UpdateInvoiceLineResult      = composable.MutateResult
	DeleteInvoiceLineResult      = composable.MutateResult
	SetInvoiceLineArchivedResult = composable.MutateResult
	GetInvoiceLineByIdResult     = composable.GetOneResult
	SearchInvoiceLinesResult     = composable.SearchResult
	InvoiceLineExistsResult      = composable.ExistsResult
)
