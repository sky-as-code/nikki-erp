package requestguard

import (
	"net/http"

	"github.com/labstack/echo/v4"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

type RequestGuardService interface {
	VerifyRequest(request *http.Request) (result *VerifyRequestResult, err error)
	GetCorsMiddleware() (echo.MiddlewareFunc, error)
}

type VerifyRequestResult struct {
	IsOk        bool
	HttpStatus  int
	ClientError *ft.ClientErrorItem
}
