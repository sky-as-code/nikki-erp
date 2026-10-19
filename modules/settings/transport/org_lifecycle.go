// Package transport subscribes Settings to the org lifecycle commands.
package transport

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/settings/app"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

func InitOrgLifecycleHandlers() error {
	svc := orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(settingsSvc it.SettingsDomainService) {
			built = app.NewOrgLifecycleAppService(settingsSvc)
		})
		return built, err
	})
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, c.SettingsModuleName, svc)
	})
}
