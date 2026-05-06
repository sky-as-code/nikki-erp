package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

type enumRestParams struct {
	dig.In

	EnumSvc it.EnumService
}

func NewEnumRest(params enumRestParams) *EnumRest {
	return &EnumRest{
		enumSvc: params.EnumSvc,
	}
}

type EnumRest struct {
	enumSvc it.EnumService
}

func (this EnumRest) Create(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create enum"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.enumSvc.CreateEnum,
		func(request CreateEnumRequest) it.CreateEnumCommand {
			cmd := it.CreateEnumCommand{}
			cmd.SetFieldData(request.DynamicFields)
			return cmd
		},
		func(data models.Enum) CreateEnumResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated,
	)
}

func (this EnumRest) Delete(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete enum"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.enumSvc.DeleteEnum,
		func(request DeleteEnumRequest) it.DeleteEnumCommand {
			return it.DeleteEnumCommand(request)
		},
		func(data dyn.MutateResultData) DeleteEnumResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this EnumRest) Exists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST enum exists"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.enumSvc.EnumExists,
		func(request EnumExistsRequest) it.EnumExistsQuery {
			return it.EnumExistsQuery(request)
		},
		func(data dyn.ExistsResultData) EnumExistsResponse {
			return EnumExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this EnumRest) GetOne(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get enum"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.enumSvc.GetEnum,
		func(request GetEnumRequest) it.GetEnumQuery {
			return it.GetEnumQuery(request)
		},
		func(data models.Enum) GetEnumResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this EnumRest) Search(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search enums"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.enumSvc.SearchEnums,
		func(request SearchEnumsRequest) it.SearchEnumsQuery {
			return it.SearchEnumsQuery(request)
		},
		func(data it.SearchEnumsResultData) SearchEnumsResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
	)
}

func (this EnumRest) Update(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update enum"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.enumSvc.UpdateEnum,
		func(request UpdateEnumRequest) it.UpdateEnumCommand {
			cmd := it.UpdateEnumCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.EnumId)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
