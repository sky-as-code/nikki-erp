package app

import (
	"context"
	"regexp"
	"time"

	"github.com/pquerna/otp/totp"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/convert"
	"github.com/sky-as-code/nikki-erp/common/crypto"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

var (
	tempPasswordLength       = 10
	tempPasswordDurationMins = 60
	otpRecoveryCodeCount     = 10
)

type PasswordServiceParams struct {
	dig.In
	Logger       logging.LoggerService
	PasswordRepo it.PasswordStoreRepository
	CqrsBus      cqrs.CqrsBus
}

func NewPasswordServiceImpl(params PasswordServiceParams) it.PasswordService {
	return &PasswordServiceImpl{
		logger:       params.Logger,
		passwordRepo: params.PasswordRepo,
		subjectHelper: subjectHelper{
			cqrsBus: params.CqrsBus,
		},
	}
}

type PasswordServiceImpl struct {
	logger        logging.LoggerService
	passwordRepo  it.PasswordStoreRepository
	subjectHelper subjectHelper
}

func (this *PasswordServiceImpl) CreateOtpPassword(ctx context.Context, cmd it.CreateOtpPasswordCommand) (_ *it.CreateOtpPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create otp password"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	var subject *loginSubject
	vErrs, err := val.StartValidationFlow(cmd).
		Step(func(vErrs *ft.ValidationErrors) error {
			passStore, subject, err = this.tryFetchPassStore(ctx, cmd.SubjectType, &cmd.SubjectRef, nil, vErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateOtpPasswordResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdOtp := this.createOtp(subject.Username)

	this.logger.Debug("create otp password", logging.Attr{
		"subjectType": cmd.SubjectType,
		"subjectRef":  cmd.SubjectRef,
		"passwordotp": createdOtp.otpUrl,
	})

	// Don't pass the existing passStore instance to avoid
	// overriding existing fields
	_, err = this.upsertPassStore(ctx, passStore, domain.PasswordStore{
		SubjectType:          &cmd.SubjectType,
		SubjectRef:           &subject.Id,
		Passwordotp:          &createdOtp.otpSecret,
		PasswordotpExpiredAt: &createdOtp.expiredAt,
	})

	ft.PanicOnErr(err)

	return &it.CreateOtpPasswordResult{
		Data: &it.CreatePasswordOtpResultData{
			CreatedAt: time.Now(),
			OtpUrl:    createdOtp.otpUrl,
			ExpiredAt: createdOtp.expiredAt,
		},
		HasData: true,
	}, nil
}

type createOtpResult struct {
	otpSecret string
	otpUrl    string
	expiredAt time.Time
}

func (this *PasswordServiceImpl) createOtp(username string) createOtpResult {
	otpGen, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "NikkiERP",
		AccountName: username,
		Period:      c.OTP_PERIOD,
		Digits:      c.OTP_CODE_LENGTH,
	})
	ft.PanicOnErr(err)

	result := createOtpResult{
		otpSecret: string(otpGen.Secret()),
		otpUrl:    otpGen.URL(),
		expiredAt: time.Now().Add(time.Duration(tempPasswordDurationMins) * time.Minute),
	}

	return result
}

func (this *PasswordServiceImpl) ConfirmOtpPassword(ctx context.Context, cmd it.ConfirmOtpPasswordCommand) (_ *it.ConfirmOtpPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "confirm otp password"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	vErrs, err := val.StartValidationFlow(cmd).
		Step(func(vErrs *ft.ValidationErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.SubjectType, &cmd.SubjectRef, nil, vErrs)
			if passStore.PasswordotpExpiredAt == nil {
				vErrs.Append("otpCode", "otp already registered")
			}
			ft.PanicOnErr(err)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			isMatched, reason := this.verifyOtpCode(cmd.OtpCode, passStore)
			if !isMatched {
				vErrs.Append("otpCode", reason)
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.ConfirmOtpPasswordResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	recoveryCodes := this.createOtpRecovery(cmd)

	_, err = this.passwordRepo.Update(ctx, domain.PasswordStore{
		ModelBase: model.ModelBase{
			Id: passStore.Id,
		},
		SubjectType:          &cmd.SubjectType,
		SubjectRef:           &cmd.SubjectRef,
		PasswordotpExpiredAt: &model.ZeroTime, // Clear expiration time
		PasswordotpRecovery:  recoveryCodes,
	})
	ft.PanicOnErr(err)

	return &it.ConfirmOtpPasswordResult{
		Data: &it.ConfirmOtpPasswordResultData{
			ConfirmedAt:   time.Now(),
			RecoveryCodes: recoveryCodes,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) createOtpRecovery(cmd it.ConfirmOtpPasswordCommand) []string {
	recoveryCodes := make([]string, otpRecoveryCodeCount)
	for i := range otpRecoveryCodeCount {
		recCode, err := crypto.GenerateRecoveryCode()
		ft.PanicOnErr(err)
		recoveryCodes[i] = recCode
	}

	this.logger.Debug("confirm otp password", logging.Attr{
		"subjectType":   cmd.SubjectType,
		"subjectRef":    cmd.SubjectRef,
		"recoveryCodes": recoveryCodes,
	})
	return recoveryCodes
}

func (this *PasswordServiceImpl) CreateTempPassword(ctx context.Context, cmd it.CreateTempPasswordCommand) (_ *it.CreateTempPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create temp password"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	var subject *loginSubject
	vErrs, err := val.StartValidationFlow(cmd).
		Step(func(vErrs *ft.ValidationErrors) error {
			passStore, subject, err = this.tryFetchPassStore(ctx, cmd.SubjectType, nil, &cmd.Username, vErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateTempPasswordResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	tmpPass, err := crypto.GenerateSecurePassword(tempPasswordLength)
	ft.PanicOnErr(err)

	this.logger.Debug("create temp password", logging.Attr{
		"subjectType": cmd.SubjectType,
		"username":    cmd.Username,
		"passwordtmp": tmpPass,
	})

	tmpPassHash, err := crypto.GenerateFromPassword([]byte(tmpPass))
	ft.PanicOnErr(err)

	tmpPassExpiredAt := time.Now().Add(time.Duration(tempPasswordDurationMins) * time.Minute)
	_, err = this.upsertPassStore(ctx, passStore, domain.PasswordStore{
		SubjectType:          &cmd.SubjectType,
		SubjectRef:           &subject.Id,
		Passwordtmp:          util.ToPtr(string(tmpPassHash)),
		PasswordtmpExpiredAt: &tmpPassExpiredAt,
	})
	ft.PanicOnErr(err)

	return &it.CreateTempPasswordResult{
		Data: &it.CreateTempPasswordResultData{
			CreatedAt: time.Now(),
			ExpiredAt: tmpPassExpiredAt,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) SetPassword(ctx context.Context, cmd it.SetPasswordCommand) (_ *it.SetPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create attempt attempt"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	var curPassHash []byte
	vErrs, err := val.StartValidationFlow(cmd).
		Step(func(vErrs *ft.ValidationErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.SubjectType, &cmd.SubjectRef, nil, vErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if passStore == nil {
				return nil
			}
			curPassHash = convert.StringPtrToBytes(passStore.Password)
			if cmd.CurrentPassword == nil {
				vErrs.Append("currentPassword", "current password is required")
				return nil
			}
			curPassMatched := this.isPasswordEqual(curPassHash, *cmd.CurrentPassword)
			if !curPassMatched {
				vErrs.Append("currentPassword", "invalid password")
			}
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.validateNewPass(curPassHash, cmd.NewPassword, vErrs)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.SetPasswordResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	newPassHash, err := crypto.GenerateFromPassword([]byte(cmd.NewPassword))
	ft.PanicOnErr(err)

	now := time.Now()
	passStore, err = this.upsertPassStore(ctx, passStore, domain.PasswordStore{
		SubjectType:       &cmd.SubjectType,
		SubjectRef:        &cmd.SubjectRef,
		Password:          util.ToPtr(string(newPassHash)),
		PasswordUpdatedAt: &now,
		// TODO: PasswordExpiredAt
	})

	ft.PanicOnErr(err)

	return &it.SetPasswordResult{
		Data: &it.SetPasswordResultData{
			UpdatedAt: *passStore.PasswordUpdatedAt,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) validateNewPass(curPassHash []byte, newPass string, vErrs *ft.ValidationErrors) {
	if curPassHash != nil {
		sameOldPass := this.isPasswordEqual(curPassHash, newPass)
		if sameOldPass {
			vErrs.Append("newPassword", "new password must not be the same as the old password")
		}
	}
	// TODO: fetch password policy from DB
	isValid := checkPasswordPolicy(newPass)
	if !isValid {
		vErrs.Append("newPassword", "password must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, and one number")
	}
}

func (this *PasswordServiceImpl) VerifyPassword(ctx context.Context, cmd it.VerifyPasswordQuery) (_ *it.VerifyPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "verify password"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	vErrs, err := val.StartValidationFlow(cmd).
		Step(func(vErrs *ft.ValidationErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.SubjectType, nil, &cmd.Username, vErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.VerifyPasswordResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	isMatched, reason := this.validateCurrentAndTempPass(cmd.Password, passStore)

	return &it.VerifyPasswordResult{
		Data: &it.VerifyPasswordResultData{
			IsVerified:   isMatched,
			FailedReason: reason,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) validateCurrentAndTempPass(candidatePass string, passStore *domain.PasswordStore) (bool, string) {
	var isMatched bool
	reason := "password mismatched"
	if passStore != nil {
		if passStore.Password != nil {
			isMatched, reason = this.validateCurrentPass([]byte(*passStore.Password), passStore.PasswordExpiredAt, candidatePass)
		}
		if !isMatched && passStore.Passwordtmp != nil {
			isMatched, reason = this.validateCurrentPass([]byte(*passStore.Passwordtmp), passStore.PasswordtmpExpiredAt, candidatePass)
		}
	}
	return isMatched, reason
}

func (this *PasswordServiceImpl) validateCurrentPass(curPassHash []byte, curPassExpireAt *time.Time, candidatePass string) (bool, string) {
	var reason string
	isExpired := curPassExpireAt != nil && time.Now().After(*curPassExpireAt)
	if isExpired {
		reason = "password expired"
		return false, reason
	}

	isMatched := this.isPasswordEqual(curPassHash, candidatePass)
	if !isMatched {
		reason = "password mismatched"
		return false, reason
	}

	return true, ""
}

func (this *PasswordServiceImpl) VerifyOtpCode(ctx context.Context, cmd it.VerifyOtpCodeQuery) (_ *it.VerifyOtpCodeResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "verify otp code"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	vErrs, err := val.StartValidationFlow(cmd).
		Step(func(vErrs *ft.ValidationErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.SubjectType, nil, &cmd.Username, vErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.VerifyOtpCodeResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	result := this.verifyOtpAndRecovery(cmd.OtpCode, passStore)

	if result.remainingRecoveryCodes != nil {
		passStore, err = this.passwordRepo.Update(ctx, domain.PasswordStore{
			ModelBase: model.ModelBase{
				Id: passStore.Id,
			},
			PasswordotpRecovery: result.remainingRecoveryCodes,
		})
	}

	return &it.VerifyOtpCodeResult{
		Data: &it.VerifyPasswordResultData{
			IsVerified:   result.isMatched,
			FailedReason: result.reason,
		},
		HasData: true,
	}, nil
}

type verifyOtpAndRecoveryResult struct {
	isMatched              bool
	reason                 string
	remainingRecoveryCodes []string
}

func (this *PasswordServiceImpl) verifyOtpAndRecovery(otpCode domain.OtpCode, passStore *domain.PasswordStore) verifyOtpAndRecoveryResult {
	result := verifyOtpAndRecoveryResult{}
	if passStore != nil {
		isRecoveryCode := crypto.IsRecoveryCodeFormat(otpCode.String())
		if isRecoveryCode {
			result.isMatched, result.remainingRecoveryCodes, result.reason = this.verifyOtpRecovery(otpCode, passStore)
		}
		result.isMatched, result.reason = this.verifyOtpCode(otpCode, passStore)
	}

	return result
}

func (this *PasswordServiceImpl) verifyOtpRecovery(otpCode domain.OtpCode, passStore *domain.PasswordStore) (bool, []string, string) {
	if passStore.PasswordotpRecovery == nil {
		return false, nil, "recovery code mismatched"
	}
	remainingRecoveries, isMatched := array.RemoveString(passStore.PasswordotpRecovery, otpCode.String())

	if !isMatched {
		return false, nil, "recovery code mismatched"
	}
	return true, remainingRecoveries, ""
}

func (this *PasswordServiceImpl) verifyOtpCode(otpCode domain.OtpCode, passStore *domain.PasswordStore) (bool, string) {
	if passStore.Passwordotp == nil {
		return false, "otp code mismatched"
	}
	if passStore.PasswordotpExpiredAt != nil && time.Now().After(*passStore.PasswordotpExpiredAt) {
		return false, "otp not successfully registered"
	}
	isMatched, err := totp.ValidateCustom(otpCode.String(), *passStore.Passwordotp, time.Now(), totp.ValidateOpts{
		Digits: c.OTP_CODE_LENGTH,
		Period: c.OTP_PERIOD,
		Skew:   c.OTP_SKEW,
	})
	ft.PanicOnErr(err)

	if !isMatched {
		return false, "otp code mismatched"
	}

	return true, ""
}

func (this *PasswordServiceImpl) tryFetchPassStore(ctx context.Context, subjectType domain.SubjectType, subjectRef *model.Id, username *string, vErrs *ft.ValidationErrors) (*domain.PasswordStore, *loginSubject, error) {
	subject, err := this.subjectHelper.assertSubjectExists(ctx, subjectType, subjectRef, username, vErrs)
	ft.PanicOnErr(err)

	if subject == nil {
		return nil, nil, nil
	}

	passStore, err := this.findPasswordStore(ctx, subjectType, subject.Id)
	ft.PanicOnErr(err)

	return passStore, subject, nil
}

func (this *PasswordServiceImpl) findPasswordStore(ctx context.Context, subjectType domain.SubjectType, subjectRef model.Id) (*domain.PasswordStore, error) {
	pass, err := this.passwordRepo.FindBySubject(ctx, it.FindBySubjectParam{
		SubjectType: subjectType,
		SubjectRef:  subjectRef,
	})
	if err != nil {
		return nil, err
	}

	return pass, nil
}

func (this *PasswordServiceImpl) isPasswordEqual(passHash []byte, candidatePass string) bool {
	isEqual, err := crypto.CompareHashAndPassword(passHash, []byte(candidatePass))
	if err != nil {
		this.logger.Warn("error with crypto.CompareHashAndPassword()", logging.Attr{
			"error": err.Error(),
		})
		return false
	}
	return isEqual
}

func checkPasswordPolicy(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	// hasSpecial := regexp.MustCompile(`[@$!%*?&]`).MatchString(password)
	isOnlyAllowedChars := regexp.MustCompile(`^[A-Za-z\d@$!%*?&]+$`).MatchString(password)

	return hasLowercase && hasUppercase && hasDigit && isOnlyAllowedChars
}

func (this *PasswordServiceImpl) upsertPassStore(ctx context.Context, curPassStore *domain.PasswordStore, newPassStore domain.PasswordStore) (*domain.PasswordStore, error) {
	var passStore *domain.PasswordStore
	var err error
	if curPassStore != nil {
		newPassStore.ModelBase = curPassStore.ModelBase
		passStore, err = this.passwordRepo.Update(ctx, newPassStore)
	} else {
		newPassStore.SetDefaults()
		passStore, err = this.passwordRepo.Create(ctx, newPassStore)
	}
	return passStore, err
}
