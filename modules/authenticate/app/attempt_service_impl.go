package app

import (
	"time"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/defense"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/app/methods"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type NewAttemptServiceParam struct {
	dig.In

	AttemptRepo it.AttemptRepository
	ConfigSvc   config.ConfigService
	UserExtSvc  external.UserExtService
	CqrsBus     cqrs.CqrsBus
}

func NewAttemptServiceImpl(param NewAttemptServiceParam) it.AttemptService {
	return &AttemptServiceImpl{
		cqrsBus:             param.CqrsBus,
		attemptRepo:         param.AttemptRepo,
		attemptDurationSecs: param.ConfigSvc.GetInt(c.LoginAttemptDurationSecs),
		userExtSvc:          param.UserExtSvc,
	}
}

type AttemptServiceImpl struct {
	cqrsBus     cqrs.CqrsBus
	attemptRepo it.AttemptRepository
	userExtSvc  external.UserExtService

	attemptDurationSecs int
}

func (this *AttemptServiceImpl) CreateLoginAttempt(ctx corectx.Context, cmd it.CreateLoginAttemptCommand) (result *it.CreateLoginAttemptResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create attempt attempt"); e != nil {
			err = e
		}
	}()

	attempt := cmd.ToLoginAttempt()
	attempt.SetDefaults()

	var subject *attemptSubject
	flow := dyn.StartValidationFlow()
	clientErrs, err := flow.
		Step(func(clientErrs *ft.ClientErrors) error {
			appendClientErrors(clientErrs, domain.ValidateLoginAttempt(attempt, false))
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			this.sanitizeAttempt(attempt)
			subject, err = this.assertSubjectExists(ctx, *attempt.GetSubjectType(), cmd.Username, clientErrs)
			return err
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			methods := []string{"password"} // TODO: load method settings from DB
			if len(methods) == 0 {
				clientErrs.Append(*ft.NewAnonymousBusinessViolation("", "no attempt methods available"))
				return nil
			}
			attempt.SetMethods(methods)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.CreateLoginAttemptResult{
			ClientErrors: clientErrs,
		}, nil
	}

	durationTime := time.Duration(this.attemptDurationSecs) * time.Second
	attempt.SetExpiredAt(util.ToPtr(time.Now().Add(durationTime)))
	m := attempt.GetMethods()
	attempt.SetCurrentMethod(util.ToPtr(m[0]))
	attempt.SetSubjectRef(util.ToPtr(subject.Id))
	attempt.SetSubjectType(&cmd.SubjectType)
	attempt.SetUsername(&cmd.Username)

	insertRes, err := this.attemptRepo.Insert(ctx, *attempt)
	ft.PanicOnErr(err)
	if insertRes.ClientErrors.Count() > 0 {
		return &it.CreateLoginAttemptResult{ClientErrors: insertRes.ClientErrors}, nil
	}

	attempt, err = this.getAttemptById(ctx, *attempt.GetId())
	ft.PanicOnErr(err)

	return &it.CreateLoginAttemptResult{
		Data: &it.CreateLoginAttemptResultData{
			Attempt:     *attempt,
			SubjectName: subject.Name,
		},
		HasData: attempt != nil,
	}, nil
}

func (this *AttemptServiceImpl) UpdateLoginAttempt(ctx corectx.Context, cmd it.UpdateLoginAttemptCommand) (result *it.UpdateLoginAttemptResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create attempt attempt"); e != nil {
			err = e
		}
	}()

	attempt := cmd.ToLoginAttempt()

	var dbAttempt *domain.LoginAttempt
	flow := dyn.StartValidationFlow()
	clientErrs, err := flow.
		Step(func(clientErrs *ft.ClientErrors) error {
			appendClientErrors(clientErrs, domain.ValidateLoginAttempt(attempt, true))
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			dbAttempt, err = this.assertAttemptExists(ctx, cmd.Id, clientErrs)
			return err
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			this.assertNewStatusValid(ctx, dbAttempt, attempt.GetStatus(), clientErrs)
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			this.assertNewMethodValid(ctx, dbAttempt, attempt.GetCurrentMethod(), clientErrs)
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			this.sanitizeAttempt(attempt)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.UpdateLoginAttemptResult{
			ClientErrors: clientErrs,
		}, nil
	}

	updateRes, err := this.attemptRepo.Update(ctx, *attempt)
	ft.PanicOnErr(err)
	if updateRes.ClientErrors.Count() > 0 {
		return &it.UpdateLoginAttemptResult{ClientErrors: updateRes.ClientErrors}, nil
	}

	attempt, err = this.getAttemptById(ctx, cmd.Id)
	ft.PanicOnErr(err)

	return &it.UpdateLoginAttemptResult{
		Data:    attempt,
		HasData: attempt != nil,
	}, nil
}

func (this *AttemptServiceImpl) sanitizeAttempt(attempt *domain.LoginAttempt) {
	if attempt.GetDeviceName() != nil {
		attempt.SetDeviceName(util.ToPtr(defense.SanitizePlainText(*attempt.GetDeviceName(), true)))
	}
	if attempt.GetDeviceLocation() != nil {
		attempt.SetDeviceLocation(util.ToPtr(defense.SanitizePlainText(*attempt.GetDeviceLocation(), true)))
	}
}

