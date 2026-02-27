package restful

import (
	"errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
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

	// Drive files
	route.POST("/files", driveFileRest.CreateDriveFile)
	route.PUT("/files/:driveFileId", driveFileRest.UpdateDriveFile)
	route.DELETE("/files/:driveFileId", driveFileRest.DeleteDriveFile)
	route.POST("/files/:driveFileId/move-to-trash", driveFileRest.MoveDriveFileToTrash)

	route.GET("/files/:driveFileId", driveFileRest.GetDriveFileById)
	route.GET("/files/:driveFileId/download", driveFileRest.DownloadDriveFile)
	route.GET("/files/:driveFileId/children", driveFileRest.GetDriveFileByParent)
	route.GET("/files", driveFileRest.SearchDriveFile)

	// Sharing management on file
	fileGroup := route.Group("/files/:driveFileId")

	fileGroup.POST("/shares", driveFileShareRest.CreateDriveFileShare)
	fileGroup.POST("/shares/bulk", driveFileShareRest.CreateBulkDriveFileShares)
	fileGroup.PUT("/shares/:driveFileShareId", driveFileShareRest.UpdateDriveFileShare)
	fileGroup.GET("/shares/:driveFileShareId", driveFileShareRest.GetDriveFileShareById)
	fileGroup.GET("/shares", driveFileShareRest.GetDriveFileShareByFileId)
	fileGroup.DELETE("/shares/:driveFileShareId", driveFileShareRest.DeleteDriveFileShare)

	// List files shared to user
	route.GET("/shares", driveFileShareRest.SearchDriveFileShare)
}
