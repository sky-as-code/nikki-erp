package repository

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
)

// START: Resource
func entToResource(dbResource *ent.Resource) *domain.Resource {
	resource := &domain.Resource{}
	model.MustCopy(dbResource, resource)

	if dbResource.Edges.Actions != nil {
		resource.Actions = entToActions(dbResource.Edges.Actions)
	}

	if dbResource.Edges.Entitlements != nil {
		resource.Entitlements = entToEntitlements(dbResource.Edges.Entitlements)
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

	if dbAction.Edges.Entitlements != nil {
		action.Entitlements = array.Map(dbAction.Edges.Entitlements, func(dbEntitlement *ent.Entitlement) domain.Entitlement {
			return *entToEntitlement(dbEntitlement)
		})
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
	role := &domain.Role{}
	model.MustCopy(dbRole, role)

	return role
}

func entToRoles(dbRoles []*ent.Role) []domain.Role {
	if dbRoles == nil {
		return nil
	}

	return array.Map(dbRoles, func(dbRole *ent.Role) domain.Role {
		return *entToRole(dbRole)
	})
}

// END: Role

// START: RoleSuite
func entToRoleSuite(dbRoleSuite *ent.RoleSuite) *domain.RoleSuite {
	roleSuite := &domain.RoleSuite{}
	model.MustCopy(dbRoleSuite, roleSuite)

	if dbRoleSuite.Edges.Roles != nil {
		roleSuite.Roles = array.Map(dbRoleSuite.Edges.Roles, func(dbRole *ent.Role) domain.Role {
			return *entToRole(dbRole)
		})
	}

	return roleSuite
}

func entToRoleSuites(dbRoleSuites []*ent.RoleSuite) []domain.RoleSuite {
	if dbRoleSuites == nil {
		return nil
	}

	return array.Map(dbRoleSuites, func(dbRoleSuite *ent.RoleSuite) domain.RoleSuite {
		return *entToRoleSuite(dbRoleSuite)
	})
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

func effectiveEntToEntitlementAssignments(dbEffectiveUserEntitlements []*ent.EffectiveUserEntitlement, dbEffectiveGroupEntitlements []*ent.EffectiveGroupEntitlement) []*domain.EntitlementAssignment {
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

func entToEntitlementAssignments(dbEntitlementAssignments []*ent.EntitlementAssignment) []*domain.EntitlementAssignment {
	entitlementAssignments := make([]*domain.EntitlementAssignment, len(dbEntitlementAssignments))
	for i, dbEntitlementAssignment := range dbEntitlementAssignments {
		entitlementAssignments[i] = entToEntitlementAssignment(dbEntitlementAssignment)
	}
	return entitlementAssignments
}

// END: EntitlementAssignment

// START: GrantRequest
func entToGrantRequest(dbGrantRequest *ent.GrantRequest) *domain.GrantRequest {
	grantRequest := &domain.GrantRequest{}
	model.MustCopy(dbGrantRequest, grantRequest)

	if dbGrantRequest.TargetType == entGrantRequest.TargetTypeRole {
		grantRequest.TargetType = domain.WrapGrantRequestTargetTypeEnt(entGrantRequest.TargetTypeRole)
		grantRequest.TargetRef = dbGrantRequest.TargetRoleID
	} else if dbGrantRequest.TargetType == entGrantRequest.TargetTypeSuite {
		grantRequest.TargetType = domain.WrapGrantRequestTargetTypeEnt(entGrantRequest.TargetTypeSuite)
		grantRequest.TargetRef = dbGrantRequest.TargetSuiteID
	}

	if dbGrantRequest.CreatedBy != "" {
		grantRequest.RequestorId = &dbGrantRequest.CreatedBy
	}

	if dbGrantRequest.Edges.Role != nil {
		grantRequest.Role = entToRole(dbGrantRequest.Edges.Role)
	}

	if dbGrantRequest.Edges.RoleSuite != nil {
		grantRequest.RoleSuite = entToRoleSuite(dbGrantRequest.Edges.RoleSuite)
	}

	if dbGrantRequest.Edges.GrantResponses != nil {
		grantRequest.GrantResponses = entToGrantResponses(dbGrantRequest.Edges.GrantResponses)
	}

	return grantRequest
}

func entToGrantRequests(dbGrantRequests []*ent.GrantRequest) []domain.GrantRequest {
	if dbGrantRequests == nil {
		return nil
	}

	return array.Map(dbGrantRequests, func(dbGrantRequest *ent.GrantRequest) domain.GrantRequest {
		return *entToGrantRequest(dbGrantRequest)
	})
}

// END: GrantRequest

// START: GrantResponse
func entToGrantResponse(dbGrantResponse *ent.GrantResponse) *domain.GrantResponse {
	grantResponse := &domain.GrantResponse{}
	model.MustCopy(dbGrantResponse, grantResponse)

	return grantResponse
}

func entToGrantResponses(dbGrantResponses []*ent.GrantResponse) []domain.GrantResponse {
	if dbGrantResponses == nil {
		return nil
	}

	return array.Map(dbGrantResponses, func(dbGrantResponse *ent.GrantResponse) domain.GrantResponse {
		return *entToGrantResponse(dbGrantResponse)
	})
}

// END: GrantResponse

// START: PermissionHistory
func entToPermissionHistory(dbPermissionHistory *ent.PermissionHistory) *domain.PermissionHistory {
	permissionHistory := &domain.PermissionHistory{}
	model.MustCopy(dbPermissionHistory, permissionHistory)

	return permissionHistory
}

// END: PermissionHistory
