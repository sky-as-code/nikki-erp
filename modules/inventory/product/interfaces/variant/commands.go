package variant

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateVariantCommand)(nil)
	req = (*DeleteVariantCommand)(nil)
	req = (*VariantExistsQuery)(nil)
	req = (*GetVariantQuery)(nil)
	req = (*SearchVariantsQuery)(nil)
	req = (*UpdateVariantCommand)(nil)
	util.Unused(req)
}

var createVariantCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "create",
}

type CreateVariantCommand struct {
	domain.Variant
}

func (CreateVariantCommand) CqrsRequestType() cqrs.RequestType {
	return createVariantCommandType
}

func (this CreateVariantCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.VariantSchemaName)
}

type CreateVariantResult = dyn.OpResult[domain.Variant]

var updateVariantCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "update",
}

type UpdateVariantCommand struct {
	domain.Variant
}

func (UpdateVariantCommand) CqrsRequestType() cqrs.RequestType {
	return updateVariantCommandType
}

func (this UpdateVariantCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.VariantSchemaName)
}

type UpdateVariantResult = dyn.OpResult[dyn.MutateResultData]

var deleteVariantCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "delete",
}

type DeleteVariantCommand dyn.DeleteOneCommand

func (DeleteVariantCommand) CqrsRequestType() cqrs.RequestType {
	return deleteVariantCommandType
}

type DeleteVariantResult = dyn.OpResult[dyn.MutateResultData]

var getVariantQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "getVariant",
}

type GetVariantQuery struct {
	Columns []string `json:"columns" query:"columns"`
	Id      *string  `json:"id" param:"id"`
}

func (GetVariantQuery) CqrsRequestType() cqrs.RequestType {
	return getVariantQueryType
}

type GetVariantResult = dyn.OpResult[domain.Variant]

var variantExistsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "variantExists",
}

type VariantExistsQuery dyn.ExistsQuery

func (VariantExistsQuery) CqrsRequestType() cqrs.RequestType {
	return variantExistsQueryType
}

type VariantExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchVariantsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "search",
}

type SearchVariantsQuery dyn.SearchQuery

func (SearchVariantsQuery) CqrsRequestType() cqrs.RequestType {
	return searchVariantsQueryType
}

type SearchVariantsResultData = dyn.PagedResultData[domain.Variant]
type SearchVariantsResult = dyn.OpResult[SearchVariantsResultData]
