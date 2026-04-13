package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
)

type attributeGroupRestParams struct {
	dig.In

	AttributeGroupSvc itAttributeGroup.AttributeGroupService
}

func NewAttributeGroupRest(params attributeGroupRestParams) *AttributeGroupRest {
	return &AttributeGroupRest{
		AttributeGroupSvc: params.AttributeGroupSvc,
	}
}

type AttributeGroupRest struct {
	httpserver.RestBase
	AttributeGroupSvc itAttributeGroup.AttributeGroupService
}

func (this AttributeGroupRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create attribute group",
		echoCtx,
		&itAttributeGroup.CreateAttributeGroupCommand{},
		this.AttributeGroupSvc.CreateAttributeGroup,
	)
}

func (this AttributeGroupRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update attribute group",
		echoCtx,
		&itAttributeGroup.UpdateAttributeGroupCommand{},
		this.AttributeGroupSvc.UpdateAttributeGroup,
	)
}

func (this AttributeGroupRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete attribute group",
		echoCtx,
		this.AttributeGroupSvc.DeleteAttributeGroup,
	)
}

func (this AttributeGroupRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get attribute group",
		echoCtx,
		this.AttributeGroupSvc.GetAttributeGroup,
	)
}

func (this AttributeGroupRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search attribute groups",
		echoCtx,
		this.AttributeGroupSvc.SearchAttributeGroups,
		true,
	)
}

func (this AttributeGroupRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"attribute group exists",
		echoCtx,
		this.AttributeGroupSvc.AttributeGroupExists,
	)
}
