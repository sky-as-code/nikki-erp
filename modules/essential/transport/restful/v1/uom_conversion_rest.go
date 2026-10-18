package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

type uomConversionRestParams struct {
	dig.In

	ConversionSvc itUom.UomConversionApplicationService
}

func NewUomConversionRest(params uomConversionRestParams) *UomConversionRest {
	return &UomConversionRest{conversionSvc: params.ConversionSvc}
}

// UomConversionRest exposes the shared conversion capability of BR-UOM-ESS-013 as two collection
// actions of essential_uom, for the frontend and out-of-process consumers. In-process Go modules
// depend on itUom.UomConversionAppService through their own interfaces/external port instead.
type UomConversionRest struct {
	conversionSvc itUom.UomConversionApplicationService
}

func (this *UomConversionRest) Convert(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "convert uom quantity", payload,
		this.conversionSvc.ConvertAction, NewConvertQuantityResponseAny)
}

func (this *UomConversionRest) ToReference(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "convert uom quantity to reference", payload,
		this.conversionSvc.ToReferenceAction, NewToReferenceResponseAny)
}
