package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildUserDescriptor()),
		orm.RegisterEntity(BuildGroupDescriptor()),
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(NewUserEntRepository),
		deps.Register(NewGroupEntRepository),
		deps.Register(NewOrganizationEntRepository),
	)

	return err
}
