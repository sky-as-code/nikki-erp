package commerce

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	invservices "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	invcqrs "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/transport/cqrs"
	purconstants "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	purcqrs "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/transport/cqrs"
	salesconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	salescqrs "github.com/sky-as-code/nikki-erp/modules/commerce/sales/transport/cqrs"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
)

// initCqrsHandlers subscribes Commerce's one usage-check answer and its one org lifecycle handler.
func initCqrsHandlers() error {
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		registry, err := newUsageCheckRegistry()
		if err != nil {
			return err
		}
		if err := usagecheck.SubscribeHandler(
			context.Background(), bus, constants.CommerceModuleName, registry); err != nil {
			return err
		}
		return orglifecycle.RegisterHandlers(context.Background(), bus, constants.CommerceModuleName,
			newOrgLifecycle())
	})
}

// newUsageCheckRegistry binds a checker to every (owner, resource) pair the owners dispatch to
// Commerce. A unit of measure is in use when any part still holds it.
func newUsageCheckRegistry() (*usagecheck.CheckerRegistry, error) {
	registry := usagecheck.NewCheckerRegistry()
	err := registry.Register(essconstants.EssentialModuleName, usagecheck.ResourceUom, anyPartUses(
		invcqrs.UomUsageChecker(), purcqrs.UomUsageChecker(), salescqrs.UomUsageChecker(),
	))
	if err != nil {
		return nil, err
	}
	return registry, nil
}

// anyPartUses answers "used" with the first part's table that still references the resource.
func anyPartUses(checkers ...usagecheck.ResourceUsageChecker) usagecheck.ResourceUsageChecker {
	return usagecheck.CheckerFunc(func(ctx corectx.Context, ref usagecheck.ResourceRef) (bool, string, error) {
		for _, checker := range checkers {
			used, usedBy, err := checker.IsUsed(ctx, ref)
			if err != nil || used {
				return used, usedBy, err
			}
		}
		return false, "", nil
	})
}

// installSameModuleUsage hands Inventory the Purchase and Sales checkers it asks in process before
// deleting a product or a stock balance. Purchase and Sales are not dependants: they are this module.
func installSameModuleUsage() error {
	purchaseRegistry, err := purcqrs.InventoryUsageRegistry()
	if err != nil {
		return err
	}
	salesRegistry, err := salescqrs.InventoryUsageRegistry()
	if err != nil {
		return err
	}
	invservices.SetSameModuleParts(
		invservices.SameModulePart{Name: purconstants.PurchaseModuleName, Registry: purchaseRegistry},
		invservices.SameModulePart{Name: salesconstants.SalesModuleName, Registry: salesRegistry},
	)
	return nil
}

// assertDependantsSubscribed runs at OnAppStarted: a dependant subscribes its usage handler during
// its own Init, and peer init order is nondeterministic, so this is the first point at which every
// module that was going to subscribe has done so.
func assertDependantsSubscribed() error {
	return deps.Invoke(func(bus cqrs.CqrsBus, dependants *modules.ModuleDependantRegistry) error {
		return usagecheck.AssertDependantsSubscribed(bus, dependants, constants.CommerceModuleName)
	})
}

// orgLifecycle runs the parts' org lifecycles as one: a new org gets Inventory's defaults before
// Sales', which name Inventory's operation types; cleanup runs the other way, children first, and
// stops at the first failure because a parent's rows cannot go while a child's still point at them.
type orgLifecycle struct {
	parts []orglifecycle.OrgLifecycleService
}

func newOrgLifecycle() *orgLifecycle {
	return &orgLifecycle{parts: []orglifecycle.OrgLifecycleService{
		invcqrs.OrgLifecycleService(), purcqrs.OrgLifecycleService(), salescqrs.OrgLifecycleService(),
	}}
}

func (this *orgLifecycle) InitOrg(ctx corectx.Context, orgId model.Id) error {
	for _, part := range this.parts {
		if err := part.InitOrg(ctx, orgId); err != nil {
			return err
		}
	}
	return nil
}

func (this *orgLifecycle) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	for i := len(this.parts) - 1; i >= 0; i-- {
		if err := this.parts[i].CleanUpOrg(ctx, orgId); err != nil {
			return err
		}
	}
	return nil
}
