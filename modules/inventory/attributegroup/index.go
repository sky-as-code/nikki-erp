package attributegroup

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/inventory/attributegroup/impl"
	"github.com/sky-as-code/nikki-erp/modules/inventory/attributegroup/transport"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildAttributeGroupDescriptor()),
		deps.Register(impl.NewAttributeGroupEntRepository),
		deps.Register(impl.NewAttributeGroupServiceImpl),

		transport.InitTransport(),
	)

	return err
}
