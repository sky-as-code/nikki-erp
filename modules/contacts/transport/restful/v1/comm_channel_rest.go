package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/comm_channel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type commChannelRestParams struct {
	dig.In

	CommChannelSvc comm_channel.CommChannelService
}

func NewCommChannelRest(params commChannelRestParams) *CommChannelRest {
	return &CommChannelRest{
		CommChannelSvc: params.CommChannelSvc,
	}
}

type CommChannelRest struct {
	httpserver.RestBase
	CommChannelSvc comm_channel.CommChannelService
}

func (this CommChannelRest) CreateCommChannel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create comm channel"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.CommChannelSvc.CreateCommChannel,
		func(request CreateCommChannelRequest) comm_channel.CreateCommChannelCommand {
			return comm_channel.CreateCommChannelCommand(request)
		},
		func(result comm_channel.CreateCommChannelResult) CreateCommChannelResponse {
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
		func(request UpdateCommChannelRequest) comm_channel.UpdateCommChannelCommand {
			return comm_channel.UpdateCommChannelCommand(request)
		},
		func(result comm_channel.UpdateCommChannelResult) UpdateCommChannelResponse {
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
		func(request DeleteCommChannelRequest) comm_channel.DeleteCommChannelCommand {
			return comm_channel.DeleteCommChannelCommand(request)
		},
		func(result comm_channel.DeleteCommChannelResult) DeleteCommChannelResponse {
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
		func(request GetCommChannelByIdRequest) comm_channel.GetCommChannelByIdQuery {
			return comm_channel.GetCommChannelByIdQuery(request)
		},
		func(result comm_channel.GetCommChannelByIdResult) GetCommChannelByIdResponse {
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
		func(request SearchCommChannelsRequest) comm_channel.SearchCommChannelsQuery {
			return comm_channel.SearchCommChannelsQuery(request)
		},
		func(result comm_channel.SearchCommChannelsResult) SearchCommChannelsResponse {
			response := SearchCommChannelsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
