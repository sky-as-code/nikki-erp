package attribute

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateAttributeCommand)(nil)
	req = (*DeleteAttributeCommand)(nil)
	req = (*AttributeExistsQuery)(nil)
	req = (*GetAttributeQuery)(nil)
	req = (*SearchAttributesQuery)(nil)
	req = (*UpdateAttributeCommand)(nil)
	util.Unused(req)
}

var createAttributeCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "create",
}

type CreateAttributeCommand struct {
	domain.Attribute
}

func (CreateAttributeCommand) CqrsRequestType() cqrs.RequestType {
	return createAttributeCommandType
}

func (this CreateAttributeCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.AttributeSchemaName)
}

type CreateAttributeResult = dyn.OpResult[domain.Attribute]

var deleteAttributeCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "delete",
}

type DeleteAttributeCommand dyn.DeleteOneCommand

func (DeleteAttributeCommand) CqrsRequestType() cqrs.RequestType {
	return deleteAttributeCommandType
}

type DeleteAttributeResult = dyn.OpResult[dyn.MutateResultData]

var getAttributeQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "getAttribute",
}

type GetAttributeQuery dyn.GetOneQuery

func (GetAttributeQuery) CqrsRequestType() cqrs.RequestType {
	return getAttributeQueryType
}

type GetAttributeResult = dyn.OpResult[domain.Attribute]

var attributeExistsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "attributeExists",
}

type AttributeExistsQuery dyn.ExistsQuery

func (AttributeExistsQuery) CqrsRequestType() cqrs.RequestType {
	return attributeExistsQueryType
}

type AttributeExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchAttributesQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "search",
}

type SearchAttributesQuery dyn.SearchQuery

func (SearchAttributesQuery) CqrsRequestType() cqrs.RequestType {
	return searchAttributesQueryType
}

type SearchAttributesResultData = dyn.PagedResultData[domain.Attribute]
type SearchAttributesResult = dyn.OpResult[SearchAttributesResultData]

var updateAttributeCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute",
	Action:    "update",
}

type UpdateAttributeCommand struct {
	domain.Attribute
}

func (UpdateAttributeCommand) CqrsRequestType() cqrs.RequestType {
	return updateAttributeCommandType
}

func (this UpdateAttributeCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.AttributeSchemaName)
}

type UpdateAttributeResult = dyn.OpResult[dyn.MutateResultData]
