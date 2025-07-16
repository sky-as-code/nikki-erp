package methods

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itLogin "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	itPass "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
)

type LoginMethodPassword struct {
}

func (this *LoginMethodPassword) Name() string {
	return "password"
}

func (this *LoginMethodPassword) SkipMethod() *itLogin.SkippedMethod {
	return nil
}

func (this *LoginMethodPassword) Execute(ctx context.Context, param itLogin.LoginParam) (bool, *string, error) {
	var result *itPass.VerifyPasswordResult
	var err error
	err = deps.Invoke(func(passwordService itPass.PasswordService) error {
		result, err = passwordService.VerifyPassword(ctx, itPass.VerifyPasswordQuery{
			SubjectType: param.SubjectType,
			Username:    param.Username,
			Password:    param.Password,
		})
		return err
	})
	if err != nil {
		return false, nil, err
	}
	return result.Data.IsVerified, result.Data.FailedReason, nil
}
