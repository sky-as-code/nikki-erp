package cron_job

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func InitSubModule() error {
	err := deps.Register(func(logger logging.LoggerService) initCronJobResult {
		return initCronJob(logger)
	})

	return err
}
