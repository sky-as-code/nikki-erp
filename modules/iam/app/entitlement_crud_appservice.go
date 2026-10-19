package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
)

// NewEntitlementCrudApplicationService is handed the composable default by the iam_entitlement
// onion. The resource has no custom action: archiving is the built-in route.
func NewEntitlementCrudApplicationService(base composable.CrudApplicationService) itEnt.EntitlementCrudApplicationService {
	return &EntitlementCrudApplicationServiceImpl{CrudApplicationService: base}
}

type EntitlementCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
}
