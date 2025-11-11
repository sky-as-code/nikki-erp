package variant

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/inventory/variant/impl"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildVariantDescriptor()),
		deps.Register(impl.NewVariantEntRepository),
		deps.Register(impl.NewVariantServiceImpl),
	)

	return err
}
