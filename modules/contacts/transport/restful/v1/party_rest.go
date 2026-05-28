package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type partyRestParams struct {
	dig.In

	PartySvc party.PartyService
}

func NewPartyRest(params partyRestParams) *PartyRest {
	return &PartyRest{
		PartySvc: params.PartySvc,
	}
}

type PartyRest struct {
	httpserver.RestBase
	PartySvc party.PartyService
}

func (this PartyRest) CreateParty(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create party",
		echoCtx,
		&party.CreatePartyCommand{},
		this.PartySvc.CreateParty,
	)
}

func (this PartyRest) UpdateParty(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update party",
		echoCtx,
		&party.UpdatePartyCommand{},
		this.PartySvc.UpdateParty,
	)
}

func (this PartyRest) DeleteParty(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete party",
		echoCtx,
		this.PartySvc.DeleteParty,
	)
}

func (this PartyRest) GetParty(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get party",
		echoCtx,
		this.PartySvc.GetParty,
	)
}

func (this PartyRest) SearchParties(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search parties",
		echoCtx,
		this.PartySvc.SearchParties,
	)
}

func (this PartyRest) PartyExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"party exists",
		echoCtx,
		this.PartySvc.PartyExists,
	)
}
