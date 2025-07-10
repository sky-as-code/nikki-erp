package enum

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildEnumDescriptor()),
		deps.Register(NewEnumEntRepository),
		deps.Register(NewEnumServiceImpl),
	)

	return err
}
