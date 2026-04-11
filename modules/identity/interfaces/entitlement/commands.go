package entitlement

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func init() {
	var req cqrs.Request
	req = (*CreateEntitlementCommand)(nil)
	req = (*DeleteEntitlementCommand)(nil)
	req = (*GetEntitlementQuery)(nil)
	req = (*EntitlementExistsQuery)(nil)
	req = (*SearchEntitlementsQuery)(nil)
	req = (*UpdateEntitlementCommand)(nil)
	req = (*SetEntitlementIsArchivedCommand)(nil)
	req = (*ManageEntitlementRolesCommand)(nil)
	util.Unused(req)
}

var createEntitlementCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "entitlement", Action: "createEntitlement",
}

type CreateEntitlementCommand struct {
	domain.Entitlement
}

func (CreateEntitlementCommand) CqrsRequestType() cqrs.RequestType {
	return createEntitlementCommandType
}

func (CreateEntitlementCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.EntitlementSchemaName)
}

type CreateEntitlementResult = dyn.OpResult[domain.Entitlement]

var deleteEntitlementCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "entitlement", Action: "deleteEntitlement",
}

type DeleteEntitlementCommand dyn.DeleteOneCommand

func (DeleteEntitlementCommand) CqrsRequestType() cqrs.RequestType {
	return deleteEntitlementCommandType
}

type DeleteEntitlementResult = dyn.OpResult[dyn.MutateResultData]

var getEntitlementQueryType = cqrs.RequestType{Module: "identity", Submodule: "entitlement", Action: "getEntitlement"}

type GetEntitlementQuery dyn.GetOneQuery

func (GetEntitlementQuery) CqrsRequestType() cqrs.RequestType { return getEntitlementQueryType }

type GetEntitlementResult = dyn.OpResult[domain.Entitlement]

var entitlementExistsQueryType = cqrs.RequestType{
	Module: "identity", Submodule: "entitlement", Action: "entitlementExists",
}

type EntitlementExistsQuery dyn.ExistsQuery

func (EntitlementExistsQuery) CqrsRequestType() cqrs.RequestType { return entitlementExistsQueryType }

type EntitlementExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchEntitlementsQueryType = cqrs.RequestType{
	Module: "identity", Submodule: "entitlement", Action: "searchEntitlements",
}

type SearchEntitlementsQuery dyn.SearchQuery

func (SearchEntitlementsQuery) CqrsRequestType() cqrs.RequestType { return searchEntitlementsQueryType }

type SearchEntitlementsResultData = dyn.PagedResultData[domain.Entitlement]
type SearchEntitlementsResult = dyn.OpResult[SearchEntitlementsResultData]

var updateEntitlementCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "entitlement", Action: "updateEntitlement",
}

type UpdateEntitlementCommand struct {
	domain.Entitlement
}

func (UpdateEntitlementCommand) CqrsRequestType() cqrs.RequestType {
	return updateEntitlementCommandType
}

func (UpdateEntitlementCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.EntitlementSchemaName)
}

type UpdateEntitlementResult = dyn.OpResult[dyn.MutateResultData]

var setEntitlementIsArchivedCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "entitlement", Action: "setEntitlementIsArchived",
}

type SetEntitlementIsArchivedCommand dyn.SetIsArchivedCommand

func (SetEntitlementIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setEntitlementIsArchivedCommandType
}

type SetEntitlementIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var manageEntitlementRolesCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "entitlement", Action: "manageEntitlementRoles",
}

type ManageEntitlementRolesCommand struct {
	EntitlementId model.Id                    `json:"entitlement_id" param:"entitlement_id"`
	Add           datastructure.Set[model.Id] `json:"add"`
	Remove        datastructure.Set[model.Id] `json:"remove"`
}

func (ManageEntitlementRolesCommand) CqrsRequestType() cqrs.RequestType {
	return manageEntitlementRolesCommandType
}

type ManageEntitlementRolesResult = dyn.OpResult[dyn.MutateResultData]
