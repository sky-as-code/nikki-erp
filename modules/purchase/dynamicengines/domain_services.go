package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"
)

// The derived services this module installs over the engines' default ones: the order stamps fields
// a client may not choose, and the line keeps the stored totals in step with what it just wrote.
// Both wrap the engine's own service rather than replacing it.

// InitDomainServices must run after InitDynamicEngines, whose engines it wraps, and after
// infra/external has bound the ports the line validator depends on.
func InitDomainServices() error {
	validator, err := resolveProductLineValidator()
	if err != nil {
		return err
	}
	references, err := resolveOrderReferenceValidator()
	if err != nil {
		return err
	}

	pricer, err := resolveLinePricer()
	if err != nil {
		return err
	}

	vendorPrices, err := resolveVendorPriceValidator()
	if err != nil {
		return err
	}
	SetVendorPriceValidator(vendorPrices)

	// Totals round to the order's own currency from here on, instead of a fixed two places.
	services.SetOrderScaleResolver(references.ScaleFor)

	if err := installDerivedService(models.PurchaseOrderSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewPurchaseOrderDomainService(base, references, validator, pricer)
		}); err != nil {
		return err
	}
	if err := installDerivedService(models.PurchaseOrderLineSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewPurchaseOrderLineDomainService(base, validator, pricer)
		}); err != nil {
		return err
	}
	// The vendor price service exists only to revalidate an unarchive: create and update go through
	// ValidateExtra, which the engine does not support for set_archived.
	if err := installDerivedService(models.VendorProductPriceSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewVendorProductPriceDomainService(base, vendorPrices)
		}); err != nil {
		return err
	}
	return installDerivedService(models.AgreementSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewPurchaseAgreementDomainService(base, references)
		})
}

// resolveProductLineValidator pulls the two ports the line rules need out of the container. A
// missing port fails Init rather than yielding a nil validator, which would silently accept lines
// for unbuyable products in units that convert to nothing.
func resolveProductLineValidator() (*services.ProductLineValidator, error) {
	var products itExt.ProductExtService
	var uoms itExt.UomExtService

	if err := deps.Invoke(func(svc itExt.ProductExtService) { products = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the product port is not registered; purchase/infra/external must bind it"), err)
	}
	if err := deps.Invoke(func(svc itExt.UomExtService) { uoms = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the UoM port is not registered; purchase/infra/external must bind it"), err)
	}
	return services.NewProductLineValidator(products, uoms), nil
}

// resolveLinePricer pulls the unit port out of the container for vendor price resolution. It
// resolves its own handle on the same port the line validator holds, because the two convert into
// different units — the validator into the product's inventory unit, the pricer into whatever unit
// each quote is written in. Missing fails Init: a pricer with no way to convert would silently skip
// every quote written in a unit other than the line's.
func resolveLinePricer() (*services.LinePricer, error) {
	var uoms itExt.UomExtService
	if err := deps.Invoke(func(svc itExt.UomExtService) { uoms = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the UoM port is not registered; purchase/infra/external must bind it"), err)
	}
	return services.NewLinePricer(uoms), nil
}

// resolveOrderReferenceValidator pulls the vendor and currency ports out of the container. Missing
// fails Init, or the module would accept orders naming unusable vendors and withdrawn currencies.
func resolveOrderReferenceValidator() (*services.OrderReferenceValidator, error) {
	var vendors itExt.VendorExtService
	var currencies itExt.CurrencyExtService

	if err := deps.Invoke(func(svc itExt.VendorExtService) { vendors = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the vendor port is not registered; purchase/infra/external must bind it"), err)
	}
	if err := deps.Invoke(func(svc itExt.CurrencyExtService) { currencies = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the currency port is not registered; purchase/infra/external must bind it"), err)
	}
	return services.NewOrderReferenceValidator(vendors, currencies), nil
}

func installDerivedService(
	schemaName string, derive func(drif.DynamicResourceService) drif.DynamicResourceService,
) error {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return errors.New("the '" + schemaName + "' engine is not registered")
	}
	engine.SetResourceService(derive(engine.ResourceService()))
	return nil
}
