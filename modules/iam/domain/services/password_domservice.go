package services

import (
	"time"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/crypto"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/password"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

var (
	tempPasswordLength       = 10
	tempPasswordDurationMins = 60
	otpRecoveryCodeCount     = 10
)

type PasswordServiceParams struct {
	dig.In

	CqrsBus           cqrs.CqrsBus
	ConfigSvc         config.ConfigService
	Logger            logging.LoggerService
	UserSvc           itUser.UserDomainService
	UserRepo          itUser.UserRepository
	PasswordStoreRepo it.PasswordStoreRepository
}

func NewPasswordDomainServiceImpl(params PasswordServiceParams) it.PasswordDomainService {
	return &PasswordDomainServiceImpl{
		configSvc:         params.ConfigSvc,
		logger:            params.Logger,
		userRepo:          params.UserRepo,
		passwordStoreRepo: params.PasswordStoreRepo,
		principalHelper: principalHelper{
			cqrsBus: params.CqrsBus,
			userSvc: params.UserSvc,
		},
	}
}

type PasswordDomainServiceImpl struct {
	configSvc         config.ConfigService
	logger            logging.LoggerService
	userRepo          itUser.UserRepository
	passwordStoreRepo it.PasswordStoreRepository
	principalHelper   principalHelper
}

func (this *PasswordDomainServiceImpl) CreatePasswordOtp(ctx corectx.Context, cmd it.CreatePasswordOtpCommand) (_ *it.CreatePasswordOtpResult, err error) {
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

	var principal *loginPrincipal
	_, principal, _, err = this.tryFetchUserForPassword(
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

	err = this.upsertPasswordStoreHash(
		ctx,
		cmd.PrincipalType,
		principal.Id,
		models.PasswordStoreTypeOtpSecret,
		&createdOtp.otpSecret,
		&createdOtp.expiresAt,
		nil,
	)
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

func (this *PasswordDomainServiceImpl) ConfirmPasswordOtp(ctx corectx.Context, cmd it.ConfirmPasswordOtpCommand) (_ *it.ConfirmPasswordOtpResult, err error) {
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

	var stores *principalPasswordStores
	var recoveryCodes []string
	cErrs, err = dyn.StartValidationFlowCopy(&cErrs).
		Step(func(cErrs *ft.ClientErrors) error {
			_, _, stores, err = this.tryFetchUserForPassword(ctx, cmd.PrincipalType, &cmd.PrincipalId, nil, cErrs)
			if err != nil {
				return err
			}
			if stores == nil || stores.otpSecret == nil || stores.otpSecret.GetExpiresAt() == nil {
				cErrs.Append(*ft.NewBusinessViolation(
					"otp_code",
					ft.ErrorKey("err_otp_register_completed", "iam"),
					"OTP register process already completed.",
				))
			}
			return nil
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			reason, err := this.verifyOtpCode(cmd.OtpCode, stores)
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
			err = this.upsertPasswordStoreHash(
				ctx,
				cmd.PrincipalType,
				cmd.PrincipalId,
				models.PasswordStoreTypeOtpSecret,
				stores.otpSecret.GetHash(),
				nil,
				nil,
			)
			if err != nil {
				return err
			}
			return this.upsertPasswordStoreHash(
				ctx,
				cmd.PrincipalType,
				cmd.PrincipalId,
				models.PasswordStoreTypeOtpRecovery,
				this.encodeRecoveryCodes(recoveryCodes),
				nil,
				nil,
			)
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

func (this *PasswordDomainServiceImpl) CreatePasswordTemp(ctx corectx.Context, cmd it.CreatePasswordTempCommand) (_ *it.CreatePasswordTempResult, err error) {
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

	var principal *loginPrincipal
	_, principal, _, err = this.tryFetchUserForPassword(ctx, cmd.PrincipalType, nil, &cmd.Username, &cErrs)

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
	err = this.upsertPasswordStoreHash(
		ctx,
		cmd.PrincipalType,
		principal.Id,
		models.PasswordStoreTypePasswordTemp,
		util.ToPtr(string(tmpPassHash)),
		&tmpPassExpiresAt,
		nil,
	)
	ft.PanicOnErr(err)

	return &it.CreatePasswordTempResult{
		Data: it.CreatePasswordTempResultData{
			CreatedAt: model.NewModelDateTime(),
			ExpiresAt: tmpPassExpiresAt,
		},
		HasData: true,
	}, nil
}

func (this *PasswordDomainServiceImpl) SetPassword(ctx corectx.Context, cmd it.SetPasswordCommand) (_ *it.SetPasswordResult, err error) {
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

	var stores *principalPasswordStores
	var curPassHash []byte
	cErrsTotal, err = dyn.StartValidationFlowCopy(&cErrsTotal).
		Step(func(cErrs *ft.ClientErrors) error {
			_, _, stores, err = this.tryFetchUserForPassword(ctx, cmd.PrincipalType, &cmd.PrincipalId, nil, cErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			if stores == nil {
				return nil
			}
			if stores.getPasswordHash() != nil {
				curPassHash = []byte(*stores.getPasswordHash())
			}
			if cmd.CurrentPassword == nil {
				cErrs.Append(*ft.NewValidationError(
					"current_password",
					ft.ErrorKey("err_current_password_required", "iam"),
					"Current password is required.",
				))
				return nil
			}
			curPassMatched := this.isPasswordEqual(curPassHash, *cmd.CurrentPassword)
			if !curPassMatched {
				cErrs.Append(*ft.NewValidationError(
					"current_password",
					ft.ErrorKey("err_current_password_incorrect", "iam"),
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
	err = this.upsertPasswordStoreHash(
		ctx,
		cmd.PrincipalType,
		cmd.PrincipalId,
		models.PasswordStoreTypePassword,
		util.ToPtr(string(newPassHash)),
		nil,
		&dateTimeType,
	)

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

func (this *PasswordDomainServiceImpl) VerifyPassword(ctx corectx.Context, cmd it.VerifyPasswordQuery) (_ *it.VerifyPasswordResult, err error) {
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

	var stores *principalPasswordStores
	_, _, stores, err = this.tryFetchUserForPassword(ctx, cmd.PrincipalType, nil, &cmd.Username, &cErrs)
	if err != nil {
		return nil, err
	}

	if cErrs.Count() > 0 {
		return &it.VerifyPasswordResult{ClientErrors: cErrs}, nil
	}

	isMatched, reason := this.validateCurrentAndTempPass(cmd.Password, stores)

	return &it.VerifyPasswordResult{
		Data: it.VerifyPasswordResultData{
			IsVerified:   isMatched,
			FailedReason: reason,
		},
		HasData: true,
	}, nil
}

func (this *PasswordDomainServiceImpl) VerifyOtpCode(ctx corectx.Context, cmd it.VerifyPasswordOtpQuery) (_ *it.VerifyOtpCodeResult, err error) {
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

	var stores *principalPasswordStores
	var principal *loginPrincipal
	_, principal, stores, err = this.tryFetchUserForPassword(ctx, cmd.PrincipalType, nil, &cmd.Username, &cErrs)

	if err != nil {
		return nil, err
	}

	if cErrs.Count() > 0 {
		return &it.VerifyOtpCodeResult{ClientErrors: cErrs}, nil
	}

	result, err := this.verifyOtpAndRecovery(cmd.OtpCode, stores)
	if err != nil {
		return nil, err
	}

	if result.remainingRecoveryCodes != nil {
		err = this.upsertPasswordStoreHash(
			ctx,
			cmd.PrincipalType,
			principal.Id,
			models.PasswordStoreTypeOtpRecovery,
			this.encodeRecoveryCodes(result.remainingRecoveryCodes),
			nil,
			nil,
		)
		if err != nil {
			return nil, err
		}
		_, err = this.findUserById(ctx, principal.Id)
		if err != nil {
			return nil, err
		}
	}

	return &it.VerifyOtpCodeResult{
		Data: it.VerifyPasswordResultData{
			IsVerified:   result.isMatched,
			FailedReason: result.reason,
		},
		HasData: true,
	}, nil
}
