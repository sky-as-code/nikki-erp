package cqrs

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/app"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
)

// OrgLifecycleService is Purchase's share of Commerce's org lifecycle. Commerce subscribes one
// handler for the whole module, so this part only builds the service it composes. Lazy, because
// Commerce composes it during Init, before the container can build the numbering ports.
func OrgLifecycleService() orglifecycle.OrgLifecycleService {
	return orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(sequences itExt.SequenceExtService, seeder itExt.OrgSettingsSeedExtService) {
			built = app.NewOrgLifecycleAppService(sequences, seeder)
		})
		return built, err
	})
}
