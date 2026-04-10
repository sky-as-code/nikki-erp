package app

import (
	"regexp"
	"time"

	"github.com/pquerna/otp/totp"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/crypto"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	coreConst "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type createOtpResult struct {
	otpSecret string
	otpUrl    string
	expiresAt model.ModelDateTime
}

func (this *PasswordServiceImpl) createOtp(username string) createOtpResult {
	otpGen, err := totp.Generate(totp.GenerateOpts{
		Issuer:      this.configSvc.GetStr(coreConst.AppName),
		AccountName: username,
		Period:      c.OtpPeriod,
		Digits:      c.OtpCodeLength,
	})
	ft.PanicOnErr(err)

	return createOtpResult{
		otpSecret: string(otpGen.Secret()),
		otpUrl:    otpGen.URL(),
		expiresAt: model.NewModelDateTime().Calc(func(t time.Time) time.Time {
			return t.Add(time.Duration(tempPasswordDurationMins) * time.Minute)
		}),
	}
}

func (this *PasswordServiceImpl) createOtpRecovery(cmd it.ConfirmPasswordOtpCommand) []string {
	recoveryCodes := make([]string, otpRecoveryCodeCount)
	for i := range otpRecoveryCodeCount {
		recCode, err := crypto.GenerateRecoveryCode()
		ft.PanicOnErr(err)
		recoveryCodes[i] = recCode
	}

	this.logger.Debug("confirm otp password", logging.Attr{
		"principalType": cmd.PrincipalType,
		"principalRef":  cmd.PrincipalId,
		"recoveryCodes": recoveryCodes,
	})
	return recoveryCodes
}

func (this *PasswordServiceImpl) validateNewPass(curPassHash []byte, newPass string, clientErrs *ft.ClientErrors) {
	if curPassHash != nil {
		sameOldPass := this.isPasswordEqual(curPassHash, newPass)
		if sameOldPass {
			clientErrs.Append(*ft.NewValidationError(
				"new_password",
				ft.ErrorKey("err_new_password_same_as_old", "authenticate"),
				"New password must not be the same as the old password",
			))
		}
	}
	if !checkPasswordPolicy(newPass) {
		clientErrs.Append(*ft.NewValidationError(
			"new_password",
			ft.ErrorKey("err_new_password_policy", "authenticate"),
			"Password must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, and one number",
		))
	}
}

func (this *PasswordServiceImpl) validateCurrentAndTempPass(
	candidatePass string, passStore *domain.PasswordStore,
) (bool, *ft.ClientErrorItem) {
	var isMatched bool
	var reason *ft.ClientErrorItem
	// Validate if a password matches either current or temporary password
	if passStore != nil {
		if passStore.GetPassword() != nil {
			isMatched, reason = this.validateCurrentPass(
				[]byte(*passStore.GetPassword()),
				passStore.GetPasswordExpiresAt(),
				candidatePass,
			)
		}
		if !isMatched && passStore.GetPasswordTmp() != nil {
			isMatched, reason = this.validateCurrentPass(
				[]byte(*passStore.GetPasswordTmp()),
				passStore.GetPasswordTmpExpiresAt(),
				candidatePass,
			)
		}
	}
	if reason != nil {
		reason.Field = "password"
	} else if !isMatched {
		reason = ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_password_incorrect", "authenticate"),
			"Incorrect password.",
		)
	}
	return isMatched, reason
}

func (this *PasswordServiceImpl) validateCurrentPass(
	curPassHash []byte, curPassExpiresAt *model.ModelDateTime, candidatePass string,
) (bool, *ft.ClientErrorItem) {
	if curPassExpiresAt != nil && (*curPassExpiresAt).BeforeT(time.Now()) {
		return false, ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_password_expired", "authenticate"),
			"Password has expired.",
		)
	}

	if !this.isPasswordEqual(curPassHash, candidatePass) {
		return false, ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_password_incorrect", "authenticate"),
			"Incorrect password.",
		)
	}

	return true, nil
}

type verifyOtpAndRecoveryResult struct {
	isMatched              bool
	reason                 *ft.ClientErrorItem
	remainingRecoveryCodes []string
}

