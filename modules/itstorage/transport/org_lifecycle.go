// Package transport subscribes IT Storage to the org lifecycle commands and mounts its routes.
package transport

import (
	"context"
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/app"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/transport/restful"
)

func InitTransport() error {
	return stdErr.Join(
		restful.InitRestfulHandlers(),
		initOrgLifecycleHandlers(),
	)
}

// initOrgLifecycleHandlers resolves the services lazily: the onions must not be built during
// Init().
func initOrgLifecycleHandlers() error {
	svc := orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(spaceSvc itSpace.SpaceDomainService, fileSvc itFile.FileDomainService) {
			built = app.NewOrgLifecycleAppService(spaceSvc, fileSvc)
		})
		return built, err
	})
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, c.ItStorageModuleName, svc)
	})
}
