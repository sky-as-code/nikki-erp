package services

import (
	"regexp"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/sky-as-code/nikki-erp/common/crypto"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	coreConst "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/password"
)

type createOtpResult struct {
	otpSecret string
	otpUrl    string
	expiresAt model.ModelDateTime
}

type principalPasswordStores struct {
	password    *models.PasswordStore
	passwordTmp *models.PasswordStore
	otpSecret   *models.PasswordStore
	otpRecovery *models.PasswordStore
}

func (this principalPasswordStores) getPasswordHash() *string {
	if this.password == nil {
		return nil
	}
	return this.password.GetHash()
}

func (this principalPasswordStores) getTempPasswordHash() *string {
	if this.passwordTmp == nil {
		return nil
	}
	return this.passwordTmp.GetHash()
}

func (this *PasswordDomainServiceImpl) createOtp(username string) createOtpResult {
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

func (this *PasswordDomainServiceImpl) createOtpRecovery(cmd it.ConfirmPasswordOtpCommand) []string {
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

func (this *PasswordDomainServiceImpl) validateNewPass(curPassHash []byte, newPass string, clientErrs *ft.ClientErrors) {
	if curPassHash != nil {
		sameOldPass := this.isPasswordEqual(curPassHash, newPass)
		if sameOldPass {
			clientErrs.Append(*ft.NewValidationError(
				"new_password",
				ft.ErrorKey("err_new_password_same_as_old", "iam"),
				"New password must not be the same as the old password",
			))
		}
	}
	if !checkPasswordPolicy(newPass) {
		clientErrs.Append(*ft.NewValidationError(
			"new_password",
			ft.ErrorKey("err_new_password_policy", "iam"),
			"Password must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, and one number",
		))
	}
}

func (this *PasswordDomainServiceImpl) validateCurrentAndTempPass(
	candidatePass string, stores *principalPasswordStores,
) (bool, *ft.ClientErrorItem) {
	var isMatched bool
	var reason *ft.ClientErrorItem
	if stores != nil {
		if stores.getPasswordHash() != nil {
			isMatched, reason = this.validateCurrentPass(
				[]byte(*stores.getPasswordHash()),
				stores.password.GetExpiresAt(),
				candidatePass,
			)
		}
		if !isMatched && stores.getTempPasswordHash() != nil {
			isMatched, reason = this.validateCurrentPass(
				[]byte(*stores.getTempPasswordHash()),
				stores.passwordTmp.GetExpiresAt(),
				candidatePass,
			)
		}
	}
	if reason != nil {
		reason.Field = "password"
	} else if !isMatched {
		reason = ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_password_incorrect", "iam"),
			"Incorrect password.",
		)
	}
	return isMatched, reason
}

func (this *PasswordDomainServiceImpl) validateCurrentPass(
	curPassHash []byte, curPassExpiresAt *model.ModelDateTime, candidatePass string,
) (bool, *ft.ClientErrorItem) {
	if curPassExpiresAt != nil && (*curPassExpiresAt).BeforeT(time.Now()) {
		return false, ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_password_expired", "iam"),
			"Password has expired.",
		)
	}

	if !this.isPasswordEqual(curPassHash, candidatePass) {
		return false, ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_password_incorrect", "iam"),
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

func (this *PasswordDomainServiceImpl) verifyOtpAndRecovery(
	otpCode models.OtpCode, stores *principalPasswordStores,
) (_ verifyOtpAndRecoveryResult, err error) {
	result := verifyOtpAndRecoveryResult{}
	if stores != nil {
		if crypto.IsRecoveryCodeFormat(string(otpCode)) {
			result.isMatched, result.remainingRecoveryCodes, result.reason = this.verifyOtpRecovery(otpCode, stores)
			if result.reason == nil {
				return result, nil
			}
		}
		result.reason, err = this.verifyOtpCode(otpCode, stores)
		result.isMatched = (result.reason == nil)
	}

	return result, nil
}

func (this *PasswordDomainServiceImpl) verifyOtpRecovery(
	otpCode models.OtpCode, stores *principalPasswordStores,
) (bool, []string, *ft.ClientErrorItem) {
	currentRecoveryHashes := decodeRecoveryCodes(stores)
	if currentRecoveryHashes == nil {
		return false, nil, ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_recovery_code_mismatched", "iam"),
			"Recovery code mismatched.",
		)
	}
	matchIndex := -1
	for index, storedHash := range currentRecoveryHashes {
		isMatched, err := crypto.CompareHashAndPassword([]byte(storedHash), []byte(otpCode))
		if err != nil {
			this.logger.Warn("error with crypto.CompareHashAndPassword()", logging.Attr{
				"error": err.Error(),
			})
			continue
		}
		if isMatched {
			matchIndex = index
			break
		}
	}
	if matchIndex < 0 {
		return false, nil, ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_recovery_code_mismatched", "iam"),
			"Recovery code mismatched.",
		)
	}
	remainingRecoveries := append(
		currentRecoveryHashes[:matchIndex],
		currentRecoveryHashes[matchIndex+1:]...,
	)
	return true, remainingRecoveries, nil
}

