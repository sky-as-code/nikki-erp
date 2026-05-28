package cqrs

// import (
// 	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
// 	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
// 	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
// )
//
// type DriveFileHandler struct {
// 	driveFileService it.DriveFileService
// }
//
// func NewDriveFileHandler(driveFileService it.DriveFileService) *DriveFileHandler {
// 	return &DriveFileHandler{
// 		driveFileService: driveFileService,
// 	}
// }
//
// func (this *DriveFileHandler) GetDriveFileById(ctx corectx.Context, packet *cqrs.RequestPacket[it.GetDriveFileByIdQuery]) (*cqrs.Reply[it.GetDriveFileByIdResult], error) {
// 	if packet != nil && packet.Request() != nil {
// 		packet.Request().UserId = ctx.GetUser()
// 	}
// 	return cqrs.HandlePacket(ctx, packet, this.driveFileService.GetDriveFileById)
// }
