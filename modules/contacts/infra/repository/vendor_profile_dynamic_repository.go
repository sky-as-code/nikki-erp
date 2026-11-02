package repository

import (
	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewVendorProfileRepository is handed the composable default by the onion. The type exists so a bespoke
// query has a home; it adds nothing today.
func NewVendorProfileRepository(base dynamicresource.CrudRepository) itVendor.VendorProfileRepository {
	return &VendorProfileRepositoryImpl{CrudRepository: base}
}

type VendorProfileRepositoryImpl struct {
	dynamicresource.CrudRepository
}
