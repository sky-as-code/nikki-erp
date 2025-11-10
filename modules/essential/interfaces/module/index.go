package module

// import (
// 	stdErr "errors"

// 	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
// 	"github.com/sky-as-code/nikki-erp/common/orm"
// 	"github.com/sky-as-code/nikki-erp/modules/essential/module/impl"
// )

// func InitSubModule() error {
// 	err := stdErr.Join(
// 		orm.RegisterEntity(impl.BuildModuleDescriptor()),
// 		deps.Register(impl.NewModuleEntRepository),
// 		deps.Register(impl.NewModuleServiceImpl),
// 	)

// 	return err
// }
