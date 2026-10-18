// Package transaction declares the CRUD layers of the transaction resource.
package transaction

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// TransactionCrudRepository reads and writes payment transactions.
type TransactionCrudRepository interface {
	composable.CrudRepository
}

// TransactionCrudDomainService is the CRUD of the resource.
type TransactionCrudDomainService interface {
	composable.CrudDomainService
}

// TransactionCrudApplicationService is the authorized surface the REST handler serves.
type TransactionCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTransactionCommand      = composable.CreateCommand
	UpdateTransactionCommand      = composable.UpdateCommand
	DeleteTransactionCommand      = composable.DeleteCommand
	SetTransactionArchivedCommand = composable.SetArchivedCommand
	GetTransactionByIdQuery       = composable.GetByIdQuery
	SearchTransactionsQuery       = composable.SearchQuery
	TransactionExistsQuery        = composable.ExistsQuery
)

type (
	CreateTransactionResult      = composable.CreateResult
	UpdateTransactionResult      = composable.MutateResult
	DeleteTransactionResult      = composable.MutateResult
	SetTransactionArchivedResult = composable.MutateResult
	GetTransactionByIdResult     = composable.GetOneResult
	SearchTransactionsResult     = composable.SearchResult
	TransactionExistsResult      = composable.ExistsResult
)
