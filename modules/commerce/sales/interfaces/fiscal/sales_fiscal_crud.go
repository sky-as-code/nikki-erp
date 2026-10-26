// Package fiscal declares the layers of the resources that turn a sale into a legal document: the
// request for an invoice, the instruction that says how to bill, and the record of each issuance
// attempt.
package fiscal

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type SalesFiscalRequestRepository interface {
	composable.CrudRepository
}

// SalesFiscalRequestApplicationService is the authorized surface the REST handler serves.
type SalesFiscalRequestApplicationService interface {
	composable.CrudApplicationService

	// RequestInvoice is collection-level: it creates the fiscal request, so there is no record to
	// address yet.
	RequestInvoice(ctx corectx.Context, cmd FiscalActionCommand) (*dyn.OpResult[any], error)
}

type SalesBillingInstructionRepository interface {
	composable.CrudRepository
}

// SalesBillingInstructionApplicationService carries the draft-to-ready lifecycle. mark_ready and
// revert_to_draft are their own permissions because releasing an instruction is what lets a legal
// document be issued from it; create, update and cancel are ordinary edits of a draft.
type SalesBillingInstructionApplicationService interface {
	composable.CrudApplicationService

	CreateInstruction(ctx corectx.Context, cmd FiscalActionCommand) (*dyn.OpResult[any], error)
	UpdateInstruction(ctx corectx.Context, cmd FiscalActionCommand) (*dyn.OpResult[any], error)
	MarkReady(ctx corectx.Context, cmd FiscalActionCommand) (*dyn.OpResult[any], error)
	RevertToDraft(ctx corectx.Context, cmd FiscalActionCommand) (*dyn.OpResult[any], error)
	Cancel(ctx corectx.Context, cmd FiscalActionCommand) (*dyn.OpResult[any], error)
}

// FiscalActionCommand is the bound request of a custom fiscal or billing action.
type FiscalActionCommand = composable.UpdateCommand

// The issuance attempt records what a provider answered, so it is read-only.
type SalesBillingIssuanceAttemptRepository interface {
	composable.CrudRepository
}

type SalesBillingIssuanceAttemptApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateSalesFiscalRequestCommand      = composable.CreateCommand
	SearchSalesFiscalRequestsQuery       = composable.SearchQuery
	SearchSalesFiscalRequestsResult      = composable.SearchResult
	CreateSalesBillingInstructionCommand = composable.CreateCommand
	SearchSalesBillingInstructionsQuery  = composable.SearchQuery
	SearchSalesBillingInstructionsResult = composable.SearchResult
)
