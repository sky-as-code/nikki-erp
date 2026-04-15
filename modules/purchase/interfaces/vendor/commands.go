package vendor

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateVendorCommand)(nil)
	req = (*DeleteVendorCommand)(nil)
	req = (*GetVendorQuery)(nil)
	req = (*SearchVendorsQuery)(nil)
	req = (*UpdateVendorCommand)(nil)
	req = (*SetVendorIsArchivedCommand)(nil)
	req = (*VendorExistsQuery)(nil)
	util.Unused(req)
}

var createCommandType = cqrs.RequestType{Module: "purchase", Submodule: "vendor", Action: "create"}

type CreateVendorCommand struct{ domain.Vendor }

func (CreateVendorCommand) CqrsRequestType() cqrs.RequestType { return createCommandType }
func (CreateVendorCommand) GetSchema() *dmodel.ModelSchema    { return dmodel.GetSchema(domain.VendorSchemaName) }

type CreateVendorResult = dyn.OpResult[domain.Vendor]

var updateCommandType = cqrs.RequestType{Module: "purchase", Submodule: "vendor", Action: "update"}

type UpdateVendorCommand struct{ domain.Vendor }

func (UpdateVendorCommand) CqrsRequestType() cqrs.RequestType { return updateCommandType }
func (UpdateVendorCommand) GetSchema() *dmodel.ModelSchema    { return dmodel.GetSchema(domain.VendorSchemaName) }

type UpdateVendorResult = dyn.OpResult[dyn.MutateResultData]

var deleteCommandType = cqrs.RequestType{Module: "purchase", Submodule: "vendor", Action: "delete"}

type DeleteVendorCommand dyn.DeleteOneCommand

func (DeleteVendorCommand) CqrsRequestType() cqrs.RequestType { return deleteCommandType }

type DeleteVendorResult = dyn.OpResult[dyn.MutateResultData]

var getQueryType = cqrs.RequestType{Module: "purchase", Submodule: "vendor", Action: "get"}

type GetVendorQuery dyn.GetOneQuery

func (GetVendorQuery) CqrsRequestType() cqrs.RequestType { return getQueryType }

type GetVendorResult = dyn.OpResult[domain.Vendor]

var searchQueryType = cqrs.RequestType{Module: "purchase", Submodule: "vendor", Action: "search"}

type SearchVendorsQuery dyn.SearchQuery

func (SearchVendorsQuery) CqrsRequestType() cqrs.RequestType { return searchQueryType }

type SearchVendorsResultData = dyn.PagedResultData[domain.Vendor]
type SearchVendorsResult = dyn.OpResult[SearchVendorsResultData]

var setArchivedCommandType = cqrs.RequestType{Module: "purchase", Submodule: "vendor", Action: "set_archived"}

type SetVendorIsArchivedCommand dyn.SetIsArchivedCommand

func (SetVendorIsArchivedCommand) CqrsRequestType() cqrs.RequestType { return setArchivedCommandType }

type SetVendorIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var existsQueryType = cqrs.RequestType{Module: "purchase", Submodule: "vendor", Action: "exists"}

type VendorExistsQuery dyn.ExistsQuery

func (VendorExistsQuery) CqrsRequestType() cqrs.RequestType { return existsQueryType }

type VendorExistsResult = dyn.OpResult[dyn.ExistsResultData]