func (this *AttemptServiceImpl) GetAttemptById(ctx corectx.Context, query it.GetAttemptByIdQuery) (result *it.GetAttemptByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get attempt by id"); e != nil {
			err = e
		}
	}()

	var dbAttempt *domain.LoginAttempt
	flow := dyn.StartValidationFlow(query)
	clientErrs, err := flow.
		Step(func(clientErrs *ft.ClientErrors) error {
			return nil
		}).
		Step(func(clientErrs *ft.ClientErrors) error {
			dbAttempt, err = this.assertAttemptExists(ctx, query.Id, clientErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if clientErrs.Count() > 0 {
		return &it.GetAttemptByIdResult{
			ClientErrors: clientErrs,
		}, nil
	}

	return &it.GetAttemptByIdResult{
		Data:    dbAttempt,
		HasData: dbAttempt != nil,
	}, nil
}

func (this *AttemptServiceImpl) assertAttemptExists(
	ctx corectx.Context, id model.Id, clientErrs *ft.ClientErrors,
) (attempt *domain.LoginAttempt, err error) {
	attempt, err = this.getAttemptById(ctx, id)
	if attempt == nil {
		appendNotFoundError(clientErrs, "id", "attempt")
	}
	return
}

func (this *AttemptServiceImpl) assertNewStatusValid(
	ctx corectx.Context, attempt *domain.LoginAttempt, newStatus *domain.AttemptStatus, clientErrs *ft.ClientErrors,
) {
	if newStatus == nil {
		return
	}
	st := attempt.GetStatus()
	if st != nil && *st != domain.AttemptStatusPending {
		appendValidationError(clientErrs, "status", "attempt already settled")
	}
	return
}

func (this *AttemptServiceImpl) assertNewMethodValid(
	ctx corectx.Context, attempt *domain.LoginAttempt, newMethod *string, clientErrs *ft.ClientErrors,
) {
	if newMethod == nil {
		return
	}
	methodImpl := methods.GetLoginMethod(*newMethod)
	notExists := methodImpl == nil

	meths := attempt.GetMethods()
	cur := attempt.GetCurrentMethod()
	var curStr string
	if cur != nil {
		curStr = *cur
	}
	newIdx := array.IndexOf(meths, *newMethod)
	notAssigned := newIdx == -1

	curIdx := array.IndexOf(meths, curStr)
	notNextStep := newIdx <= curIdx

	if notExists || notAssigned || notNextStep {
		appendValidationError(clientErrs, "current_method", "invalid login method")
	}
}

type attemptSubject struct {
	Id       model.Id
	Name     string
	Username string
}

func (this *AttemptServiceImpl) assertSubjectExists(
	ctx corectx.Context, subjectType domain.SubjectType, username string, clientErrs *ft.ClientErrors,
) (subject *attemptSubject, err error) {
	switch subjectType {
	case domain.SubjectTypeUser:
		subject, err = this.assertUserExists(ctx, username, clientErrs)
	}
	// case domain.SubjectTypeCustomer:
	// 	subject, err = this.assertCustomerExists(ctx, username, vErrs)
	// }
	if err != nil {
		return nil, err
	}
	return subject, nil
}

func (this *AttemptServiceImpl) assertUserExists(
	ctx corectx.Context, username string, clientErrs *ft.ClientErrors,
) (*attemptSubject, error) {
	// result := itUser.GetUserByEmailResult{}
	// err := this.cqrsBus.Request(ctx, &itUser.GetUserByEmailQuery{
	// 	Email: username,
	// }, &result)
	// if err != nil {
	// 	return nil, err
	// }
	// If not validation error but another client error
	// if !vErrs.MergeClientError(result.ClientError) {
	// 	return nil, result.ClientError
	// }
	// if result.Data == nil {
	// 	vErrs.Append("user: ", "username not found")
	// 	return nil, nil
	// }

	// if vErrs.Count() > 0 {
	// 	vErrs.RenameKey("email", "username")
	// 	return nil, nil
	// }
	query := external.GetUserQuery{
		Email: &username,
	}
	coreCtx := ctx.(corectx.Context)
	userRes, err := this.userExtSvc.GetUser(coreCtx, query)
	if err != nil {
		return nil, err
	}
	if userRes.ClientErrors.Count() > 0 {
		appendClientErrors(clientErrs, userRes.ClientErrors)
		return nil, nil
	}
	if !userRes.HasData {
		return nil, nil
	}
	user := userRes.Data
	return &attemptSubject{
		Id:       *user.GetId(),
		Name:     *user.GetDisplayName(),
		Username: *user.GetEmail(),
	}, nil
}

func (this *AttemptServiceImpl) getAttemptById(ctx corectx.Context, id model.Id) (*domain.LoginAttempt, error) {
	result, err := this.attemptRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{basemodel.FieldId: string(id)},
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 || !result.HasData {
		return nil, nil
	}
	return &result.Data, nil
}
