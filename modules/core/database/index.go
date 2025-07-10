package database

import (
	"time"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/database/dialects"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type InitParams struct {
	dig.In

	Config config.ConfigService
	Logger logging.LoggerService
}

func InitSubModule(params InitParams) error {
	configSvc := params.Config
	logger := params.Logger

	dbDialect := configSvc.GetStr(c.DbDialect)
	user := configSvc.GetStr(c.DbUser)
	password := configSvc.GetStr(c.DbPassword)
	host := configSvc.GetStr(c.DbHostPort)
	dbname := configSvc.GetStr(c.DbName)
	tlsEnabled := configSvc.GetBool(c.DbTlsEnabled)
	debugEnabled := configSvc.GetBool(c.DbDebugEnabled)

	maxIdleConns := configSvc.GetUint(c.DbMaxIdleConns)
	maxOpenConns := configSvc.GetUint(c.DbMaxOpenConns)
	connMaxLifetimeSecs := configSvc.GetUint(c.DbConnMaxLifetimeSecs)

	driver, err := dialects.NewEntDriver(dialects.EntDriverOptions{
		DialectName:     dbDialect,
		MaxIdleConns:    maxIdleConns,
		MaxOpenConns:    maxOpenConns,
		ConnMaxLifetime: time.Duration(connMaxLifetimeSecs) * time.Second,

		DialectOptions: dialects.DialectOptions{
			User:         user,
			Password:     password,
			HostPort:     host,
			Database:     dbname,
			IsTlsEnabled: tlsEnabled,
		},
	})

	if err != nil {
		logger.Error("failed to open database connection", err)
		return err
	}

	err = deps.Register(func() *EntClientOptions {
		return &EntClientOptions{
			Driver:       driver,
			DebugEnabled: debugEnabled,
		}
	})

	return err
}

type EntClientOptions struct {
	Driver       *entsql.Driver
	DebugEnabled bool
}
