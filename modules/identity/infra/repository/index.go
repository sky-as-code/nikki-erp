package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	dbOrm "github.com/sky-as-code/nikki-erp/modules/core/database/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
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
		deps.Register(newIdentityClient),
		deps.Register(NewUserEntRepository),
		deps.Register(NewGroupEntRepository),
		deps.Register(NewOrganizationEntRepository),
		deps.Register(NewHierarchyLevelEntRepository),
	)

	return err
}

func newIdentityClient(clientOpts *dbOrm.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
