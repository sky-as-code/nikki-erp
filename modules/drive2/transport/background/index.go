package background

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	cronJob "github.com/sky-as-code/nikki-erp/modules/core/cron_job"
	"github.com/sky-as-code/nikki-erp/modules/drive2/constants"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
)

const drive2ModuleName = "drive2"

func InitBackgroundHandler() error {
	return initJobHandler()
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
			"Delete Trashed File (drive2)",
			driveFileHandler.DeleteTrashedFile)
	})

	return nil
}

type DriveFileHandler interface {
	DeleteTrashedFile(ctx context.Context) error
}

type driveFileHandler struct {
	driveFileService it.DriveFileService
}

func NewDriveFileHandler(driveFileService it.DriveFileService) DriveFileHandler {
	return &driveFileHandler{driveFileService: driveFileService}
}

func (this *driveFileHandler) DeleteTrashedFile(ctx context.Context) error {
	reqCtx := corectx.NewRequestContextM(ctx, drive2ModuleName)
	return this.driveFileService.DeleteTrashedDriveFile(reqCtx)
}