func (this *PasswordDomainServiceImpl) verifyOtpCode(
	otpCode models.OtpCode, stores *principalPasswordStores,
) (*ft.ClientErrorItem, error) {
	if stores == nil || stores.otpSecret == nil || stores.otpSecret.GetHash() == nil {
		return ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_code_mismatched", "iam"),
			"OTP code mismatched.",
		), nil
	}
	otpExpiresAt := stores.otpSecret.GetExpiresAt()
	if otpExpiresAt != nil && (*otpExpiresAt).BeforeT(time.Now()) {
		return ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_register_timeout", "iam"),
			"OTP register process timed out. Please start over.",
		), nil
	}
	isMatched, err := totp.ValidateCustom(
		string(otpCode),
		stores.otpSecret.MustGetHash(),
		time.Now(),
		totp.ValidateOpts{
			Digits: c.OtpCodeLength,
			Period: c.OtpPeriod,
			Skew:   c.OtpSkew,
		},
	)
	if err != nil {
		return nil, err
	}

	if !isMatched {
		return ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_otp_code_mismatched", "iam"),
			"OTP code mismatched.",
		), nil
	}

	return nil, nil
}

func (this *PasswordDomainServiceImpl) tryFetchUserForPassword(
	ctx corectx.Context, principalType models.PrincipalType, principalId *model.Id,
	username *string, cErrs *ft.ClientErrors,
) (*models.User, *loginPrincipal, *principalPasswordStores, error) {
	principal, err := this.principalHelper.assertPrincipalExists(ctx, principalType, principalId, username, cErrs)
	if err != nil {
		return nil, nil, nil, err
	}

	if principal == nil {
		return nil, nil, nil, nil
	}

	user, err := this.findUserById(ctx, principal.Id)
	ft.PanicOnErr(err)
	if user == nil {
		return nil, nil, nil, nil
	}

	stores, err := this.findPasswordStoresByPrincipal(ctx, principalType, principal.Id)
	if err != nil {
		return nil, nil, nil, err
	}

	return user, principal, stores, nil
}

func (this *PasswordDomainServiceImpl) findUserById(
	ctx corectx.Context, userId model.Id,
) (*models.User, error) {
	userResult, err := this.userRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			basemodel.FieldId: string(userId),
		},
	})
	if err != nil {
		return nil, err
	}
	if userResult.ClientErrors.Count() > 0 || !userResult.HasData {
		return nil, nil
	}
	return &userResult.Data, nil
}

