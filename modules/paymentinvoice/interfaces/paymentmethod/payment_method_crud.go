package paymentmethod

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// PaymentMethodCrudRepository reads and writes payment methods.
type PaymentMethodCrudRepository interface {
	dynamicresource.CrudRepository
}

// PaymentMethodCrudDomainService is the CRUD of the resource.
type PaymentMethodCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// PaymentMethodCrudApplicationService is the authorized surface the REST handler serves.
type PaymentMethodCrudApplicationService interface {
	dynamicresource.CrudApplicationService

	// The published port is served by the same instance, so the usability gates and the CRUD
	// read the same rows through the same repository.
	PaymentMethodAppService
}

type (
	CreatePaymentMethodCommand      = dynamicresource.CreateCommand
	UpdatePaymentMethodCommand      = dynamicresource.UpdateCommand
	DeletePaymentMethodCommand      = dynamicresource.DeleteCommand
	SetPaymentMethodArchivedCommand = dynamicresource.SetArchivedCommand
	GetPaymentMethodByIdQuery       = dynamicresource.GetByIdQuery
	SearchPaymentMethodsQuery       = dynamicresource.SearchQuery
	PaymentMethodExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreatePaymentMethodResult      = dynamicresource.CreateResult
	UpdatePaymentMethodResult      = dynamicresource.MutateResult
	DeletePaymentMethodResult      = dynamicresource.MutateResult
	SetPaymentMethodArchivedResult = dynamicresource.MutateResult
	GetPaymentMethodByIdResult     = dynamicresource.GetOneResult
	SearchPaymentMethodsResult     = dynamicresource.SearchResult
	PaymentMethodExistsResult      = dynamicresource.ExistsResult
)
