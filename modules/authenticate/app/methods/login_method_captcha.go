package methods

import (
	"context"

	itLogin "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
)

var (
	errExpired    = "expired"
	errMismatched = "mismatched"
)

type LoginMethodCaptcha struct {
}

func (this *LoginMethodCaptcha) Name() string {
	return "captcha"
}

func (this *LoginMethodCaptcha) SkipMethod() *itLogin.SkippedMethod {
	return nil
}

func (this *LoginMethodCaptcha) Execute(ctx context.Context, param itLogin.LoginParam) (bool, *string, error) {
	switch param.Password {
	case "NIKKI":
		return true, nil, nil
	case "EXPIRED":
		return false, &errExpired, nil
	default:
		return false, &errMismatched, nil
	}
}
