package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/infra/ent"
	dbOrm "github.com/sky-as-code/nikki-erp/modules/core/database/orm"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildPartyDescriptor()),
		orm.RegisterEntity(BuildCommChannelDescriptor()),
		orm.RegisterEntity(BuildRelationshipDescriptor()),
	)

	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(newContactClient),
		deps.Register(NewPartyEntRepository),
		deps.Register(NewCommChannelEntRepository),
		deps.Register(NewRelationshipEntRepository),
	)

	return err
}

func newContactClient(clientOpts *dbOrm.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
