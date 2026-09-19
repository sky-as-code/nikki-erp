// Package external binds Inventory's outbound ports to the services other modules and the core
// publish. It is the only package in Inventory that may import another module; everything else
// depends on interfaces/, so splitting the module into its own process changes this file alone.
package external

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"

	invMessage "github.com/sky-as-code/nikki-erp/modules/inventory/infra/external/message"
	itExt "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/external"
	itMessage "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/message"
)

// InitExternal binds every port Inventory consumes. It runs before the engines are created,
// because a derived service resolves its ports at construction time.
func InitExternal() error {
	return deps.Register(
		func(publisher pubsub.Publisher) itMessage.IntegrationEventPublisher {
			// An adapter, not a hand-over: the broker takes bytes on a topic, and nothing above this
			// layer knows integration events are JSON.
			return invMessage.NewPublisher(publisher)
		},
		func(uomSvc itUom.UomConversionAppService) itExt.UomConversionExtService {
			// A hand-over: the upstream service already has exactly the method the port declares.
			return uomSvc
		},
	)
}
