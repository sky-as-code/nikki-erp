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
	return installDerivedService(models.SalesPointSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSalesPointDomainService(base)
		})
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
