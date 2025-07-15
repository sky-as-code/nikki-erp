package enum

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/enum/impl"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildEnumDescriptor()),
		deps.Register(impl.NewEnumEntRepository),
		deps.Register(impl.NewEnumServiceImpl),
	)

	return err
}
