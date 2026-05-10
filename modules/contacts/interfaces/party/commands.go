package party

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreatePartyCommand)(nil)
	req = (*UpdatePartyCommand)(nil)
	req = (*DeletePartyCommand)(nil)
	req = (*GetPartyQuery)(nil)
	req = (*SearchPartiesQuery)(nil)
	req = (*PartyExistsQuery)(nil)
	util.Unused(req)
}

var createPartyCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "create",
}

type CreatePartyCommand struct {
	domain.Party
}

func (CreatePartyCommand) CqrsRequestType() cqrs.RequestType {
	return createPartyCommandType
}

func (this CreatePartyCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.PartySchemaName)
}

type CreatePartyResult = dyn.OpResult[domain.Party]

var updatePartyCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "update",
}

type UpdatePartyCommand struct {
	domain.Party
}

func (UpdatePartyCommand) CqrsRequestType() cqrs.RequestType {
	return updatePartyCommandType
}

func (this UpdatePartyCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.PartySchemaName)
}

type UpdatePartyResult = dyn.OpResult[dyn.MutateResultData]

var deletePartyCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "delete",
}

type DeletePartyCommand dyn.DeleteOneCommand

func (DeletePartyCommand) CqrsRequestType() cqrs.RequestType {
	return deletePartyCommandType
}

type DeletePartyResult = dyn.OpResult[dyn.MutateResultData]

var getPartyQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "getParty",
}

type GetPartyQuery struct {
	Columns     []string  `json:"columns" query:"columns"`
	Id          *model.Id `json:"id" param:"id"`
	DisplayName *string   `json:"displayName" query:"displayName"`
}

func (GetPartyQuery) CqrsRequestType() cqrs.RequestType {
	return getPartyQueryType
}

type GetPartyResult = dyn.OpResult[domain.Party]

var searchPartiesQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "search",
}

type SearchPartiesQuery dyn.SearchQuery

func (SearchPartiesQuery) CqrsRequestType() cqrs.RequestType {
	return searchPartiesQueryType
}

type SearchPartiesResultData = dyn.PagedResultData[domain.Party]
type SearchPartiesResult = dyn.OpResult[SearchPartiesResultData]

var partyExistsQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "exists",
}

type PartyExistsQuery dyn.ExistsQuery

func (PartyExistsQuery) CqrsRequestType() cqrs.RequestType {
	return partyExistsQueryType
}

type PartyExistsResult = dyn.OpResult[dyn.ExistsResultData]
