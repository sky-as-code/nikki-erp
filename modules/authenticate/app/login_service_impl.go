package app

import (
	"fmt"
	"time"

	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	m "github.com/sky-as-code/nikki-erp/modules/authenticate/app/methods"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	coreConstants "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type NewLoginServiceParam struct {
	dig.In

	AttemptSvc it.AttemptService
	ConfigSvc  config.ConfigService
	CqrsBus    cqrs.CqrsBus
}

func NewLoginServiceImpl(param NewLoginServiceParam) it.LoginService {
	return &LoginServiceImpl{
		cqrsBus:             param.CqrsBus,
		attemptSvc:          param.AttemptSvc,
		configSvc:           param.ConfigSvc,
		attemptDurationSecs: param.ConfigSvc.GetInt(c.LoginAttemptDurationSecs),
		subjectHelper: subjectHelper{
			cqrsBus: param.CqrsBus,
		},
	}
}

type LoginServiceImpl struct {
	cqrsBus       cqrs.CqrsBus
	attemptSvc    it.AttemptService
	subjectHelper subjectHelper
	configSvc     config.ConfigService

	attemptDurationSecs int
}

func (s *LoginServiceImpl) Authenticate(ctx corectx.Context, cmd it.AuthenticateCommand) (result *it.AuthenticateResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "authenticate"); e != nil {
			err = e
		}
	}()

	attempt, clientErrs, err := s.validateAuthInput(ctx, cmd)
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.AuthenticateResult{ClientErrors: clientErrs}, nil
	}

	done, err := s.performLoginMethods(ctx, cmd, attempt, &clientErrs)
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.AuthenticateResult{ClientErrors: clientErrs}, nil
	}

	if err := s.updateAttemptStatus(ctx, attempt); err != nil {
		return nil, err
	}

	return s.buildAuthenticateResult(done, attempt), nil
}

func (s *LoginServiceImpl) RefreshToken(ctx corectx.Context, cmd it.RefreshTokenCommand) (result *it.RefreshTokenResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "refresh token"); e != nil {
			err = e
		}
	}()

	clientErrs := ft.NewClientErrors()
	if len(cmd.RefreshToken) == 0 {
		appendValidationError(clientErrs, "refresh_token", "required")
		return &it.RefreshTokenResult{ClientErrors: *clientErrs}, nil
	}

	payload, err := util.ParseGJWToken(cmd.RefreshToken, s.configSvc.GetStr(coreConstants.TokenSecretKey))
	if err != nil {
		appendValidationError(clientErrs, "refresh_token", err.Error())
		return &it.RefreshTokenResult{ClientErrors: *clientErrs}, nil
	}

	accessExpireSeconds := int64(time.Hour.Seconds())
	refreshExpireSeconds := int64((24 * 7) * time.Hour.Seconds())

	accessToken, genErr := util.GenerateGJWToken(
		s.configSvc.GetStr(coreConstants.TokenSecretKey),
		payload.DId,
		payload.UserId,
		"nikki-erp",
		payload.Roles,
		accessExpireSeconds,
	)
	if genErr != nil {
		return nil, genErr
	}

	refreshToken, genErr := util.GenerateGJWToken(
		s.configSvc.GetStr(coreConstants.TokenSecretKey),
		payload.DId,
		payload.UserId,
		"nikki-erp",
		payload.Roles,
		refreshExpireSeconds,
	)
	if genErr != nil {
		return nil, genErr
	}

	now := time.Now()
	return &it.RefreshTokenResult{
		Data: &it.RefreshTokenResultData{
			AccessToken:           accessToken,
			AccessTokenExpiresAt:  now.Add(time.Duration(accessExpireSeconds) * time.Second),
			RefreshToken:          refreshToken,
			RefreshTokenExpiresAt: now.Add(time.Duration(refreshExpireSeconds) * time.Second),
		},
		HasData: true,
	}, nil
}

func (s *LoginServiceImpl) validateAuthInput(ctx corectx.Context, cmd it.AuthenticateCommand) (*domain.LoginAttempt, ft.ClientErrors, error) {
	var attempt *domain.LoginAttempt

	flow := dyn.StartValidationFlow(cmd)
	clientErrs, err := flow.
		Step(func(clientErrs *ft.ClientErrors) error {
			var err error
			attempt, err = s.assertAttemptExists(ctx, cmd.AttemptId, clientErrs)
			return err
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			s.assertAttemptValid(ctx, attempt, clientErrs)
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			var err error
			_, err = s.subjectHelper.assertSubjectExists(ctx, *attempt.GetSubjectType(), nil, attempt.GetUsername(), clientErrs)
			return err
		}).
		End()

	if err != nil {
		return nil, nil, err
	}
	return attempt, clientErrs, nil
}

