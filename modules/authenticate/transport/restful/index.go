package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/authenticate/transport/restful/v1"
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

		v1.POST("/login/start", loginRest.StartLoginFlow)
		v1.POST("/login", loginRest.Authenticate)
		v1.POST("/refresh", loginRest.RefreshToken)

		v1.POST("/passwords/password", passwordRest.SetPassword)
		v1.POST("/passwords/passwordtmp", passwordRest.CreatePasswordTemp)
		v1.POST("/passwords/passwordotp", passwordRest.CreatePasswordOtp)
		v1.POST("/passwords/passwordotp/confirm", passwordRest.ConfirmPasswordOtp)
	})
}
