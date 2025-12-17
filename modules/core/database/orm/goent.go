package orm

import (
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/pkg/errors"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/database/dialects"
)

type EntClientOptions struct {
	Driver       *entsql.Driver
	DebugEnabled bool
}

func InitEntOrm(dbDialect dialects.DialectName, conn *sql.DB, isDebugEnabled bool) error {
	var driver *entsql.Driver
	switch dbDialect {
	case dialect.MySQL:
		driver = entsql.OpenDB(dialect.MySQL, conn)
	case dialect.Postgres:
		driver = entsql.OpenDB(dialect.Postgres, conn)
	default:
		return errors.Errorf("unsupported dialect: %s", dbDialect)
	}

	return deps.Register(func() *EntClientOptions {
		return &EntClientOptions{
			Driver:       driver,
			DebugEnabled: isDebugEnabled,
		}
	})
}
