package repository

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
)

// START: Resource
func entToResource(dbResource *ent.Resource) *domain.Resource {
	resource := &domain.Resource{}
	model.MustCopy(dbResource, resource)

	if dbResource.Edges.Actions != nil {
		resource.Actions = entToActions(dbResource.Edges.Actions)
	}

	return resource
}

func entToResources(dbResources []*ent.Resource) []domain.Resource {
	if dbResources == nil {
		return nil
	}

	return array.Map(dbResources, func(dbResource *ent.Resource) domain.Resource {
		return *entToResource(dbResource)
	})
}
// END: Resource

// START: Action
func entToAction(dbAction *ent.Action) *domain.Action {
	action := &domain.Action{}
	model.MustCopy(dbAction, action)

	if dbAction.Edges.Resource != nil {
		action.Resource = entToResource(dbAction.Edges.Resource)
	}

	return action
}

func entToActions(dbActions []*ent.Action) []domain.Action {
	if dbActions == nil {
		return nil
	}

	return array.Map(dbActions, func(dbAction *ent.Action) domain.Action {
		return *entToAction(dbAction)
	})
}
// END: Action

// START: Entitlement
func entToEntitlement(dbEntitlement *ent.Entitlement) *domain.Entitlement {
	entitlement := &domain.Entitlement{}
	model.MustCopy(dbEntitlement, entitlement)

	if dbEntitlement.Edges.Action != nil {
		entitlement.Action = entToAction(dbEntitlement.Edges.Action)
	}

	if dbEntitlement.Edges.Resource != nil {
		entitlement.Resource = entToResource(dbEntitlement.Edges.Resource)
	}

	return entitlement
}

func entToEntitlements(dbEntitlements []*ent.Entitlement) []domain.Entitlement {
	if dbEntitlements == nil {
		return nil
	}

	return array.Map(dbEntitlements, func(dbEntitlement *ent.Entitlement) domain.Entitlement {
		return *entToEntitlement(dbEntitlement)
	})
}
// END: Entitlement

// START: Role
func entToRole(dbRole *ent.Role) *domain.Role {
	role := &domain.Role{
		ModelBase: model.ModelBase{
			Id:   &dbRole.ID,
			Etag: &dbRole.Etag,
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &dbRole.CreatedAt,
		},
		Name:                 &dbRole.Name,
		Description:          dbRole.Description,
		OwnerType:            domain.WrapRoleOwnerTypeEnt(dbRole.OwnerType),
		OwnerRef:             &dbRole.OwnerRef,
		IsRequestable:        &dbRole.IsRequestable,
		IsRequiredAttachment: &dbRole.IsRequiredAttachment,
		IsRequiredComment:    &dbRole.IsRequiredComment,
		CreatedBy:            &dbRole.CreatedBy,
	}

	return role
}

func entToRoles(dbRoles []*ent.Role) []*domain.Role {
	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = entToRole(dbRole)
	}
	return roles
}
// END: Role

// START: RoleSuite
func entToRoleSuite(dbRoleSuite *ent.RoleSuite) *domain.RoleSuite {
	roleSuite := &domain.RoleSuite{
		ModelBase: model.ModelBase{
			Id:   &dbRoleSuite.ID,
			Etag: &dbRoleSuite.Etag,
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &dbRoleSuite.CreatedAt,
		},
		Name:                 &dbRoleSuite.Name,
		Description:          &dbRoleSuite.Description,
		OwnerType:            domain.WrapRoleSuiteOwnerTypeEnt(dbRoleSuite.OwnerType),
		OwnerRef:             &dbRoleSuite.OwnerRef,
		IsRequestable:        &dbRoleSuite.IsRequestable,
		IsRequiredAttachment: &dbRoleSuite.IsRequiredAttachment,
		IsRequiredComment:    &dbRoleSuite.IsRequiredComment,
		CreatedBy:            &dbRoleSuite.CreatedBy,
	}

	if dbRoleSuite.Edges.Roles != nil {
		roleSuite.Roles = array.Map(dbRoleSuite.Edges.Roles, func(dbRole *ent.Role) domain.Role {
			return *entToRole(dbRole)
		})
	}

	return roleSuite
}

