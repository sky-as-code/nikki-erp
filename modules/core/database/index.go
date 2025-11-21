package database

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/database/dialects"
	"github.com/sky-as-code/nikki-erp/modules/core/database/orm"
)

type InitParams struct {
	dig.In

	Config config.ConfigService
}

func InitSubModule(params InitParams) error {
	configSvc := params.Config

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

	return orm.InitEntOrm(orm.EntDriverOptions{
		DialectName:  dialects.DialectName(dbDialect),
		DebugEnabled: debugEnabled,
		DialectOptions: dialects.DialectOptions{
			User:         user,
			Password:     password,
			HostPort:     host,
			Database:     dbname,
			IsTlsEnabled: tlsEnabled,

			MaxIdleConns:        maxIdleConns,
			MaxOpenConns:        maxOpenConns,
			ConnMaxLifetimeSecs: connMaxLifetimeSecs,
			ConnMaxIdleTimeSecs: connMaxIdleTimeSecs,
		},
	})
}
