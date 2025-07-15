package v1

import (
	"github.com/labstack/echo/v4"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type entitlementRestParams struct {
	dig.In

	Config  config.ConfigService
	Logger  logging.LoggerService
	CqrsBus cqrs.CqrsBus
}

func NewEntitlementRest(params entitlementRestParams) *EntitlementRest {
	return &EntitlementRest{
		RestBase: httpserver.RestBase{
			ConfigSvc: params.Config,
			Logger:    params.Logger,
			CqrsBus:   params.CqrsBus,
		},
	}
}

type EntitlementRest struct {
	httpserver.RestBase
}

func (this EntitlementRest) CreateEntitlement(echoCtx echo.Context) (err error) {
	request := &CreateEntitlementRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.CreateEntitlementResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := CreateEntitlementResponse{}
	response.FromEntitlement(*result.Data)

	return httpserver.JsonCreated(echoCtx, response)
}

func (this EntitlementRest) UpdateEntitlement(echoCtx echo.Context) (err error) {
	request := &UpdateEntitlementRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.UpdateEntitlementResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := UpdateEntitlementResponse{}
	response.FromEntitlement(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this EntitlementRest) GetEntitlementById(echoCtx echo.Context) (err error) {
	request := &GetEntitlementByIdRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.GetEntitlementByIdResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := GetEntitlementByIdResponse{}
	response.FromEntitlement(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this EntitlementRest) SearchEntitlements(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list entitlements"); e != nil {
			err = e
		}
	}()

	request := &SearchEntitlementsRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.SearchEntitlementsResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := SearchEntitlementsResponse{}
	response.FromResult(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}
