package methods

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	itLogin "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	itPass "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
)

type LoginMethodOtpCode struct {
}

func (this *LoginMethodOtpCode) Name() string {
	return "otpCode"
}

func (this *LoginMethodOtpCode) SkipMethod() *itLogin.SkippedMethod {
	return nil
}

func (this *LoginMethodOtpCode) Execute(ctx context.Context, param itLogin.LoginParam) (*itLogin.ExecuteResult, error) {
	var result *itPass.VerifyPasswordResult
	var err error
	err = deps.Invoke(func(passwordSvc itPass.PasswordService) error {
		result, err = passwordSvc.VerifyOtpCode(ctx, itPass.VerifyOtpCodeQuery{
			SubjectType: param.SubjectType,
			Username:    param.Username,
			OtpCode:     domain.OtpCode(param.Password),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	if result.ClientError != nil {
		return &itLogin.ExecuteResult{
			ClientErr: result.ClientError,
		}, nil
	}
	return &itLogin.ExecuteResult{
		IsVerified:   result.Data.IsVerified,
		FailedReason: result.Data.FailedReason,
	}, nil
}
