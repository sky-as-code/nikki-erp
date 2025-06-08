package database

import (
	"fmt"
	"net/url"

	_ "github.com/lib/pq"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
)

type InitParams struct {
	dig.In

	Config config.ConfigService
	Logger logging.LoggerService
}

func InitSubModule(params InitParams) (*ent.Client, error) {
	configSvc := params.Config
	logger := params.Logger

	user := configSvc.GetStr(c.DbUser)
	password := configSvc.GetStr(c.DbPassword)
	host := configSvc.GetStr(c.DbHostPort)
	dbname := configSvc.GetStr(c.DbName)
	sslmode := configSvc.GetStr(c.DbPgSslMode)

	// Properly encode the password (optional but safer)
	escapedPassword := url.QueryEscape(password)

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		user, escapedPassword, host, dbname, sslmode)

	client, err := ent.Open("postgres", dsn, ent.Debug())
	if err != nil {
		logger.Errorf("failed opening connection to postgres: %v", err)
		return nil, err
	}
	// defer client.Close()

	return client, nil
}
