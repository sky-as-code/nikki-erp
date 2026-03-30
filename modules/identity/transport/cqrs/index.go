package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initUserHandlers(),
		initGroupHandlers(),
		initOrganizationHandlers(),
		initHierarchyHandlers(),
	)
	return err
}

func initUserHandlers() error {
	deps.Register(NewUserHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *UserHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateUser),
			cqrs.NewHandler(handler.DeleteUser),
			cqrs.NewHandler(handler.UserExists),
			cqrs.NewHandler(handler.GetActiveUser),
			cqrs.NewHandler(handler.GetUser),
			cqrs.NewHandler(handler.SearchUsers),
			cqrs.NewHandler(handler.UpdateUser),
		)
	})
}

func initGroupHandlers() error {
	deps.Register(NewGroupHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *GroupHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateGroup),
			cqrs.NewHandler(handler.DeleteGroup),
			cqrs.NewHandler(handler.GetGroup),
			cqrs.NewHandler(handler.GroupExists),
			cqrs.NewHandler(handler.ManageGroupUsers),
			cqrs.NewHandler(handler.SearchGroups),
			cqrs.NewHandler(handler.UpdateGroup),
		)
	})
}

func initOrganizationHandlers() error {
	deps.Register(NewOrganizationHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *OrganizationHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateOrg),
			cqrs.NewHandler(handler.DeleteOrg),
			cqrs.NewHandler(handler.GetOrg),
			cqrs.NewHandler(handler.OrgExists),
			cqrs.NewHandler(handler.ManageOrgUsers),
			cqrs.NewHandler(handler.SearchOrgs),
			cqrs.NewHandler(handler.UpdateOrg),
		)
	})
}

func initHierarchyHandlers() error {
	deps.Register(NewHierarchyHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *HierarchyHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateHierarchyLevel),
			cqrs.NewHandler(handler.DeleteHierarchyLevel),
			cqrs.NewHandler(handler.GetHierarchyLevel),
			cqrs.NewHandler(handler.HierarchyLevelExists),
			cqrs.NewHandler(handler.ManageHierarchyLevelUsers),
			cqrs.NewHandler(handler.SearchHierarchyLevels),
			cqrs.NewHandler(handler.UpdateHierarchyLevel),
		)
	})
}
