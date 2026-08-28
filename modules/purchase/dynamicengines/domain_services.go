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

// The derived services this module installs over the engines' default ones.
//
// A derived service is how a resource gets behavior that the built-in CRUD cannot express: the
// order stamps the fields a client may not choose, and the line keeps the stored totals in step
// with what it just wrote. Both wrap the engine's own service rather than replacing it, so
// ordinary CRUD keeps running through the default implementation underneath.

// InitDomainServices installs them. It runs after InitDynamicEngines, whose engines it wraps, and
// after infra/external has bound the ports the line validator depends on.
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
	// The vendor price service exists only to revalidate an UNARCHIVE (section 25). Create and
	// update are guarded through ValidateExtra, which the engine supports for them; set_archived it
	// explicitly does not, so that one check has to be a derived service.
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

// resolveProductLineValidator pulls the two ports the line rules need out of the container.
//
// A missing port is a hard failure rather than a nil validator: the alternative is a module that
// boots and then silently accepts purchase lines for products nobody may buy, in units that convert
// to nothing. Failing at Init names the missing binding while somebody is still looking at it.
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

// resolveLinePricer pulls the unit port out of the container for vendor price resolution.
//
// It resolves the SAME port the line validator holds rather than sharing that validator's copy,
// because the two ask different questions of it — the validator converts into the product's
// inventory unit, the pricer into whatever unit each quote is written in — and a shared field
// would tempt a later change to make one of those depend on the other.
//
// Missing is a hard failure, matching every other Init here: a pricer with no way to convert would
// silently skip every quote written in a unit other than the line's, and a buyer would see prices
// appear for some products and not others with nothing to explain the difference.
func resolveLinePricer() (*services.LinePricer, error) {
	var uoms itExt.UomExtService
	if err := deps.Invoke(func(svc itExt.UomExtService) { uoms = svc }); err != nil {
		return nil, stdErr.Join(
			errors.New("the UoM port is not registered; purchase/infra/external must bind it"), err)
	}
	return services.NewLinePricer(uoms), nil
}

// resolveOrderReferenceValidator pulls the vendor and currency ports out of the container.
//
// A missing port fails Init for the same reason the line ports do: a module that booted without
// them would accept orders naming vendors nobody may order from, denominated in currencies that
// were withdrawn.
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
