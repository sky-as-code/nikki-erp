package authorize

// import (
// 	"errors"

// 	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
// 	"github.com/sky-as-code/nikki-erp/common/semver"
// 	"github.com/sky-as-code/nikki-erp/modules"
// 	app "github.com/sky-as-code/nikki-erp/modules/authorize/app"
// 	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	repo "github.com/sky-as-code/nikki-erp/modules/authorize/infra/repository"
// 	transport "github.com/sky-as-code/nikki-erp/modules/authorize/transport"
// )

// // ModuleSingleton is the exported symbol that will be looked up by the plugin loader
// var ModuleSingleton modules.InCodeModule = &AuthorizeModule{}

// type AuthorizeModule struct {
// }

// // LabelKey implements NikkiModule.
// func (*AuthorizeModule) LabelKey() string {
// 	return "authorize.moduleLabel"
// }

// // Name implements NikkiModule.
// func (*AuthorizeModule) Name() string {
// 	return "authorize"
// }

// // Deps implements NikkiModule.
// func (*AuthorizeModule) Deps() []string {
// 	return []string{
// 		"identity",
// 	}
// }

// // Version implements NikkiModule.
// func (*AuthorizeModule) Version() semver.SemVer {
// 	return *semver.MustParseSemVer("v1.0.0")
// }

// // Init implements NikkiModule.
// func (*AuthorizeModule) Init() error {
// 	err := errors.Join(
// 		repo.InitRepositories(),
// 		app.InitServices(),
// 		transport.InitTransport(),
// 	)

// 	return err
// }

// func (*AuthorizeModule) RegisterModels() error {
// 	return errors.Join(
// 		dmodel.RegisterSchemaB(domain.ResourceSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.ActionSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.EntitlementSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.EntitlementRoleRelSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.RoleUserRelSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.RoleSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.GrantRequestSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.GrantResponseSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.RevokeRequestSchemaBuilder()),
// 		dmodel.RegisterSchemaB(domain.PermissionHistorySchemaBuilder()),
// 	)
// }
