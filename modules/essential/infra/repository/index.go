package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	dbOrm "github.com/sky-as-code/nikki-erp/modules/core/database/orm"
	"github.com/sky-as-code/nikki-erp/modules/essential/infra/ent"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildModuleDescriptor()),
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(newEssentialClient),
		deps.Register(NewModuleEntRepository),
	)

	return err
}

func newEssentialClient(clientOpts *dbOrm.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
