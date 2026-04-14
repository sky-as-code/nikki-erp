package module

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateModuleCommand)(nil)
	req = (*DeleteModuleCommand)(nil)
	req = (*GetModuleQuery)(nil)
	req = (*SearchModulesQuery)(nil)
	req = (*UpdateModuleCommand)(nil)
	req = (*ModuleExistsQuery)(nil)
	util.Unused(req)
}

var createModuleCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "module_metadata",
	Action:    "create",
}

type CreateModuleCommand struct {
	domain.ModuleMetadata
}

func (CreateModuleCommand) CqrsRequestType() cqrs.RequestType {
	return createModuleCommandType
}

func (CreateModuleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ModuleMetadataSchemaName)
}

type CreateModuleResult = dyn.OpResult[domain.ModuleMetadata]

var updateModuleCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "module_metadata",
	Action:    "update",
}

type UpdateModuleCommand struct {
	domain.ModuleMetadata
}

func (UpdateModuleCommand) CqrsRequestType() cqrs.RequestType {
	return updateModuleCommandType
}

func (UpdateModuleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ModuleMetadataSchemaName)
}

type UpdateModuleResult = dyn.OpResult[dyn.MutateResultData]

var deleteModuleCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "module_metadata",
	Action:    "delete",
}

type DeleteModuleCommand dyn.DeleteOneCommand

func (DeleteModuleCommand) CqrsRequestType() cqrs.RequestType {
	return deleteModuleCommandType
}

type DeleteModuleResult = dyn.OpResult[dyn.MutateResultData]

var getModuleQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "module_metadata",
	Action:    "get",
}

type GetModuleQuery dyn.GetOneQuery

func (GetModuleQuery) CqrsRequestType() cqrs.RequestType {
	return getModuleQueryType
}

type GetModuleResult = dyn.OpResult[domain.ModuleMetadata]

var searchModulesQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "module_metadata",
	Action:    "search",
}

type SearchModulesQuery dyn.SearchQuery

func (SearchModulesQuery) CqrsRequestType() cqrs.RequestType {
	return searchModulesQueryType
}

type SearchModulesResultData = dyn.PagedResultData[domain.ModuleMetadata]
type SearchModulesResult = dyn.OpResult[SearchModulesResultData]

var moduleExistsQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "module_metadata",
	Action:    "exists",
}

type ModuleExistsQuery dyn.ExistsQuery

func (ModuleExistsQuery) CqrsRequestType() cqrs.RequestType {
	return moduleExistsQueryType
}

type ModuleExistsResult = dyn.OpResult[dyn.ExistsResultData]
