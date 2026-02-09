package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
)

type attributeRestParams struct {
	dig.In

	AttributeSvc itAttribute.AttributeService
}

func NewAttributeRest(params attributeRestParams) *AttributeRest {
	return &AttributeRest{
		AttributeSvc: params.AttributeSvc,
	}
}

type AttributeRest struct {
	httpserver.RestBase
	AttributeSvc itAttribute.AttributeService
}

func (this AttributeRest) CreateAttribute(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create attribute"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeSvc.CreateAttribute,
		func(request CreateAttributeRequest) itAttribute.CreateAttributeCommand {
			return itAttribute.CreateAttributeCommand(request)
		},
		func(result itAttribute.CreateAttributeResult) CreateAttributeResponse {
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
		func(request UpdateAttributeRequest) itAttribute.UpdateAttributeCommand {
			return itAttribute.UpdateAttributeCommand(request)
		},
		func(result itAttribute.UpdateAttributeResult) UpdateAttributeResponse {
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
		func(request DeleteAttributeRequest) itAttribute.DeleteAttributeCommand {
			return itAttribute.DeleteAttributeCommand(request)
		},
		func(result itAttribute.DeleteAttributeResult) DeleteAttributeResponse {
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
		func(request GetAttributeByIdRequest) itAttribute.GetAttributeByIdQuery {
			return itAttribute.GetAttributeByIdQuery(request)
		},
		func(result itAttribute.GetAttributeByIdResult) GetAttributeByIdResponse {
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
		func(request SearchAttributesRequest) itAttribute.SearchAttributesQuery {
			return itAttribute.SearchAttributesQuery(request)
		},
		func(result itAttribute.SearchAttributesResult) SearchAttributesResponse {
			response := SearchAttributesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
