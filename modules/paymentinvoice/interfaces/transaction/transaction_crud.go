// Package transaction declares the CRUD layers of the transaction resource.
package transaction

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// TransactionCrudRepository reads and writes payment transactions.
type TransactionCrudRepository interface {
	dynamicresource.CrudRepository
}

// TransactionCrudDomainService is the CRUD of the resource.
type TransactionCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// TransactionCrudApplicationService is the authorized surface the REST handler serves.
type TransactionCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTransactionCommand      = dynamicresource.CreateCommand
	UpdateTransactionCommand      = dynamicresource.UpdateCommand
	DeleteTransactionCommand      = dynamicresource.DeleteCommand
	SetTransactionArchivedCommand = dynamicresource.SetArchivedCommand
	GetTransactionByIdQuery       = dynamicresource.GetByIdQuery
	SearchTransactionsQuery       = dynamicresource.SearchQuery
	TransactionExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateTransactionResult      = dynamicresource.CreateResult
	UpdateTransactionResult      = dynamicresource.MutateResult
	DeleteTransactionResult      = dynamicresource.MutateResult
	SetTransactionArchivedResult = dynamicresource.MutateResult
	GetTransactionByIdResult     = dynamicresource.GetOneResult
	SearchTransactionsResult     = dynamicresource.SearchResult
	TransactionExistsResult      = dynamicresource.ExistsResult
)
