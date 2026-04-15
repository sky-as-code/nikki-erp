package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

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

func (this AttributeRest) Create(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create attribute",
		echoCtx,
		&itAttribute.CreateAttributeCommand{},
		this.AttributeSvc.CreateAttribute,
	)
}

func (this AttributeRest) Update(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update attribute",
		echoCtx,
		&itAttribute.UpdateAttributeCommand{},
		this.AttributeSvc.UpdateAttribute,
	)
}

func (this AttributeRest) Delete(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete attribute",
		echoCtx,
		this.AttributeSvc.DeleteAttribute,
	)
}

func (this AttributeRest) GetOne(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get attribute",
		echoCtx,
		this.AttributeSvc.GetAttribute,
	)
}

func (this AttributeRest) Search(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search attributes",
		echoCtx,
		this.AttributeSvc.SearchAttributes,
		true,
	)
}

func (this AttributeRest) Exists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"attribute exists",
		echoCtx,
		this.AttributeSvc.AttributeExists,
	)
}
