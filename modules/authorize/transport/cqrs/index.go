package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initResourceHandlers(),
		initActionHandlers(),
		initEntitlementHandlers(),
		initRoleHandlers(),
		// initRoleSuiteHandlers(),
	)
	return err
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
		)
	})
}

// func initRoleSuiteHandlers() error {
// 	deps.Register(NewRoleSuiteHandler)

// 	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *RoleSuiteHandler) error {
// 		ctx := context.Background()
// 		return cqrsBus.SubscribeRequests(
// 			ctx,
// 		)
// 	})
// }
