package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// InitDomainServices installs the derived services that carry behavior built-in CRUD cannot express
// (mostly lifecycle transition rules). Each wraps the engine's own service rather than replacing it,
// so ordinary CRUD still runs underneath. Must run after InitDynamicEngines.
func InitDomainServices() error {
	if err := installDerivedService(models.SalesChannelSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSalesChannelDomainService(base)
		}); err != nil {
		return err
	}
	if err := installDerivedService(models.SalesPointSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSalesPointDomainService(base)
		}); err != nil {
		return err
	}
	if err := installDerivedService(models.SalesOrderSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSalesOrderDomainService(base)
		}); err != nil {
		return err
	}
	// The line service carries the quantity rules and snapshot immutability. The framework declares
	// no CHECK constraints, so installing it is the only enforcement.
	if err := installDerivedService(models.SalesOrderLineSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSalesOrderLineDomainService(base)
		}); err != nil {
		return err
	}
	// The pricelist service carries default-uniqueness ("at most one default per org among
	// non-archived rows", a partial uniqueness the framework cannot declare) and the currency guard.
	if err := installDerivedService(models.SalesPricelistSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSalesPricelistDomainService(base)
		}); err != nil {
		return err
	}
	// The item service holds the rules conditional on another field, which the schema cannot
	// express: required target column depends on applies_to, required price fields on
	// calculation_method, and base-pricelist acceptability on the whole derivation graph.
	if err := installDerivedService(models.SalesPricelistItemSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSalesPricelistItemDomainService(base)
		}); err != nil {
		return err
	}
	return initChannelPaymentService()
}

// channelPaymentService is a package variable rather than an engine ResourceService: the junction
// has an engine only to get a repository (see junctionSchemas) and nothing routes to it, so there is
// no request for a derived service to hang off.
var channelPaymentService *services.ChannelPaymentDomainServiceImpl

// ChannelPaymentService is nil until InitDomainServices has run; that is a wiring bug, not a
// request problem, so callers report it as one.
func ChannelPaymentService() (*services.ChannelPaymentDomainServiceImpl, error) {
	if channelPaymentService == nil {
		return nil, errors.New(
			"the sales channel payment service is not initialized; " +
				"SalesModule.Init must run dynamicengines.InitDomainServices")
	}
	return channelPaymentService, nil
}

func initChannelPaymentService() error {
	engine, ok := dynamicresource.Registry().GetEngine(models.SalesChannelPaymentRelSchemaName)
	if !ok {
		return errors.New(
			"the '" + models.SalesChannelPaymentRelSchemaName + "' engine is not registered")
	}
	channelPaymentService = services.NewChannelPaymentDomainService(engine.ResourceRepository())
	return nil
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
