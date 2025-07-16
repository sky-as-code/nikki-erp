package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/authenticate/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewLoginRest,
		v1.NewPasswordRest,
	)
	return deps.Invoke(func(
		route *echo.Group,
		loginRest *v1.LoginRest,
		passwordRest *v1.PasswordRest,
	) {
		v1 := route.Group("/v1/authn")
		initV1(v1, loginRest, passwordRest)
	})
}

func initV1(route *echo.Group, loginRest *v1.LoginRest, passwordRest *v1.PasswordRest) {
	// route.POST("/attempts", loginRest.CreateLoginAttempt)
	route.POST("/login/start", loginRest.StartLoginFlow)
	route.POST("/login", loginRest.Authenticate)

	route.POST("/passwords/password", passwordRest.SetPassword)
	route.POST("/passwords/passwordtmp", passwordRest.CreateTempPassword)
	// route.POST("/passwords/otp", passwordRest.SetPassword)
	// route.POST("/passwords/tmp", passwordRest.SetPassword)
}
