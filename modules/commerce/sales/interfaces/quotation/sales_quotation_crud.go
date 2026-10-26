// Package quotation declares the layers of the quotation resources: the offer a customer is given
// before it becomes a sale, and the lines it quotes.
package quotation

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type SalesQuotationRepository interface {
	composable.CrudRepository
}

type SalesQuotationDomainService interface {
	composable.CrudDomainService
}

// SalesQuotationApplicationService is the authorized surface the REST handler serves.
type SalesQuotationApplicationService interface {
	composable.CrudApplicationService

	// Convert carries its own permission because it creates a sales order and commits the business
	// to a sale: a role that may draft and correct an offer must not thereby be able to bind one.
	Convert(ctx corectx.Context, cmd QuotationActionCommand) (*dyn.OpResult[any], error)

	// Sending and cancelling are ordinary handling of a document that binds nobody yet, so they
	// ride on update.
	Send(ctx corectx.Context, cmd QuotationActionCommand) (*dyn.OpResult[any], error)
	Cancel(ctx corectx.Context, cmd QuotationActionCommand) (*dyn.OpResult[any], error)
}

type QuotationActionCommand = composable.UpdateCommand

type SalesQuotationLineRepository interface {
	composable.CrudRepository
}

type SalesQuotationLineApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateSalesQuotationCommand = composable.CreateCommand
	UpdateSalesQuotationCommand = composable.UpdateCommand
	SearchSalesQuotationsQuery  = composable.SearchQuery
	SearchSalesQuotationsResult = composable.SearchResult
)