func (s *LoginServiceImpl) performLoginMethods(
	ctx corectx.Context,
	cmd it.AuthenticateCommand,
	attempt *domain.LoginAttempt,
	clientErrs *ft.ClientErrors,
) (done bool, err error) {
	currentMethod := attempt.GetCurrentMethod()
	if currentMethod == nil {
		appendValidationError(clientErrs, "attempt_id", "attempt has no current method")
		return false, nil
	}
	requiredMethod := *currentMethod

	for {
		currentMethod = attempt.GetCurrentMethod()
		if currentMethod == nil {
			return false, nil
		}
		methodName := *currentMethod
		submittedPassword, ok := cmd.Passwords[methodName]

		if !ok {
			if methodName == requiredMethod {
				appendValidationError(clientErrs, fmt.Sprintf("passwords.%s", methodName), fmt.Sprintf("%s mismatched", methodName))
			}
			break
		}

		method := m.GetLoginMethod(methodName)
		var exeResult *it.ExecuteResult
		exeResult, err = method.Execute(ctx, it.LoginParam{
			SubjectType: *attempt.GetSubjectType(),
			Username:    *attempt.GetUsername(),
			Password:    submittedPassword,
		})
		if err != nil {
			return false, err
		}

		if exeResult.ClientErrors.Count() > 0 {
			appendClientErrors(clientErrs, exeResult.ClientErrors)
			return false, nil
		}

		if !exeResult.IsVerified {
			appendValidationError(clientErrs, fmt.Sprintf("passwords.%s", methodName), exeResult.FailedReason)
			return false, nil
		}

		if nextMethod := attempt.NextMethod(); nextMethod == nil {
			attempt.SetCurrentMethod(nil)
			attempt.SetStatus(util.ToPtr(domain.AttemptStatusSuccess))
			return true, nil
		} else {
			attempt.SetCurrentMethod(nextMethod)
		}
	}
	return false, nil
}

func (s *LoginServiceImpl) updateAttemptStatus(ctx corectx.Context, attempt *domain.LoginAttempt) error {
	attResult, err := s.attemptSvc.UpdateLoginAttempt(ctx, it.UpdateLoginAttemptCommand{
		Id:            *attempt.GetId(),
		CurrentMethod: attempt.GetCurrentMethod(),
		Status:        attempt.GetStatus(),
	})
	if err != nil {
		return err
	}
	if attResult.ClientErrors.Count() > 0 {
		return errors.Wrap(clientErrorsToError(attResult.ClientErrors, "failed to update attempt status"), "failed to update attempt status")
	}
	return nil
}

func (s *LoginServiceImpl) buildAuthenticateResult(done bool, attempt *domain.LoginAttempt) *dyn.OpResult[*it.AuthenticateResultData] {
	// accessExpireSeconds := int64(s.configSvc.GetInt(coreConstants.TokenExpiryHours) * 1)
	// refreshExpireSeconds := int64(s.configSvc.GetInt(coreConstants.TokenExpiryHours) * 12)
	accessExpireSeconds := int64(time.Hour.Seconds())
	refreshExpireSeconds := int64((24 * 7) * time.Hour.Seconds())

	accessToken, _ := util.GenerateGJWToken(
		s.configSvc.GetStr(coreConstants.TokenSecretKey),
		*attempt.GetDeviceIp(),
		*attempt.GetSubjectRef(),
		"nikki-erp",
		attempt.GetMethods(),
		accessExpireSeconds,
	)
	refreshToken, _ := util.GenerateGJWToken(
		s.configSvc.GetStr(coreConstants.TokenSecretKey),
		*attempt.GetDeviceIp(),
		*attempt.GetSubjectRef(),
		"nikki-erp",
		attempt.GetMethods(),
		refreshExpireSeconds,
	)
	if done {
		now := time.Now()
		return &it.AuthenticateResult{
			Data: &it.AuthenticateResultData{
				Done: true,
				Data: &it.AuthenticateSuccessData{
					AccessToken:           accessToken,
					AccessTokenExpiresAt:  now.Add(time.Duration(accessExpireSeconds) * time.Second),
					RefreshToken:          refreshToken,
					RefreshTokenExpiresAt: now.Add(time.Duration(refreshExpireSeconds) * time.Second),
				},
			},
			HasData: true,
		}
	}
	return &it.AuthenticateResult{
		Data: &it.AuthenticateResultData{
			Done:     false,
			NextStep: attempt.GetCurrentMethod(),
		},
		HasData: true,
	}
}

func (this *LoginServiceImpl) assertAttemptExists(
	ctx corectx.Context, id model.Id, clientErrs *ft.ClientErrors,
) (attempt *domain.LoginAttempt, err error) {
	result, err := this.attemptSvc.GetAttemptById(ctx, it.GetAttemptByIdQuery{Id: id})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		appendClientErrors(clientErrs, result.ClientErrors)
	}
	attempt = result.Data
	return
}

func (this *LoginServiceImpl) assertAttemptValid(
	ctx corectx.Context, attempt *domain.LoginAttempt, clientErrs *ft.ClientErrors,
) {
	if attempt.GetExpiredAt() != nil && attempt.GetExpiredAt().Before(time.Now()) {
		appendValidationError(clientErrs, "attempt_id", "attempt expired")
	} else if attempt.GetStatus() != nil && *attempt.GetStatus() != domain.AttemptStatusPending {
		appendValidationError(clientErrs, "attempt_id", "attempt already settled")
	}
}
