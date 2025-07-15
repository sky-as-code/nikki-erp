package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildResourceDescriptor()),
		orm.RegisterEntity(BuildActionDescriptor()),
		orm.RegisterEntity(BuildEntitlementDescriptor()),
		orm.RegisterEntity(BuildEntitlementAssignmentDescriptor()),
		orm.RegisterEntity(BuildRoleDescriptor()),
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(newAuthorizeClient),
		deps.Register(NewResourceEntRepository),
		deps.Register(NewActionEntRepository),
		deps.Register(NewEntitlementEntRepository),
		deps.Register(NewEntitlementAssignmentEntRepository),
		deps.Register(NewRoleEntRepository),
	)

	return err
}

func newAuthorizeClient(clientOpts *db.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