func (this *PasswordDomainServiceImpl) isPasswordEqual(passHash []byte, candidatePass string) bool {
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

func (this *PasswordDomainServiceImpl) findPasswordStoresByPrincipal(
	ctx corectx.Context, principalType models.PrincipalType, principalId model.Id,
) (*principalPasswordStores, error) {
	// The search must be expressed as a Graph: Search ignores the Filter field
	// entirely, so the previous filter selected nothing and the query returned an
	// arbitrary page of OTHER principals' stores. Every password check against a
	// freshly provisioned account therefore failed as "incorrect password".
	//
	// Page numbering is zero-based, as everywhere else in the codebase.
	graph := dmodel.NewSearchGraph()
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.PasswordStoreFieldPrincipalType, dmodel.Equals, string(principalType),
		),
		*dmodel.NewSearchNode().NewCondition(
			models.PasswordStoreFieldPrincipalId, dmodel.Equals, string(principalId),
		),
	)
	searchResult, err := this.passwordStoreRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  10,
	})
	if err != nil {
		return nil, err
	}
	if searchResult.ClientErrors.Count() > 0 {
		return nil, searchResult.ClientErrors.ToError()
	}

	stores := &principalPasswordStores{}
	for index := range searchResult.Data.Items {
		item := &searchResult.Data.Items[index]
		if item.GetType() == nil {
			continue
		}
		switch *item.GetType() {
		case models.PasswordStoreTypePassword:
			stores.password = item
		case models.PasswordStoreTypePasswordTemp:
			stores.passwordTmp = item
		case models.PasswordStoreTypeOtpSecret:
			stores.otpSecret = item
		case models.PasswordStoreTypeOtpRecovery:
			stores.otpRecovery = item
		}
	}
	return stores, nil
}

func (this *PasswordDomainServiceImpl) upsertPasswordStoreHash(
	ctx corectx.Context,
	principalType models.PrincipalType,
	principalId model.Id,
	passwordType models.PasswordStoreType,
	hash *string,
	expiresAt *model.ModelDateTime,
	lastUsedAt *model.ModelDateTime,
) error {
	existingResult, err := this.passwordStoreRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			models.PasswordStoreFieldPrincipalType: string(principalType),
			models.PasswordStoreFieldPrincipalId:   string(principalId),
			models.PasswordStoreFieldType:          string(passwordType),
		},
	})
	if err != nil {
		return err
	}
	if existingResult.ClientErrors.Count() > 0 {
		return existingResult.ClientErrors.ToError()
	}

	var record models.PasswordStore
	if existingResult.HasData {
		record = existingResult.Data
	} else {
		record = *models.NewPasswordStore()
		// The id is generated here rather than by the database: the column has no
		// default, so inserting without one fails the NOT NULL constraint and no
		// password could ever be stored for a principal that had none before.
		newId, err := model.NewId()
		if err != nil {
			return err
		}
		record.SetId(newId)
		record.SetPrincipalType(&principalType)
		record.SetPrincipalId(&principalId)
		record.SetType(&passwordType)
	}
	record.SetHash(hash)
	record.SetExpiresAt(expiresAt)
	record.SetLastUsedAt(lastUsedAt)

	if existingResult.HasData {
		updateResult, err := this.passwordStoreRepo.Update(ctx, record)
		if err != nil {
			return err
		}
		if updateResult.ClientErrors.Count() > 0 {
			return updateResult.ClientErrors.ToError()
		}
		return nil
	}

	insertResult, err := this.passwordStoreRepo.Insert(ctx, record)
	if err != nil {
		return err
	}
	if insertResult.ClientErrors.Count() > 0 {
		return insertResult.ClientErrors.ToError()
	}
	return nil
}

func (this *PasswordDomainServiceImpl) encodeRecoveryCodes(recoveryCodes []string) *string {
	if len(recoveryCodes) == 0 {
		return nil
	}
	hashedRecoveryCodes := make([]string, len(recoveryCodes))
	for index := range recoveryCodes {
		hashedRecoveryCodes[index] = this.hashRecoveryCode(recoveryCodes[index])
	}
	recoveryCodeStr := strings.Join(hashedRecoveryCodes, ",")
	return &recoveryCodeStr
}

func (this *PasswordDomainServiceImpl) hashRecoveryCode(recoveryCode string) string {
	recoveryHash, err := crypto.GenerateFromPassword([]byte(recoveryCode))
	ft.PanicOnErr(err)
	return string(recoveryHash)
}

func decodeRecoveryCodes(stores *principalPasswordStores) []string {
	if stores == nil || stores.otpRecovery == nil || stores.otpRecovery.GetHash() == nil {
		return nil
	}
	return strings.Split(*stores.otpRecovery.GetHash(), ",")
}
