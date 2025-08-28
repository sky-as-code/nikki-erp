package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

func (this PartyRest) CreateParty(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create party"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.PartySvc.CreateParty,
		func(request CreatePartyRequest) party.CreatePartyCommand {
			return party.CreatePartyCommand(request)
		},
		func(result party.CreatePartyResult) CreatePartyResponse {
			response := CreatePartyResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this PartyRest) UpdateParty(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update party"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.PartySvc.UpdateParty,
		func(request UpdatePartyRequest) party.UpdatePartyCommand {
			return party.UpdatePartyCommand(request)
		},
		func(result party.UpdatePartyResult) UpdatePartyResponse {
			response := UpdatePartyResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this PartyRest) DeleteParty(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete party"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.PartySvc.DeleteParty,
		func(request DeletePartyRequest) party.DeletePartyCommand {
			return party.DeletePartyCommand(request)
		},
		func(result party.DeletePartyResult) DeletePartyResponse {
			response := DeletePartyResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this PartyRest) GetPartyById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get party by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.PartySvc.GetPartyById,
		func(request GetPartyByIdRequest) party.GetPartyByIdQuery {
			return party.GetPartyByIdQuery(request)
		},
		func(result party.GetPartyByIdResult) GetPartyByIdResponse {
			response := GetPartyByIdResponse{}
			response.FromParty(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this PartyRest) SearchParties(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search parties"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.PartySvc.SearchParties,
		func(request SearchPartiesRequest) party.SearchPartiesQuery {
			return party.SearchPartiesQuery(request)
		},
		func(result party.SearchPartiesResult) SearchPartiesResponse {
			response := SearchPartiesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
