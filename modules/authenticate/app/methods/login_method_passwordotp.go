package methods

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itLogin "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	itPass "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
)

type LoginMethodPasswordOtp struct {
}

func (this *LoginMethodPasswordOtp) Name() string {
	return "passwordotp"
}

func (this *LoginMethodPasswordOtp) SkipMethod() *itLogin.SkippedMethod {
	return nil
}

func (this *LoginMethodPasswordOtp) Execute(ctx context.Context, param itLogin.LoginParam) (bool, *string, error) {
	var result *itPass.VerifyPasswordResult
	var err error
	err = deps.Invoke(func(passwordotpService itPass.PasswordService) error {
		result, err = passwordotpService.VerifyPasswordOtp(ctx, itPass.VerifyPasswordOtpQuery{
			SubjectType: param.SubjectType,
			Username:    param.Username,
			OtpCode:     param.Password,
		})
		return err
	})
	if err != nil {
		return false, nil, err
	}
	return result.Data.IsVerified, result.Data.FailedReason, nil
}
