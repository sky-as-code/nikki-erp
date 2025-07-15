package app

import (
	"context"
	"time"

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

func (this *LoginServiceImpl) Authenticate(ctx context.Context, cmd it.AuthenticateCommand) (result *it.AuthenticateResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "authenticate"); e != nil {
			err = e
		}
	}()

	var attempt *domain.LoginAttempt
	var subject *loginSubject
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			attempt, err = this.assertAttemptExists(ctx, cmd.AttemptId, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertAttemptValid(ctx, attempt, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			subject, err = this.subjectHelper.assertSubjectExists(ctx, *attempt.SubjectType, cmd.Username, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.AuthenticateResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	method := m.GetLoginMethod(*attempt.CurrentMethod)
	isAuthenticated, err := method.Execute(ctx, it.LoginParam{
		SubjectType: *attempt.SubjectType,
		Username:    subject.Username,
		Password:    cmd.Password,
	})
	ft.PanicOnErr(err)

	if !isAuthenticated {
		return &it.AuthenticateResult{
			Data: &it.AuthenticateResultData{
				Done: false,
			},
			HasData: true,
		}, nil
	}

	nextMethod := attempt.NextMethod()
	done := false
	if nextMethod == nil {
		done = true
		attempt.CurrentMethod = nil
		attempt.Status = util.ToPtr(domain.AttemptStatusSuccess)
	} else {
		attempt.CurrentMethod = nextMethod
	}

	attResult, err := this.attemptSvc.UpdateLoginAttempt(ctx, it.UpdateLoginAttemptCommand{
		Id:            *attempt.Id,
		CurrentMethod: attempt.CurrentMethod,
		Status:        attempt.Status,
	})
	ft.PanicOnErr(err)
	ft.PanicOnErr(attResult.ClientError)

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
		}, nil
	}
	return &it.AuthenticateResult{
		Data: &it.AuthenticateResultData{
			Done:     false,
			NextStep: nextMethod,
		},
		HasData: true,
	}, nil
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
