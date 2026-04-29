package external

import (
	identModels "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itUsr "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

const (
	UserStatusInvited = identModels.UserStatusInvited
	UserStatusActive  = identModels.UserStatusActive
)

type UserExtService = itUsr.UserDomainService
type OrganizationExtService = itOrg.OrganizationDomainService
type OrgUnitExtService = itOrgUnit.OrgUnitDomainService
type GroupExtService = itGrp.GroupDomainService

type GetOrgQuery = itOrg.GetOrgQuery
type GetUserQuery = itUsr.GetUserQuery

type Organization = identModels.Organization
