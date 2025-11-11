package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	itCommChannel "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/commchannel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type commChannelRestParams struct {
	dig.In

	CommChannelSvc itCommChannel.CommChannelService
}

func NewCommChannelRest(params commChannelRestParams) *CommChannelRest {
	return &CommChannelRest{
		CommChannelSvc: params.CommChannelSvc,
	}
}

type CommChannelRest struct {
	httpserver.RestBase
	CommChannelSvc itCommChannel.CommChannelService
}

func (this CommChannelRest) CreateCommChannel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create comm channel"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.CommChannelSvc.CreateCommChannel,
		func(request CreateCommChannelRequest) itCommChannel.CreateCommChannelCommand {
			return itCommChannel.CreateCommChannelCommand(request)
		},
		func(result itCommChannel.CreateCommChannelResult) CreateCommChannelResponse {
			response := CreateCommChannelResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this CommChannelRest) UpdateCommChannel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update comm channel"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.CommChannelSvc.UpdateCommChannel,
		func(request UpdateCommChannelRequest) itCommChannel.UpdateCommChannelCommand {
			return itCommChannel.UpdateCommChannelCommand(request)
		},
		func(result itCommChannel.UpdateCommChannelResult) UpdateCommChannelResponse {
			response := UpdateCommChannelResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this CommChannelRest) DeleteCommChannel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete comm channel"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.CommChannelSvc.DeleteCommChannel,
		func(request DeleteCommChannelRequest) itCommChannel.DeleteCommChannelCommand {
			return itCommChannel.DeleteCommChannelCommand(request)
		},
		func(result itCommChannel.DeleteCommChannelResult) DeleteCommChannelResponse {
			response := DeleteCommChannelResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this CommChannelRest) GetCommChannelById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get comm channel by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.CommChannelSvc.GetCommChannelById,
		func(request GetCommChannelByIdRequest) itCommChannel.GetCommChannelByIdQuery {
			return itCommChannel.GetCommChannelByIdQuery(request)
		},
		func(result itCommChannel.GetCommChannelByIdResult) GetCommChannelByIdResponse {
			response := GetCommChannelByIdResponse{}
			response.FromCommChannel(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this CommChannelRest) SearchCommChannels(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search comm channels"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.CommChannelSvc.SearchCommChannels,
		func(request SearchCommChannelsRequest) itCommChannel.SearchCommChannelsQuery {
			return itCommChannel.SearchCommChannelsQuery(request)
		},
		func(result itCommChannel.SearchCommChannelsResult) SearchCommChannelsResponse {
			response := SearchCommChannelsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
