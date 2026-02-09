package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

type variantRestParams struct {
	dig.In

	VariantSvc itVariant.VariantService
}

func NewVariantRest(params variantRestParams) *VariantRest {
	return &VariantRest{
		VariantSvc: params.VariantSvc,
	}
}

type VariantRest struct {
	httpserver.RestBase
	VariantSvc itVariant.VariantService
}

func (this VariantRest) CreateVariant(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create variant"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.VariantSvc.CreateVariant,
		func(request CreateVariantRequest) itVariant.CreateVariantCommand {
			return itVariant.CreateVariantCommand(request)
		},
		func(result itVariant.CreateVariantResult) CreateVariantResponse {
			response := CreateVariantResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this VariantRest) UpdateVariant(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update variant"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.VariantSvc.UpdateVariant,
		func(request UpdateVariantRequest) itVariant.UpdateVariantCommand {
			return itVariant.UpdateVariantCommand(request)
		},
		func(result itVariant.UpdateVariantResult) UpdateVariantResponse {
			response := UpdateVariantResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this VariantRest) DeleteVariant(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete variant"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.VariantSvc.DeleteVariant,
		func(request DeleteVariantRequest) itVariant.DeleteVariantCommand {
			return itVariant.DeleteVariantCommand(request)
		},
		func(result itVariant.DeleteVariantResult) DeleteVariantResponse {
			response := DeleteVariantResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this VariantRest) GetVariantById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get variant by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.VariantSvc.GetVariantById,
		func(request GetVariantByIdRequest) itVariant.GetVariantByIdQuery {
			return itVariant.GetVariantByIdQuery(request)
		},
		func(result itVariant.GetVariantByIdResult) GetVariantByIdResponse {
			response := GetVariantByIdResponse{}
			response.FromVariant(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this VariantRest) SearchVariants(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search variants"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.VariantSvc.SearchVariants,
		func(request SearchVariantsRequest) itVariant.SearchVariantsQuery {
			return itVariant.SearchVariantsQuery(request)
		},
		func(result itVariant.SearchVariantsResult) SearchVariantsResponse {
			response := SearchVariantsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
