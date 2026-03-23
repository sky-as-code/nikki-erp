package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type DriveFileServiceImpl struct {
	logger         logging.LoggerService
	config         config.ConfigService
	driveFileRepo  it.DriveFileRepository
	storageAdapter file_storage.FileStorageAdapter
}

func NewDriveFileService(
	logger logging.LoggerService,
	config config.ConfigService,
	driveFileRepo it.DriveFileRepository,
	storageAdapter file_storage.FileStorageAdapter,
) it.DriveFileService {
	return &DriveFileServiceImpl{
		logger:         logger,
		config:         config,
		driveFileRepo:  driveFileRepo,
		storageAdapter: storageAdapter,
	}
}
