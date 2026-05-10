package contact

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateContactCommand)(nil)
	req = (*DeleteContactCommand)(nil)
	req = (*GetContactQuery)(nil)
	req = (*SearchContactsQuery)(nil)
	req = (*UpdateContactCommand)(nil)
	req = (*ContactExistsQuery)(nil)
	util.Unused(req)
}

var createContactCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "contact",
	Action:    "create",
}

type CreateContactCommand struct {
	models.Contact
}

func (CreateContactCommand) CqrsRequestType() cqrs.RequestType {
	return createContactCommandType
}

func (CreateContactCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.ContactSchemaName)
}

type CreateContactResult = dyn.OpResult[models.Contact]

var updateContactCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "contact",
	Action:    "update",
}

type UpdateContactCommand struct {
	models.Contact
}

func (UpdateContactCommand) CqrsRequestType() cqrs.RequestType {
	return updateContactCommandType
}

func (UpdateContactCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.ContactSchemaName)
}

type UpdateContactResult = dyn.OpResult[dyn.MutateResultData]

var deleteContactCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "contact",
	Action:    "delete",
}

type DeleteContactCommand dyn.DeleteOneCommand

func (DeleteContactCommand) CqrsRequestType() cqrs.RequestType {
	return deleteContactCommandType
}

type DeleteContactResult = dyn.OpResult[dyn.MutateResultData]

var getContactQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "contact",
	Action:    "get",
}

type GetContactQuery dyn.GetOneQuery

func (GetContactQuery) CqrsRequestType() cqrs.RequestType {
	return getContactQueryType
}

type GetContactResult = dyn.OpResult[models.Contact]

var searchContactsQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "contact",
	Action:    "search",
}

type SearchContactsQuery dyn.SearchQuery

func (SearchContactsQuery) CqrsRequestType() cqrs.RequestType {
	return searchContactsQueryType
}

type SearchContactsResultData = dyn.PagedResultData[models.Contact]
type SearchContactsResult = dyn.OpResult[SearchContactsResultData]

var contactExistsQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "contact",
	Action:    "exists",
}

type ContactExistsQuery dyn.ExistsQuery

func (ContactExistsQuery) CqrsRequestType() cqrs.RequestType {
	return contactExistsQueryType
}

type ContactExistsResult = dyn.OpResult[dyn.ExistsResultData]
