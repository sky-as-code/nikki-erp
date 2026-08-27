package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// The derived services this module installs over the engines' default ones.
//
// A derived service is how a resource gets behavior the built-in CRUD cannot express: the channel
// and the sales point each carry a lifecycle whose transitions are refused or allowed by rules, and
// a plain update would let a client write any status over any other. Both wrap the engine's own
// service rather than replacing it, so ordinary CRUD keeps running through the default
// implementation underneath.

// InitDomainServices installs them. It runs after InitDynamicEngines, whose engines it wraps.
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
	// The line service carries the two invariants the database cannot: BR 55's quantity rules and
	// BR 11's snapshot immutability. Installing it is what makes them enforced at all - the
	// framework declares no CHECK constraints, so there is no second line of defence.
	if err := installDerivedService(models.SalesOrderLineSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSalesOrderLineDomainService(base)
		}); err != nil {
		return err
	}
	return initChannelPaymentService()
}

// channelPaymentService is the junction's service, reached through a package variable rather than
// through an engine's ResourceService.
//
// The junction has an engine only so that it has a repository (see junctionSchemas), and nothing
// routes to it, so there is no request for a derived service to be installed on. Resolving it once
// here is what connects the application layer to it — the same shape paymentinvoice uses for its
// order service.
var channelPaymentService *services.ChannelPaymentDomainServiceImpl

// ChannelPaymentService answers the application layer's mapping operations. It is nil until
// InitDomainServices has run, which is a wiring bug rather than a request problem, so callers
// report it as one.
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
