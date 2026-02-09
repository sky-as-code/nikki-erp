package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildUnitDescriptor()),
		orm.RegisterEntity(BuildUnitCategoryDescriptor()),
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(NewUnitEntRepository),
		deps.Register(NewUnitCategoryEntRepository),
	)

	return err
}
