package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	shareIt "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

type driveFileShareRestParams struct {
	dig.In

	DriveFileShareSvc shareIt.DriveFileShareService
}

func NewDriveFileShareRest(params driveFileShareRestParams) *DriveFileShareRest {
	return &DriveFileShareRest{
		DriveFileShareSvc: params.DriveFileShareSvc,
	}
}

type DriveFileShareRest struct {
	httpserver.RestBase
	DriveFileShareSvc shareIt.DriveFileShareService
}

func (this DriveFileShareRest) CreateDriveFileShare(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create drive file share"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileShareSvc.CreateDriveFileShare,
		func(request CreateDriveFileShareRequest) shareIt.CreateDriveFileShareCommand {
			return request
		},
		func(result shareIt.CreateDriveFileShareResult) CreateDriveFileShareResponse {
			response := CreateDriveFileShareResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this DriveFileShareRest) CreateBulkDriveFileShares(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create bulk drive file shares"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileShareSvc.CreateBulkDriveFileShares,
		func(request CreateBulkDriveFileShareRequest) shareIt.CreateBulkDriveFileShareCommand {
			return request
		},
		func(result shareIt.CreateBulkDriveFileShareResult) CreateBulkDriveFileShareResponse {
			resp := CreateBulkDriveFileShareResponse{}
			if result.Data != nil {
				resp.Items = make([]httpserver.RestCreateResponse, 0, len(result.Data))
				for _, created := range result.Data {
					item := httpserver.RestCreateResponse{}
					item.FromEntity(created)
					resp.Items = append(resp.Items, item)
				}
			}
			return resp
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this DriveFileShareRest) UpdateDriveFileShare(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST update drive file share"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileShareSvc.UpdateDriveFileShare,
		func(request UpdateDriveFileShareRequest) shareIt.UpdateDriveFileShareCommand {
			return request
		},
		func(result shareIt.UpdateDriveFileShareResult) UpdateDriveFileShareResponse {
			response := UpdateDriveFileShareResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileShareRest) GetDriveFileShareById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get drive file share by id"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileShareSvc.GetDriveFileShareById,
		func(request GetDriveFileShareByIdRequest) shareIt.GetDriveFileShareByIdQuery {
			return request
		},
		func(result shareIt.GetDriveFileShareByIdResult) GetDriveFileShareByIdResponse {
			response := GetDriveFileShareByIdResponse{}
			response.FromDriveFileShare(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileShareRest) GetDriveFileShareByFileId(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get drive file shares file id"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileShareSvc.GetDriveFileShareByFileId,
		func(request GetDriveFileShareByFileIdRequest) shareIt.GetDriveFileShareByFileIdQuery {
			return request
		},
		func(result shareIt.GetDriveFileShareByFileIdResult) GetDriveFileShareByFileIdResponse {
			response := GetDriveFileShareByFileIdResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileShareRest) GetDriveFileShareByUser(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get drive file shares by user"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileShareSvc.GetDriveFileShareByUser,
		func(request GetDriveFileShareByUserRequest) shareIt.GetDriveFileShareByUserQuery {
			return request
		},
		func(result shareIt.GetDriveFileShareByUserResult) GetDriveFileShareByUserResponse {
			response := GetDriveFileShareByUserResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileShareRest) SearchDriveFileShare(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST search drive file shares"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileShareSvc.SearchDriveFileShare,
		func(request SearchDriveFileShareRequest) shareIt.SearchDriveFileShareQuery {
			return request
		},
		func(result shareIt.SearchDriveFileShareResult) SearchDriveFileShareResponse {
			response := SearchDriveFileShareResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this DriveFileShareRest) DeleteDriveFileShare(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete drive file share"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.DriveFileShareSvc.DeleteDriveFileShare,
		func(request DeleteDriveFileShareRequest) shareIt.DeleteDriveFileShareCommand {
			return request
		},
		func(result shareIt.DeleteDriveFileShareResult) DeleteDriveFileShareResponse {
			response := DeleteDriveFileShareResponse{}
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
