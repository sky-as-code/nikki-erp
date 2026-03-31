package external

import (
	identDomain "github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itHier "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itUsr "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type UserExtService = itUsr.UserService
type OrganizationExtService = itOrg.OrganizationService
type HierarchyExtService = itHier.HierarchyService
type GroupExtService = itGrp.GroupService

type GetOrgQuery = itOrg.GetOrgQuery
type GetUserQuery = itUsr.GetUserQuery

type Organization = identDomain.Organization
