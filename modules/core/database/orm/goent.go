package orm

import (
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/pkg/errors"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/database/dialects"
)

type EntDriverOptions struct {
	dialects.DialectOptions

	DialectName  dialects.DialectName
	DebugEnabled bool
}

type EntClientOptions struct {
	Driver       *entsql.Driver
	DebugEnabled bool
}

func InitEntOrm(opts EntDriverOptions) error {
	conn, err := dialects.OpenConnection(opts.DialectName, opts.DialectOptions)
	if err != nil {
		return err
	}

	var driver *entsql.Driver
	switch opts.DialectName {
	case dialect.MySQL:
		driver = entsql.OpenDB(dialect.MySQL, conn)
	case dialect.Postgres:
		driver = entsql.OpenDB(dialect.Postgres, conn)
	default:
		return errors.Errorf("unsupported dialect: %s", opts.DialectName)
	}

	return deps.Register(func() *EntClientOptions {
		return &EntClientOptions{
			Driver:       driver,
			DebugEnabled: opts.DebugEnabled,
		}
	})
}
