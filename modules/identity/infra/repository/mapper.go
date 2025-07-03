package repository

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
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
			Id:   &dbOrg.ID,
			Etag: &dbOrg.Etag,
		},
		AuditableBase: model.AuditableBase{
			CreatedAt: &dbOrg.CreatedAt,
			UpdatedAt: dbOrg.UpdatedAt,
		},
		Address:     dbOrg.Address,
		DisplayName: &dbOrg.DisplayName,
		LegalName:   dbOrg.LegalName,
		PhoneNumber: dbOrg.PhoneNumber,
		Slug:        &dbOrg.Slug,
		Status:      domain.WrapOrgStatusEnt(dbOrg.Status),
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
	user := &domain.User{
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
		StatusId:            &dbUser.StatusID,
	}

	if dbUser.Edges.Groups != nil {
		user.Groups = entToGroups(dbUser.Edges.Groups)
	}

	if dbUser.Edges.Hierarchy != nil {
		user.HierarchyId = &dbUser.Edges.Hierarchy.ID
	}

	if dbUser.Edges.Orgs != nil {
		user.Orgs = entToOrganizations(dbUser.Edges.Orgs)
	}

	if dbUser.Edges.UserStatus != nil {
		user.StatusValue = dbUser.Edges.UserStatus.Value
		user.Status = entToUserStatus(dbUser.Edges.UserStatus)
	}
	return user
}

func entToUsers(dbUsers []*ent.User) []domain.User {
	if dbUsers == nil {
		return nil
	}
	return array.Map(dbUsers, func(entUser *ent.User) domain.User {
		return *entToUser(entUser)
	})
}

func entToUserStatus(dbStatus *ent.UserStatusEnum) *domain.UserStatus {
	return &domain.UserStatus{
		Enum: *enum.AnyToEnum(*dbStatus),
	}
}
