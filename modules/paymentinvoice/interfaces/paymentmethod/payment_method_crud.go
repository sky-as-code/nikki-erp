package paymentmethod

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// PaymentMethodCrudRepository reads and writes payment methods.
type PaymentMethodCrudRepository interface {
	composable.CrudRepository
}

// PaymentMethodCrudDomainService is the CRUD of the resource.
type PaymentMethodCrudDomainService interface {
	composable.CrudDomainService
}

// PaymentMethodCrudApplicationService is the authorized surface the REST handler serves.
type PaymentMethodCrudApplicationService interface {
	composable.CrudApplicationService

	// The published port is served by the same instance, so the usability gates and the CRUD
	// read the same rows through the same repository.
	PaymentMethodAppService
}

type (
	CreatePaymentMethodCommand      = composable.CreateCommand
	UpdatePaymentMethodCommand      = composable.UpdateCommand
	DeletePaymentMethodCommand      = composable.DeleteCommand
	SetPaymentMethodArchivedCommand = composable.SetArchivedCommand
	GetPaymentMethodByIdQuery       = composable.GetByIdQuery
	SearchPaymentMethodsQuery       = composable.SearchQuery
	PaymentMethodExistsQuery        = composable.ExistsQuery
)

type (
	CreatePaymentMethodResult      = composable.CreateResult
	UpdatePaymentMethodResult      = composable.MutateResult
	DeletePaymentMethodResult      = composable.MutateResult
	SetPaymentMethodArchivedResult = composable.MutateResult
	GetPaymentMethodByIdResult     = composable.GetOneResult
	SearchPaymentMethodsResult     = composable.SearchResult
	PaymentMethodExistsResult      = composable.ExistsResult
)
