package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initSignInV1() error {
	return deps.Invoke(func(route *echo.Group, loginRest *v1.LoginRest, passwordRest *v1.PasswordRest) {
		routeV1 := route.Group("/v1/iam")
		routeV1.POST("/signin/start", loginRest.StartSignInFlow, m.PublicUnauthorized)
		routeV1.POST("/signin/continue", loginRest.ContinueSignInFlow, m.PublicUnauthorized)
		routeV1.POST("/signin/refresh", loginRest.RefreshToken, m.PublicUnauthorized)

		routeV1.POST("/passwords/password", passwordRest.SetPassword, m.UserSmokeAuthz())
		routeV1.POST("/passwords/passwordtmp", passwordRest.CreatePasswordTemp, m.UserSmokeAuthz())
		routeV1.POST("/passwords/passwordotp", passwordRest.CreatePasswordOtp, m.UserSmokeAuthz())
		routeV1.POST("/passwords/passwordotp/confirm", passwordRest.ConfirmPasswordOtp, m.UserSmokeAuthz())
	})
}
