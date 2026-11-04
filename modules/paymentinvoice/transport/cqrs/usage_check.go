package cqrs

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	modconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	pinvsettings "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/external"
)

// initUsageCheckHandlers subscribes the module's answers to Essential's usage checks. Essential
// asks every dependant about each resource it guards, and a dependant without a checker answers
// with an error that blocks the delete, so the unit of measure is answered too.
func initUsageCheckHandlers() error {
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		registry, err := newUsageCheckRegistry()
		if err != nil {
			return err
		}
		return usagecheck.SubscribeHandler(context.Background(), bus, modconstants.PaymentInvoiceModuleName, registry)
	})
}

func newUsageCheckRegistry() (*usagecheck.CheckerRegistry, error) {
	registry := usagecheck.NewCheckerRegistry()
	// No table of this module holds a uom_id: an invoice line carries a quantity, not a unit.
	if err := registry.Register(essconstants.EssentialModuleName, usagecheck.ResourceUom,
		usagecheck.CheckerFunc(func(corectx.Context, usagecheck.ResourceRef) (bool, string, error) {
			return false, "", nil
		})); err != nil {
		return nil, err
	}
	err := registry.Register(essconstants.EssentialModuleName, usagecheck.ResourceSequenceDefinition,
		SequenceUsageChecker(resolveSettingsReader))
	return registry, err
}

// ReaderResolver returns the settings reader at check time: checkers are registered during Init,
// before the container can build the settings services.
type ReaderResolver func() (itExt.OrgSettingsReaderExtService, error)

// SequenceUsageChecker answers "used" while the org's settings hold the definition as its invoice
// series. The invoice series is org-scoped, so a definition without an org is never held here.
func SequenceUsageChecker(resolve ReaderResolver) usagecheck.ResourceUsageChecker {
	return usagecheck.CheckerFunc(func(ctx corectx.Context, ref usagecheck.ResourceRef) (bool, string, error) {
		id := ref.Identifier[usagecheck.IdentifierKeyId]
		orgId := ref.Identifier[usagecheck.IdentifierKeyOrgId]
		if id == "" || orgId == "" {
			return false, "", nil
		}
		reader, err := resolve()
		if err != nil {
			return false, "", err
		}
		held, err := services.InvoiceSequenceIdOf(ctx, reader, model.Id(orgId))
		if err != nil {
			return false, "", err
		}
		if string(held) == id {
			return true, modconstants.PaymentInvoiceModuleName + "." + pinvsettings.OrgSettingInvoiceNumberSequenceId, nil
		}
		return false, "", nil
	})
}

func resolveSettingsReader() (itExt.OrgSettingsReaderExtService, error) {
	var reader itExt.OrgSettingsReaderExtService
	err := deps.Invoke(func(found itExt.OrgSettingsReaderExtService) {
		reader = found
	})
	return reader, err
}
