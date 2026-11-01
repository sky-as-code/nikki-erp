package invoice

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// InvoiceCrudRepository reads and writes invoices.
type InvoiceCrudRepository interface {
	dynamicresource.CrudRepository
}

// InvoiceCrudDomainService is the CRUD of the resource. Issuing and raising an invoice from a
// source document live on the InvoiceDomainService port.
type InvoiceCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// InvoiceCrudApplicationService is the authorized surface the REST handler serves.
type InvoiceCrudApplicationService interface {
	dynamicresource.CrudApplicationService

	// IssueAction closes a draft. It carries its own permission: closing a draft and minting its
	// number is not the same authority as correcting its note.
	IssueAction(ctx corectx.Context, cmd InvoiceActionCommand) (*InvoiceActionResult, error)
}

// InvoiceActionCommand is the bound request of a custom invoice action.
type InvoiceActionCommand = dynamicresource.UpdateCommand

// InvoiceActionResult answers a custom invoice action with the flat body the legacy action returned.
type InvoiceActionResult = dyn.OpResult[map[string]any]

// InvoiceLineCrudRepository reads and writes invoice lines.
type InvoiceLineCrudRepository interface {
	dynamicresource.CrudRepository
}

// InvoiceLineCrudDomainService is the CRUD of the resource.
type InvoiceLineCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// InvoiceLineCrudApplicationService is the authorized surface the REST handler serves.
type InvoiceLineCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateInvoiceCommand      = dynamicresource.CreateCommand
	UpdateInvoiceCommand      = dynamicresource.UpdateCommand
	DeleteInvoiceCommand      = dynamicresource.DeleteCommand
	SetInvoiceArchivedCommand = dynamicresource.SetArchivedCommand
	GetInvoiceByIdQuery       = dynamicresource.GetByIdQuery
	SearchInvoicesQuery       = dynamicresource.SearchQuery
	InvoiceExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateInvoiceResult      = dynamicresource.CreateResult
	UpdateInvoiceResult      = dynamicresource.MutateResult
	DeleteInvoiceResult      = dynamicresource.MutateResult
	SetInvoiceArchivedResult = dynamicresource.MutateResult
	GetInvoiceByIdResult     = dynamicresource.GetOneResult
	SearchInvoicesResult     = dynamicresource.SearchResult
	InvoiceExistsResult      = dynamicresource.ExistsResult
)

type (
	CreateInvoiceLineCommand      = dynamicresource.CreateCommand
	UpdateInvoiceLineCommand      = dynamicresource.UpdateCommand
	DeleteInvoiceLineCommand      = dynamicresource.DeleteCommand
	SetInvoiceLineArchivedCommand = dynamicresource.SetArchivedCommand
	GetInvoiceLineByIdQuery       = dynamicresource.GetByIdQuery
	SearchInvoiceLinesQuery       = dynamicresource.SearchQuery
	InvoiceLineExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateInvoiceLineResult      = dynamicresource.CreateResult
	UpdateInvoiceLineResult      = dynamicresource.MutateResult
	DeleteInvoiceLineResult      = dynamicresource.MutateResult
	SetInvoiceLineArchivedResult = dynamicresource.MutateResult
	GetInvoiceLineByIdResult     = dynamicresource.GetOneResult
	SearchInvoiceLinesResult     = dynamicresource.SearchResult
	InvoiceLineExistsResult      = dynamicresource.ExistsResult
)
