package repository

import (
	"github.com/thoas/go-funk"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
)

func entToGroup(entGroup *ent.Group) *domain.Group {
	return &domain.Group{
		ModelBase: model.ModelBase{
			Id: model.WrapId(entGroup.ID),
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &entGroup.CreatedAt,
			CreatedBy: model.WrapId(entGroup.CreatedBy),
			UpdatedAt: &entGroup.UpdatedAt,
			UpdatedBy: model.WrapNillableId(entGroup.UpdatedBy),
		},
		Name:        &entGroup.Name,
		Description: entGroup.Description,
	}
}

func entToGroups(entGroups []*ent.Group) []*domain.Group {
	if entGroups == nil {
		return nil
	}
	groups := funk.Map(entGroups, entToGroup)
	return groups.([]*domain.Group)
}

func entToOrganization(entOrg *ent.Organization) *domain.Organization {
	return &domain.Organization{
		ModelBase: model.ModelBase{
			Id: model.WrapId(entOrg.ID),
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &entOrg.CreatedAt,
			CreatedBy: model.WrapId(entOrg.CreatedBy),
			UpdatedAt: &entOrg.UpdatedAt,
			UpdatedBy: model.WrapNillableId(entOrg.UpdatedBy),
		},
		DisplayName: &entOrg.DisplayName,
		Slug:        model.WrapSlug(entOrg.Slug),
	}
}

func entToOrganizations(entOrgs []*ent.Organization) []*domain.Organization {
	if entOrgs == nil {
		return nil
	}
	orgs := funk.Map(entOrgs, entToOrganization)
	return orgs.([]*domain.Organization)
}

func entToUser(entUser *ent.User) *domain.User {
	return &domain.User{
		ModelBase: model.ModelBase{
			Id:   model.WrapId(entUser.ID),
			Etag: model.WrapEtag(entUser.Etag),
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &entUser.CreatedAt,
			CreatedBy: model.WrapId(entUser.CreatedBy),
			UpdatedAt: &entUser.UpdatedAt,
			UpdatedBy: model.WrapNillableId(entUser.UpdatedBy),
		},
		AvatarUrl:           entUser.AvatarURL,
		DisplayName:         &entUser.DisplayName,
		Email:               &entUser.Email,
		FailedLoginAttempts: &entUser.FailedLoginAttempts,
		LastLoginAt:         entUser.LastLoginAt,
		LockedUntil:         entUser.LockedUntil,
		MustChangePassword:  &entUser.MustChangePassword,
		PasswordChangedAt:   &entUser.PasswordChangedAt,
		PasswordHash:        &entUser.PasswordHash,
		Status:              domain.WrapUserStatusEnt(entUser.Status),

		Groups: entToGroups(entUser.Edges.Groups),
		Orgs:   entToOrganizations(entUser.Edges.Orgs),
	}
}

func entToUsers(entUsers []*ent.User) []*domain.User {
	if entUsers == nil {
		return nil
	}
	users := funk.Map(entUsers, entToUser)
	return users.([]*domain.User)
}
