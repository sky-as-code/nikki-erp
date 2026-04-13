package unitcategory

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateUnitCategoryCommand)(nil)
	req = (*DeleteUnitCategoryCommand)(nil)
	req = (*GetUnitCategoryQuery)(nil)
	req = (*SearchUnitCategoriesQuery)(nil)
	req = (*UpdateUnitCategoryCommand)(nil)
	req = (*UnitCategoryExistsQuery)(nil)
	util.Unused(req)
}

var createUnitCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "create",
}

type CreateUnitCategoryCommand struct {
	domain.UnitCategory
}

func (CreateUnitCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return createUnitCategoryCommandType
}

func (this CreateUnitCategoryCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UnitCategorySchemaName)
}

type CreateUnitCategoryResult = dyn.OpResult[domain.UnitCategory]

var updateUnitCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "update",
}

type UpdateUnitCategoryCommand struct {
	domain.UnitCategory
}

func (UpdateUnitCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return updateUnitCategoryCommandType
}

func (this UpdateUnitCategoryCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UnitCategorySchemaName)
}

type UpdateUnitCategoryResult = dyn.OpResult[dyn.MutateResultData]

var deleteUnitCategoryCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "delete",
}

type DeleteUnitCategoryCommand dyn.DeleteOneCommand

func (DeleteUnitCategoryCommand) CqrsRequestType() cqrs.RequestType {
	return deleteUnitCategoryCommandType
}

type DeleteUnitCategoryResult = dyn.OpResult[dyn.MutateResultData]

var getUnitCategoryQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "getUnitCategory",
}

type GetUnitCategoryQuery struct {
	Columns []string `json:"columns" query:"columns"`
	Id      *string  `json:"id" param:"id"`
}

func (GetUnitCategoryQuery) CqrsRequestType() cqrs.RequestType {
	return getUnitCategoryQueryType
}

type GetUnitCategoryResult = dyn.OpResult[domain.UnitCategory]

var searchUnitCategoriesQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "search",
}

type SearchUnitCategoriesQuery dyn.SearchQuery

func (SearchUnitCategoriesQuery) CqrsRequestType() cqrs.RequestType {
	return searchUnitCategoriesQueryType
}

type SearchUnitCategoriesResultData = dyn.PagedResultData[domain.UnitCategory]
type SearchUnitCategoriesResult = dyn.OpResult[SearchUnitCategoriesResultData]

var unitCategoryExistsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit_category",
	Action:    "exists",
}

type UnitCategoryExistsQuery dyn.ExistsQuery

func (UnitCategoryExistsQuery) CqrsRequestType() cqrs.RequestType {
	return unitCategoryExistsQueryType
}

type UnitCategoryExistsResult = dyn.OpResult[dyn.ExistsResultData]
