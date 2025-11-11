package attributevalue

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/inventory/attributevalue/impl"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildAttributeValueDescriptor()),
		deps.Register(impl.NewAttributeValueEntRepository),
		deps.Register(impl.NewAttributeValueServiceImpl),
	)

	return err
}
