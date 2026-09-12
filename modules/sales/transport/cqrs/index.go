package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	invconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
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
	)
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
		registry := usagecheck.NewCheckerRegistry()

		if err := registry.Register(
			essconstants.EssentialModuleName,
			usagecheck.ResourceUom,
			usagecheck.NewTableChecker(resolveRepository, uomTables()...),
		); err != nil {
			return err
		}

		if err := registry.Register(
			invconstants.InventoryModuleName,
			usagecheck.ResourceProductVariant,
			usagecheck.NewTableChecker(resolveRepository, variantTables()...),
		); err != nil {
			return err
		}

		return usagecheck.SubscribeHandler(
			context.Background(), bus, salesconstants.SalesModuleName, registry)
	})
}
