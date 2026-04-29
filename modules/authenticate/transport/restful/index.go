package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/authenticate/transport/restful/v1"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewLoginRest,
		v1.NewPasswordRest,
	)
	err = stdErr.Join(
		err,
		initAuthnV1(),
	)
	return err
}

func initAuthnV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		loginRest *v1.LoginRest,
		passwordRest *v1.PasswordRest,
	) {
		v1 := route.Group("/v1/authn")

		v1.POST("/login/start", loginRest.StartLoginFlow, m.PublicUnauthorized)
		v1.POST("/login", loginRest.Authenticate, m.PublicUnauthorized)
		v1.POST("/refresh", loginRest.RefreshToken, m.PublicUnauthorized)

		v1.POST("/passwords/password", passwordRest.SetPassword, m.SmokeAuthz())
		v1.POST("/passwords/passwordtmp", passwordRest.CreatePasswordTemp, m.SmokeAuthz())
		v1.POST("/passwords/passwordotp", passwordRest.CreatePasswordOtp, m.SmokeAuthz())
		v1.POST("/passwords/passwordotp/confirm", passwordRest.ConfirmPasswordOtp, m.SmokeAuthz())
	})
}
