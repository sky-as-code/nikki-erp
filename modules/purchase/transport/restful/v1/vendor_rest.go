package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/vendor"
)

type vendorRestParams struct {
	dig.In
	Svc it.VendorService
}

func NewVendorRest(params vendorRestParams) *VendorRest {
	return &VendorRest{svc: params.Svc}
}

type VendorRest struct{ svc it.VendorService }

func (this VendorRest) CreateVendor(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create vendor"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(echoCtx, this.svc.CreateVendor,
		func(fields dmodel.DynamicFields) it.CreateVendorCommand {
			cmd := it.CreateVendorCommand{Vendor: *domain.NewVendor()}
			cmd.SetFieldData(fields)
			return cmd
		},
		func(data domain.Vendor) CreateVendorResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated)
}
func (this VendorRest) DeleteVendor(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete vendor"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.DeleteVendor,
		func(request DeleteVendorRequest) it.DeleteVendorCommand { return it.DeleteVendorCommand(request) },
		func(data dyn.MutateResultData) DeleteVendorResponse { return httpserver.NewRestDeleteResponse2(data) },
		httpserver.JsonOk)
}
func (this VendorRest) GetVendor(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get vendor"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.GetVendor,
		func(request GetVendorRequest) it.GetVendorQuery { return it.GetVendorQuery(request) },
		func(data domain.Vendor) GetVendorResponse { return data.GetFieldData() }, httpserver.JsonOk)
}
func (this VendorRest) VendorExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST vendor exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.VendorExists,
		func(request VendorExistsRequest) it.VendorExistsQuery { return it.VendorExistsQuery(request) },
		func(data dyn.ExistsResultData) VendorExistsResponse { return VendorExistsResponse(data) },
		httpserver.JsonOk)
}
func (this VendorRest) SearchVendors(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search vendors"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SearchVendors,
		func(request SearchVendorsRequest) it.SearchVendorsQuery { return it.SearchVendorsQuery(request) },
		func(data it.SearchVendorsResultData) SearchVendorsResponse { return httpserver.NewSearchResponseDyn(data) },
		httpserver.JsonOk, true)
}
func (this VendorRest) SetVendorIsArchived(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set vendor archived"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SetVendorIsArchived,
		func(request SetVendorIsArchivedRequest) it.SetVendorIsArchivedCommand {
			return it.SetVendorIsArchivedCommand(request)
		},
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
func (this VendorRest) UpdateVendor(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update vendor"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.UpdateVendor,
		func(request UpdateVendorRequest) it.UpdateVendorCommand {
			cmd := it.UpdateVendorCommand{Vendor: *domain.NewVendor()}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.VendorId)))
			return cmd
		},
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
