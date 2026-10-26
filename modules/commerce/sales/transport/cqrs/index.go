package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	commconstants "github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// InitCqrsHandlers subscribes Sales' command handlers. Usage checks and the org lifecycle are
// subscribed once for the whole of Commerce, which composes UomUsageChecker, InventoryUsageRegistry
// and OrgLifecycleService.
//
// Billing must be subscribed BEFORE the job is registered in OnAppStarted: the scheduler
// validates that a job's command name is a registered request type, and rejects the registration
// otherwise.
func InitCqrsHandlers() error {
	return initBillingHandlers()
}

// OrgLifecycleService is Sales' share of Commerce's org lifecycle.
func OrgLifecycleService() orglifecycle.OrgLifecycleService {
	return orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(param app.OrgLifecycleParam) {
			built = app.NewOrgLifecycleAppService(param)
		})
		return built, err
	})
}

func initBillingHandlers() error {
	if err := deps.Register(NewBillingHandler); err != nil {
		return err
	}
	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *BillingHandler) error {
		return cqrsBus.SubscribeRequests(
			context.Background(),
			cqrs.NewHandler(handler.IssueEinvoices),
		)
	})
}

// UomUsageChecker answers whether a Sales table still holds a unit of measure.
func UomUsageChecker() usagecheck.ResourceUsageChecker {
	return usagecheck.NewTableChecker(resolveRepository, uomTables()...)
}

// InventoryUsageRegistry binds a checker to every Inventory resource Commerce guards before a
// delete. Sales and Inventory are parts of one module, so Inventory asks this registry in process
// rather than over the bus. A resource Sales holds no reference to is still registered, with no
// tables, so the registry states every resource it was asked about.
func InventoryUsageRegistry() (*usagecheck.CheckerRegistry, error) {
	registry := usagecheck.NewCheckerRegistry()
	register := func(resource string, tables ...usagecheck.ReferencingTable) error {
		return registry.Register(commconstants.CommerceModuleName, resource,
			usagecheck.NewTableChecker(resolveRepository, tables...))
	}

	err := errors.Join(
		register(usagecheck.ResourceProductVariant, variantTables()...),
		register(usagecheck.ResourceProductTemplate, templateTables()...),
		// No Sales schema names a stock quant: its documents name the variant and the location.
		register(usagecheck.ResourceStockQuant),
	)
	if err != nil {
		return nil, err
	}
	return registry, nil
}
