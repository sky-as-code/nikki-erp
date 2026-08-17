package methods

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itLogin "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/login"
	itPass "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/password"
)

const LoginOtpCode = "otpCode"

type LoginMethodOtpCode struct {
}

func (this *LoginMethodOtpCode) Name() string {
	return LoginOtpCode
}

func (this *LoginMethodOtpCode) SkipMethod() *itLogin.SkippedMethod {
	return nil
}

func (this *LoginMethodOtpCode) Execute(ctx corectx.Context, param itLogin.LoginParam) (*itLogin.ExecuteResult, error) {
	var result *itPass.VerifyPasswordResult
	var err error
	err = deps.Invoke(func(passwordSvc itPass.PasswordAppService) error {
		result, err = passwordSvc.VerifyOtpCode(ctx, itPass.VerifyPasswordOtpQuery{
			PrincipalType: param.PrincipalType,
			Username:      param.Username,
			OtpCode:       models.OtpCode(param.Password),
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
