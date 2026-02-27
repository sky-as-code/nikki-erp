package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type driveFileRestParams struct {
	dig.In

	DriveFileSvc it.DriveFileService
}

func NewDriveFileRest(params driveFileRestParams) *DriveFileRest {
	return &DriveFileRest{
		DriveFileSvc: params.DriveFileSvc,
	}
}

type DriveFileRest struct {
	httpserver.RestBase
	DriveFileSvc it.DriveFileService
}

func (this DriveFileRest) CreateDriveFile(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create drive file"); e != nil {
			err = e
		}
	}()

	var cmd it.CreateDriveFileCommand
	if err = echoCtx.Bind(&cmd); err != nil {
		return err
	}

	fileHeader, formErr := echoCtx.FormFile("file")
	if formErr == nil {
		file, openErr := fileHeader.Open()
		if openErr != nil {
			return openErr
		}
		defer file.Close()
		cmd.File = file
		cmd.FileHeader = *fileHeader
	} else if !cmd.IsFolder {
		return httpserver.JsonBadRequest(echoCtx, &fault.ClientError{Code: "file_required", Details: "file is required when creating a file (not folder)"})
	}

	reqCtx := echoCtx.Request().Context().(crud.Context)
	result, err := this.DriveFileSvc.CreateDriveFile(reqCtx, cmd)
	if err != nil {
		return err
	}

	if result.GetClientError() != nil {
		return httpserver.JsonBadRequest(echoCtx, result.GetClientError())
	}
	if !result.GetHasData() {
		return httpserver.JsonBadRequest(echoCtx, &fault.ClientError{Code: "not_found", Details: "resource not found"})
	}

	response := CreateDriveFileResponse{}
	response.FromEntity(result.Data)
	return httpserver.JsonCreated(echoCtx, response)
}

func (this DriveFileRest) UpdateDriveFile(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST update drive file"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileSvc.UpdateDriveFile,
		func(request UpdateDriveFileRequest) it.UpdateDriveFileCommand {
			return request
		},
		func(result it.UpdateDriveFileResult) UpdateDriveFileResponse {
			response := UpdateDriveFileResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileRest) DeleteDriveFile(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete drive file"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileSvc.DeleteDriveFile,
		func(request DeleteDriveFileRequest) it.DeleteDriveFileCommand {
			return request
		},
		func(result it.DeleteDriveFileResult) DeleteDriveFileResponse {
			response := DeleteDriveFileResponse{}
			if result.Data != nil {
				response.Id = result.Data.Id
				response.DeletedAt = result.Data.DeletedAt.UnixMilli()
			}
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileRest) MoveDriveFileToTrash(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST move drive file to trash"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileSvc.MoveDriveFileToTrash,
		func(request MoveDriveFileToTrashRequest) it.MoveDriveFileToTrashCommand {
			return request
		},
		func(result it.MoveDriveFileToTrashResult) MoveDriveFileToTrashResponse {
			response := MoveDriveFileToTrashResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileRest) GetDriveFileById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get drive file by id"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileSvc.GetDriveFileById,
		func(request GetDriveFileByIdRequest) it.GetDriveFileByIdQuery {
			return request
		},
		func(result it.GetDriveFileByIdResult) GetDriveFileByIdResponse {
			response := GetDriveFileByIdResponse{}
			response.FromDriveFile(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileRest) DownloadDriveFile(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST download drive file"); e != nil {
			err = e
		}
	}()

	var query it.GetDriveFileByIdQuery
	if err = echoCtx.Bind(&query); err != nil {
		return
	}

	reqCtx := echoCtx.Request().Context().(crud.Context)
	stream, err := this.DriveFileSvc.DownloadDriveFile(reqCtx, query)
	if err != nil {
		return
	}
	defer stream.Close()

	return echoCtx.Stream(http.StatusOK, "application/octet-stream", stream)
}

func (this DriveFileRest) GetDriveFileByParent(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get drive files by parent"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileSvc.GetDriveFileByParent,
		func(request GetDriveFileByParentRequest) it.GetDriveFileByParentQuery {
			return request
		},
		func(result it.GetDriveFileByParentResult) GetDriveFileByParentResponse {
			response := GetDriveFileByParentResponse{}
			response.FromResult(&result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileRest) SearchDriveFile(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST search drive files"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileSvc.SearchDriveFile,
		func(request SearchDriveFileRequest) it.SearchDriveFileQuery {
			return request
		},
		func(result it.SearchDriveFileResult) SearchDriveFileResponse {
			response := SearchDriveFileResponse{}
			response.FromResult(&result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
