package dialects

import (
	"database/sql"
	"fmt"
	"net/url"

	"entgo.io/ent/dialect"
)

type PostgresqlDialect struct {
}

func (PostgresqlDialect) BuildDataSourceName(opts DialectOptions) (string, error) {
	var sslmode string

	if opts.IsTlsEnabled {
		sslmode = "require"
	} else {
		sslmode = "disable"
	}

	// Properly encode the password (optional but safer)
	escapedPassword := url.QueryEscape(opts.Password)

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		opts.User, escapedPassword, opts.HostPort, opts.Database, sslmode)

	return dsn, nil
}

func (this PostgresqlDialect) Open(opts DialectOptions) (*sql.DB, error) {
	dsn, err := this.BuildDataSourceName(opts)
	if err != nil {
		return nil, err
	}

	conn, err := sql.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
