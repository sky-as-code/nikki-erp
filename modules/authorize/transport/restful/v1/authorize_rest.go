package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize"
)

type authorizeRestParams struct {
	dig.In

	AuthorizeSvc it.AuthorizeService
}

func NewAuthorizeRest(params authorizeRestParams) *AuthorizeRest {
	return &AuthorizeRest{
		AuthorizeSvc: params.AuthorizeSvc,
	}
}

type AuthorizeRest struct {
	httpserver.RestBase
	AuthorizeSvc it.AuthorizeService
}

func (this AuthorizeRest) IsAuthorized(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST is authorized"); e != nil {
			err = e
		}
	}()

	query := IsAuthorizedRequest{}
	err = echoCtx.Bind(&query)
	if err != nil {
		return err
	}

	result, err := this.AuthorizeSvc.IsAuthorized(echoCtx.Request().Context(), query)
	if err != nil {
		return err
	}

	response := IsAuthorizedResponse{}
	response.FromResult(result)
	return echoCtx.JSON(http.StatusOK, response)
}