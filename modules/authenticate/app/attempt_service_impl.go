package app

import (
	"context"
	"time"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/app/methods"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type NewAttemptServiceParam struct {
	dig.In

	AttemptRepo it.AttemptRepository
	ConfigSvc   config.ConfigService
	CqrsBus     cqrs.CqrsBus
}

func NewAttemptServiceImpl(param NewAttemptServiceParam) it.AttemptService {
	return &AttemptServiceImpl{
		cqrsBus:             param.CqrsBus,
		attemptRepo:         param.AttemptRepo,
		attemptDurationSecs: param.ConfigSvc.GetInt(c.LoginAttemptDurationSecs),
	}
}

type AttemptServiceImpl struct {
	cqrsBus     cqrs.CqrsBus
	attemptRepo it.AttemptRepository

	attemptDurationSecs int
}

func (this *AttemptServiceImpl) CreateLoginAttempt(ctx context.Context, cmd it.CreateLoginAttemptCommand) (result *it.CreateLoginAttemptResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create attempt attempt"); e != nil {
			err = e
		}
	}()

	attempt := cmd.ToLoginAttempt()
	attempt.SetDefaults()

	var subject *attemptSubject
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = attempt.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeAttempt(attempt)
			subject, err = this.assertSubjectExists(ctx, *attempt.SubjectType, cmd.Username, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			methods := []string{"password", "captcha"} // TODO: load method settings from DB
			if len(methods) == 0 {
				return ft.ClientError{
					Code:    "unauthorized",
					Details: "no attempt methods available",
				}
			}
			attempt.Methods = methods
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateLoginAttemptResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	durationTime := time.Duration(this.attemptDurationSecs) * time.Second
	attempt.ExpiredAt = util.ToPtr(time.Now().Add(durationTime))
	attempt.CurrentMethod = util.ToPtr(attempt.Methods[0])
	attempt.SubjectRef = util.ToPtr(subject.Id)
	attempt.SubjectType = &cmd.SubjectType
	attempt.Username = &cmd.Username

	attempt, err = this.attemptRepo.Create(ctx, *attempt)
	ft.PanicOnErr(err)

	return &it.CreateLoginAttemptResult{
		Data: &it.CreateLoginAttemptResultData{
			Attempt:     *attempt,
			SubjectName: subject.Name,
		},
		HasData: attempt != nil,
	}, nil
}

func (this *AttemptServiceImpl) UpdateLoginAttempt(ctx context.Context, cmd it.UpdateLoginAttemptCommand) (result *it.UpdateLoginAttemptResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create attempt attempt"); e != nil {
			err = e
		}
	}()

	attempt := cmd.ToLoginAttempt()

	var dbAttempt *domain.LoginAttempt
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = attempt.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbAttempt, err = this.assertAttemptExists(ctx, cmd.Id, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertNewStatusValid(ctx, dbAttempt, attempt.Status, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertNewMethodValid(ctx, dbAttempt, attempt.CurrentMethod, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeAttempt(attempt)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateLoginAttemptResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	attempt, err = this.attemptRepo.Update(ctx, *attempt)
	ft.PanicOnErr(err)

	return &it.UpdateLoginAttemptResult{
		Data:    attempt,
		HasData: attempt != nil,
	}, nil
}

func (this *AttemptServiceImpl) sanitizeAttempt(attempt *domain.LoginAttempt) {
	if attempt.DeviceName != nil {
		attempt.DeviceName = util.ToPtr(defense.SanitizePlainText(*attempt.DeviceName, true))
	}
	if attempt.DeviceLocation != nil {
		attempt.DeviceLocation = util.ToPtr(defense.SanitizePlainText(*attempt.DeviceLocation, true))
	}
}

func (this *AttemptServiceImpl) GetAttemptById(ctx context.Context, query it.GetAttemptByIdQuery) (result *it.GetAttemptByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get attempt by id"); e != nil {
			err = e
		}
	}()

	var dbAttempt *domain.LoginAttempt
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbAttempt, err = this.assertAttemptExists(ctx, query.Id, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetAttemptByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetAttemptByIdResult{
		Data:    dbAttempt,
		HasData: dbAttempt != nil,
	}, nil
}

func (this *AttemptServiceImpl) assertAttemptExists(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (attempt *domain.LoginAttempt, err error) {
	attempt, err = this.attemptRepo.FindById(ctx, it.FindByIdParam{Id: id})
	if attempt == nil {
		vErrs.AppendIdNotFound("attempt")
	}
	return
}

func (this *AttemptServiceImpl) assertNewStatusValid(ctx context.Context, attempt *domain.LoginAttempt, newStatus *domain.AttemptStatus, vErrs *ft.ValidationErrors) {
	if newStatus == nil {
		return
	}
	if *attempt.Status != domain.AttemptStatusPending {
		vErrs.Append("status", "attempt already settled")
	}
	return
}

func (this *AttemptServiceImpl) assertNewMethodValid(ctx context.Context, attempt *domain.LoginAttempt, newMethod *string, vErrs *ft.ValidationErrors) {
	if newMethod == nil {
		return
	}
	methodImpl := methods.GetLoginMethod(*newMethod)
	notExists := methodImpl == nil

	newIdx := array.IndexOf(attempt.Methods, *newMethod)
	notAssigned := newIdx == -1

	curIdx := array.IndexOf(attempt.Methods, *attempt.CurrentMethod)
	notNextStep := newIdx <= curIdx

	if notExists || notAssigned || notNextStep {
		vErrs.Append("currentMethod", "invalid login method")
	}
}

type attemptSubject struct {
	Id       model.Id
	Name     string
	Username string
}

func (this *AttemptServiceImpl) assertSubjectExists(ctx context.Context, subjectType domain.SubjectType, username string, vErrs *ft.ValidationErrors) (subject *attemptSubject, err error) {
	switch subjectType {
	case domain.SubjectTypeUser:
		subject, err = this.assertUserExists(ctx, username, vErrs)
	}
	// case domain.SubjectTypeCustomer:
	// 	subject, err = this.assertCustomerExists(ctx, username, vErrs)
	// }
	if err != nil {
		return nil, err
	}
	return subject, nil
}

func (this *AttemptServiceImpl) assertUserExists(ctx context.Context, username string, vErrs *ft.ValidationErrors) (*attemptSubject, error) {
	result := itUser.GetUserByEmailResult{}
	err := this.cqrsBus.Request(ctx, &itUser.GetUserByEmailQuery{
		Email: username,
	}, &result)
	if err != nil {
		return nil, err
	}
	// If not validation error but another client error
	if !vErrs.MergeClientError(result.ClientError) {
		return nil, result.ClientError
	}
	if vErrs.Count() > 0 {
		vErrs.RenameKey("email", "username")
		return nil, nil
	}
	return &attemptSubject{
		Id:       *result.Data.Id,
		Name:     *result.Data.DisplayName,
		Username: *result.Data.Email,
	}, nil
}
