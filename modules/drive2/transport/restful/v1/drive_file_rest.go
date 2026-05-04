package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/middleware"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
)

type driveFileRestParams struct {
	dig.In

	DriveFileSvc it.DriveFileService
}

func NewDriveFileRest(params driveFileRestParams) *DriveFileRest {
	return &DriveFileRest{driveFileSvc: params.DriveFileSvc}
}

type DriveFileRest struct {
	httpserver.RestBase
	driveFileSvc it.DriveFileService
}

func userIdFromEcho(echoCtx *echo.Context) model.Id {
	return model.Id(middleware.GetUserIdFromContext(echoCtx.Request().Context()))
}

func (this DriveFileRest) CreateDriveFile(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 create drive file"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestFormData(
		echoCtx,
		this.driveFileSvc.CreateDriveFile,
		func(req CreateDriveFileRequest) it.CreateDriveFileCommand {
			cmd := it.CreateDriveFileCommand{}

			cmd.DriveFile = *domain.NewDriveFile()
			cmd.SetName(&req.Name)
			cmd.SetIsFolder(&req.IsFolder)
			cmd.SetParentFileRef(req.ParentFileRef)
			cmd.SetVisibility(&req.Visibility)
			fakeUserID := "01JWNMZ36QHC7CQQ748H9NQ6J6"
			cmd.SetOwnerRef(&fakeUserID)

			if req.FileHeader != nil {
				file, openErr := req.FileHeader.Open()
				if openErr != nil {
					panic(httpserver.JsonBadRequest(echoCtx,
						ft.ClientErrors{*ft.NewValidationError("file", "file_error", openErr.Error())}))
				}
				cmd.File = file
				cmd.FileHeader = req.FileHeader
			}

			return cmd
		},
		func(data domain.DriveFile) CreateDriveFileResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this DriveFileRest) UpdateDriveFileMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 update drive file metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.driveFileSvc.UpdateDriveFileMetadata,
		func(cmd UpdateDriveFileMetadataRequest) UpdateDriveFileMetadataRequest {
			id := echoCtx.Param("id")
			cmd.SetId(&id)
			return cmd
		},
		func(data dynamicmodel.MutateResultData) dynamicmodel.MutateResultData {
			return data
		},
		httpserver.JsonOk,
	)
}

func (this DriveFileRest) UpdateDriveFileContent(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 update drive file content"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestFormData(
		echoCtx,
		this.driveFileSvc.UpdateDriveFileContent,
		func(req UpdateDriveFileContentRequest) it.UpdateDriveFileContentCommand {
			cmd := it.UpdateDriveFileContentCommand{}

			dFile := domain.NewDriveFile()
			dFile.SetId(&req.Id)
			dFile.SetEtag(&req.Etag)
			dFile.SetName(&req.Name)
			cmd.DriveFile = *dFile

			if req.FileHeader != nil {
				file, openErr := req.FileHeader.Open()
				if openErr != nil {
					panic(httpserver.JsonBadRequest(echoCtx,
						ft.ClientErrors{*ft.NewValidationError("file", "file_error", openErr.Error())}))
				}
				cmd.File = file
				cmd.FileHeader = req.FileHeader
			}

			return cmd
		},
		func(data dynamicmodel.MutateResultData) UpdateDriveFileContentResponse {
			return data
		},
		httpserver.JsonOk,
	)
}

