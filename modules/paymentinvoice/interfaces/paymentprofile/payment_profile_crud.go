// Package paymentprofile declares the CRUD layers of the paymentprofile resource.
package paymentprofile

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// PaymentProfileCrudRepository reads and writes payment profiles.
type PaymentProfileCrudRepository interface {
	composable.CrudRepository
}

// PaymentProfileCrudDomainService is the CRUD of the resource. Its overrides encrypt "config" into
// "encrypted_config" on every write and decrypt it on every read.
type PaymentProfileCrudDomainService interface {
	composable.CrudDomainService
}

// PaymentProfileCrudApplicationService is the authorized surface the REST handler serves.
type PaymentProfileCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreatePaymentProfileCommand      = composable.CreateCommand
	UpdatePaymentProfileCommand      = composable.UpdateCommand
	DeletePaymentProfileCommand      = composable.DeleteCommand
	SetPaymentProfileArchivedCommand = composable.SetArchivedCommand
	GetPaymentProfileByIdQuery       = composable.GetByIdQuery
	SearchPaymentProfilesQuery       = composable.SearchQuery
	PaymentProfileExistsQuery        = composable.ExistsQuery
)

type (
	CreatePaymentProfileResult      = composable.CreateResult
	UpdatePaymentProfileResult      = composable.MutateResult
	DeletePaymentProfileResult      = composable.MutateResult
	SetPaymentProfileArchivedResult = composable.MutateResult
	GetPaymentProfileByIdResult     = composable.GetOneResult
	SearchPaymentProfilesResult     = composable.SearchResult
	PaymentProfileExistsResult      = composable.ExistsResult
)
