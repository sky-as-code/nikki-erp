package methods

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	itLogin "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	itPass "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type LoginMethodOtpCode struct {
}

func (this *LoginMethodOtpCode) Name() string {
	return "otpCode"
}

func (this *LoginMethodOtpCode) SkipMethod() *itLogin.SkippedMethod {
	return nil
}

func (this *LoginMethodOtpCode) Execute(ctx corectx.Context, param itLogin.LoginParam) (*itLogin.ExecuteResult, error) {
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
	if result.ClientErrors.Count() > 0 {
		return &itLogin.ExecuteResult{
			ClientErrors: result.ClientErrors,
		}, nil
	}
	return &itLogin.ExecuteResult{
		IsVerified:   result.Data.IsVerified,
		FailedReason: result.Data.FailedReason,
	}, nil
}
