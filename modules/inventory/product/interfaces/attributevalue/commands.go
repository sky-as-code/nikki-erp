package attributevalue

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateAttributeValueCommand)(nil)
	req = (*DeleteAttributeValueCommand)(nil)
	req = (*AttributeValueExistsQuery)(nil)
	req = (*GetAttributeValueQuery)(nil)
	req = (*SearchAttributeValuesQuery)(nil)
	req = (*UpdateAttributeValueCommand)(nil)
	util.Unused(req)
}

var createAttributeValueCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "create",
}

type CreateAttributeValueCommand struct {
	domain.AttributeValue
}

func (CreateAttributeValueCommand) CqrsRequestType() cqrs.RequestType {
	return createAttributeValueCommandType
}

func (this CreateAttributeValueCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.AttributeValueSchemaName)
}

type CreateAttributeValueResult = dyn.OpResult[domain.AttributeValue]

var updateAttributeValueCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "update",
}

type UpdateAttributeValueCommand struct {
	domain.AttributeValue
}

func (UpdateAttributeValueCommand) CqrsRequestType() cqrs.RequestType {
	return updateAttributeValueCommandType
}

func (this UpdateAttributeValueCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.AttributeValueSchemaName)
}

type UpdateAttributeValueResult = dyn.OpResult[dyn.MutateResultData]

var deleteAttributeValueCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "delete",
}

type DeleteAttributeValueCommand dyn.DeleteOneCommand

func (DeleteAttributeValueCommand) CqrsRequestType() cqrs.RequestType {
	return deleteAttributeValueCommandType
}

type DeleteAttributeValueResult = dyn.OpResult[dyn.MutateResultData]

var getAttributeValueQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "getAttributeValue",
}

type GetAttributeValueQuery struct {
	Columns []string `json:"columns" query:"columns"`
	Id      *string  `json:"id" param:"id"`
}

func (GetAttributeValueQuery) CqrsRequestType() cqrs.RequestType {
	return getAttributeValueQueryType
}

type GetAttributeValueResult = dyn.OpResult[domain.AttributeValue]

var attributeValueExistsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "attributeValueExists",
}

type AttributeValueExistsQuery dyn.ExistsQuery

func (AttributeValueExistsQuery) CqrsRequestType() cqrs.RequestType {
	return attributeValueExistsQueryType
}

type AttributeValueExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchAttributeValuesQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "search",
}

type SearchAttributeValuesQuery dyn.SearchQuery

func (SearchAttributeValuesQuery) CqrsRequestType() cqrs.RequestType {
	return searchAttributeValuesQueryType
}

type SearchAttributeValuesResultData = dyn.PagedResultData[domain.AttributeValue]
type SearchAttributeValuesResult = dyn.OpResult[SearchAttributeValuesResultData]
