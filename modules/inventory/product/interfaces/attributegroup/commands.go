package attributegroup

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateAttributeGroupCommand)(nil)
	req = (*DeleteAttributeGroupCommand)(nil)
	req = (*AttributeGroupExistsQuery)(nil)
	req = (*GetAttributeGroupQuery)(nil)
	req = (*SearchAttributeGroupsQuery)(nil)
	req = (*UpdateAttributeGroupCommand)(nil)
	util.Unused(req)
}

var createAttributeGroupCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_group",
	Action:    "create",
}

type CreateAttributeGroupCommand struct {
	domain.AttributeGroup
}

func (CreateAttributeGroupCommand) CqrsRequestType() cqrs.RequestType {
	return createAttributeGroupCommandType
}

func (this CreateAttributeGroupCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.AttributeGroupSchemaName)
}

type CreateAttributeGroupResult = dyn.OpResult[domain.AttributeGroup]

var deleteAttributeGroupCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_group",
	Action:    "delete",
}

type DeleteAttributeGroupCommand dyn.DeleteOneCommand

func (DeleteAttributeGroupCommand) CqrsRequestType() cqrs.RequestType {
	return deleteAttributeGroupCommandType
}

type DeleteAttributeGroupResult = dyn.OpResult[dyn.MutateResultData]

var getAttributeGroupQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_group",
	Action:    "getAttributeGroup",
}

type GetAttributeGroupQuery dyn.GetOneQuery

func (GetAttributeGroupQuery) CqrsRequestType() cqrs.RequestType {
	return getAttributeGroupQueryType
}

type GetAttributeGroupResult = dyn.OpResult[domain.AttributeGroup]

var attributeGroupExistsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_group",
	Action:    "attributeGroupExists",
}

type AttributeGroupExistsQuery dyn.ExistsQuery

func (AttributeGroupExistsQuery) CqrsRequestType() cqrs.RequestType {
	return attributeGroupExistsQueryType
}

type AttributeGroupExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchAttributeGroupsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_group",
	Action:    "search",
}

type SearchAttributeGroupsQuery dyn.SearchQuery

func (SearchAttributeGroupsQuery) CqrsRequestType() cqrs.RequestType {
	return searchAttributeGroupsQueryType
}

type SearchAttributeGroupsResultData = dyn.PagedResultData[domain.AttributeGroup]
type SearchAttributeGroupsResult = dyn.OpResult[SearchAttributeGroupsResultData]

var updateAttributeGroupCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_group",
	Action:    "update",
}

type UpdateAttributeGroupCommand struct {
	domain.AttributeGroup
}

func (UpdateAttributeGroupCommand) CqrsRequestType() cqrs.RequestType {
	return updateAttributeGroupCommandType
}

func (this UpdateAttributeGroupCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.AttributeGroupSchemaName)
}

type UpdateAttributeGroupResult = dyn.OpResult[dyn.MutateResultData]
