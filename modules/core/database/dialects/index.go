package dialects

import (
	"database/sql"
	"time"

	"go.bryk.io/pkg/errors"
)

type DialectName string

const (
	MySql    DialectName = "mysql"
	Sqlite   DialectName = "sqlite3"
	Postgres DialectName = "postgres"
)

type DialectOptions struct {
	DialectName    DialectName
	IsDebugEnabled bool

	Database     string
	HostPort     string
	User         string
	Password     string
	IsTlsEnabled bool

	ConnMaxLifetimeSecs uint
	ConnMaxIdleTimeSecs uint
	MaxIdleConns        uint
	MaxOpenConns        uint
}

type DbDialect interface {
	Open(opts DialectOptions) (*sql.DB, error)
}

func OpenConnection(opts DialectOptions) (conn *sql.DB, err error) {
	switch opts.DialectName {
	case MySql:
		conn, err = MysqlDialect{}.Open(opts)
		setConnOptions(conn, opts)
	case Postgres:
		conn, err = PostgresqlDialect{}.Open(opts)
		setConnOptions(conn, opts)
	default:
		err = errors.Errorf("unsupported dialect: %s", opts.DialectName)
		return nil, err
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to open database connection")
	}

	return conn, nil
}

func setConnOptions(conn *sql.DB, opts DialectOptions) {
	conn.SetMaxIdleConns(int(opts.MaxIdleConns))
	conn.SetMaxOpenConns(int(opts.MaxOpenConns))
	conn.SetConnMaxLifetime(time.Duration(opts.ConnMaxLifetimeSecs) * time.Second)
	conn.SetConnMaxIdleTime(time.Duration(opts.ConnMaxIdleTimeSecs) * time.Second)
}
