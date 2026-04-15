package purchaseorder

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreatePurchaseOrderCommand)(nil)
	req = (*DeletePurchaseOrderCommand)(nil)
	req = (*GetPurchaseOrderQuery)(nil)
	req = (*SearchPurchaseOrdersQuery)(nil)
	req = (*UpdatePurchaseOrderCommand)(nil)
	req = (*SetPurchaseOrderIsArchivedCommand)(nil)
	req = (*PurchaseOrderExistsQuery)(nil)
	util.Unused(req)
}

var createCommandType = cqrs.RequestType{Module: "purchase", Submodule: "purchaseorder", Action: "create"}

type CreatePurchaseOrderCommand struct{ domain.PurchaseOrder }

func (CreatePurchaseOrderCommand) CqrsRequestType() cqrs.RequestType { return createCommandType }
func (CreatePurchaseOrderCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.PurchaseOrderSchemaName)
}

type CreatePurchaseOrderResult = dyn.OpResult[domain.PurchaseOrder]

var updateCommandType = cqrs.RequestType{Module: "purchase", Submodule: "purchaseorder", Action: "update"}

type UpdatePurchaseOrderCommand struct{ domain.PurchaseOrder }

func (UpdatePurchaseOrderCommand) CqrsRequestType() cqrs.RequestType { return updateCommandType }
func (UpdatePurchaseOrderCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.PurchaseOrderSchemaName)
}

type UpdatePurchaseOrderResult = dyn.OpResult[dyn.MutateResultData]

var deleteCommandType = cqrs.RequestType{Module: "purchase", Submodule: "purchaseorder", Action: "delete"}

type DeletePurchaseOrderCommand dyn.DeleteOneCommand

func (DeletePurchaseOrderCommand) CqrsRequestType() cqrs.RequestType { return deleteCommandType }

type DeletePurchaseOrderResult = dyn.OpResult[dyn.MutateResultData]

var getQueryType = cqrs.RequestType{Module: "purchase", Submodule: "purchaseorder", Action: "get"}

type GetPurchaseOrderQuery dyn.GetOneQuery

func (GetPurchaseOrderQuery) CqrsRequestType() cqrs.RequestType { return getQueryType }

type GetPurchaseOrderResult = dyn.OpResult[domain.PurchaseOrder]

var searchQueryType = cqrs.RequestType{Module: "purchase", Submodule: "purchaseorder", Action: "search"}

type SearchPurchaseOrdersQuery dyn.SearchQuery

func (SearchPurchaseOrdersQuery) CqrsRequestType() cqrs.RequestType { return searchQueryType }

type SearchPurchaseOrdersResultData = dyn.PagedResultData[domain.PurchaseOrder]
type SearchPurchaseOrdersResult = dyn.OpResult[SearchPurchaseOrdersResultData]

var setArchivedCommandType = cqrs.RequestType{Module: "purchase", Submodule: "purchaseorder", Action: "set_archived"}

type SetPurchaseOrderIsArchivedCommand dyn.SetIsArchivedCommand

func (SetPurchaseOrderIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setArchivedCommandType
}

type SetPurchaseOrderIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var existsQueryType = cqrs.RequestType{Module: "purchase", Submodule: "purchaseorder", Action: "exists"}

type PurchaseOrderExistsQuery dyn.ExistsQuery

func (PurchaseOrderExistsQuery) CqrsRequestType() cqrs.RequestType { return existsQueryType }

type PurchaseOrderExistsResult = dyn.OpResult[dyn.ExistsResultData]
