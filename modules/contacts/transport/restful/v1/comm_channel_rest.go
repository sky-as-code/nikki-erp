package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

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

func (this CommChannelRest) CreateCommChannel(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create comm channel",
		echoCtx,
		&itCommChannel.CreateCommChannelCommand{},
		this.CommChannelSvc.CreateCommChannel,
	)
}

func (this CommChannelRest) UpdateCommChannel(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update comm channel",
		echoCtx,
		&itCommChannel.UpdateCommChannelCommand{},
		this.CommChannelSvc.UpdateCommChannel,
	)
}

func (this CommChannelRest) DeleteCommChannel(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete comm channel",
		echoCtx,
		this.CommChannelSvc.DeleteCommChannel,
	)
}

func (this CommChannelRest) GetCommChannel(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get comm channel",
		echoCtx,
		this.CommChannelSvc.GetCommChannel,
	)
}

func (this CommChannelRest) SearchCommChannels(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search comm channels",
		echoCtx,
		this.CommChannelSvc.SearchCommChannels,
	)
}

func (this CommChannelRest) CommChannelExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"comm channel exists",
		echoCtx,
		this.CommChannelSvc.CommChannelExists,
	)
}
