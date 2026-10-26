package cqrs

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
)

// OrgLifecycleService is Purchase's share of Commerce's org lifecycle. Commerce subscribes one
// handler for the whole module, so this part only builds the service it composes.
func OrgLifecycleService() orglifecycle.OrgLifecycleService {
	return app.NewOrgLifecycleAppService()
}
