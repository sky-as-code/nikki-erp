package app

import (
	"context"
	"regexp"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/crypto"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

type PasswordServiceParams struct {
	dig.In
	PasswordRepo it.PasswordStoreRepository
	CqrsBus      cqrs.CqrsBus
}

func NewPasswordServiceImpl(params PasswordServiceParams) it.PasswordService {
	return &PasswordServiceImpl{
		passwordRepo: params.PasswordRepo,
		subjectHelper: subjectHelper{
			cqrsBus: params.CqrsBus,
		},
	}
}

type PasswordServiceImpl struct {
	passwordRepo  it.PasswordStoreRepository
	subjectHelper subjectHelper
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
		HasData: passStore != nil,
	}, nil
}

func (this *PasswordServiceImpl) IsPasswordMatched(ctx context.Context, cmd it.IsPasswordMatchedQuery) (_ *it.IsPasswordMatchedResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "check password matched"); e != nil {
			err = e
		}
	}()

	var subject *loginSubject
	var curPassHash []byte
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			subject, err = this.subjectHelper.assertSubjectExists(ctx, cmd.SubjectType, cmd.Username, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			_, curPassHash, err = this.getCurrentPassword(ctx, cmd.SubjectType, subject.Id)
			ft.PanicOnErr(err)

			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.IsPasswordMatchedResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	curPassMatched := this.isPasswordEqual(curPassHash, cmd.Password)

	return &it.IsPasswordMatchedResult{
		Data: &it.IsPasswordMatchedResultData{
			IsMatched: curPassMatched,
		},
		HasData: true,
	}, nil
}

func (this *PasswordServiceImpl) getCurrentPassword(ctx context.Context, subjectType domain.SubjectType, subjectRef model.Id) (*model.Id, []byte, error) {
	pass, err := this.passwordRepo.FindBySubject(ctx, it.FindBySubjectParam{
		SubjectType: subjectType,
		SubjectRef:  subjectRef,
	})
	if err != nil {
		return nil, nil, err
	}

	noPass := (pass == nil)
	if noPass {
		return nil, nil, nil
	}

	return pass.Id, []byte(*pass.Password), nil
}

func (this *PasswordServiceImpl) isPasswordEqual(passHash []byte, candidatePass string) bool {
	isEqual, _ := crypto.CompareHashAndPassword(passHash, []byte(candidatePass))
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
