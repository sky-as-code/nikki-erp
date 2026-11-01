// Package paymentprofile declares the CRUD layers of the paymentprofile resource.
package paymentprofile

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// PaymentProfileCrudRepository reads and writes payment profiles.
type PaymentProfileCrudRepository interface {
	dynamicresource.CrudRepository
}

// PaymentProfileCrudDomainService is the CRUD of the resource. Its overrides encrypt "config" into
// "encrypted_config" on every write and decrypt it on every read.
type PaymentProfileCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// PaymentProfileCrudApplicationService is the authorized surface the REST handler serves.
type PaymentProfileCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreatePaymentProfileCommand      = dynamicresource.CreateCommand
	UpdatePaymentProfileCommand      = dynamicresource.UpdateCommand
	DeletePaymentProfileCommand      = dynamicresource.DeleteCommand
	SetPaymentProfileArchivedCommand = dynamicresource.SetArchivedCommand
	GetPaymentProfileByIdQuery       = dynamicresource.GetByIdQuery
	SearchPaymentProfilesQuery       = dynamicresource.SearchQuery
	PaymentProfileExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreatePaymentProfileResult      = dynamicresource.CreateResult
	UpdatePaymentProfileResult      = dynamicresource.MutateResult
	DeletePaymentProfileResult      = dynamicresource.MutateResult
	SetPaymentProfileArchivedResult = dynamicresource.MutateResult
	GetPaymentProfileByIdResult     = dynamicresource.GetOneResult
	SearchPaymentProfilesResult     = dynamicresource.SearchResult
	PaymentProfileExistsResult      = dynamicresource.ExistsResult
)
