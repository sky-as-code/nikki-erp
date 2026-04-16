package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

type fieldMetadataRestParams struct {
	dig.In
	FieldMetadataSvc it.FieldMetadataService
}

func NewFieldMetadataRest(params fieldMetadataRestParams) *FieldMetadataRest {
	return &FieldMetadataRest{svc: params.FieldMetadataSvc}
}

type FieldMetadataRest struct{ svc it.FieldMetadataService }

func (this FieldMetadataRest) CreateFieldMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create field metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.CreateFieldMetadata,
		func(request CreateFieldMetadataRequest) it.CreateFieldMetadataCommand {
			cmd := it.CreateFieldMetadataCommand{}
			cmd.SetFieldData(request.DynamicFields)
			return cmd
		},
		func(data domain.FieldMetadata) CreateFieldMetadataResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated)
}

func (this FieldMetadataRest) DeleteFieldMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete field metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.DeleteFieldMetadata,
		func(request DeleteFieldMetadataRequest) it.DeleteFieldMetadataCommand {
			return it.DeleteFieldMetadataCommand(request)
		},
		func(data dyn.MutateResultData) DeleteFieldMetadataResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk)
}

func (this FieldMetadataRest) GetFieldMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get field metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.GetFieldMetadata,
		func(request GetFieldMetadataRequest) it.GetFieldMetadataQuery {
			return it.GetFieldMetadataQuery(request)
		},
		func(data domain.FieldMetadata) GetFieldMetadataResponse { return data.GetFieldData() },
		httpserver.JsonOk)
}

func (this FieldMetadataRest) FieldMetadataExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST field metadata exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.FieldMetadataExists,
		func(request FieldMetadataExistsRequest) it.FieldMetadataExistsQuery {
			return it.FieldMetadataExistsQuery(request)
		},
		func(data dyn.ExistsResultData) FieldMetadataExistsResponse { return FieldMetadataExistsResponse(data) },
		httpserver.JsonOk)
}

func (this FieldMetadataRest) SearchFieldMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search field metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SearchFieldMetadata,
		func(request SearchFieldMetadataRequest) it.SearchFieldMetadataQuery {
			return it.SearchFieldMetadataQuery(request)
		},
		func(data it.SearchFieldMetadataResultData) SearchFieldMetadataResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk, true)
}

func (this FieldMetadataRest) UpdateFieldMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update field metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.UpdateFieldMetadata,
		func(request UpdateFieldMetadataRequest) it.UpdateFieldMetadataCommand {
			cmd := it.UpdateFieldMetadataCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.Id)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk)
}
