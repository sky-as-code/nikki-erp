package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
)

// NewEntitlementCrudRepository is handed the composable default by the iam_entitlement onion.
func NewEntitlementCrudRepository(base composable.CrudRepository) itEnt.EntitlementCrudRepository {
	return &EntitlementCrudRepositoryImpl{CrudRepository: base}
}

type EntitlementCrudRepositoryImpl struct {
	composable.CrudRepository
}
