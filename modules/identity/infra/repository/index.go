package repository

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/common/orm"
	entGroup "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/group"
	entUser "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/user"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(entUser.Label, BuildUserDescriptor()),
		orm.RegisterEntity(entGroup.Label, BuildGroupDescriptor()),
	)
	return err
}
