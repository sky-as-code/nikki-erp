package v1

import (
	"github.com/labstack/echo/v5"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/middleware"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
	shareIt "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_share"
)

type driveFileShareRestParams struct {
	dig.In

	DriveFileShareSvc shareIt.DriveFileShareService
}

func NewDriveFileShareRest(params driveFileShareRestParams) *DriveFileShareRest {
	return &DriveFileShareRest{DriveFileShareSvc: params.DriveFileShareSvc}
}

type DriveFileShareRest struct {
	httpserver.RestBase
	DriveFileShareSvc shareIt.DriveFileShareService
}

func shareActorUserId(echoCtx *echo.Context) model.Id {
	return model.Id(middleware.GetUserIdFromContext(echoCtx.Request().Context()))
}

func (this DriveFileShareRest) CreateDriveFileShare(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 create drive file share"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.DriveFileShareSvc.CreateDriveFileShare,
		func(cmd CreateDriveFileShareRequest) CreateDriveFileShareRequest {
			cmd.UserId = shareActorUserId(echoCtx)
			return cmd
		},
		func(data domain.DriveFileShare) *httpserver.RestCreateResponse {
			return httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated,
	)
}

func (this DriveFileShareRest) CreateBulkDriveFileShares(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 create bulk drive file shares"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.DriveFileShareSvc.CreateBulkDriveFileShares,
		func(cmd CreateBulkDriveFileShareRequest) CreateBulkDriveFileShareRequest {
			cmd.UserId = shareActorUserId(echoCtx)
			return cmd
		},
		func(rows []domain.DriveFileShare) []httpserver.RestCreateResponse {
			out := make([]httpserver.RestCreateResponse, 0, len(rows))
			for i := range rows {
				out = append(out, *httpserver.NewRestCreateResponseDyn(rows[i].GetFieldData()))
			}
			return out
		},
		httpserver.JsonCreated,
	)
}

func (this DriveFileShareRest) UpdateDriveFileShare(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 update drive file share"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.DriveFileShareSvc.UpdateDriveFileShare,
		func(cmd UpdateDriveFileShareRequest) UpdateDriveFileShareRequest {
			cmd.UserId = shareActorUserId(echoCtx)
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}

func (this DriveFileShareRest) GetDriveFileShareById(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 get drive file share by id"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeGetOne(
		"drive2 get drive file share by id",
		echoCtx,
		this.DriveFileShareSvc.GetDriveFileShareById,
	)
}

func (this DriveFileShareRest) GetDriveFileShareByFileId(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 list drive file shares by file"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeSearch(
		"drive2 get drive file shares by file id",
		echoCtx,
		this.DriveFileShareSvc.GetDriveFileShareByFileId,
	)
}

func (this DriveFileShareRest) GetDriveFileAncestorOwnersByFileId(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 get ancestor owners"); e != nil {
			err = e
		}
	}()
	var query GetDriveFileAncestorOwnersByFileIdRequest
	if err = echoCtx.Bind(&query); err != nil {
		return httpserver.JsonBadRequest(
			echoCtx,
			[]any{ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request")},
		)
	}
	reqCtx := echoCtx.Request().Context().(corectx.Context)
	result, err := this.DriveFileShareSvc.GetDriveFileAncestorOwnersByFileId(reqCtx, query)
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

func (this DriveFileShareRest) GetDriveFileResolvedSharesByFileId(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 get resolved shares"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeSearch(
		"drive2 get resolved shares by file id",
		echoCtx,
		this.DriveFileShareSvc.GetDriveFileResolvedSharesByFileId,
	)
}

func (this DriveFileShareRest) GetDriveFileUserShareDetails(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 get file-user share details"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.DriveFileShareSvc.GetDriveFileUserShareDetails,
		func(q GetDriveFileUserShareDetailsRequest) GetDriveFileUserShareDetailsRequest {
			return q
		},
		func(data []shareIt.DriveFileUserShareDetail) []shareIt.DriveFileUserShareDetail {
			return data
		},
		httpserver.JsonOk,
	)
}

func (this DriveFileShareRest) SearchDriveFileShare(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 search drive file shares"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeSearch(
		"drive2 search drive file shares",
		echoCtx,
		this.DriveFileShareSvc.SearchDriveFileShare,
	)
}

func (this DriveFileShareRest) GetDriveFileShareByUser(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 list shares by user"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeSearch(
		"drive2 get drive file shares by user",
		echoCtx,
		this.DriveFileShareSvc.GetDriveFileShareByUser,
	)
}

func (this DriveFileShareRest) DeleteDriveFileShare(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST drive2 delete drive file share"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.DriveFileShareSvc.DeleteDriveFileShare,
		func(cmd DeleteDriveFileShareRequest) DeleteDriveFileShareRequest {
			cmd.UserId = shareActorUserId(echoCtx)
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
