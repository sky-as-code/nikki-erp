package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

type organizationRestParams struct {
	dig.In

	Config  config.ConfigService
	Logger  logging.LoggerService
	CqrsBus cqrs.CqrsBus
}

func NewOrganizationRest(params organizationRestParams) *OrganizationRest {
	return &OrganizationRest{
		RestBase: httpserver.RestBase{
			ConfigSvc: params.Config,
			Logger:    params.Logger,
			CqrsBus:   params.CqrsBus,
		},
	}
}

type OrganizationRest struct {
	httpserver.RestBase
}

func (this OrganizationRest) CreateOrganization(echoCtx echo.Context) (err error) {
	request := &CreateOrganizationRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.CreateOrganizationResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := CreateOrganizationResponse{}
	response.FromEntity(result.Data)

	return httpserver.JsonCreated(echoCtx, response)
}

func (this OrganizationRest) UpdateOrganization(echoCtx echo.Context) (err error) {
	request := &UpdateOrganizationRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.UpdateOrganizationResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := UpdateOrganizationResponse{}
	response.FromEntity(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this OrganizationRest) DeleteOrganization(echoCtx echo.Context) (err error) {
	request := &DeleteOrganizationRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.DeleteOrganizationResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := DeleteOrganizationResponse{}
	response.FromNonEntity(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this OrganizationRest) GetOrganizationBySlug(echoCtx echo.Context) (err error) {
	request := &GetOrganizationBySlugRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.GetOrganizationBySlugResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := GetOrganizationBySlugResponse{}
	response.FromOrg(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this OrganizationRest) SearchOrganizations(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to search organizations"); e != nil {
			err = e
		}
	}()

	request := &SearchOrganizationsRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.SearchOrganizationsResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := SearchOrganizationsResponse{}
	response.FromResult(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}
