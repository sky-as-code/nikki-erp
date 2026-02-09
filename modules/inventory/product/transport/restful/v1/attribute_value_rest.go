package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
)

type attributeValueRestParams struct {
	dig.In

	AttributeValueSvc itAttributeValue.AttributeValueService
}

func NewAttributeValueRest(params attributeValueRestParams) *AttributeValueRest {
	return &AttributeValueRest{
		AttributeValueSvc: params.AttributeValueSvc,
	}
}

type AttributeValueRest struct {
	httpserver.RestBase
	AttributeValueSvc itAttributeValue.AttributeValueService
}

func (this AttributeValueRest) CreateAttributeValue(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create attribute value"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.AttributeValueSvc.CreateAttributeValue,
		func(request CreateAttributeValueRequest) itAttributeValue.CreateAttributeValueCommand {
			return itAttributeValue.CreateAttributeValueCommand(request)
		},
		func(result itAttributeValue.CreateAttributeValueResult) CreateAttributeValueResponse {
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
		func(request UpdateAttributeValueRequest) itAttributeValue.UpdateAttributeValueCommand {
			return itAttributeValue.UpdateAttributeValueCommand(request)
		},
		func(result itAttributeValue.UpdateAttributeValueResult) UpdateAttributeValueResponse {
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
		func(request DeleteAttributeValueRequest) itAttributeValue.DeleteAttributeValueCommand {
			return itAttributeValue.DeleteAttributeValueCommand(request)
		},
		func(result itAttributeValue.DeleteAttributeValueResult) DeleteAttributeValueResponse {
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
		func(request GetAttributeValueByIdRequest) itAttributeValue.GetAttributeValueByIdQuery {
			return itAttributeValue.GetAttributeValueByIdQuery(request)
		},
		func(result itAttributeValue.GetAttributeValueByIdResult) GetAttributeValueByIdResponse {
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
		func(request SearchAttributeValuesRequest) itAttributeValue.SearchAttributeValuesQuery {
			return itAttributeValue.SearchAttributeValuesQuery(request)
		},
		func(result itAttributeValue.SearchAttributeValuesResult) SearchAttributeValuesResponse {
			response := SearchAttributeValuesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