func (this DriveFileRest) DeleteDriveFile(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 delete drive file"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.driveFileSvc.DeleteDriveFile,
		func(cmd DeleteDriveFileRequest) DeleteDriveFileRequest {
			cmd.UserId = userIdFromEcho(echoCtx)
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}

func (this DriveFileRest) MoveDriveFileToTrash(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 move drive file to trash"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.driveFileSvc.MoveDriveFileToTrash,
		func(cmd MoveDriveFileToTrashRequest) MoveDriveFileToTrashRequest {
			cmd.UserId = userIdFromEcho(echoCtx)
			return cmd
		},
		func(data domain.DriveFile) dmodel.DynamicFields {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this DriveFileRest) RestoreDriveFile(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 restore drive file"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.driveFileSvc.RestoreDriveFile,
		func(cmd RestoreDriveFileRequest) RestoreDriveFileRequest {
			cmd.UserId = userIdFromEcho(echoCtx)
			return cmd
		},
		func(data domain.DriveFile) dmodel.DynamicFields {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this DriveFileRest) MoveDriveFile(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 move drive file"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.driveFileSvc.MoveDriveFile,
		func(cmd MoveDriveFileRequest) MoveDriveFileRequest {
			cmd.UserId = userIdFromEcho(echoCtx)
			return cmd
		},
		func(data domain.DriveFile) dmodel.DynamicFields {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this DriveFileRest) GetDriveFileAncestors(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 get drive file ancestors"); e != nil {
			err = e
		}
	}()
	var query GetDriveFileAncestorsRequest
	if err = echoCtx.Bind(&query); err != nil {
		return httpserver.JsonBadRequest(
			echoCtx,
			[]any{ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request")},
		)
	}
	query.UserId = userIdFromEcho(echoCtx)
	reqCtx := echoCtx.Request().Context().(corectx.Context)
	result, err := this.driveFileSvc.GetDriveFileAncestors(reqCtx, query)
	if err != nil {
		return httpserver.HandleServiceError(echoCtx, err)
	}
	if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
	}
	out := make([]dmodel.DynamicFields, 0, len(result.Data))
	for i := range result.Data {
		out = append(out, result.Data[i].GetFieldData())
	}
	return httpserver.JsonOk(echoCtx, out)
}

func (this DriveFileRest) GetDriveFileById(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 get drive file by id"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeGetOne(
		"drive2 get drive file by id",
		echoCtx,
		func(ctx corectx.Context, q GetDriveFileByIdRequest) (*it.GetDriveFileByIdResult, error) {
			q.UserId = userIdFromEcho(echoCtx)
			return this.driveFileSvc.GetDriveFileById(ctx, q)
		},
	)
}

func (this DriveFileRest) StreamDriveFile(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 stream drive file"); e != nil {
			err = e
		}
	}()
	var query GetDriveFileByIdRequest
	if err = echoCtx.Bind(&query); err != nil {
		return httpserver.JsonBadRequest(
			echoCtx,
			[]any{ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request")},
		)
	}
	if query.DriveFileId == "" {
		query.DriveFileId = model.Id(echoCtx.Param("drive_file_id"))
	}
	query.UserId = userIdFromEcho(echoCtx)
	reqCtx := echoCtx.Request().Context().(corectx.Context)
	driveFile, stream, err := this.driveFileSvc.DownloadDriveFile(reqCtx, query)
	if err != nil {
		return httpserver.HandleServiceError(echoCtx, err)
	}
	if stream == nil {
		return httpserver.JsonBadRequest(echoCtx, &ft.ClientError{
			Code:    "not_found",
			Details: ft.ValidationErrors{"drive_file_id": "drive file not found or not downloadable"},
		})
	}
	defer stream.Close()
	fname := "file"
	if driveFile != nil {
		if n := driveFile.GetName(); n != nil {
			fname = *n
		}
	}
	disposition := "inline"
	if query.IsDownload {
		disposition = "attachment"
	}
	echoCtx.Response().Header().Set(
		echo.HeaderContentDisposition,
		fmt.Sprintf("%s; filename=%q", disposition, fname),
	)
	mimeType := "application/octet-stream"
	if driveFile != nil {
		if m := driveFile.GetMime(); m != nil && *m != "" {
			mimeType = *m
		}
	}
	return echoCtx.Stream(http.StatusOK, mimeType, stream)
}

func (this DriveFileRest) GetDriveFileByParent(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 get drive files by parent"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeSearch(
		"drive2 get drive files by parent",
		echoCtx,
		func(ctx corectx.Context, q GetDriveFileByParentRequest) (*it.GetDriveFileByParentResult, error) {
			q.UserId = userIdFromEcho(echoCtx)
			return this.driveFileSvc.GetDriveFileByParent(ctx, q)
		},
	)
}

func (this DriveFileRest) SearchDriveFile(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 search drive files"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeSearch(
		"drive2 search drive files",
		echoCtx,
		func(ctx corectx.Context, q SearchDriveFileRequest) (*it.SearchDriveFileResult, error) {
			q.UserId = userIdFromEcho(echoCtx)
			return this.driveFileSvc.SearchDriveFile(ctx, q)
		},
	)
}

func (this DriveFileRest) SearchDriveFilesShared(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 search drive files shared"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeSearch(
		"drive2 search drive files shared",
		echoCtx,
		func(ctx corectx.Context, q SearchDriveFilesSharedRequest) (*it.SearchDriveFilesSharedResult, error) {
			q.UserId = userIdFromEcho(echoCtx)
			return this.driveFileSvc.SearchDriveFilesShared(ctx, q)
		},
	)
}
