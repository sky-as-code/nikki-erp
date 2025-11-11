package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attributevalue/interfaces"
)

type attributeValueRestParams struct {
	dig.In

	AttributeValueSvc it.AttributeValueService
}

func NewAttributeValueRest(params attributeValueRestParams) *AttributeValueRest {
	return &AttributeValueRest{
		AttributeValueSvc: params.AttributeValueSvc,
	}
}

type AttributeValueRest struct {
	httpserver.RestBase
	AttributeValueSvc it.AttributeValueService
}

func (this AttributeValueRest) CreateAttributeValue(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create attribute value"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeValueSvc.CreateAttributeValue,
		func(request CreateAttributeValueRequest) it.CreateAttributeValueCommand {
			return it.CreateAttributeValueCommand(request)
		},
		func(result it.CreateAttributeValueResult) CreateAttributeValueResponse {
			response := CreateAttributeValueResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this AttributeValueRest) UpdateAttributeValue(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update attribute value"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeValueSvc.UpdateAttributeValue,
		func(request UpdateAttributeValueRequest) it.UpdateAttributeValueCommand {
			return it.UpdateAttributeValueCommand(request)
		},
		func(result it.UpdateAttributeValueResult) UpdateAttributeValueResponse {
			response := UpdateAttributeValueResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeValueRest) DeleteAttributeValue(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete attribute value"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeValueSvc.DeleteAttributeValue,
		func(request DeleteAttributeValueRequest) it.DeleteAttributeValueCommand {
			return it.DeleteAttributeValueCommand(request)
		},
		func(result it.DeleteAttributeValueResult) DeleteAttributeValueResponse {
			response := DeleteAttributeValueResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeValueRest) GetAttributeValueById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get attribute value by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeValueSvc.GetAttributeValueById,
		func(request GetAttributeValueByIdRequest) it.GetAttributeValueByIdQuery {
			return it.GetAttributeValueByIdQuery(request)
		},
		func(result it.GetAttributeValueByIdResult) GetAttributeValueByIdResponse {
			response := GetAttributeValueByIdResponse{}
			response.FromAttributeValue(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this AttributeValueRest) SearchAttributeValues(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search attribute values"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeValueSvc.SearchAttributeValues,
		func(request SearchAttributeValuesRequest) it.SearchAttributeValuesQuery {
			return it.SearchAttributeValuesQuery(request)
		},
		func(result it.SearchAttributeValuesResult) SearchAttributeValuesResponse {
			response := SearchAttributeValuesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
