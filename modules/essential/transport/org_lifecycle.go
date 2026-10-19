package transport

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	c "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

func initOrgLifecycleHandlers() error {
	svc := orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(uomCatSvc itUomCat.UomCatDomainService, uomSvc itUom.UomDomainService) {
			built = app.NewOrgLifecycleAppService(uomCatSvc, uomSvc)
		})
		return built, err
	})
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, c.EssentialModuleName, svc)
	})
}
