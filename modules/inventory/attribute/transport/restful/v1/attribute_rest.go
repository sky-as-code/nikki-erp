package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attribute/interfaces"
)

type attributeRestParams struct {
	dig.In

	AttributeSvc it.AttributeService
}

func NewAttributeRest(params attributeRestParams) *AttributeRest {
	return &AttributeRest{
		AttributeSvc: params.AttributeSvc,
	}
}

type AttributeRest struct {
	httpserver.RestBase
	AttributeSvc it.AttributeService
}

func (this AttributeRest) CreateAttribute(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create attribute"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeSvc.CreateAttribute,
		func(request CreateAttributeRequest) it.CreateAttributeCommand {
			return it.CreateAttributeCommand(request)
		},
		func(result it.CreateAttributeResult) CreateAttributeResponse {
			response := CreateAttributeResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this AttributeRest) UpdateAttribute(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update attribute"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeSvc.UpdateAttribute,
		func(request UpdateAttributeRequest) it.UpdateAttributeCommand {
			return it.UpdateAttributeCommand(request)
		},
		func(result it.UpdateAttributeResult) UpdateAttributeResponse {
			response := UpdateAttributeResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeRest) DeleteAttribute(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete attribute"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeSvc.DeleteAttribute,
		func(request DeleteAttributeRequest) it.DeleteAttributeCommand {
			return it.DeleteAttributeCommand(request)
		},
		func(result it.DeleteAttributeResult) DeleteAttributeResponse {
			response := DeleteAttributeResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeRest) GetAttributeById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get attribute by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeSvc.GetAttributeById,
		func(request GetAttributeByIdRequest) it.GetAttributeByIdQuery {
			return it.GetAttributeByIdQuery(request)
		},
		func(result it.GetAttributeByIdResult) GetAttributeByIdResponse {
			response := GetAttributeByIdResponse{}
			response.FromAttribute(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeRest) SearchAttributes(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search attributes"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeSvc.SearchAttributes,
		func(request SearchAttributesRequest) it.SearchAttributesQuery {
			return it.SearchAttributesQuery(request)
		},
		func(result it.SearchAttributesResult) SearchAttributesResponse {
			response := SearchAttributesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
