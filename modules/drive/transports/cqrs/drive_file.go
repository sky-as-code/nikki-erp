package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/common/middleware"
	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type DriveFileHandler struct {
	driveFileService it.DriveFileService
}

func NewDriveFileHandler(driveFileService it.DriveFileService) *DriveFileHandler {
	return &DriveFileHandler{
		driveFileService: driveFileService,
	}
}

func (this *DriveFileHandler) GetDriveFileById(ctx context.Context, packet *cqrs.RequestPacket[it.GetDriveFileByIdQuery]) (*cqrs.Reply[it.GetDriveFileByIdResult], error) {
	if packet != nil && packet.Request() != nil {
		packet.Request().UserId = model.Id(middleware.GetUserIdFromContext(ctx))
	}
	return cqrs.HandlePacket(ctx, packet, this.driveFileService.GetDriveFileById)
}
