package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

type uomConversionRestParams struct {
	dig.In

	ConversionSvc itUom.UomConversionAppService
}

func NewUomConversionRest(params uomConversionRestParams) *UomConversionRest {
	return &UomConversionRest{
		conversionSvc: params.ConversionSvc,
	}
}

// UomConversionRest exposes the shared conversion capability of BR-UOM-ESS-013 over HTTP,
// for the frontend and for out-of-process consumers. In-process Go modules should depend on
// itUom.UomConversionAppService through their own interfaces/external port instead.
type UomConversionRest struct {
	conversionSvc itUom.UomConversionAppService
}

func (this UomConversionRest) Convert(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST convert uom quantity"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.conversionSvc.Convert,
		ConvertQuantityRequest.ToQuery,
		NewConvertQuantityResponse,
		httpserver.JsonOk,
	)
}

func (this UomConversionRest) ToReference(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST convert uom quantity to reference"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.conversionSvc.ToReference,
		ToReferenceRequest.ToQuery,
		NewToReferenceResponse,
		httpserver.JsonOk,
	)
}
