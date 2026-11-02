// Package quotation declares the layers of the quotation resources: the offer a customer is given
// before it becomes a sale, and the lines it quotes.
package quotation

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type SalesQuotationRepository interface {
	dynamicresource.CrudRepository
}

type SalesQuotationDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesQuotationApplicationService is the authorized surface the REST handler serves.
type SalesQuotationApplicationService interface {
	dynamicresource.CrudApplicationService

	// Convert carries its own permission because it creates a sales order and commits the business
	// to a sale: a role that may draft and correct an offer must not thereby be able to bind one.
	Convert(ctx corectx.Context, cmd QuotationActionCommand) (*dyn.OpResult[any], error)

	// Sending and cancelling are ordinary handling of a document that binds nobody yet, so they
	// ride on update.
	Send(ctx corectx.Context, cmd QuotationActionCommand) (*dyn.OpResult[any], error)
	Cancel(ctx corectx.Context, cmd QuotationActionCommand) (*dyn.OpResult[any], error)
}

type QuotationActionCommand = dynamicresource.UpdateCommand

type SalesQuotationLineRepository interface {
	dynamicresource.CrudRepository
}

type SalesQuotationLineApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateSalesQuotationCommand = dynamicresource.CreateCommand
	UpdateSalesQuotationCommand = dynamicresource.UpdateCommand
	SearchSalesQuotationsQuery  = dynamicresource.SearchQuery
	SearchSalesQuotationsResult = dynamicresource.SearchResult
)
