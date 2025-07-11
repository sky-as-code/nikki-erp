package repository

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
)

// START: Resource
func entToResource(dbResource *ent.Resource) *domain.Resource {
	resource := &domain.Resource{
		ModelBase: model.ModelBase{
			Id:   &dbResource.ID,
			Etag: &dbResource.Etag,
		},
		Name:         &dbResource.Name,
		Description:  &dbResource.Description,
		ResourceType: domain.WrapResourceTypeEnt(dbResource.ResourceType),
		ResourceRef:  &dbResource.ResourceRef,
		ScopeType:    domain.WrapResourceScopeTypeEnt(dbResource.ScopeType),
		Actions:      []domain.Action{},
	}

	// Convert actions if they are loaded
	if dbResource.Edges.Actions != nil {
		resource.Actions = array.Map(dbResource.Edges.Actions, func(dbAction *ent.Action) domain.Action {
			return domain.Action{
				ModelBase: model.ModelBase{
					Id:   &dbAction.ID,
					Etag: &dbAction.Etag,
				},
				Name: &dbAction.Name,
			}
		})
	}

	return resource
}

func entToResources(dbResources []*ent.Resource) []domain.Resource {
	resources := make([]domain.Resource, len(dbResources))
	for i, dbResource := range dbResources {
		resources[i] = *entToResource(dbResource)
	}
	return resources
}

// END: Resource

// START: Action
func entToAction(dbAction *ent.Action) *domain.Action {
	if dbAction == nil {
		return nil
	}

	action := &domain.Action{
		ModelBase: model.ModelBase{
			Id:   &dbAction.ID,
			Etag: &dbAction.Etag,
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &dbAction.CreatedAt,
		},
		Name:        &dbAction.Name,
		ResourceId:  &dbAction.ResourceID,
		Description: &dbAction.Description,
		CreatedBy:   &dbAction.CreatedBy,
	}

	if dbAction.Edges.Resource != nil {
		action.Resource = entToResource(dbAction.Edges.Resource)
	}

	return action
}

func entToActions(dbActions []*ent.Action) []domain.Action {
	actions := make([]domain.Action, len(dbActions))
	for i, dbAction := range dbActions {
		actions[i] = *entToAction(dbAction)
	}
	return actions
}

// END: Action

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

// END: Role

// START: Entitlement
func entToEntitlement(dbEntitlement *ent.Entitlement) *domain.Entitlement {
	entitlement := &domain.Entitlement{
		ModelBase: model.ModelBase{
			Id:   &dbEntitlement.ID,
			Etag: &dbEntitlement.Etag,
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &dbEntitlement.CreatedAt,
		},
		ActionId:    dbEntitlement.ActionID,
		ActionExpr:  &dbEntitlement.ActionExpr,
		Name:        dbEntitlement.Name,
		Description: dbEntitlement.Description,
		// SubjectType: domain.WrapEntitlementSubjectTypeEnt(dbEntitlement.SubjectType),
		// SubjectRef:  &dbEntitlement.SubjectRef,
		ScopeRef:    dbEntitlement.ScopeRef,
		ResourceId:  dbEntitlement.ResourceID,
		CreatedBy:   &dbEntitlement.CreatedBy,
	}

	return entitlement
}

// END: Entitlement