func (this *PasswordServiceImpl) verifyOtpAndRecovery(
	otpCode domain.OtpCode, passStore *domain.PasswordStore,
) (_ verifyOtpAndRecoveryResult, err error) {
	result := verifyOtpAndRecoveryResult{}
	if passStore != nil {
		if crypto.IsRecoveryCodeFormat(string(otpCode)) {
			result.isMatched, result.remainingRecoveryCodes, result.reason = this.verifyOtpRecovery(otpCode, passStore)
		}
		result.reason, err = this.verifyOtpCode(otpCode, passStore)
		result.isMatched = (result.reason == nil)
	}

	return result, nil
}

func (this *PasswordServiceImpl) verifyOtpRecovery(
	otpCode domain.OtpCode, passStore *domain.PasswordStore,
) (bool, []string, *ft.ClientErrorItem) {
	if passStore.GetPasswordOtpRecovery() == nil {
		return false, nil, ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_recovery_code_mismatched", "authenticate"),
			"Recovery code mismatched.",
		)
	}
	remainingRecoveries, isMatched := array.RemoveString(passStore.GetPasswordOtpRecovery(), string(otpCode))

	if !isMatched {
		return false, nil, ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_recovery_code_mismatched", "authenticate"),
			"Recovery code mismatched.",
		)
	}
	return true, remainingRecoveries, nil
}

func (this *PasswordServiceImpl) verifyOtpCode(
	otpCode domain.OtpCode, passStore *domain.PasswordStore,
) (*ft.ClientErrorItem, error) {
	if passStore.GetPasswordOtp() == nil {
		return ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_code_mismatched", "authenticate"),
			"OTP code mismatched.",
		), nil
	}
	otpExpiresAt := passStore.GetPasswordOtpExpiresAt()
	if otpExpiresAt != nil && (*otpExpiresAt).BeforeT(time.Now()) {
		return ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_register_timeout", "authenticate"),
			"OTP register process timed out. Please start over.",
		), nil
	}
	isMatched, err := totp.ValidateCustom(string(otpCode), passStore.MustGetPasswordOtp(), time.Now(), totp.ValidateOpts{
		Digits: c.OtpCodeLength,
		Period: c.OtpPeriod,
		Skew:   c.OtpSkew,
	})
	if err != nil {
		return nil, err
	}

	if !isMatched {
		return ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_code_mismatched", "authenticate"),
			"OTP code mismatched.",
		), nil
	}

	return nil, nil
}

// Fetches password store with either `principalId` or `username` as the filter.
func (this *PasswordServiceImpl) tryFetchPassStore(
	ctx corectx.Context, principalType domain.PrincipalType, principalId *model.Id,
	username *string, cErrs *ft.ClientErrors,
) (*domain.PasswordStore, *loginPrincipal, error) {
	principal, err := this.principalHelper.assertPrincipalExists(ctx, principalType, principalId, username, cErrs)
	if err != nil {
		return nil, nil, err
	}

	if principal == nil {
		return nil, nil, nil
	}

	passStore, err := this.findPasswordStore(ctx, principalType, principal.Id)
	ft.PanicOnErr(err)

	return passStore, principal, nil
}

func (this *PasswordServiceImpl) findPasswordStore(
	ctx corectx.Context, principalType domain.PrincipalType, principalRef model.Id,
) (*domain.PasswordStore, error) {
	passResult, err := this.passwordRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			domain.PasswordStoreFieldPrincipalType: string(principalType),
			domain.PasswordStoreFieldPrincipalId:   string(principalRef),
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
	ctx corectx.Context, curPassStore *domain.PasswordStore, newPassStore domain.PasswordStore,
) error {
	if curPassStore != nil {
		newPassStore.SetId(curPassStore.GetId())
		resUpdate, err := this.passwordRepo.Update(ctx, newPassStore)
		if err != nil {
			return err
		}
		if resUpdate.ClientErrors.Count() > 0 {
			return errors.Wrap(resUpdate.ClientErrors.ToError(), "upsertPassStore")
		}
		return nil
	}

	defaultFields, _ := this.passwordRepo.GetBaseRepo().Schema().Validate(newPassStore.GetFieldData())
	newPassStore.SetFieldData(defaultFields)

	insertRes, err := this.passwordRepo.Insert(ctx, newPassStore)
	if err != nil {
		return err
	}
	if insertRes.ClientErrors.Count() > 0 {
		return errors.Wrap(insertRes.ClientErrors.ToError(), "upsertPassStore")
	}
	return nil
}
