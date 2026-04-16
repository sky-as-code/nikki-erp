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
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

type modelMetadataRestParams struct {
	dig.In
	ModelMetadataSvc it.ModelMetadataService
}

func NewModelMetadataRest(params modelMetadataRestParams) *ModelMetadataRest {
	return &ModelMetadataRest{svc: params.ModelMetadataSvc}
}

type ModelMetadataRest struct{ svc it.ModelMetadataService }

func (this ModelMetadataRest) CreateModelMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create model metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.CreateModelMetadata,
		func(request CreateModelMetadataRequest) it.CreateModelMetadataCommand {
			cmd := it.CreateModelMetadataCommand{}
			cmd.SetFieldData(request.DynamicFields)
			return cmd
		},
		func(data domain.ModelMetadata) CreateModelMetadataResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated)
}

func (this ModelMetadataRest) DeleteModelMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete model metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.DeleteModelMetadata,
		func(request DeleteModelMetadataRequest) it.DeleteModelMetadataCommand {
			return it.DeleteModelMetadataCommand(request)
		},
		func(data dyn.MutateResultData) DeleteModelMetadataResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk)
}

func (this ModelMetadataRest) GetModelMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get model metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.GetModelMetadata,
		func(request GetModelMetadataRequest) it.GetModelMetadataQuery {
			return it.GetModelMetadataQuery(request)
		},
		func(data domain.ModelMetadata) GetModelMetadataResponse { return data.GetFieldData() },
		httpserver.JsonOk)
}

func (this ModelMetadataRest) ModelMetadataExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST model metadata exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.ModelMetadataExists,
		func(request ModelMetadataExistsRequest) it.ModelMetadataExistsQuery {
			return it.ModelMetadataExistsQuery(request)
		},
		func(data dyn.ExistsResultData) ModelMetadataExistsResponse { return ModelMetadataExistsResponse(data) },
		httpserver.JsonOk)
}

func (this ModelMetadataRest) SearchModelMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search model metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SearchModelMetadata,
		func(request SearchModelMetadataRequest) it.SearchModelMetadataQuery {
			return it.SearchModelMetadataQuery(request)
		},
		func(data it.SearchModelMetadataResultData) SearchModelMetadataResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk, true)
}

func (this ModelMetadataRest) UpdateModelMetadata(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update model metadata"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.UpdateModelMetadata,
		func(request UpdateModelMetadataRequest) it.UpdateModelMetadataCommand {
			cmd := it.UpdateModelMetadataCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.Id)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk)
}
