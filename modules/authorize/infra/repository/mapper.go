package repository

// import (
// 	"github.com/sky-as-code/nikki-erp/common/array"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
// 	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
// 	entRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/revokerequest"
// )

// // START: Resource
// func entToResource(dbResource *ent.Resource) *domain.Resource {
// 	resource := domain.NewResource()
// 	id := model.Id(dbResource.ID)
// 	resource.SetId(&id)
// 	etag := model.Etag(dbResource.Etag)
// 	resource.SetEtag(etag)
// 	name := dbResource.Name
// 	resource.SetName(&name)
// 	if dbResource.Description != "" {
// 		d := dbResource.Description
// 		resource.SetDescription(&d)
// 	}
// 	rt := domain.ResourceOwnerType(dbResource.ResourceType)
// 	resource.SetResourceType(&rt)
// 	ref := dbResource.ResourceRef
// 	resource.SetResourceRef(&ref)
// 	st := domain.ResourceScope(dbResource.ScopeType)
// 	resource.SetScopeType(&st)
// 	resource.GetFieldData().SetAny(basemodel.FieldCreatedAt, model.ModelDateTime(dbResource.CreatedAt.UTC()))

// 	return resource
// }

// func entToResources(dbResources []*ent.Resource) []domain.Resource {
// 	if dbResources == nil {
// 		return nil
// 	}

// 	return array.Map(dbResources, func(dbResource *ent.Resource) domain.Resource {
// 		return *entToResource(dbResource)
// 	})
// }

// // END: Resource

// // START: Action
// func entToAction(dbAction *ent.Action) *domain.Action {
// 	action := domain.NewAction()
// 	id := model.Id(dbAction.ID)
// 	action.SetId(&id)
// 	etag := model.Etag(dbAction.Etag)
// 	action.SetEtag(etag)
// 	n := dbAction.Name
// 	action.SetName(&n)
// 	if dbAction.Description != "" {
// 		d := dbAction.Description
// 		action.SetDescription(&d)
// 	}
// 	rid := model.Id(dbAction.ResourceID)
// 	action.SetResourceId(&rid)
// 	cb := model.Id(dbAction.CreatedBy)
// 	action.SetCreatedBy(&cb)
// 	action.GetFieldData().SetAny(basemodel.FieldCreatedAt, model.ModelDateTime(dbAction.CreatedAt.UTC()))

// 	return action
// }

// func entToActions(dbActions []*ent.Action) []domain.Action {
// 	if dbActions == nil {
// 		return nil
// 	}

// 	return array.Map(dbActions, func(dbAction *ent.Action) domain.Action {
// 		return *entToAction(dbAction)
// 	})
// }

// // END: Action

// // START: Entitlement
// func entToEntitlement(dbEntitlement *ent.Entitlement) *domain.Entitlement {
// 	entitlement := domain.NewEntitlement()
// 	id := model.Id(dbEntitlement.ID)
// 	entitlement.SetId(&id)
// 	etag := model.Etag(dbEntitlement.Etag)
// 	entitlement.SetEtag(etag)
// 	nm := dbEntitlement.Name
// 	entitlement.SetName(&nm)
// 	entitlement.SetDescription(dbEntitlement.Description)
// 	ae := dbEntitlement.ActionExpr
// 	entitlement.SetActionExpr(&ae)
// 	cb := model.Id(dbEntitlement.CreatedBy)
// 	entitlement.SetCreatedBy(&cb)
// 	if dbEntitlement.ActionID != nil && *dbEntitlement.ActionID != "" {
// 		aid := model.Id(*dbEntitlement.ActionID)
// 		entitlement.SetActionId(&aid)
// 	}
// 	if dbEntitlement.ResourceID != nil && *dbEntitlement.ResourceID != "" {
// 		rid := model.Id(*dbEntitlement.ResourceID)
// 		entitlement.SetResourceId(&rid)
// 	}
// 	entitlement.GetFieldData().SetAny(basemodel.FieldCreatedAt, model.ModelDateTime(dbEntitlement.CreatedAt.UTC()))

// 	return entitlement
// }

// func entToEntitlements(dbEntitlements []*ent.Entitlement) []domain.Entitlement {
// 	if dbEntitlements == nil {
// 		return nil
// 	}

