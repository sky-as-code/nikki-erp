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
