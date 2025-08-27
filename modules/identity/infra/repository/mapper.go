package repository

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
)

func entToGroup(dbGroup *ent.Group) *domain.Group {
	group := &domain.Group{}
	model.MustCopy(dbGroup, group)

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
	org := &domain.Organization{}
	model.MustCopy(dbOrg, org)
	return org
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
	user := &domain.User{}
	model.MustCopy(dbUser, user)
	user.AvatarUrl = dbUser.AvatarURL

	if dbUser.Edges.Groups != nil {
		user.Groups = entToGroups(dbUser.Edges.Groups)
	}

	if dbUser.Edges.Hierarchy != nil {
		user.HierarchyId = &dbUser.Edges.Hierarchy.ID
	}

	if dbUser.Edges.Orgs != nil {
		user.Orgs = entToOrganizations(dbUser.Edges.Orgs)
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

func entToHierarchyLevel(dbHierarchy *ent.HierarchyLevel) *domain.HierarchyLevel {
	hierarchy := &domain.HierarchyLevel{}
	model.MustCopy(dbHierarchy, hierarchy)

	if dbHierarchy.Edges.Org != nil {
		hierarchy.Org = entToOrganization(dbHierarchy.Edges.Org)
	}

	if dbHierarchy.Edges.Parent != nil {
		hierarchy.ParentId = &dbHierarchy.Edges.Parent.ID
		hierarchy.Parent = entToHierarchyLevel(dbHierarchy.Edges.Parent)
	}

	if dbHierarchy.Edges.Children != nil {
		hierarchy.Children = entToHierarchyLevels(dbHierarchy.Edges.Children)
	}

	return hierarchy
}

func entToHierarchyLevels(dbHierarchies []*ent.HierarchyLevel) []domain.HierarchyLevel {
	if dbHierarchies == nil {
		return nil
	}
	return array.Map(dbHierarchies, func(entHierarchy *ent.HierarchyLevel) domain.HierarchyLevel {
		return *entToHierarchyLevel(entHierarchy)
	})
}
