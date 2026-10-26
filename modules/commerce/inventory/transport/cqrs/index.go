package cqrs

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// OrgLifecycleService is Inventory's share of Commerce's org lifecycle. Commerce subscribes one
// handler for the whole module, so this part only builds the service it composes.
func OrgLifecycleService() orglifecycle.OrgLifecycleService {
	return orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(param app.OrgLifecycleParam) {
			built = app.NewOrgLifecycleAppService(param)
		})
		return built, err
	})
}

// UomUsageChecker answers whether an Inventory table still holds a unit of measure. Inventory
// holds a uom_id on every stock movement and balance, so Essential asks Commerce before letting
// a unit be edited or deleted.
func UomUsageChecker() usagecheck.ResourceUsageChecker {
	return usagecheck.NewTableChecker(resolveRepository, uomTables()...)
}
