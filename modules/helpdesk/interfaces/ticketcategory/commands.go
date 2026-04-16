package ticketcategory

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateTicketCategoryCommand)(nil)
	req = (*DeleteTicketCategoryCommand)(nil)
	req = (*GetTicketCategoryQuery)(nil)
	req = (*TicketCategoryExistsQuery)(nil)
	req = (*SearchTicketCategoriesQuery)(nil)
	req = (*UpdateTicketCategoryCommand)(nil)
	req = (*SetTicketCategoryIsArchivedCommand)(nil)
	util.Unused(req)
}

var createTicketCategoryCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketcategory", Action: "createTicketCategory"}

type CreateTicketCategoryCommand struct{ domain.TicketCategory }

func (CreateTicketCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return createTicketCategoryCommandType
}
func (CreateTicketCategoryCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TicketCategorySchemaName)
}

type CreateTicketCategoryResult = dyn.OpResult[domain.TicketCategory]

var deleteTicketCategoryCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketcategory", Action: "deleteTicketCategory"}

type DeleteTicketCategoryCommand dyn.DeleteOneCommand

func (DeleteTicketCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return deleteTicketCategoryCommandType
}

type DeleteTicketCategoryResult = dyn.OpResult[dyn.MutateResultData]

var getTicketCategoryQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketcategory", Action: "getTicketCategory"}

type GetTicketCategoryQuery dyn.GetOneQuery

func (GetTicketCategoryQuery) CqrsRequestType() cqrs.RequestType { return getTicketCategoryQueryType }

type GetTicketCategoryResult = dyn.OpResult[domain.TicketCategory]

var ticketCategoryExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketcategory", Action: "ticketCategoryExists"}

type TicketCategoryExistsQuery dyn.ExistsQuery

func (TicketCategoryExistsQuery) CqrsRequestType() cqrs.RequestType {
	return ticketCategoryExistsQueryType
}

type TicketCategoryExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchTicketCategoriesQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketcategory", Action: "searchTicketCategories"}

type SearchTicketCategoriesQuery dyn.SearchQuery

func (SearchTicketCategoriesQuery) CqrsRequestType() cqrs.RequestType {
	return searchTicketCategoriesQueryType
}

type SearchTicketCategoriesResultData = dyn.PagedResultData[domain.TicketCategory]
type SearchTicketCategoriesResult = dyn.OpResult[SearchTicketCategoriesResultData]

var updateTicketCategoryCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketcategory", Action: "updateTicketCategory"}

type UpdateTicketCategoryCommand struct{ domain.TicketCategory }

func (UpdateTicketCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return updateTicketCategoryCommandType
}
func (UpdateTicketCategoryCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TicketCategorySchemaName)
}

type UpdateTicketCategoryResult = dyn.OpResult[dyn.MutateResultData]

var setTicketCategoryIsArchivedCommandType = cqrs.RequestType{
	Module:    "helpdesk",
	Submodule: "ticketcategory",
	Action:    "setTicketCategoryIsArchived",
}

type SetTicketCategoryIsArchivedCommand dyn.SetIsArchivedCommand

func (SetTicketCategoryIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setTicketCategoryIsArchivedCommandType
}

type SetTicketCategoryIsArchivedResult = dyn.OpResult[dyn.MutateResultData]
