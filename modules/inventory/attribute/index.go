package attribute

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/inventory/attribute/impl"
	"github.com/sky-as-code/nikki-erp/modules/inventory/attribute/transport"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildAttributeDescriptor()),
		deps.Register(impl.NewAttributeEntRepository),
		deps.Register(impl.NewAttributeServiceImpl),

		transport.InitTransport(),
	)

	return err

}
