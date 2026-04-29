package external

// import (
// 	stdErr "errors"

// 	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
// 	itExt "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/external"
// 	"github.com/sky-as-code/nikki-erp/modules/identity/app"
// 	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
// 	itHier "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
// 	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
// 	itUsr "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
// )

// // This will be replaced with the actual implementation when this application is
// // split into separate microservices.
// type OrganizationExtServiceImpl = app.OrganizationServiceImpl
// type HierarchyExtServiceImpl = app.HierarchyServiceImpl
// type GroupExtServiceImpl = app.GroupServiceImpl
// type UserExtServiceImpl = app.UserServiceImpl

// func InitExternal() error {
// 	err := stdErr.Join(
// 		deps.Register(func(orgSvc itOrg.OrganizationDomainService) itExt.OrganizationExtService {
// 			return orgSvc
// 		}),
// 		deps.Register(func(hierSvc itHier.HierarchyService) itExt.HierarchyExtService {
// 			return hierSvc
// 		}),
// 		deps.Register(func(groupSvc itGrp.GroupDomainService) itExt.GroupExtService {
// 			return groupSvc
// 		}),
// 		deps.Register(func(userSvc itUsr.UserService) itExt.UserExtService {
// 			return userSvc
// 		}),
// 	)

// 	return err
// }
