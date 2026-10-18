package methods

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itLogin "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/login"
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
		result.FailedReason = ft.NewBusinessViolation("password", ft.ErrorKey("err_captcha_expired", "iam"),
			"Captcha expired")
	default:
		result.FailedReason = ft.NewBusinessViolation("password", ft.ErrorKey("err_captcha_mismatched", "iam"),
			"Captcha mismatched")
	}
	return result, nil
}
