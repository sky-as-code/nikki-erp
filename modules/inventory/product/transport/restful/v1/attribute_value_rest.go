package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

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

func (this AttributeValueRest) Create(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create attribute value",
		echoCtx,
		&itAttributeValue.CreateAttributeValueCommand{},
		this.AttributeValueSvc.CreateAttributeValue,
	)
}

func (this AttributeValueRest) Update(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update attribute value",
		echoCtx,
		&itAttributeValue.UpdateAttributeValueCommand{},
		this.AttributeValueSvc.UpdateAttributeValue,
	)
}

func (this AttributeValueRest) Delete(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete attribute value",
		echoCtx,
		this.AttributeValueSvc.DeleteAttributeValue,
	)
}

func (this AttributeValueRest) GetOne(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get attribute value",
		echoCtx,
		this.AttributeValueSvc.GetAttributeValue,
	)
}

func (this AttributeValueRest) Search(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search attribute values",
		echoCtx,
		this.AttributeValueSvc.SearchAttributeValues,
	)
}

func (this AttributeValueRest) Exists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"attribute value exists",
		echoCtx,
		this.AttributeValueSvc.AttributeValueExists,
	)
}
