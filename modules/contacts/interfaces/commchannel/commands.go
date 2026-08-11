package commchannel

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
	req = (*CreateCommChannelCommand)(nil)
	req = (*UpdateCommChannelCommand)(nil)
	req = (*DeleteCommChannelCommand)(nil)
	req = (*GetCommChannelQuery)(nil)
	req = (*SearchCommChannelsQuery)(nil)
	req = (*CommChannelExistsQuery)(nil)
	util.Unused(req)
}

var createCommChannelCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "create",
}

type CreateCommChannelCommand struct {
	domain.CommChannel
}

func (CreateCommChannelCommand) CqrsRequestType() cqrs.RequestType {
	return createCommChannelCommandType
}

func (this CreateCommChannelCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.CommChannelSchemaName)
}

type CreateCommChannelResult = dyn.OpResult[domain.CommChannel]

var updateCommChannelCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "update",
}

type UpdateCommChannelCommand struct {
	domain.CommChannel
}

func (UpdateCommChannelCommand) CqrsRequestType() cqrs.RequestType {
	return updateCommChannelCommandType
}

func (this UpdateCommChannelCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.CommChannelSchemaName)
}

type UpdateCommChannelResult = dyn.OpResult[dyn.MutateResultData]

var deleteCommChannelCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "delete",
}

type DeleteCommChannelCommand dyn.DeleteOneCommand

func (DeleteCommChannelCommand) CqrsRequestType() cqrs.RequestType {
	return deleteCommChannelCommandType
}

type DeleteCommChannelResult = dyn.OpResult[dyn.MutateResultData]

var getCommChannelQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "getCommChannel",
}

type GetCommChannelQuery struct {
	Columns []string `json:"columns" query:"columns"`
	Id      *string  `json:"id" param:"id"`
}

func (GetCommChannelQuery) CqrsRequestType() cqrs.RequestType {
	return getCommChannelQueryType
}

type GetCommChannelResult = dyn.OpResult[domain.CommChannel]

var searchCommChannelsQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "search",
}

type SearchCommChannelsQuery struct {
	Fields          []string            `json:"fields" query:"fields"`
	Graph           *dmodel.SearchGraph `json:"graph" query:"graph"`
	Page            int                 `json:"page" query:"page"`
	Size            int                 `json:"size" query:"size"`
	IncludeArchived *bool               `json:"include_archived" query:"include_archived"`
	PartyId         model.Id            `json:"party_id" param:"party_id"`
}

func (SearchCommChannelsQuery) CqrsRequestType() cqrs.RequestType {
	return searchCommChannelsQueryType
}

func (SearchCommChannelsQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"contacts.search_comm_channels_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dyn.DefineFieldSearchFields()).
				Field(dyn.DefineFieldSearchGraph()).
				Field(dyn.DefineFieldSearchPage()).
				Field(dyn.DefineFieldSearchSize()).
				Field(dyn.DefineFieldIncludeArchived()).
				Field(dmodel.DefineField().
					Name("party_id").
					DataType(dmodel.FieldDataTypeUlid()))
		},
	)
}

type SearchCommChannelsResultData = dyn.PagedResultData[domain.CommChannel]
type SearchCommChannelsResult = dyn.OpResult[SearchCommChannelsResultData]

var commChannelExistsQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "comm_channel",
	Action:    "exists",
}

type CommChannelExistsQuery dyn.ExistsQuery

func (CommChannelExistsQuery) CqrsRequestType() cqrs.RequestType {
	return commChannelExistsQueryType
}

type CommChannelExistsResult = dyn.OpResult[dyn.ExistsResultData]
