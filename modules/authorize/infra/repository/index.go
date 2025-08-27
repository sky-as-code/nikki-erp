package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildResourceDescriptor()),
		orm.RegisterEntity(BuildActionDescriptor()),
		orm.RegisterEntity(BuildEntitlementDescriptor()),
		orm.RegisterEntity(BuildEntitlementAssignmentDescriptor()),
		orm.RegisterEntity(BuildRoleDescriptor()),
		orm.RegisterEntity(BuildRoleSuiteDescriptor()),
		orm.RegisterEntity(BuildGrantRequestDescriptor()),
		orm.RegisterEntity(BuildRevokeRequestDescriptor()),
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
		deps.Register(NewRoleSuiteEntRepository),
		deps.Register(NewGrantRequestEntRepository),
		deps.Register(NewRevokeRequestEntRepository),
	)

	return err
}

func newAuthorizeClient(clientOpts *database.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
