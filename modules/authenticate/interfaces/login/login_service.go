package login

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type LoginService interface {
	Authenticate(ctx corectx.Context, cmd AuthenticateCommand) (result *AuthenticateResult, err error)
	RefreshToken(ctx corectx.Context, cmd RefreshTokenCommand) (result *RefreshTokenResult, err error)
}
