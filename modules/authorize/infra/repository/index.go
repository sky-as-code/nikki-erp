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
		orm.RegisterEntity(BuildActionDescriptor()),
		orm.RegisterEntity(BuildEntitlementAssignmentDescriptor()),
		orm.RegisterEntity(BuildEntitlementDescriptor()),
		orm.RegisterEntity(BuildGrantRequestDescriptor()),
		orm.RegisterEntity(BuildGrantResponseDescriptor()),
		orm.RegisterEntity(BuildPermissionHistoryDescriptor()),
		orm.RegisterEntity(BuildResourceDescriptor()),
		orm.RegisterEntity(BuildRevokeRequestDescriptor()),
		orm.RegisterEntity(BuildRoleDescriptor()),
		orm.RegisterEntity(BuildRoleSuiteDescriptor()),
		orm.RegisterEntity(BuildRoleUserDescriptor()),
		orm.RegisterEntity(BuildRoleSuiteUserDescriptor()),
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(newAuthorizeClient),
		deps.Register(NewActionEntRepository),
		deps.Register(NewEntitlementAssignmentEntRepository),
		deps.Register(NewEntitlementEntRepository),
		deps.Register(NewGrantRequestEntRepository),
		deps.Register(NewGrantResponseEntRepository),
		deps.Register(NewPermissionHistoryEntRepository),
		deps.Register(NewResourceEntRepository),
		deps.Register(NewRevokeRequestEntRepository),
		deps.Register(NewRoleEntRepository),
		deps.Register(NewRoleSuiteEntRepository),
		deps.Register(NewRoleUserEntRepository),
		deps.Register(NewRoleSuiteUserEntRepository),
	)

	return err
}

func newAuthorizeClient(clientOpts *database.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
