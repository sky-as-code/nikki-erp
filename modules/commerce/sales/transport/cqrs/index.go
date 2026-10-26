package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	invconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	salesconstants "github.com/sky-as-code/nikki-erp/modules/sales/constants"
)

// InitCqrsHandlers subscribes Sales' command handlers.
//
// Billing must be subscribed BEFORE the job is registered in OnAppStarted: the scheduler
// validates that a job's command name is a registered request type, and rejects the registration
// otherwise.
func InitCqrsHandlers() error {
	return errors.Join(
		initBillingHandlers(),
		initUsageCheckHandlers(),
		initOrgLifecycleHandlers(),
	)
}

func initOrgLifecycleHandlers() error {
	svc := orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(param app.OrgLifecycleParam) {
			built = app.NewOrgLifecycleAppService(param)
		})
		return built, err
	})
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, salesconstants.SalesModuleName, svc)
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

// initUsageCheckHandlers subscribes Sales' answer to usage checks. One subscription serves every
// upstream resource Sales references; the registry routes by (source module, resource).
func initUsageCheckHandlers() error {
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		registry, err := newUsageCheckRegistry()
		if err != nil {
			return err
		}
		return usagecheck.SubscribeHandler(
			context.Background(), bus, salesconstants.SalesModuleName, registry)
	})
}

// newUsageCheckRegistry binds a checker to every (owner, resource) pair the owners dispatch to
// Sales. Inventory asks each of its dependants about every resource it guards, so a pair Sales
// holds no reference to is still registered, with no tables, to answer "unused": an unregistered
// resource is answered with an error, and an unanswered check refuses the delete.
func newUsageCheckRegistry() (*usagecheck.CheckerRegistry, error) {
	registry := usagecheck.NewCheckerRegistry()
	register := func(owner, resource string, tables ...usagecheck.ReferencingTable) error {
		return registry.Register(owner, resource, usagecheck.NewTableChecker(resolveRepository, tables...))
	}

	err := errors.Join(
		register(essconstants.EssentialModuleName, usagecheck.ResourceUom, uomTables()...),
		register(invconstants.InventoryModuleName, usagecheck.ResourceProductVariant, variantTables()...),
		register(invconstants.InventoryModuleName, usagecheck.ResourceProductTemplate, templateTables()...),
		// No Sales schema names a stock quant: its documents name the variant and the location.
		register(invconstants.InventoryModuleName, usagecheck.ResourceStockQuant),
	)
	if err != nil {
		return nil, err
	}
	return registry, nil
}
