package unit

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/impl"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/transport"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildUnitDescriptor()),
		deps.Register(impl.NewUnitEntRepository),
		deps.Register(impl.NewUnitServiceImpl),
		transport.InitTransport(),
	)

	return err
}
