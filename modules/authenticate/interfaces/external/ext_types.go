package external

import (
	identDomain "github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itUsr "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

const (
	UserStatusInvited = identDomain.UserStatusInvited
	UserStatusActive  = identDomain.UserStatusActive
)

type UserExtService = itUsr.UserService
type OrganizationExtService = itOrg.OrganizationService
type OrgUnitExtService = itOrgUnit.OrgUnitService
type GroupExtService = itGrp.GroupService

type GetOrgQuery = itOrg.GetOrgQuery
type GetUserQuery = itUsr.GetUserQuery

type Organization = identDomain.Organization