// 	return array.Map(dbEntitlements, func(dbEntitlement *ent.Entitlement) domain.Entitlement {
// 		return *entToEntitlement(dbEntitlement)
// 	})
// }

// // END: Entitlement

// // START: Role
// func entToRole(dbRole *ent.Role) *domain.Role {
// 	role := domain.NewRole()
// 	id := model.Id(dbRole.ID)
// 	role.SetId(&id)
// 	etag := model.Etag(dbRole.Etag)
// 	role.SetEtag(etag)
// 	name := dbRole.Name
// 	role.SetName(&name)
// 	role.SetDescription(dbRole.Description)
// 	ot := domain.RoleOwnerType(dbRole.OwnerType)
// 	role.SetOwnerType(&ot)
// 	oref := model.Id(dbRole.OwnerRef)
// 	role.SetOwnerRef(&oref)
// 	role.SetIsRequestable(&dbRole.IsRequestable)
// 	role.SetIsRequiredAttachment(&dbRole.IsRequiredAttachment)
// 	role.SetIsRequiredComment(&dbRole.IsRequiredComment)
// 	cb := model.Id(dbRole.CreatedBy)
// 	role.SetCreatedBy(&cb)
// 	if dbRole.OrgID != nil && *dbRole.OrgID != "" {
// 		oid := model.Id(*dbRole.OrgID)
// 		role.SetOrgId(&oid)
// 	}
// 	role.GetFieldData().SetAny(basemodel.FieldCreatedAt, model.ModelDateTime(dbRole.CreatedAt.UTC()))
// 	return role
// }

// func entToRoles(dbRoles []*ent.Role) []domain.Role {
// 	if dbRoles == nil {
// 		return nil
// 	}

// 	return array.Map(dbRoles, func(dbRole *ent.Role) domain.Role {
// 		return *entToRole(dbRole)
// 	})
// }

// // END: Role

// // START: RoleSuite
// func entToRoleSuite(dbRoleSuite *ent.RoleSuite) *domain.RoleSuite {
// 	roleSuite := &domain.RoleSuite{}
// 	model.MustCopy(dbRoleSuite, roleSuite)

// 	if dbRoleSuite.Edges.Roles != nil {
// 		roleSuite.Roles = array.Map(dbRoleSuite.Edges.Roles, func(dbRole *ent.Role) domain.Role {
// 			return *entToRole(dbRole)
// 		})
// 	}

// 	return roleSuite
// }

// func entToRoleSuites(dbRoleSuites []*ent.RoleSuite) []domain.RoleSuite {
// 	if dbRoleSuites == nil {
// 		return nil
// 	}

// 	return array.Map(dbRoleSuites, func(dbRoleSuite *ent.RoleSuite) domain.RoleSuite {
// 		return *entToRoleSuite(dbRoleSuite)
// 	})
// }

// // END: RoleSuite

// // START: EntitlementAssignment
// func entToEntitlementAssignment(dbEntitlementAssignment *ent.EntitlementAssignment) *domain.EntitlementGrant {
// 	out := domain.NewEntitlementAssignment()
// 	id := model.Id(dbEntitlementAssignment.ID)
// 	out.SetId(&id)
// 	st := domain.EntitlementAssignmentSubjectType(dbEntitlementAssignment.SubjectType)
// 	out.SetSubjectType(&st)
// 	sref := dbEntitlementAssignment.SubjectRef
// 	out.SetSubjectRef(&sref)
// 	if dbEntitlementAssignment.ActionName != nil {
// 		out.SetActionName(dbEntitlementAssignment.ActionName)
// 	}
// 	if dbEntitlementAssignment.ResourceName != nil {
// 		out.SetResourceName(dbEntitlementAssignment.ResourceName)
// 	}
// 	re := dbEntitlementAssignment.ResolvedExpr
// 	out.SetResolvedExpr(&re)
// 	eid := model.Id(dbEntitlementAssignment.EntitlementID)
// 	out.SetEntitlementId(&eid)
// 	out.SetScopeRef(dbEntitlementAssignment.ScopeRef)

// 	return out
// }

