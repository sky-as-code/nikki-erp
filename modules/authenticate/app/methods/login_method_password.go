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

func (this *LoginMethodPassword) Execute(ctx context.Context, param itLogin.LoginParam) (bool, error) {
	var result *itPass.IsPasswordMatchedResult
	var err error
	err = deps.Invoke(func(passwordService itPass.PasswordService) error {
		result, err = passwordService.IsPasswordMatched(ctx, itPass.IsPasswordMatchedQuery{
			SubjectType: param.SubjectType,
			Username:    param.Username,
			Password:    param.Password,
		})
		return err
	})
	if err != nil {
		return false, err
	}
	if !result.Data.IsMatched {
		return false, nil
	}

	return true, nil
}
