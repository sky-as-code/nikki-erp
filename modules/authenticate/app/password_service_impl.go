package app

import (
	"time"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/convert"
	"github.com/sky-as-code/nikki-erp/common/crypto"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	ext "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

var (
	tempPasswordLength       = 10
	tempPasswordDurationMins = 60
	otpRecoveryCodeCount     = 10
)

type PasswordServiceParams struct {
	dig.In

	CqrsBus      cqrs.CqrsBus
	ConfigSvc    config.ConfigService
	Logger       logging.LoggerService
	UserSvc      ext.UserExtService
	PasswordRepo it.PasswordStoreRepository
}

func NewPasswordServiceImpl(params PasswordServiceParams) it.PasswordService {
	return &PasswordServiceImpl{
		configSvc:    params.ConfigSvc,
		logger:       params.Logger,
		passwordRepo: params.PasswordRepo,
		principalHelper: principalHelper{
			cqrsBus: params.CqrsBus,
			userSvc: params.UserSvc,
		},
	}
}

type PasswordServiceImpl struct {
	configSvc       config.ConfigService
	logger          logging.LoggerService
	passwordRepo    it.PasswordStoreRepository
	principalHelper principalHelper
}

func (this *PasswordServiceImpl) CreatePasswordOtp(ctx corectx.Context, cmd it.CreatePasswordOtpCommand) (_ *it.CreatePasswordOtpResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create password OTP"); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &it.CreatePasswordOtpResult{ClientErrors: cErrs}, nil
	}
	cmd = *sanitized.(*it.CreatePasswordOtpCommand)

	var passStore *domain.PasswordStore
	var principal *loginPrincipal
	passStore, principal, err = this.tryFetchPassStore(
		ctx, cmd.PrincipalType, &cmd.PrincipalId, nil, &cErrs,
	)
	if err != nil {
		return nil, err
	}

	if cErrs.Count() > 0 {
		return &it.CreatePasswordOtpResult{ClientErrors: cErrs}, nil
	}

	createdOtp := this.createOtp(principal.Username)

	this.logger.Debug("create otp password", logging.Attr{
		"principalType": cmd.PrincipalType,
		"PrincipalId":   cmd.PrincipalId,
		"passwordotp":   createdOtp.otpUrl,
	})

	np := domain.NewPasswordStore()
	np.SetPrincipalType(&cmd.PrincipalType)
	np.SetPrincipalId(&principal.Id)
	np.SetPasswordOtp(&createdOtp.otpSecret)
	np.SetPasswordOtpExpiresAt(&createdOtp.expiresAt)

	err = this.upsertPassStore(ctx, passStore, *np)
	if err != nil {
		return nil, err
	}

	return &it.CreatePasswordOtpResult{
		Data: it.CreatePasswordOtpResultData{
			CreatedAt: model.NewModelDateTime(),
			OtpUrl:    createdOtp.otpUrl,
			ExpiredAt: createdOtp.expiresAt,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) ConfirmPasswordOtp(ctx corectx.Context, cmd it.ConfirmPasswordOtpCommand) (_ *it.ConfirmPasswordOtpResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "confirm password OTP"); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &it.ConfirmPasswordOtpResult{ClientErrors: cErrs}, nil
	}
	cmd = *sanitized.(*it.ConfirmPasswordOtpCommand)

	var passStore *domain.PasswordStore
	var recoveryCodes []string
	cErrs, err = dyn.StartValidationFlowCopy(&cErrs).
		Step(func(cErrs *ft.ClientErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.PrincipalType, &cmd.PrincipalId, nil, cErrs)
			if err != nil {
				return err
			}
			if passStore.GetPasswordOtpExpiresAt() == nil {
				cErrs.Append(*ft.NewBusinessViolation(
					"otp_code",
					ft.ErrorKey("err_otp_register_completed", "authenticate"),
					"OTP register process already completed.",
				))
			}
			return nil
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			reason, err := this.verifyOtpCode(cmd.OtpCode, passStore)
			if err != nil {
				return err
			}
			if reason != nil {
				cErrs.Append(*reason)
			}
			return nil
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			recoveryCodes = this.createOtpRecovery(cmd)
			updatedStore := domain.NewPasswordStore()
			updatedStore.SetId(passStore.GetId())
			updatedStore.SetPrincipalType(&cmd.PrincipalType)
			updatedStore.SetPrincipalId(&cmd.PrincipalId)
			updatedStore.SetPasswordOtpExpiresAt(nil)
			updatedStore.SetPasswordOtpRecovery(recoveryCodes)

			updateRes, err := this.passwordRepo.Update(ctx, *updatedStore)
			cErrs.Concat(updateRes.ClientErrors)
			return err
		}).
		End()

	if err != nil {
		return nil, err
	}

	if cErrs.Count() > 0 {
		return &it.ConfirmPasswordOtpResult{ClientErrors: cErrs}, nil
	}

	return &it.ConfirmPasswordOtpResult{
		Data: it.ConfirmPasswordOtpResultData{
			ConfirmedAt:   model.NewModelDateTime(),
			RecoveryCodes: recoveryCodes,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) CreatePasswordTemp(ctx corectx.Context, cmd it.CreatePasswordTempCommand) (_ *it.CreatePasswordTempResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create password temp"); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &it.CreatePasswordTempResult{ClientErrors: cErrs}, nil
	}
	cmd = *sanitized.(*it.CreatePasswordTempCommand)

	var dbPassStore *domain.PasswordStore
	var principal *loginPrincipal
	dbPassStore, principal, err = this.tryFetchPassStore(ctx, cmd.PrincipalType, nil, &cmd.Username, &cErrs)

	if err != nil {
		return nil, err
	}

	if cErrs.Count() > 0 {
		return &it.CreatePasswordTempResult{ClientErrors: cErrs}, nil
	}

	tmpPass, err := crypto.GenerateSecurePassword(tempPasswordLength)
	if err != nil {
		return nil, err
	}

	this.logger.Debug("create temp password", logging.Attr{
		"principalType": cmd.PrincipalType,
		"username":      cmd.Username,
		"passwordtmp":   tmpPass,
	})

	tmpPassHash, err := crypto.GenerateFromPassword([]byte(tmpPass))
	if err != nil {
		return nil, err
	}

	tmpPassExpiresAt := model.NewModelDateTime().Calc(func(t time.Time) time.Time {
		return t.Add(time.Duration(tempPasswordDurationMins) * time.Minute)
	})
	updatedStore := domain.NewPasswordStore()
	updatedStore.SetPrincipalType(&cmd.PrincipalType)
	updatedStore.SetPrincipalId(&principal.Id)
	updatedStore.SetPasswordTmp(util.ToPtr(string(tmpPassHash)))
	updatedStore.SetPasswordTmpExpiresAt(&tmpPassExpiresAt)

	err = this.upsertPassStore(ctx, dbPassStore, *updatedStore)
	ft.PanicOnErr(err)

	return &it.CreatePasswordTempResult{
		Data: it.CreatePasswordTempResultData{
			CreatedAt: model.NewModelDateTime(),
			ExpiresAt: tmpPassExpiresAt,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) SetPassword(ctx corectx.Context, cmd it.SetPasswordCommand) (_ *it.SetPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "set password"); e != nil {
			err = e
		}
	}()

	sanitized, cErrsTotal := cmd.GetSchema().ValidateStruct(cmd)
	if cErrsTotal.Count() > 0 {
		return &it.SetPasswordResult{ClientErrors: cErrsTotal}, nil
	}
	cmd = *sanitized.(*it.SetPasswordCommand)

	var passStore *domain.PasswordStore
	var curPassHash []byte
	cErrsTotal, err = dyn.StartValidationFlowCopy(&cErrsTotal).
		Step(func(cErrs *ft.ClientErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.PrincipalType, &cmd.PrincipalId, nil, cErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			if passStore == nil {
				return nil
			}
			curPassHash = convert.StringPtrToBytes(passStore.GetPassword())
			if cmd.CurrentPassword == nil {
				cErrs.Append(*ft.NewValidationError(
					"current_password",
					ft.ErrorKey("err_current_password_required", "authenticate"),
					"Current password is required.",
				))
				return nil
			}
			curPassMatched := this.isPasswordEqual(curPassHash, *cmd.CurrentPassword)
			if !curPassMatched {
				cErrs.Append(*ft.NewValidationError(
					"current_password",
					ft.ErrorKey("err_current_password_incorrect", "authenticate"),
					"Incorrect current password.",
				))
			}
			return nil
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			this.validateNewPass(curPassHash, cmd.NewPassword, cErrs)
			return nil
		}).
		End()

	if err != nil {
		return nil, err
	}

	if cErrsTotal.Count() > 0 {
		return &it.SetPasswordResult{ClientErrors: cErrsTotal}, nil
	}

	newPassHash, err := crypto.GenerateFromPassword([]byte(cmd.NewPassword))
	if err != nil {
		return nil, err
	}

	dateTimeType := model.NewModelDateTime()
	updatedStore := domain.NewPasswordStore()
	updatedStore.SetPrincipalType(&cmd.PrincipalType)
	updatedStore.SetPrincipalId(&cmd.PrincipalId)
	updatedStore.SetPassword(util.ToPtr(string(newPassHash)))
	updatedStore.SetPasswordUpdatedAt(&dateTimeType)

	err = this.upsertPassStore(ctx, passStore, *updatedStore)

	if err != nil {
		return nil, err
	}

	return &it.SetPasswordResult{
		Data: dyn.MutateResultData{
			AffectedCount: 1,
			AffectedAt:    dateTimeType,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) VerifyPassword(ctx corectx.Context, cmd it.VerifyPasswordQuery) (_ *it.VerifyPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "verify password"); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &it.VerifyPasswordResult{ClientErrors: cErrs}, nil
	}
	cmd = *sanitized.(*it.VerifyPasswordQuery)

	var dbPassStore *domain.PasswordStore
	dbPassStore, _, err = this.tryFetchPassStore(ctx, cmd.PrincipalType, nil, &cmd.Username, &cErrs)
	if err != nil {
		return nil, err
	}

	if cErrs.Count() > 0 {
		return &it.VerifyPasswordResult{ClientErrors: cErrs}, nil
	}

	isMatched, reason := this.validateCurrentAndTempPass(cmd.Password, dbPassStore)

	return &it.VerifyPasswordResult{
		Data: it.VerifyPasswordResultData{
			IsVerified:   isMatched,
			FailedReason: reason,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) VerifyOtpCode(ctx corectx.Context, cmd it.VerifyPasswordOtpQuery) (_ *it.VerifyOtpCodeResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "verify otp code"); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &it.VerifyOtpCodeResult{ClientErrors: cErrs}, nil
	}
	cmd = *sanitized.(*it.VerifyPasswordOtpQuery)

	var passStore *domain.PasswordStore
	passStore, _, err = this.tryFetchPassStore(ctx, cmd.PrincipalType, nil, &cmd.Username, &cErrs)

	if err != nil {
		return nil, err
	}

	if cErrs.Count() > 0 {
		return &it.VerifyOtpCodeResult{ClientErrors: cErrs}, nil
	}

	result, err := this.verifyOtpAndRecovery(cmd.OtpCode, passStore)
	if err != nil {
		return nil, err
	}

	if result.remainingRecoveryCodes != nil {
		updatedStore := domain.NewPasswordStore()
		updatedStore.SetId(passStore.GetId())
		updatedStore.SetPasswordOtpRecovery(result.remainingRecoveryCodes)
		updateRes, err := this.passwordRepo.Update(ctx, *updatedStore)
		if err != nil {
			return nil, err
		}
		if updateRes.ClientErrors.Count() > 0 {
			return &it.VerifyOtpCodeResult{ClientErrors: updateRes.ClientErrors}, nil
		}
		passStore, err = this.findPasswordStore(ctx, *passStore.GetPrincipalType(), *passStore.GetPrincipalId())
	}

	return &it.VerifyOtpCodeResult{
		Data: it.VerifyPasswordResultData{
			IsVerified:   result.isMatched,
			FailedReason: result.reason,
		},
		HasData: true,
	}, nil
}
