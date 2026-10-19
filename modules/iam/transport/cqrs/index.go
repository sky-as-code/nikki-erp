package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initUserHandlers(),
		initPermissionHandlers(),
		initOrgLifecycleHandlers(),
	)
	return err
}

func initOrgLifecycleHandlers() error {
	svc := orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(entitlementSvc itEnt.EntitlementCrudDomainService, orgUnitSvc itOrgUnit.OrgUnitCrudDomainService) {
			built = app.NewOrgLifecycleAppService(entitlementSvc, orgUnitSvc)
		})
		return built, err
	})
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, c.IamModuleName, svc)
	})
}

func initUserHandlers() error {
	deps.Register(NewUserHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *UserHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.GetEnabledUser),
		)
	})
}

func initPermissionHandlers() error {
	deps.Register(NewPermissionHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *PermissionHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.IsAuthorized),
			cqrs.NewHandler(handler.GetUserEntitlements),
		)
	})
}
