package password

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type PasswordService interface {
	CreatePasswordOtp(ctx corectx.Context, cmd CreatePasswordOtpCommand) (*CreatePasswordOtpResult, error)
	ConfirmPasswordOtp(ctx corectx.Context, cmd ConfirmPasswordOtpCommand) (*ConfirmPasswordOtpResult, error)
	CreatePasswordTemp(ctx corectx.Context, cmd CreatePasswordTempCommand) (*CreatePasswordTempResult, error)
	SetPassword(ctx corectx.Context, cmd SetPasswordCommand) (*SetPasswordResult, error)
	VerifyPassword(ctx corectx.Context, cmd VerifyPasswordQuery) (*VerifyPasswordResult, error)
	VerifyOtpCode(ctx corectx.Context, cmd VerifyPasswordOtpQuery) (*VerifyOtpCodeResult, error)
}
