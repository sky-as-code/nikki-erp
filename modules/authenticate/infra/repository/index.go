package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/infra/ent"
	dbOrm "github.com/sky-as-code/nikki-erp/modules/core/database/orm"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildAttemptDescriptor()),
		orm.RegisterEntity(BuildPasswordStoreDescriptor()),
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(newAuthenticateClient),
		deps.Register(NewAttemptEntRepository),
		deps.Register(NewPasswordStoreEntRepository),
	)

	return err
}

func newAuthenticateClient(clientOpts *dbOrm.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
