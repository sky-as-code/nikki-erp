package app

import (
	"regexp"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/crypto"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

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

	return createOtpResult{
		otpSecret: string(otpGen.Secret()),
		otpUrl:    otpGen.URL(),
		expiredAt: time.Now().Add(time.Duration(tempPasswordDurationMins) * time.Minute),
	}
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

func (this *PasswordServiceImpl) validateNewPass(curPassHash []byte, newPass string, clientErrs *ft.ClientErrors) {
	if curPassHash != nil {
		sameOldPass := this.isPasswordEqual(curPassHash, newPass)
		if sameOldPass {
			appendValidationError(clientErrs, "new_password", "new password must not be the same as the old password")
		}
	}
	if !checkPasswordPolicy(newPass) {
		appendValidationError(
			clientErrs,
			"new_password",
			"password must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, and one number",
		)
	}
}

func (this *PasswordServiceImpl) validateCurrentAndTempPass(
	candidatePass string, passStore *domain.PasswordStore,
) (bool, string) {
	var isMatched bool
	reason := "password mismatched"
	if passStore != nil {
		if passStore.GetPassword() != nil {
			isMatched, reason = this.validateCurrentPass([]byte(*passStore.GetPassword()), passStore.GetPasswordExpiredAt(), candidatePass)
		}
		if !isMatched && passStore.GetPasswordtmp() != nil {
			isMatched, reason = this.validateCurrentPass([]byte(*passStore.GetPasswordtmp()), passStore.GetPasswordtmpExpiredAt(), candidatePass)
		}
	}
	return isMatched, reason
}

func (this *PasswordServiceImpl) validateCurrentPass(
	curPassHash []byte, curPassExpireAt *time.Time, candidatePass string,
) (bool, string) {
	if curPassExpireAt != nil && time.Now().After(*curPassExpireAt) {
		return false, "password expired"
	}

	if !this.isPasswordEqual(curPassHash, candidatePass) {
		return false, "password mismatched"
	}

	return true, ""
}

type verifyOtpAndRecoveryResult struct {
	isMatched              bool
	reason                 string
	remainingRecoveryCodes []string
}

func (this *PasswordServiceImpl) verifyOtpAndRecovery(
	otpCode domain.OtpCode, passStore *domain.PasswordStore,
) verifyOtpAndRecoveryResult {
	result := verifyOtpAndRecoveryResult{}
	if passStore != nil {
		if crypto.IsRecoveryCodeFormat(otpCode.String()) {
			result.isMatched, result.remainingRecoveryCodes, result.reason = this.verifyOtpRecovery(otpCode, passStore)
		}
		result.isMatched, result.reason = this.verifyOtpCode(otpCode, passStore)
	}

	return result
}

func (this *PasswordServiceImpl) verifyOtpRecovery(
	otpCode domain.OtpCode, passStore *domain.PasswordStore,
) (bool, []string, string) {
	if passStore.GetPasswordotpRecovery() == nil {
		return false, nil, "recovery code mismatched"
	}
	remainingRecoveries, isMatched := array.RemoveString(passStore.GetPasswordotpRecovery(), otpCode.String())

	if !isMatched {
		return false, nil, "recovery code mismatched"
	}
	return true, remainingRecoveries, ""
}

func (this *PasswordServiceImpl) verifyOtpCode(
	otpCode domain.OtpCode, passStore *domain.PasswordStore,
) (bool, string) {
	if passStore.GetPasswordotp() == nil {
		return false, "otp code mismatched"
	}
	if passStore.GetPasswordotpExpiredAt() != nil && time.Now().After(*passStore.GetPasswordotpExpiredAt()) {
		return false, "otp not successfully registered"
	}
	isMatched, err := totp.ValidateCustom(otpCode.String(), *passStore.GetPasswordotp(), time.Now(), totp.ValidateOpts{
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

func (this *PasswordServiceImpl) tryFetchPassStore(
	ctx corectx.Context,
	subjectType domain.SubjectType,
	subjectRef *model.Id,
	username *string,
	clientErrs *ft.ClientErrors,
) (*domain.PasswordStore, *loginSubject, error) {
	subject, err := this.subjectHelper.assertSubjectExists(ctx, subjectType, subjectRef, username, clientErrs)
	ft.PanicOnErr(err)

	if subject == nil {
		return nil, nil, nil
	}

	passStore, err := this.findPasswordStore(ctx, subjectType, subject.Id)
	ft.PanicOnErr(err)

	return passStore, subject, nil
}

func (this *PasswordServiceImpl) findPasswordStore(
	ctx corectx.Context, subjectType domain.SubjectType, subjectRef model.Id,
) (*domain.PasswordStore, error) {
	passResult, err := this.passwordRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			domain.PasswordStoreFieldSubjectType: subjectType.String(),
			domain.PasswordStoreFieldSubjectRef:  string(subjectRef),
		},
	})
	if err != nil {
		return nil, err
	}
	if passResult.ClientErrors.Count() > 0 || !passResult.HasData {
		return nil, nil
	}
	return &passResult.Data, nil
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
	isOnlyAllowedChars := regexp.MustCompile(`^[A-Za-z\d@$!%*?&]+$`).MatchString(password)

	return hasLowercase && hasUppercase && hasDigit && isOnlyAllowedChars
}

func (this *PasswordServiceImpl) upsertPassStore(
	ctx corectx.Context,
	curPassStore *domain.PasswordStore,
	newPassStore domain.PasswordStore,
) (*domain.PasswordStore, error) {
	if curPassStore != nil {
		newPassStore.SetId(curPassStore.GetId())
		updateRes, err := this.passwordRepo.Update(ctx, newPassStore)
		if err != nil {
			return nil, err
		}
		if updateRes.ClientErrors.Count() > 0 {
			return nil, clientErrorsToError(updateRes.ClientErrors, "update password store failed")
		}
		return this.findPasswordStore(ctx, *curPassStore.GetSubjectType(), *curPassStore.GetSubjectRef())
	}

	newPassStore.SetDefaults()
	insertRes, err := this.passwordRepo.Insert(ctx, newPassStore)
	if err != nil {
		return nil, err
	}
	if insertRes.ClientErrors.Count() > 0 {
		return nil, clientErrorsToError(insertRes.ClientErrors, "create password store failed")
	}
	return this.findPasswordStore(ctx, *newPassStore.GetSubjectType(), *newPassStore.GetSubjectRef())
}
