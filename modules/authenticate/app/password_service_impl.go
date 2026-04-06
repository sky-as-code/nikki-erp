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
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
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

func (this *PasswordServiceImpl) CreateOtpPassword(ctx corectx.Context, cmd it.CreateOtpPasswordCommand) (_ *it.CreateOtpPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create otp password"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	var subject *loginSubject
	clientErrs, err := dyn.StartValidationFlow(cmd).
		Step(func(clientErrs *ft.ClientErrors) error {
			passStore, subject, err = this.tryFetchPassStore(ctx, cmd.SubjectType, &cmd.SubjectRef, nil, clientErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.CreateOtpPasswordResult{
			ClientErrors: clientErrs,
		}, nil
	}

	createdOtp := this.createOtp(subject.Username)

	this.logger.Debug("create otp password", logging.Attr{
		"subjectType": cmd.SubjectType,
		"subjectRef":  cmd.SubjectRef,
		"passwordotp": createdOtp.otpUrl,
	})

	np := domain.NewPasswordStore()
	np.SetSubjectType(&cmd.SubjectType)
	np.SetSubjectRef(&subject.Id)
	np.SetPasswordotp(&createdOtp.otpSecret)
	np.SetPasswordotpExpiredAt(&createdOtp.expiredAt)
	_, err = this.upsertPassStore(ctx, passStore, *np)

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

func (this *PasswordServiceImpl) ConfirmOtpPassword(ctx corectx.Context, cmd it.ConfirmOtpPasswordCommand) (_ *it.ConfirmOtpPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "confirm otp password"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	clientErrs, err := dyn.StartValidationFlow(cmd).
		Step(func(clientErrs *ft.ClientErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.SubjectType, &cmd.SubjectRef, nil, clientErrs)
			if passStore.GetPasswordotpExpiredAt() == nil {
				appendValidationError(clientErrs, "otp_code", "otp already registered")
			}
			ft.PanicOnErr(err)
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			isMatched, reason := this.verifyOtpCode(cmd.OtpCode, passStore)
			if !isMatched {
				appendValidationError(clientErrs, "otp_code", reason)
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.ConfirmOtpPasswordResult{
			ClientErrors: clientErrs,
		}, nil
	}

	recoveryCodes := this.createOtpRecovery(cmd)

	updatedStore := domain.NewPasswordStore()
	updatedStore.SetId(passStore.GetId())
	updatedStore.SetSubjectType(&cmd.SubjectType)
	updatedStore.SetSubjectRef(&cmd.SubjectRef)
	updatedStore.SetPasswordotpExpiredAt(&model.ZeroTime)
	updatedStore.SetPasswordotpRecovery(recoveryCodes)
	updateRes, err := this.passwordRepo.Update(ctx, *updatedStore)
	ft.PanicOnErr(err)
	if updateRes.ClientErrors.Count() > 0 {
		return &it.ConfirmOtpPasswordResult{ClientErrors: updateRes.ClientErrors}, nil
	}

	return &it.ConfirmOtpPasswordResult{
		Data: &it.ConfirmOtpPasswordResultData{
			ConfirmedAt:   time.Now(),
			RecoveryCodes: recoveryCodes,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) CreateTempPassword(ctx corectx.Context, cmd it.CreateTempPasswordCommand) (_ *it.CreateTempPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create temp password"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	var subject *loginSubject
	clientErrs, err := dyn.StartValidationFlow(cmd).
		Step(func(clientErrs *ft.ClientErrors) error {
			passStore, subject, err = this.tryFetchPassStore(ctx, cmd.SubjectType, nil, &cmd.Username, clientErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.CreateTempPasswordResult{
			ClientErrors: clientErrs,
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
	tp := domain.NewPasswordStore()
	tp.SetSubjectType(&cmd.SubjectType)
	tp.SetSubjectRef(&subject.Id)
	tp.SetPasswordtmp(util.ToPtr(string(tmpPassHash)))
	tp.SetPasswordtmpExpiredAt(&tmpPassExpiredAt)
	_, err = this.upsertPassStore(ctx, passStore, *tp)
	ft.PanicOnErr(err)

	return &it.CreateTempPasswordResult{
		Data: &it.CreateTempPasswordResultData{
			CreatedAt: time.Now(),
			ExpiredAt: tmpPassExpiredAt,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) SetPassword(ctx corectx.Context, cmd it.SetPasswordCommand) (_ *it.SetPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create attempt attempt"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	var curPassHash []byte
	clientErrs, err := dyn.StartValidationFlow(cmd).
		Step(func(clientErrs *ft.ClientErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.SubjectType, &cmd.SubjectRef, nil, clientErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			if passStore == nil {
				return nil
			}
			curPassHash = convert.StringPtrToBytes(passStore.GetPassword())
			if cmd.CurrentPassword == nil {
				appendValidationError(clientErrs, "current_password", "current password is required")
				return nil
			}
			curPassMatched := this.isPasswordEqual(curPassHash, *cmd.CurrentPassword)
			if !curPassMatched {
				appendValidationError(clientErrs, "current_password", "invalid password")
			}
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			this.validateNewPass(curPassHash, cmd.NewPassword, clientErrs)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.SetPasswordResult{
			ClientErrors: clientErrs,
		}, nil
	}

	newPassHash, err := crypto.GenerateFromPassword([]byte(cmd.NewPassword))
	ft.PanicOnErr(err)

	now := time.Now()
	updatedStore := domain.NewPasswordStore()
	updatedStore.SetSubjectType(&cmd.SubjectType)
	updatedStore.SetSubjectRef(&cmd.SubjectRef)
	updatedStore.SetPassword(util.ToPtr(string(newPassHash)))
	updatedStore.SetPasswordUpdatedAt(&now)
	passStore, err = this.upsertPassStore(ctx, passStore, *updatedStore)

	ft.PanicOnErr(err)

	return &it.SetPasswordResult{
		Data: &it.SetPasswordResultData{
			UpdatedAt: *passStore.GetPasswordUpdatedAt(),
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

	var passStore *domain.PasswordStore
	clientErrs, err := dyn.StartValidationFlow(cmd).
		Step(func(clientErrs *ft.ClientErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.SubjectType, nil, &cmd.Username, clientErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.VerifyPasswordResult{
			ClientErrors: clientErrs,
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

func (this *PasswordServiceImpl) VerifyOtpCode(ctx corectx.Context, cmd it.VerifyOtpCodeQuery) (_ *it.VerifyOtpCodeResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "verify otp code"); e != nil {
			err = e
		}
	}()

	var passStore *domain.PasswordStore
	clientErrs, err := dyn.StartValidationFlow(cmd).
		Step(func(clientErrs *ft.ClientErrors) error {
			passStore, _, err = this.tryFetchPassStore(ctx, cmd.SubjectType, nil, &cmd.Username, clientErrs)
			ft.PanicOnErr(err)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.VerifyOtpCodeResult{
			ClientErrors: clientErrs,
		}, nil
	}

	result := this.verifyOtpAndRecovery(cmd.OtpCode, passStore)

	if result.remainingRecoveryCodes != nil {
		updatedStore := domain.NewPasswordStore()
		updatedStore.SetId(passStore.GetId())
		updatedStore.SetPasswordotpRecovery(result.remainingRecoveryCodes)
		updateRes, err := this.passwordRepo.Update(ctx, *updatedStore)
		if err != nil {
			return nil, err
		}
		if updateRes.ClientErrors.Count() > 0 {
			return &it.VerifyOtpCodeResult{ClientErrors: updateRes.ClientErrors}, nil
		}
		passStore, err = this.findPasswordStore(ctx, *passStore.GetSubjectType(), *passStore.GetSubjectRef())
	}

	return &it.VerifyOtpCodeResult{
		Data: &it.VerifyPasswordResultData{
			IsVerified:   result.isMatched,
			FailedReason: result.reason,
		},
		HasData: true,
	}, nil
}
