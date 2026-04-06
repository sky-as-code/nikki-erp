package methods

import (
	itLogin "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
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

func (this *LoginMethodCaptcha) Execute(ctx corectx.Context, param itLogin.LoginParam) (*itLogin.ExecuteResult, error) {
	result := &itLogin.ExecuteResult{}
	switch param.Password {
	case "NIKKI":
		result.IsVerified = true
	case "EXPIRED":
		result.FailedReason = errExpired
	default:
		result.FailedReason = errMismatched
	}
	return result, nil
}
