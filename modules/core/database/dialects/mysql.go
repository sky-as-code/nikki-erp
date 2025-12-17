package dialects

import (
	"database/sql"
	"fmt"
	"net/url"
)

type MysqlDialect struct {
}

func (this MysqlDialect) Open(opts DialectOptions) (*sql.DB, error) {
	dsn, err := this.buildDataSourceName(opts)
	if err != nil {
		return nil, err
	}

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (MysqlDialect) buildDataSourceName(opts DialectOptions) (string, error) {
	var tls string

	if opts.IsTlsEnabled {
		tls = "true"
	} else {
		tls = "false"
	}

	// Properly encode the password (optional but safer)
	escapedPassword := url.QueryEscape(opts.Password)

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&tls=%s",
		opts.User, escapedPassword, opts.HostPort, opts.Database, tls)

	return dsn, nil
}
