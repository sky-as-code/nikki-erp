package dialects

import (
	"database/sql"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"go.bryk.io/pkg/errors"
)

type DialectOptions struct {
	Database     string
	HostPort     string
	User         string
	Password     string
	IsTlsEnabled bool
}

type DbDialect interface {
	BuildDataSourceName(opts DialectOptions) (string, error)
	Open(opts DialectOptions) (*sql.DB, error)
}

type EntDriverOptions struct {
	DialectOptions

	DialectName     string
	ConnMaxLifetime time.Duration
	MaxIdleConns    uint
	MaxOpenConns    uint
}

func NewEntDriver(opts EntDriverOptions) (*entsql.Driver, error) {
	dialectName := opts.DialectName

	switch dialectName {
	case dialect.MySQL:
		conn, err := MysqlDialect{}.Open(opts.DialectOptions)
		if err != nil {
			return nil, err
		}

		setConnOptions(conn, opts)
		return entsql.OpenDB(dialect.MySQL, conn), nil
	case dialect.Postgres:
		conn, err := PostgresqlDialect{}.Open(opts.DialectOptions)
		if err != nil {
			return nil, err
		}

		setConnOptions(conn, opts)
		return entsql.OpenDB(dialect.Postgres, conn), nil
	// case dialect.SQLite:
	default:
		return nil, errors.Errorf("unsupported dialect: %s", dialectName)
	}
}

func setConnOptions(conn *sql.DB, opts EntDriverOptions) {
	conn.SetMaxIdleConns(int(opts.MaxIdleConns))
	conn.SetMaxOpenConns(int(opts.MaxOpenConns))
	conn.SetConnMaxLifetime(opts.ConnMaxLifetime)
}
