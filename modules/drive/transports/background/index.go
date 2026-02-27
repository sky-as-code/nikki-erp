package background

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	cronJob "github.com/sky-as-code/nikki-erp/modules/core/cron_job"
	"github.com/sky-as-code/nikki-erp/modules/drive/constants"
)

func InitBackgroundHandler() error {
	err := errors.Join(
		initJobHandler(),
	)
	return err
}

func initJobHandler() error {
	deps.Register(NewDriveFileHandler)

	deps.Invoke(func(
		config config.ConfigService,
		cronJobRegistry cronJob.JobRegistry,
		driveFileHandler DriveFileHandler,
	) {
		cronJobRegistry.Register(
			config.GetStr(constants.CrontabDeleteTrashedFile),
			"Delete Trashed File",
			driveFileHandler.DeleteTrashedFile)
	})

	return nil
}
