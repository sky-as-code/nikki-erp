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
			cqrs.NewHandler(handler.Create),
			cqrs.NewHandler(handler.Delete),
			cqrs.NewHandler(handler.GetUserById),
			cqrs.NewHandler(handler.GetUserByEmail),
			cqrs.NewHandler(handler.MustGetActiveUser),
			cqrs.NewHandler(handler.SearchUsers),
			cqrs.NewHandler(handler.Update),
			cqrs.NewHandler(handler.UserExists),
			cqrs.NewHandler(handler.UserExistsMulti),
			cqrs.NewHandler(handler.FindDirectApprover),
		)
	})
}

func initGroupHandlers() error {
	deps.Register(NewGroupHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *GroupHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.AddRemoveUsers),
			cqrs.NewHandler(handler.CreateGroup),
			cqrs.NewHandler(handler.DeleteGroup),
			cqrs.NewHandler(handler.GetGroupById),
			cqrs.NewHandler(handler.SearchGroups),
			cqrs.NewHandler(handler.UpdateGroup),
			cqrs.NewHandler(handler.SearchGroups),
			cqrs.NewHandler(handler.GroupExists),
		)
	})
}

func initOrganizationHandlers() error {
	deps.Register(NewOrganizationHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *OrganizationHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateOrganization),
			cqrs.NewHandler(handler.UpdateOrganization),
			cqrs.NewHandler(handler.DeleteOrganization),
			cqrs.NewHandler(handler.GetOrganizationBySlug),
			cqrs.NewHandler(handler.SearchOrganizations),
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
			cqrs.NewHandler(handler.UpdateHierarchyLevel),
			cqrs.NewHandler(handler.DeleteHierarchyLevel),
			cqrs.NewHandler(handler.GetHierarchyLevelById),
			cqrs.NewHandler(handler.SearchHierarchyLevels),
		)
	})
}
