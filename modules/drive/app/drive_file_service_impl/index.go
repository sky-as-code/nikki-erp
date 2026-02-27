package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/cqrs_bus/identity_cqrs"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type DriveFileServiceImpl struct {
	logger         logging.LoggerService
	config         config.ConfigService
	driveFileRepo  it.DriveFileRepository
	permissionSvc  it.DriveFilePermissionService
	storageAdapter file_storage.FileStorageAdapter
	identityCqrs   identity_cqrs.IdentityCqrsAdapter
}

func NewDriveFileService(
	logger logging.LoggerService,
	config config.ConfigService,
	driveFileRepo it.DriveFileRepository,
	permissionSvc it.DriveFilePermissionService,
	storageAdapter file_storage.FileStorageAdapter,
	identityCqrsAdapter identity_cqrs.IdentityCqrsAdapter,
) it.DriveFileService {
	return &DriveFileServiceImpl{
		logger:         logger,
		config:         config,
		driveFileRepo:  driveFileRepo,
		permissionSvc:  permissionSvc,
		storageAdapter: storageAdapter,
		identityCqrs:   identityCqrsAdapter,
	}
}
