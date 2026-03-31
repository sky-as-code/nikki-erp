package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
)

func InitRepositories() error {
	err := stdErr.Join(
	// orm.RegisterEntity(BuildUserDescriptor()),
	// orm.RegisterEntity(BuildGroupDescriptor()),
	// orm.RegisterEntity(BuildOrganizationDescriptor()),
	// orm.RegisterEntity(BuildHierarchyLevelDescriptor()),
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(newIdentityClient),
		// deps.Invoke(registerIdentitySearchPredicates),
		// deps.Register(NewUserEntRepository),
		deps.Register(NewUserDynamicRepository),
		deps.Register(NewGroupDynamicRepository),
		deps.Register(NewOrganizationDynamicRepository),
		deps.Register(NewHierarchyDynamicRepository),
	)

	return err
}

func newIdentityClient(clientOpts *db.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