// func entFromEffectiveUserEntitlement(dbEffectiveUserEntitlement *ent.EffectiveUserEntitlement) domain.EntitlementGrant {
// 	out := domain.NewEntitlementAssignment()
// 	st := domain.EntitlementAssignmentSubjectTypeNikkiUser
// 	out.SetSubjectType(&st)
// 	uid := dbEffectiveUserEntitlement.UserID
// 	out.SetSubjectRef(&uid)
// 	out.SetActionName(dbEffectiveUserEntitlement.ActionName)
// 	out.SetResourceName(dbEffectiveUserEntitlement.ResourceName)
// 	re := dbEffectiveUserEntitlement.ActionExpr
// 	out.SetResolvedExpr(&re)
// 	if dbEffectiveUserEntitlement.ScopeType != nil {
// 		rst := domain.ResourceScope(*dbEffectiveUserEntitlement.ScopeType)
// 		out.EffectiveResourceScopeType = &rst
// 	}
// 	out.EffectiveEntitlementScopeRef = dbEffectiveUserEntitlement.ScopeRef
// 	return *out
// }

// func entFromEffectiveGroupEntitlement(dbEffectiveGroupEntitlement *ent.EffectiveGroupEntitlement) domain.EntitlementGrant {
// 	out := domain.NewEntitlementAssignment()
// 	st := domain.EntitlementAssignmentSubjectTypeNikkiGroup
// 	out.SetSubjectType(&st)
// 	gid := dbEffectiveGroupEntitlement.GroupID
// 	out.SetSubjectRef(&gid)
// 	out.SetActionName(dbEffectiveGroupEntitlement.ActionName)
// 	out.SetResourceName(dbEffectiveGroupEntitlement.ResourceName)
// 	re := dbEffectiveGroupEntitlement.ActionExpr
// 	out.SetResolvedExpr(&re)
// 	if dbEffectiveGroupEntitlement.ScopeType != nil {
// 		rst := domain.ResourceScope(*dbEffectiveGroupEntitlement.ScopeType)
// 		out.EffectiveResourceScopeType = &rst
// 	}
// 	out.EffectiveEntitlementScopeRef = dbEffectiveGroupEntitlement.ScopeRef
// 	return *out
// }

// func effectiveEntToEntitlementAssignments(dbEffectiveUserEntitlements []*ent.EffectiveUserEntitlement, dbEffectiveGroupEntitlements []*ent.EffectiveGroupEntitlement) []domain.EntitlementGrant {
// 	assignments := make([]domain.EntitlementGrant, 0)

// 	if dbEffectiveUserEntitlements != nil {
// 		assignments = append(assignments, array.Map(dbEffectiveUserEntitlements, func(dbEffectiveUserEntitlement *ent.EffectiveUserEntitlement) domain.EntitlementGrant {
// 			return entFromEffectiveUserEntitlement(dbEffectiveUserEntitlement)
// 		})...)
// 	}

// 	if dbEffectiveGroupEntitlements != nil {
// 		assignments = append(assignments, array.Map(dbEffectiveGroupEntitlements, func(dbEffectiveGroupEntitlement *ent.EffectiveGroupEntitlement) domain.EntitlementGrant {
// 			return entFromEffectiveGroupEntitlement(dbEffectiveGroupEntitlement)
// 		})...)
// 	}
// 	return assignments
// }

// func entToEntitlementAssignments(dbEntitlementAssignments []*ent.EntitlementAssignment) []domain.EntitlementGrant {
// 	entitlementAssignments := make([]domain.EntitlementGrant, len(dbEntitlementAssignments))
// 	for i, dbEntitlementAssignment := range dbEntitlementAssignments {
// 		entitlementAssignments[i] = *entToEntitlementAssignment(dbEntitlementAssignment)
// 	}
// 	return entitlementAssignments
// }

// // END: EntitlementAssignment

// // START: GrantRequest
// func entToGrantRequest(dbGrantRequest *ent.GrantRequest) *domain.GrantRequest {
// 	grantRequest := &domain.GrantRequest{}
// 	model.MustCopy(dbGrantRequest, grantRequest)

// 	if dbGrantRequest.TargetType == entGrantRequest.TargetTypeRole {
// 		grantRequest.TargetType = domain.WrapGrantRequestTargetTypeEnt(entGrantRequest.TargetTypeRole)
// 		grantRequest.TargetRef = dbGrantRequest.TargetRoleID
// 	} else if dbGrantRequest.TargetType == entGrantRequest.TargetTypeSuite {
// 		grantRequest.TargetType = domain.WrapGrantRequestTargetTypeEnt(entGrantRequest.TargetTypeSuite)
// 		grantRequest.TargetRef = dbGrantRequest.TargetSuiteID
// 	}

