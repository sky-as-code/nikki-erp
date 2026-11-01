package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
)

// NewEntitlementCrudRepository is handed the composable default by the iam_entitlement onion.
func NewEntitlementCrudRepository(base dynamicresource.CrudRepository) itEnt.EntitlementCrudRepository {
	return &EntitlementCrudRepositoryImpl{CrudRepository: base}
}

type EntitlementCrudRepositoryImpl struct {
	dynamicresource.CrudRepository
}