func entToRoleSuites(dbRoleSuites []*ent.RoleSuite) []domain.RoleSuite {
	roleSuites := make([]domain.RoleSuite, len(dbRoleSuites))
	for i, dbRoleSuite := range dbRoleSuites {
		roleSuites[i] = *entToRoleSuite(dbRoleSuite)
	}
	return roleSuites
}
// END: RoleSuite

// START: EntitlementAssignment
func entToEntitlementAssignment(dbEntitlementAssignment *ent.EntitlementAssignment) *domain.EntitlementAssignment {
	entitlementAssignment := &domain.EntitlementAssignment{
		ModelBase: model.ModelBase{
			Id: &dbEntitlementAssignment.ID,
		},
		SubjectType:   domain.WrapEntitlementAssignmentSubjectTypeEnt(dbEntitlementAssignment.SubjectType),
		SubjectRef:    &dbEntitlementAssignment.SubjectRef,
		ActionName:    dbEntitlementAssignment.ActionName,
		ResourceName:  dbEntitlementAssignment.ResourceName,
		ResolvedExpr:  &dbEntitlementAssignment.ResolvedExpr,
		EntitlementId: &dbEntitlementAssignment.EntitlementID,
	}

	if dbEntitlementAssignment.Edges.Entitlement != nil {
		entitlementAssignment.Entitlement = entToEntitlement(dbEntitlementAssignment.Edges.Entitlement)
	}

	return entitlementAssignment
}

func entFromEffectiveUserEntitlement(dbEffectiveUserEntitlement *ent.EffectiveUserEntitlement) *domain.EntitlementAssignment {
	entitlementAssignment := &domain.EntitlementAssignment{
		SubjectRef:   &dbEffectiveUserEntitlement.UserID,
		ActionName:   dbEffectiveUserEntitlement.ActionName,
		ResourceName: dbEffectiveUserEntitlement.ResourceName,
		Entitlement: &domain.Entitlement{
			Resource: &domain.Resource{},
			ScopeRef: dbEffectiveUserEntitlement.ScopeRef,
		},
	}

	if dbEffectiveUserEntitlement.ScopeType != nil {
		entitlementAssignment.Entitlement.Resource.ScopeType = domain.WrapResourceScopeType(*dbEffectiveUserEntitlement.ScopeType)
	}

	return entitlementAssignment
}

func entFromEffectiveGroupEntitlement(dbEffectiveGroupEntitlement *ent.EffectiveGroupEntitlement) *domain.EntitlementAssignment {
	entitlementAssignment := &domain.EntitlementAssignment{
		SubjectRef:   &dbEffectiveGroupEntitlement.GroupID,
		ActionName:   dbEffectiveGroupEntitlement.ActionName,
		ResourceName: dbEffectiveGroupEntitlement.ResourceName,
		Entitlement: &domain.Entitlement{
			Resource: &domain.Resource{},
			ScopeRef: dbEffectiveGroupEntitlement.ScopeRef,
		},
	}

	if dbEffectiveGroupEntitlement.ScopeType != nil {
		entitlementAssignment.Entitlement.Resource.ScopeType = domain.WrapResourceScopeType(*dbEffectiveGroupEntitlement.ScopeType)
	}

	return entitlementAssignment
}

func entToEntitlementAssignments(dbEffectiveUserEntitlements []*ent.EffectiveUserEntitlement, dbEffectiveGroupEntitlements []*ent.EffectiveGroupEntitlement) []*domain.EntitlementAssignment {
	assignments := make([]*domain.EntitlementAssignment, 0)

	if dbEffectiveUserEntitlements != nil {
		assignments = append(assignments, array.Map(dbEffectiveUserEntitlements, func(dbEffectiveUserEntitlement *ent.EffectiveUserEntitlement) *domain.EntitlementAssignment {
			return entFromEffectiveUserEntitlement(dbEffectiveUserEntitlement)
		})...)
	}

	if dbEffectiveGroupEntitlements != nil {
		assignments = append(assignments, array.Map(dbEffectiveGroupEntitlements, func(dbEffectiveGroupEntitlement *ent.EffectiveGroupEntitlement) *domain.EntitlementAssignment {
			return entFromEffectiveGroupEntitlement(dbEffectiveGroupEntitlement)
		})...)
	}
	return assignments
}
// END: EntitlementAssignment
