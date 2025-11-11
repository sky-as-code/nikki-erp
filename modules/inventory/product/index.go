package product

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/impl"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/transport"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildProductDescriptor()),
		deps.Register(impl.NewProductEntRepository),
		deps.Register(impl.NewProductServiceImpl),

		transport.InitTransport(),
	)

	return err
}
