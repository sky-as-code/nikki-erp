package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
)

// NewEntitlementCrudApplicationService is handed the composable default by the iam_entitlement
// onion. The resource has no custom action: archiving is the built-in route.
func NewEntitlementCrudApplicationService(
	base dynamicresource.CrudApplicationService,
) itEnt.EntitlementCrudApplicationService {
	return &EntitlementCrudApplicationServiceImpl{CrudApplicationService: base}
}

type EntitlementCrudApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
