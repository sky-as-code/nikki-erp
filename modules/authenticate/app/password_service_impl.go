package app

import (
	"context"
	"regexp"
	"time"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/crypto"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

var (
	tempPasswordLength       = 10
	tempPasswordDurationMins = 60

	errExpired    = "expired"
	errMismatched = "mismatched"
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

func (this *PasswordServiceImpl) CreateTempPassword(ctx context.Context, cmd it.CreateTempPasswordCommand) (_ *it.CreateTempPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create temp password"); e != nil {
			err = e
		}
	}()

	var subject *loginSubject
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			subject, err = this.subjectHelper.assertSubjectExists(ctx, cmd.SubjectType, cmd.Username, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateTempPasswordResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	passStore, err := this.findPasswordStore(ctx, cmd.SubjectType, subject.Id)
	ft.PanicOnErr(err)

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

	if passStore != nil {
		passStore, err = this.passwordRepo.Update(ctx, domain.PasswordStore{
			ModelBase: model.ModelBase{
				Id: passStore.Id,
			},
			Passwordtmp:          util.ToPtr(string(tmpPassHash)),
			PasswordtmpExpiredAt: &tmpPassExpiredAt,
		})
	} else {
		newPassStore := domain.PasswordStore{
			SubjectType:          &cmd.SubjectType,
			SubjectRef:           &subject.Id,
			Passwordtmp:          util.ToPtr(string(tmpPassHash)),
			PasswordtmpExpiredAt: &tmpPassExpiredAt,
		}
		newPassStore.SetDefaults()
		passStore, err = this.passwordRepo.Create(ctx, newPassStore)
	}

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

	var curPassId *model.Id
	var curPassHash []byte
	var hasCurPass bool
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			curPassId, curPassHash, err = this.getCurrentPassword(ctx, cmd.SubjectType, cmd.SubjectRef)
			ft.PanicOnErr(err)

			hasCurPass = (curPassId != nil)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if !hasCurPass {
				return nil
			}
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

	var passStore *domain.PasswordStore
	if hasCurPass {
		passStore, err = this.passwordRepo.Update(ctx, domain.PasswordStore{
			ModelBase: model.ModelBase{
				Id: curPassId,
			},
			Password: util.ToPtr(string(newPassHash)),
			// TODO: PasswordExpiredAt
		})
	} else {
		newPassStore := domain.PasswordStore{
			SubjectType: &cmd.SubjectType,
			SubjectRef:  &cmd.SubjectRef,
			Password:    util.ToPtr(string(newPassHash)),
			// TODO: PasswordExpiredAt
		}
		newPassStore.SetDefaults()
		passStore, err = this.passwordRepo.Create(ctx, newPassStore)
	}

	ft.PanicOnErr(err)

	return &it.SetPasswordResult{
		Data: &it.SetPasswordResultData{
			UpdatedAt: *passStore.PasswordUpdatedAt,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) VerifyPassword(ctx context.Context, cmd it.VerifyPasswordQuery) (_ *it.VerifyPasswordResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "check password matched"); e != nil {
			err = e
		}
	}()

	var subject *loginSubject
	var passStore *domain.PasswordStore
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			subject, err = this.subjectHelper.assertSubjectExists(ctx, cmd.SubjectType, cmd.Username, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			passStore, err = this.findPasswordStore(ctx, cmd.SubjectType, subject.Id)
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

	var isMatched bool
	var reason *string = &errMismatched
	if passStore != nil {
		if passStore.Password != nil {
			isMatched, reason = this.validateCurrentPass([]byte(*passStore.Password), passStore.PasswordExpiredAt, cmd.Password)
		}
		if !isMatched && passStore.Passwordtmp != nil {
			isMatched, reason = this.validateCurrentPass([]byte(*passStore.Passwordtmp), passStore.PasswordtmpExpiredAt, cmd.Password)
		}
	}

	return &it.VerifyPasswordResult{
		Data: &it.VerifyPasswordResultData{
			IsVerified:   isMatched,
			FailedReason: reason,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) getCurrentPassword(ctx context.Context, subjectType domain.SubjectType, subjectRef model.Id) (*model.Id, []byte, error) {
	pass, err := this.findPasswordStore(ctx, subjectType, subjectRef)
	if err != nil {
		return nil, nil, err
	}

	if pass == nil {
		return nil, nil, nil
	}

	return pass.Id, []byte(*pass.Password), nil
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

func (this *PasswordServiceImpl) validateCurrentPass(curPassHash []byte, curPassExpireAt *time.Time, candidatePass string) (bool, *string) {
	var reason string
	isExpired := curPassExpireAt != nil && time.Now().After(*curPassExpireAt)
	if isExpired {
		reason = errExpired
		return false, &reason
	}

	isMatched := this.isPasswordEqual(curPassHash, candidatePass)
	if !isMatched {
		reason = errMismatched
		return false, &reason
	}

	return true, nil
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
