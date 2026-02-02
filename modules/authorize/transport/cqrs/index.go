package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initAuthorizeHandlers(),
		initResourceHandlers(),
		initActionHandlers(),
		initEntitlementHandlers(),
		initEntitlementAssignmentHandlers(),
		initRoleHandlers(),
		initRoleSuiteHandlers(),
	)
	return err
}

func initAuthorizeHandlers() error {
	deps.Register(NewAuthorizeHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *AuthorizeHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.IsAuthorized),
			cqrs.NewHandler(handler.PermissionSnapshot),
		)
	})
}

func initResourceHandlers() error {
	deps.Register(NewResourceHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *ResourceHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateResource),
			cqrs.NewHandler(handler.UpdateResource),
			cqrs.NewHandler(handler.GetResourceByName),
			cqrs.NewHandler(handler.SearchResources),
		)
	})
}

func initActionHandlers() error {
	deps.Register(NewActionHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *ActionHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateAction),
			cqrs.NewHandler(handler.UpdateAction),
			cqrs.NewHandler(handler.GetActionById),
			cqrs.NewHandler(handler.SearchActions),
		)
	})
}

func initEntitlementHandlers() error {
	deps.Register(NewEntitlementHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *EntitlementHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateEntitlement),
			cqrs.NewHandler(handler.EntitlementExists),
			cqrs.NewHandler(handler.UpdateEntitlement),
			cqrs.NewHandler(handler.GetEntitlementById),
			cqrs.NewHandler(handler.GetAllEntitlementByIds),
			cqrs.NewHandler(handler.SearchEntitlements),
		)
	})
}

func initEntitlementAssignmentHandlers() error {
	deps.Register(NewEntitlementAssignmentHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *EntitlementAssignmentHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.GetAllEntitlementAssignmentBySubject),
		)
	})
}

func initRoleHandlers() error {
	deps.Register(NewRoleHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *RoleHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateRole),
			cqrs.NewHandler(handler.GetRoleById),
			cqrs.NewHandler(handler.SearchRoles),
			cqrs.NewHandler(handler.GetRolesBySubject),
		)
	})
}

func initRoleSuiteHandlers() error {
	deps.Register(NewRoleSuiteHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *RoleSuiteHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.GetRoleSuitesBySubject),
			cqrs.NewHandler(handler.GetRoleSuiteById),
		)
	})
}
