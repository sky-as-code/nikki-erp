package repository

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
)

func entToGroup(dbGroup *ent.Group) *domain.Group {
	group := &domain.Group{
		ModelBase: model.ModelBase{
			Id:   &dbGroup.ID,
			Etag: &dbGroup.Etag,
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &dbGroup.CreatedAt,
			UpdatedAt: dbGroup.UpdatedAt,
		},

		Name:        &dbGroup.Name,
		Description: dbGroup.Description,
		OrgId:       dbGroup.OrgID,
	}

	if dbGroup.Edges.Org != nil {
		group.Org = entToOrganization(dbGroup.Edges.Org)
	}

	return group
}

func entToGroups(dbGroups []*ent.Group) []domain.Group {
	if dbGroups == nil {
		return nil
	}
	return array.Map(dbGroups, func(entGroup *ent.Group) domain.Group {
		return *entToGroup(entGroup)
	})
}

func entToOrganization(dbOrg *ent.Organization) *domain.Organization {
	return &domain.Organization{
		ModelBase: model.ModelBase{
			Id: &dbOrg.ID,
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &dbOrg.CreatedAt,
			UpdatedAt: dbOrg.UpdatedAt,
		},
		DisplayName: &dbOrg.DisplayName,
		Slug:        &dbOrg.Slug,
	}
}

func entToOrganizations(dbOrgs []*ent.Organization) []domain.Organization {
	if dbOrgs == nil {
		return nil
	}
	return array.Map(dbOrgs, func(entOrg *ent.Organization) domain.Organization {
		return *entToOrganization(entOrg)
	})
}

func entToUser(dbUser *ent.User) *domain.User {
	return &domain.User{
		ModelBase: model.ModelBase{
			Id:   &dbUser.ID,
			Etag: &dbUser.Etag,
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		},
		AvatarUrl:           dbUser.AvatarURL,
		DisplayName:         &dbUser.DisplayName,
		Email:               &dbUser.Email,
		FailedLoginAttempts: &dbUser.FailedLoginAttempts,
		LastLoginAt:         dbUser.LastLoginAt,
		LockedUntil:         dbUser.LockedUntil,
		MustChangePassword:  &dbUser.MustChangePassword,
		PasswordChangedAt:   &dbUser.PasswordChangedAt,
		PasswordHash:        &dbUser.PasswordHash,
		Status:              domain.WrapUserStatusEnt(dbUser.Status),

		Groups: entToGroups(dbUser.Edges.Groups),
		Orgs:   entToOrganizations(dbUser.Edges.Orgs),
	}
}

func entToUsers(dbUsers []*ent.User) []domain.User {
	if dbUsers == nil {
		return nil
	}
	return array.Map(dbUsers, func(entUser *ent.User) domain.User {
		return *entToUser(entUser)
	})
}
