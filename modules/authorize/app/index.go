package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/fault"
	itAction "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/action"
	itEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement"
	itAssignment "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement_assignment"
	itPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/permission_history"
	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/resource"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitServices() error {
	var (
		ent itEntitlement.EntitlementService
		asg itAssignment.EntitlementAssignmentService
	)

	err := errors.Join(
		deps.Register(NewActionServiceImpl),
		deps.Register(NewAuthorizeServiceImpl),
		deps.Register(NewGrantRequestServiceImpl),
		deps.Register(NewResourceServiceImpl),
		deps.Register(NewRevokeRequestServiceImpl),
		deps.Register(NewRoleServiceImpl),
		deps.Register(NewRoleSuiteServiceImpl),
	)
	fault.PanicOnErr(err)

	err = errors.Join(err,
		deps.Register(func(
			actionSvc itAction.ActionService,
			cqrsBus cqrs.CqrsBus,
			entRepo itEntitlement.EntitlementRepository,
			permHistRepo itPermissionHistory.PermissionHistoryRepository,
			resourceSvc itResource.ResourceService,
		) itEntitlement.EntitlementService {
			ent = &EntitlementServiceImpl{
				actionService:         actionSvc,
				cqrsBus:               cqrsBus,
				entitlementRepo:       entRepo,
				permissionHistoryRepo: permHistRepo,
				resourceService:       resourceSvc,
				assignmentService:     nil,
			}
			return ent
		}),
		deps.Register(func(
			entAsgRepo itAssignment.EntitlementAssignmentRepository,
			permHistRepo itPermissionHistory.PermissionHistoryRepository,
			cqrsBus cqrs.CqrsBus,
			resourceSvc itResource.ResourceService,
		) itAssignment.EntitlementAssignmentService {
			asg = &EntitlementAssignmentServiceImpl{
				cqrsBus:                   cqrsBus,
				entitlementService:        nil,
				entitlementAssignmentRepo: entAsgRepo,
				permissionHistoryRepo:     permHistRepo,
				resourceService:           resourceSvc,
			}
			return asg
		}),
	)
	fault.PanicOnErr(err)

	err = errors.Join(
		deps.Invoke(func(svc itEntitlement.EntitlementService) {}),
		deps.Invoke(func(svc itAssignment.EntitlementAssignmentService) {}),
	)
	fault.PanicOnErr(err)

	return deps.Invoke(func() {
		if e, ok := ent.(*EntitlementServiceImpl); ok {
			e.assignmentService = asg
		}
		if a, ok := asg.(*EntitlementAssignmentServiceImpl); ok {
			a.entitlementService = ent
		}
	})
}
