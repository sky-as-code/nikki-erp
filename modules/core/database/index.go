package database

import (
	"database/sql"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/database/dialects"
	"github.com/sky-as-code/nikki-erp/modules/core/database/orm"
)

// type InitParams struct {
// 	dig.In

// 	Config config.ConfigService
// }

// func InitSubModule(params InitParams) error {
func InitSubModule() error {
	err := registerDbConnection()
	if err != nil {
		return err
	}

	return deps.Invoke(func(configSvc config.ConfigService, dbConn *DbConnection) error {
		dbDialect := configSvc.GetStr(c.DbDialect)
		debugEnabled := configSvc.GetBool(c.DbDebugEnabled)
		return orm.InitEntOrm(dialects.DialectName(dbDialect), dbConn.Db, debugEnabled)
	})
}

type DbConnection struct {
	Db             *sql.DB
	IsDebugEnabled bool
}

func registerDbConnection() error {
	return deps.Register(func(configSvc config.ConfigService) (*DbConnection, error) {
		dbDialect := configSvc.GetStr(c.DbDialect)
		user := configSvc.GetStr(c.DbUser)
		password := configSvc.GetStr(c.DbPassword)
		host := configSvc.GetStr(c.DbHostPort)
		dbname := configSvc.GetStr(c.DbName)
		tlsEnabled := configSvc.GetBool(c.DbTlsEnabled)
		debugEnabled := configSvc.GetBool(c.DbDebugEnabled)

		maxIdleConns := configSvc.GetUint(c.DbMaxIdleConns)
		maxOpenConns := configSvc.GetUint(c.DbMaxOpenConns)
		connMaxLifetimeSecs := configSvc.GetUint(c.DbConnMaxLifetimeMins)
		connMaxIdleTimeSecs := configSvc.GetUint(c.DbConnMaxIdleTimeMins)

		opts := dialects.DialectOptions{
			DialectName:    dialects.DialectName(dbDialect),
			IsDebugEnabled: debugEnabled,

			User:         user,
			Password:     password,
			HostPort:     host,
			Database:     dbname,
			IsTlsEnabled: tlsEnabled,

			MaxIdleConns:        maxIdleConns,
			MaxOpenConns:        maxOpenConns,
			ConnMaxLifetimeSecs: connMaxLifetimeSecs,
			ConnMaxIdleTimeSecs: connMaxIdleTimeSecs,
		}

		conn, err := dialects.OpenConnection(opts)
		if err != nil {
			return nil, err
		}

		return &DbConnection{
			Db:             conn,
			IsDebugEnabled: debugEnabled,
		}, nil
	})
}
