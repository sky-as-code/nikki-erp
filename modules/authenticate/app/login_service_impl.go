package app

import (
	"context"
	"fmt"
	"time"

	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	m "github.com/sky-as-code/nikki-erp/modules/authenticate/app/methods"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
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

	attemptDurationSecs int
}

func (s *LoginServiceImpl) Authenticate(ctx context.Context, cmd it.AuthenticateCommand) (result *it.AuthenticateResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "authenticate"); e != nil {
			err = e
		}
	}()

	attempt, vErrs, err := s.validateAuthInput(ctx, cmd)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.AuthenticateResult{ClientError: vErrs.ToClientError()}, nil
	}

	done, err := s.performLoginMethods(ctx, cmd, attempt, vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.AuthenticateResult{ClientError: vErrs.ToClientError()}, nil
	}

	if err := s.updateAttemptStatus(ctx, attempt); err != nil {
		return nil, err
	}

	return s.buildAuthenticateResult(done, attempt), nil
}

func (s *LoginServiceImpl) validateAuthInput(ctx context.Context, cmd it.AuthenticateCommand) (*domain.LoginAttempt, *ft.ValidationErrors, error) {
	var attempt *domain.LoginAttempt

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			var err error
			attempt, err = s.assertAttemptExists(ctx, cmd.AttemptId, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			s.assertAttemptValid(ctx, attempt, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			var err error
			_, err = s.subjectHelper.assertSubjectExists(ctx, *attempt.SubjectType, nil, attempt.Username, vErrs)
			return err
		}).
		End()

	if err != nil {
		return nil, nil, err
	}
	return attempt, &vErrs, nil
}

func (s *LoginServiceImpl) performLoginMethods(
	ctx context.Context,
	cmd it.AuthenticateCommand,
	attempt *domain.LoginAttempt,
	vErrs *ft.ValidationErrors,
) (done bool, err error) {

	requiredMethod := *attempt.CurrentMethod

	for {
		methodName := *attempt.CurrentMethod
		submittedPassword, ok := cmd.Passwords[methodName]

		if !ok {
			if methodName == requiredMethod {
				vErrs.Appendf(fmt.Sprintf("passwords.%s", methodName), "%s mismatched", methodName)
			}
			break
		}

		method := m.GetLoginMethod(methodName)
		var exeResult *it.ExecuteResult
		exeResult, err = method.Execute(ctx, it.LoginParam{
			SubjectType: *attempt.SubjectType,
			Username:    *attempt.Username,
			Password:    submittedPassword,
		})
		if err != nil {
			return false, err
		}

		if exeResult.ClientErr != nil {
			if vErrs.MergeClientError(exeResult.ClientErr) {
				return false, nil
			} else {
				return false, exeResult.ClientErr
			}
		}

		if !exeResult.IsVerified {
			vErrs.Append(fmt.Sprintf("passwords.%s", methodName), exeResult.FailedReason)
			return false, nil
		}

		if nextMethod := attempt.NextMethod(); nextMethod == nil {
			attempt.CurrentMethod = nil
			attempt.Status = util.ToPtr(domain.AttemptStatusSuccess)
			return true, nil
		} else {
			attempt.CurrentMethod = nextMethod
		}
	}
	return false, nil
}

func (s *LoginServiceImpl) updateAttemptStatus(ctx context.Context, attempt *domain.LoginAttempt) error {
	attResult, err := s.attemptSvc.UpdateLoginAttempt(ctx, it.UpdateLoginAttemptCommand{
		Id:            *attempt.Id,
		CurrentMethod: attempt.CurrentMethod,
		Status:        attempt.Status,
	})
	if err != nil {
		return err
	}
	if attResult.ClientError != nil {
		return errors.Wrap(attResult.ClientError, "failed to update attempt status")
	}
	return nil
}

func (s *LoginServiceImpl) buildAuthenticateResult(done bool, attempt *domain.LoginAttempt) *it.AuthenticateResult {
	if done {
		return &it.AuthenticateResult{
			Data: &it.AuthenticateResultData{
				Done: true,
				Data: &it.AuthenticateSuccessData{
					AccessToken:           "TODO",
					AccessTokenExpiredAt:  time.Now(),
					RefreshToken:          "TODO",
					RefreshTokenExpiredAt: time.Now(),
				},
			},
			HasData: true,
		}
	}
	return &it.AuthenticateResult{
		Data: &it.AuthenticateResultData{
			Done:     false,
			NextStep: attempt.CurrentMethod,
		},
		HasData: true,
	}
}

func (this *LoginServiceImpl) assertAttemptExists(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (attempt *domain.LoginAttempt, err error) {
	result, err := this.attemptSvc.GetAttemptById(ctx, it.GetAttemptByIdQuery{Id: id})
	if err != nil {
		return nil, err
	}
	vErrs.MergeClientError(result.ClientError)
	attempt = result.Data
	return
}

func (this *LoginServiceImpl) assertAttemptValid(ctx context.Context, attempt *domain.LoginAttempt, vErrs *ft.ValidationErrors) {
	if attempt.ExpiredAt.Before(time.Now()) {
		vErrs.Append("attemptId", "attempt expired")
	} else if *attempt.Status != domain.AttemptStatusPending {
		vErrs.Append("attemptId", "attempt already settled")
	}
}
