package app

import (
	"fmt"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	m "github.com/sky-as-code/nikki-erp/modules/authenticate/app/methods"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	ext "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	coretoken "github.com/sky-as-code/nikki-erp/modules/core/authtoken"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type NewLoginServiceParam struct {
	dig.In

	AttemptSvc it.AttemptService
	ConfigSvc  config.ConfigService
	CqrsBus    cqrs.CqrsBus
	Logger     logging.LoggerService
	TokenSvc   coretoken.AuthTokenService
	UserSvc    ext.UserExtService
}

func NewLoginServiceImpl(param NewLoginServiceParam) it.LoginService {
	return &LoginServiceImpl{
		cqrsBus:    param.CqrsBus,
		attemptSvc: param.AttemptSvc,
		configSvc:  param.ConfigSvc,
		logger:     param.Logger,
		tokenSvc:   param.TokenSvc,
		userSvc:    param.UserSvc,

		attemptDurationSecs: param.ConfigSvc.GetInt(c.LoginAttemptDurationSecs),
		principalHelper: principalHelper{
			cqrsBus: param.CqrsBus,
			userSvc: param.UserSvc,
		},
	}
}

type LoginServiceImpl struct {
	cqrsBus    cqrs.CqrsBus
	attemptSvc it.AttemptService
	configSvc  config.ConfigService
	logger     logging.LoggerService
	tokenSvc   coretoken.AuthTokenService
	userSvc    ext.UserExtService

	attemptDurationSecs int
	principalHelper     principalHelper
}

func (this *LoginServiceImpl) Authenticate(ctx corectx.Context, cmd it.AuthenticateCommand) (result *it.AuthenticateResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "authenticate"); e != nil {
			err = e
		}
	}()

	// TODO: Prevent login with "system" account.
	dbAttempt, cErrs, err := this.validateAuthInput(ctx, cmd)
	ft.PanicOnErr(err)

	if cErrs.Count() > 0 {
		return &it.AuthenticateResult{ClientErrors: cErrs}, nil
	}

	done, err := this.performLoginMethods(ctx, cmd, dbAttempt, &cErrs)
	ft.PanicOnErr(err)

	if cErrs.Count() > 0 {
		return &it.AuthenticateResult{ClientErrors: cErrs}, nil
	}

	if err := this.updateAttemptStatus(ctx, dbAttempt); err != nil {
		return nil, err
	}

	if !done {
		return &it.AuthenticateResult{
			Data: it.AuthenticateResultData{
				Done:     false,
				NextStep: dbAttempt.GetCurrentMethod(),
			},
			HasData: true,
		}, nil
	}

	tokenPack, err := this.buildTokenPack(
		ctx,
		dbAttempt.MustGetUsername(),
		dbAttempt.MustGetPrincipalType(),
	)
	if err != nil {
		return nil, err
	}

	return &it.AuthenticateResult{
		Data: it.AuthenticateResultData{
			Done: true,
			Data: tokenPack,
		},
		HasData: true,
	}, nil
}

func (this *LoginServiceImpl) RefreshToken(ctx corectx.Context, cmd it.RefreshTokenCommand) (result *it.RefreshTokenResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "refresh token"); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &it.RefreshTokenResult{ClientErrors: cErrs}, nil
	}
	cmd = *sanitized.(*it.RefreshTokenCommand)

	resVerify, err := this.tokenSvc.VerifyJwt(ctx, coretoken.VerifyJwtParam{
		Token: cmd.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	if !resVerify.IsOk {
		cErrs.Append(*ft.NewValidationError("refresh_token", ft.ErrorKey("err_invalid_refresh_token", "authenticate"), "Invalid refresh token"))
		return &it.RefreshTokenResult{ClientErrors: cErrs}, nil
	}

	sub := resVerify.Claims["sub"].(string)
	subParts := strings.Split(sub, ":")
	username := subParts[0]
	principalType := domain.PrincipalType(subParts[1])

	_, err = this.principalHelper.assertPrincipalExists(
		ctx, principalType, nil, &username, &cErrs,
	)
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &it.RefreshTokenResult{ClientErrors: cErrs}, nil
	}
	tokenPack, err := this.buildTokenPack(ctx, username, principalType)
	if err != nil {
		return nil, err
	}

	return &it.RefreshTokenResult{
		Data: it.RefreshTokenResultData{
			AccessToken:           tokenPack.AccessToken,
			AccessTokenExpiresAt:  tokenPack.AccessTokenExpiresAt,
			RefreshToken:          tokenPack.RefreshToken,
			RefreshTokenExpiresAt: tokenPack.RefreshTokenExpiresAt,
		},
		HasData: true,
	}, nil
}

func (this *LoginServiceImpl) validateAuthInput(ctx corectx.Context, cmd it.AuthenticateCommand) (*domain.LoginAttempt, ft.ClientErrors, error) {
	var dbAttempt *domain.LoginAttempt

	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return nil, cErrs, nil
	}
	cmd = *sanitized.(*it.AuthenticateCommand)

	flow := dyn.StartValidationFlowCopy(&cErrs)
	cErrs, err := flow.
		Step(func(cErrs *ft.ClientErrors) error {
			var err error
			dbAttempt, err = this.assertAttemptExists(ctx, cmd.AttemptId, cErrs)
			return err
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			this.assertAttemptValid(dbAttempt, cErrs)
			return nil
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			var err error
			_, err = this.principalHelper.assertPrincipalExists(
				ctx, dbAttempt.MustGetPrincipalType(), nil, dbAttempt.GetUsername(), cErrs,
			)
			return err
		}).
		End()

	if err != nil {
		return nil, nil, err
	}
	return dbAttempt, cErrs, nil
}

