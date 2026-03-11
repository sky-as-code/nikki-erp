package restful

import (
	"errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	commonMiddleware "github.com/sky-as-code/nikki-erp/common/middleware"
	v1 "github.com/sky-as-code/nikki-erp/modules/drive/transports/restful/v1"
)

func InitRestfulHandlers() error {
	err := errors.Join(
		initDriveRest(),
	)
	return err
}

func initDriveRest() error {
	deps.Register(
		v1.NewDriveFileRest,
		v1.NewDriveFileShareRest,
	)

	return deps.Invoke(
		func(
			route *echo.Group,
			driveFileRest *v1.DriveFileRest,
			driveFileShareRest *v1.DriveFileShareRest,
		) {
			v1 := route.Group("/v1/drive")
			initV1(v1, driveFileRest, driveFileShareRest)
		},
	)
}

func initV1(
	route *echo.Group,
	driveFileRest *v1.DriveFileRest,
	driveFileShareRest *v1.DriveFileShareRest,
) {
	protected := route.Group("", commonMiddleware.RequireAuthMiddleware())

	// Drive files
	protected.POST("/files", driveFileRest.CreateDriveFile)
	protected.PATCH("/files/:driveFileId", driveFileRest.UpdateDriveFileMetadata)
	protected.PUT("/files/:driveFileId/content", driveFileRest.UpdateDriveFileContent)
	protected.DELETE("/files/:driveFileId", driveFileRest.DeleteDriveFile)
	protected.PUT("/files/:driveFileId/move-to-trash", driveFileRest.MoveDriveFileToTrash)
	protected.PUT("/files/:driveFileId/restore", driveFileRest.RestoreDriveFile)
	protected.PUT("/files/:driveFileId/move", driveFileRest.MoveDriveFile)
	protected.GET("/files/:driveFileId/ancestors", driveFileRest.GetDriveFileAncestors)
	protected.GET("/files/root", driveFileRest.GetDriveFileByParent)
	protected.GET("/files/:driveFileId", driveFileRest.GetDriveFileById)

	// TODO: add an API to create token for stream
	route.GET("/files/:driveFileId/stream", driveFileRest.StreamDriveFile)
	protected.GET("/files/:driveFileId/children", driveFileRest.GetDriveFileByParent)
	protected.GET("/files", driveFileRest.SearchDriveFile)

	// Sharing management on file
	fileGroup := protected.Group("/files/:driveFileId")

	fileGroup.POST("/shares", driveFileShareRest.CreateDriveFileShare)
	fileGroup.POST("/shares/bulk", driveFileShareRest.CreateBulkDriveFileShares)
	fileGroup.PUT("/shares/:driveFileShareId", driveFileShareRest.UpdateDriveFileShare)
	fileGroup.GET("/shares/:driveFileShareId", driveFileShareRest.GetDriveFileShareById)
	fileGroup.GET("/shares", driveFileShareRest.GetDriveFileShareByFileId)
	fileGroup.DELETE("/shares/:driveFileShareId", driveFileShareRest.DeleteDriveFileShare)

}
