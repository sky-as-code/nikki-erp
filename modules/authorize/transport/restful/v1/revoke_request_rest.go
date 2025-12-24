package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/revoke_request"
)

type revokeRequestRestParams struct {
	dig.In

	RevokeRequestSvc it.RevokeRequestService
}

func NewRevokeRequestRest(params revokeRequestRestParams) *RevokeRequestRest {
	return &RevokeRequestRest{
		RevokeRequestSvc: params.RevokeRequestSvc,
	}
}

type RevokeRequestRest struct {
	httpserver.RestBase
	RevokeRequestSvc it.RevokeRequestService
}

func (this RevokeRequestRest) Create(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create revoke request"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.RevokeRequestSvc.Create,
		func(request CreateRevokeRequestRequest) it.CreateRevokeRequestCommand {
			return it.CreateRevokeRequestCommand(request)
		},
		func(result it.CreateRevokeRequestResult) CreateRevokeRequestResponse {
			response := CreateRevokeRequestResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this RevokeRequestRest) CreateBulk(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create bulk revoke requests"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.RevokeRequestSvc.CreateBulk,
		func(request CreateBulkRevokeRequestsRequest) it.CreateBulkRevokeRequestsCommand {
			return it.CreateBulkRevokeRequestsCommand(request)
		},
		func(result it.CreateBulkRevokeRequestsResult) CreateBulkRevokeRequestsResponse {
			resp := CreateBulkRevokeRequestsResponse{}
			if result.Data == nil {
				return resp
			}
			resp.Items = make([]httpserver.RestCreateResponse, 0, len(result.Data))
			for _, created := range result.Data {
				item := httpserver.RestCreateResponse{}
				item.FromEntity(created)
				resp.Items = append(resp.Items, item)
			}
			return resp
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this RevokeRequestRest) GetById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get revoke request by id"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.RevokeRequestSvc.GetById,
		func(request GetRevokeRequestByIdRequest) it.GetRevokeRequestByIdQuery {
			return it.GetRevokeRequestByIdQuery(request)
		},
		func(result it.GetRevokeRequestByIdResult) GetRevokeRequestByIdResponse {
			response := GetRevokeRequestByIdResponse{}
			response.FromRevokeRequest(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this RevokeRequestRest) Search(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST search revoke requests"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.RevokeRequestSvc.Search,
		func(request SearchRevokeRequestsRequest) it.SearchRevokeRequestsQuery {
			return it.SearchRevokeRequestsQuery(request)
		},
		func(result it.SearchRevokeRequestsResult) SearchRevokeRequestsResponse {
			response := SearchRevokeRequestsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this RevokeRequestRest) Delete(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete revoke request"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.RevokeRequestSvc.Delete,
		func(request DeleteRevokeRequestRequest) it.DeleteRevokeRequestCommand {
			return it.DeleteRevokeRequestCommand(request)
		},
		func(result it.DeleteRevokeRequestResult) DeleteRevokeRequestResponse {
			response := DeleteRevokeRequestResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