func (this *LoginServiceImpl) performLoginMethods(
	ctx corectx.Context, cmd it.AuthenticateCommand,
	dbAttempt *domain.LoginAttempt, cErrs *ft.ClientErrors,
) (done bool, err error) {
	currentMethod := dbAttempt.GetCurrentMethod()
	requiredMethod := *currentMethod

	for {
		currentMethod = dbAttempt.GetCurrentMethod()
		if currentMethod == nil {
			return false, nil
		}
		methodName := *currentMethod
		submittedPassword, isPasswordSent := cmd.Passwords[methodName]

		if !isPasswordSent {
			if methodName == requiredMethod {
				cErrs.Append(*ft.NewValidationError(
					"passwords.s"+methodName,
					ft.ErrorKey("err_password_mismatched_"+methodName, "authenticate"),
					"Incorrect password",
				))
			}
			break
		}

		method := m.GetLoginMethod(methodName)
		var exeResult *it.ExecuteResult
		exeResult, err = method.Execute(ctx, it.LoginParam{
			PrincipalType: dbAttempt.MustGetPrincipalType(),
			Username:      dbAttempt.MustGetUsername(),
			Password:      submittedPassword,
		})
		if err != nil {
			return false, err
		}

		if exeResult.ClientErrors.Count() > 0 {
			cErrs.Append(exeResult.ClientErrors...)
			return false, nil
		}

		if !exeResult.IsVerified {
			cErrs.Append(*ft.NewValidationError(
				"passwords."+methodName,
				exeResult.FailedReason.Key,
				exeResult.FailedReason.Message,
			))
			return false, nil
		}

		if nextMethod := dbAttempt.NextMethod(); nextMethod == nil {
			dbAttempt.SetCurrentMethod(nil)
			dbAttempt.SetStatus(util.ToPtr(domain.AttemptStatusSuccess))
			return true, nil
		} else {
			dbAttempt.SetCurrentMethod(nextMethod)
		}
	}
	return false, nil
}

func (this *LoginServiceImpl) updateAttemptStatus(ctx corectx.Context, dbAttempt *domain.LoginAttempt) error {
	updatedAtt := domain.NewLoginAttempt()
	updatedAtt.SetId(dbAttempt.GetId())
	updatedAtt.SetCurrentMethod(dbAttempt.GetCurrentMethod())
	updatedAtt.SetStatus(dbAttempt.GetStatus())
	attResult, err := this.attemptSvc.UpdateLoginAttempt(ctx, it.UpdateLoginAttemptCommand{
		LoginAttempt: *updatedAtt,
	})
	if err != nil {
		return err
	}
	if attResult.ClientErrors.Count() > 0 {
		return errors.Wrap(attResult.ClientErrors.ToError(), "updateAttemptStatus")
	}
	return nil
}

func (this *LoginServiceImpl) buildTokenPack(
	ctx corectx.Context, username string, principalType domain.PrincipalType,
) (*it.AuthenticateSuccessData, error) {
	sub := fmt.Sprintf("%s:%s", username, principalType)
	jwtAccess, err := this.tokenSvc.CreateJwt(ctx, coretoken.CreateJwtParam{
		Sub:     sub,
		Purpose: coretoken.JwtPurposeAccessToken,
	})

	if err != nil {
		return nil, err
	}

	jwtId := jwtAccess.Claims["jti"].(string)
	jwtRefresh, err := this.tokenSvc.CreateJwt(ctx, coretoken.CreateJwtParam{
		Sub:     sub,
		Jti:     &jwtId,
		Purpose: coretoken.JwtPurposeRefreshToken,
	})

	if err != nil {
		return nil, err
	}

	accessExpTime, _ := jwtAccess.Claims.GetExpirationTime()
	refreshExpTime, _ := jwtRefresh.Claims.GetExpirationTime()

	return &it.AuthenticateSuccessData{
		AccessToken:           jwtAccess.Token,
		AccessTokenExpiresAt:  model.WrapModelDateTime(accessExpTime.Time),
		RefreshToken:          jwtRefresh.Token,
		RefreshTokenExpiresAt: model.WrapModelDateTime(refreshExpTime.Time),
	}, nil
}

func (this *LoginServiceImpl) assertAttemptExists(
	ctx corectx.Context, id model.Id, cErrs *ft.ClientErrors,
) (*domain.LoginAttempt, error) {
	result, err := this.attemptSvc.GetAttempt(ctx, it.GetAttemptQuery{Id: id})
	if err != nil {
		return nil, err
	}
	if !result.HasData {
		cErrs.Append(*ft.NewNotFoundError("attempt_id"))
		return nil, nil
	}
	return &result.Data, nil
}

func (this *LoginServiceImpl) assertAttemptValid(dbAttempt *domain.LoginAttempt, cErrs *ft.ClientErrors) {
	expiresAt := dbAttempt.GetExpiresAt()
	status := dbAttempt.GetStatus()

	if expiresAt != nil && expiresAt.BeforeT(time.Now()) {
		cErrs.Append(*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_login_attempt_expired", "authenticate"),
			"Login attempt expired",
		))
	} else if status != nil && *status != domain.AttemptStatusPending {
		cErrs.Append(*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_login_attempt_already_settled", "authenticate"),
			"Login attempt already settled",
		))
	}
}