// 	if dbGrantRequest.CreatedBy != "" {
// 		grantRequest.RequestorId = &dbGrantRequest.CreatedBy
// 	}

// 	return grantRequest
// }

// func entToGrantRequests(dbGrantRequests []*ent.GrantRequest) []domain.GrantRequest {
// 	if dbGrantRequests == nil {
// 		return nil
// 	}

// 	return array.Map(dbGrantRequests, func(dbGrantRequest *ent.GrantRequest) domain.GrantRequest {
// 		return *entToGrantRequest(dbGrantRequest)
// 	})
// }

// // END: GrantRequest

// // START: GrantResponse
// func entToGrantResponse(dbGrantResponse *ent.GrantResponse) *domain.GrantResponse {
// 	grantResponse := &domain.GrantResponse{}
// 	model.MustCopy(dbGrantResponse, grantResponse)

// 	return grantResponse
// }

// func entToGrantResponses(dbGrantResponses []*ent.GrantResponse) []domain.GrantResponse {
// 	if dbGrantResponses == nil {
// 		return nil
// 	}

// 	return array.Map(dbGrantResponses, func(dbGrantResponse *ent.GrantResponse) domain.GrantResponse {
// 		return *entToGrantResponse(dbGrantResponse)
// 	})
// }

// // END: GrantResponse

// // START: PermissionHistory
// func entToPermissionHistory(dbPermissionHistory *ent.PermissionHistory) *domain.PermissionHistory {
// 	permissionHistory := &domain.PermissionHistory{}
// 	model.MustCopy(dbPermissionHistory, permissionHistory)

// 	return permissionHistory
// }

// func entToPermissionHistories(dbPermissionHistories []*ent.PermissionHistory) []domain.PermissionHistory {
// 	if dbPermissionHistories == nil {
// 		return nil
// 	}
// 	return array.Map(dbPermissionHistories, func(dbPermissionHistory *ent.PermissionHistory) domain.PermissionHistory {
// 		return *entToPermissionHistory(dbPermissionHistory)
// 	})
// }

// // END: PermissionHistory

// // START: RevokeRequest
// func entToRevokeRequest(dbRevokeRequest *ent.RevokeRequest) *domain.RevokeRequest {
// 	revokeRequest := &domain.RevokeRequest{}
// 	model.MustCopy(dbRevokeRequest, revokeRequest)

// 	if dbRevokeRequest.CreatedBy != "" {
// 		revokeRequest.RequestorId = &dbRevokeRequest.CreatedBy
// 	}

// 	if dbRevokeRequest.TargetType == entRevokeRequest.TargetTypeRole {
// 		revokeRequest.TargetType = domain.WrapRevokeRequestTargetTypeEnt(entRevokeRequest.TargetTypeRole)
// 		revokeRequest.TargetRef = dbRevokeRequest.TargetRoleID
// 	} else if dbRevokeRequest.TargetType == entRevokeRequest.TargetTypeSuite {
// 		revokeRequest.TargetType = domain.WrapRevokeRequestTargetTypeEnt(entRevokeRequest.TargetTypeSuite)
// 		revokeRequest.TargetRef = dbRevokeRequest.TargetSuiteID
// 	}

// 	return revokeRequest
// }

// func entToRevokeRequests(dbRevokeRequests []*ent.RevokeRequest) []domain.RevokeRequest {
// 	if dbRevokeRequests == nil {
// 		return nil
// 	}

// 	return array.Map(dbRevokeRequests, func(dbRevokeRequest *ent.RevokeRequest) domain.RevokeRequest {
// 		return *entToRevokeRequest(dbRevokeRequest)
// 	})
// }

// func entToRevokeRequestPtrs(dbRevokeRequests []*ent.RevokeRequest) []*domain.RevokeRequest {
// 	if dbRevokeRequests == nil {
// 		return nil
// 	}
// 	return array.Map(dbRevokeRequests, func(dbRevokeRequest *ent.RevokeRequest) *domain.RevokeRequest {
// 		return entToRevokeRequest(dbRevokeRequest)
// 	})
// }

// // END:
