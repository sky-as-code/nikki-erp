package dialects

import (
	"database/sql"
	"time"

	"go.bryk.io/pkg/errors"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
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
	case orm.DialectPostgres:
		dsn := buildPgDsn(opts.DialectOptions)
		conn, err := sql.Open(orm.DialectPostgres, dsn)
		if err != nil {
			return nil, err
		}
		setConnOptions(conn, opts)
		return orm.NewPgClient(conn), nil

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
	case orm.DialectMySql:
		dsn = buildMysqlDsn(opts.DialectOptions)
		conn, err = sql.Open(orm.DialectMySql, dsn)
	case orm.DialectPostgres:
		dsn = buildPgDsn(opts.DialectOptions)
		conn, err = sql.Open(orm.DialectPostgres, dsn)
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
