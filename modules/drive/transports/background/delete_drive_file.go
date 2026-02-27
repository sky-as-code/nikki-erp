package background

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type DriveFileHandler interface {
	DeleteTrashedFile(ctx context.Context) error
}

type driveFileHandler struct {
	driveFileService it.DriveFileService
}

func NewDriveFileHandler(driveFileService it.DriveFileService) DriveFileHandler {
	return &driveFileHandler{
		driveFileService: driveFileService,
	}
}

func (this *driveFileHandler) DeleteTrashedFile(ctx context.Context) error {
	crudCtx := ctx.(crud.Context)
	return this.driveFileService.DeleteTrashedDriveFile(crudCtx)
}
