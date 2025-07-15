package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	it "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	tagder "github.com/sky-as-code/nikki-erp/modules/core/tag/derived"
)

type partyRestParams struct {
	dig.In

	Config   config.ConfigService
	Logger   logging.LoggerService
	PartySvc it.PartyService
	CqrsBus  cqrs.CqrsBus
}

func NewPartyRest(params partyRestParams) *PartyRest {
	return &PartyRest{
		DerivedRest: tagder.DerivedRest{
			RestBase: httpserver.RestBase{
				ConfigSvc: params.Config,
				Logger:    params.Logger,
				CqrsBus:   params.CqrsBus,
			},
			TagSvc: params.PartySvc.TagSvc(),
		},
		PartySvc: params.PartySvc,
	}
}

type PartyRest struct {
	tagder.DerivedRest
	PartySvc it.PartyTagService
}

func (this PartyRest) CreatePartyTag(echoCtx echo.Context) error {
	return this.DerivedRest.CreateDerivedTag(echoCtx)
}

func (this PartyRest) UpdatePartyTag(echoCtx echo.Context) error {
	return this.DerivedRest.UpdateDerivedTag(echoCtx)
}

func (this PartyRest) DeletePartyTag(echoCtx echo.Context) error {
	return this.DerivedRest.DeleteDerivedTag(echoCtx)
}

func (this PartyRest) GetPartyTagById(echoCtx echo.Context) error {
	return this.DerivedRest.GetDerivedTagById(echoCtx)
}

func (this PartyRest) ListPartyTags(echoCtx echo.Context) error {
	return this.DerivedRest.ListDerivedTags(echoCtx)
}
