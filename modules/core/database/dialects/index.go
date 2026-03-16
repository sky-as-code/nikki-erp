package dialects

import (
	"database/sql"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/orm"
	"go.bryk.io/pkg/errors"
)

const (
	DialectMySql    = "mysql"
	DialectPostgres = "postgres"
)

type DialectOptions struct {
	Database     string
	HostPort     string
	User         string
	Password     string
	IsTlsEnabled bool
}

func InitDbClient(opts ClientOptions) (orm.DbClient, error) {
	switch opts.DialectName {
	case dialect.Postgres:
		dsn := buildPgDsn(opts.DialectOptions)
		conn, err := sql.Open(DialectPostgres, dsn)
		if err != nil {
			return nil, err
		}
		setConnOptions(conn, opts)
		return orm.NewPgClient(conn), nil

	// case dialect.MySQL:
	// case dialect.SQLite:
	default:
		return nil, errors.Errorf("unsupported dialect: %s", opts.DialectName)
	}
}

type ClientOptions struct {
	DialectOptions

	DialectName     string
	ConnMaxLifetime time.Duration
	MaxIdleConns    uint
	MaxOpenConns    uint
}

func NewEntDriver(opts ClientOptions) (*entsql.Driver, error) {
	conn, err := OpenConnection(opts)
	if err != nil {
		return nil, err
	}

	setConnOptions(conn, opts)
	return entsql.OpenDB(opts.DialectName, conn), nil
}

func OpenConnection(opts ClientOptions) (*sql.DB, error) {
	var conn *sql.DB
	var err error
	var dsn string

	switch opts.DialectName {
	case dialect.MySQL:
		dsn = buildMysqlDsn(opts.DialectOptions)
		conn, err = sql.Open(DialectMySql, dsn)
	case dialect.Postgres:
		dsn = buildPgDsn(opts.DialectOptions)
		conn, err = sql.Open(DialectPostgres, dsn)
	// case dialect.SQLite:
	default:
		return nil, errors.Errorf("unsupported dialect: %s", opts.DialectName)
	}

	return conn, err
}

func setConnOptions(conn *sql.DB, opts ClientOptions) {
	conn.SetMaxIdleConns(int(opts.MaxIdleConns))
	conn.SetMaxOpenConns(int(opts.MaxOpenConns))
	conn.SetConnMaxLifetime(opts.ConnMaxLifetime)
}
