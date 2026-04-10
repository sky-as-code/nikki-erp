package app

import (
	"time"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/app/methods"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
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

func (this *AttemptServiceImpl) CreateLoginAttempt(
	ctx corectx.Context, cmd it.CreateLoginAttemptCommand,
) (*it.CreateLoginAttemptResult, error) {
	var principal *attemptPrincipal

	resAttempt, err := corecrud.Create(ctx, corecrud.CreateParam[domain.LoginAttempt, *domain.LoginAttempt]{
		Action:         "create login attempt",
		BaseRepoGetter: this.attemptRepo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, attempt *domain.LoginAttempt, cErrsTotal *ft.ClientErrors) error {
			var err error
			cErrs, err := dyn.StartValidationFlowCopy(cErrsTotal).
				StepS(func(cErrs *ft.ClientErrors, stop func()) error {
					principal, err = this.assertPrincipalExists(ctx, attempt, cErrs)
					return err
				}).
				Step(func(cErrs *ft.ClientErrors) error {
					methods := []string{methods.LoginPassword} // TODO: load method settings from DB
					if len(methods) == 0 {
						cErrs.Append(*ft.NewAnonymousBusinessViolation(
							ft.ErrorKey("no_available_methods", "authenticate"),
							"no available login methods for this account"))
						return nil
					}
					attempt.SetMethods(methods)
					return nil
				}).
				End()
			cErrsTotal.Concat(cErrs)
			return err
		},
		AfterValidationSuccess: func(ctx corectx.Context, attempt *domain.LoginAttempt) (*domain.LoginAttempt, error) {
			durationTime := time.Duration(this.attemptDurationSecs) * time.Second
			expiresAt := model.NewModelDateTime().Calc(func(t time.Time) time.Time {
				return t.Add(durationTime)
			})
			attempt.SetExpiresAt(&expiresAt)
			m := attempt.MustGetMethods()
			attempt.SetCurrentMethod(&m[0])
			return attempt, nil
		},
	})

	if err != nil {
		return nil, err
	}
	if resAttempt.ClientErrors.Count() > 0 {
		return &it.CreateLoginAttemptResult{
			ClientErrors: resAttempt.ClientErrors,
		}, nil
	}

	return &it.CreateLoginAttemptResult{
		Data: it.CreateLoginAttemptResultData{
			Attempt:       resAttempt.Data,
			PrincipalName: principal.Name,
		},
		HasData: true,
	}, nil
}

func (this *AttemptServiceImpl) UpdateLoginAttempt(
	ctx corectx.Context, cmd it.UpdateLoginAttemptCommand,
) (*it.UpdateLoginAttemptResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.LoginAttempt, *domain.LoginAttempt]{
		Action:       "update login attempt",
		DbRepoGetter: this.attemptRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, attempt *domain.LoginAttempt, foundAttempt *domain.LoginAttempt, cErrsTotal *ft.ClientErrors) error {
			cErrs, err := dyn.StartValidationFlowCopy(cErrsTotal).
				Step(func(cErrs *ft.ClientErrors) error {
					this.assertNewStatusValid(foundAttempt, attempt.GetStatus(), cErrs)
					return nil
				}).
				Step(func(cErrs *ft.ClientErrors) error {
					this.assertNewMethodValid(foundAttempt, attempt.GetCurrentMethod(), cErrs)
					return nil
				}).
				End()
			cErrsTotal.Concat(cErrs)
			return err
		},
	})
}

func (this *AttemptServiceImpl) GetAttempt(ctx corectx.Context, query it.GetAttemptQuery) (*it.GetAttemptResult, error) {
	return corecrud.GetOne[domain.LoginAttempt](ctx, corecrud.GetOneParam{
		Action:       "get login attempt",
		DbRepoGetter: this.attemptRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *AttemptServiceImpl) assertNewStatusValid(
	dbAttempt *domain.LoginAttempt, newStatus *domain.AttemptStatus, clientErrs *ft.ClientErrors,
) {
	if newStatus == nil {
		return
	}
	st := dbAttempt.GetStatus()
	if st != nil && *st != domain.AttemptStatusPending {
		clientErrs.Append(*ft.NewValidationError(
			"status", ft.ErrorKey("err_attempt_already_settled", "authenticate"), "attempt already settled",
		))
	}
}

func (this *AttemptServiceImpl) assertNewMethodValid(
	dbAttempt *domain.LoginAttempt, newMethod *string, clientErrs *ft.ClientErrors,
) {
	if newMethod == nil {
		return
	}
	methodImpl := methods.GetLoginMethod(*newMethod)
	notExists := methodImpl == nil

	methods := dbAttempt.GetMethods()
	cur := dbAttempt.GetCurrentMethod()
	var curStr string
	if cur != nil {
		curStr = *cur
	}
	newIdx := array.IndexOf(methods, *newMethod)
	notAssigned := newIdx == -1

	curIdx := array.IndexOf(methods, curStr)
	notNextStep := newIdx <= curIdx

	if notExists || notAssigned || notNextStep {
		clientErrs.Append(*ft.NewValidationError(
			"current_method", ft.ErrorKey("err_not_applicable_login_method", "authenticate"), "not applicable login method",
		))
	}
}

type attemptPrincipal struct {
	Id       model.Id
	Name     string
	Username string
}

func (this *AttemptServiceImpl) assertPrincipalExists(
	ctx corectx.Context, attempt *domain.LoginAttempt, clientErrs *ft.ClientErrors,
) (subject *attemptPrincipal, err error) {
	switch attempt.MustGetPrincipalType() {
	case domain.PrincipalTypeNikkiUser:
		subject, err = this.assertUserExists(ctx, attempt.MustGetUsername(), clientErrs)
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
	ctx corectx.Context, username string, cErrs *ft.ClientErrors,
) (*attemptPrincipal, error) {
	query := external.GetUserQuery{
		Email: &username,
	}
	coreCtx := ctx.(corectx.Context)
	userRes, err := this.userExtSvc.GetUser(coreCtx, query)
	if err != nil {
		return nil, err
	}
	if userRes.ClientErrors.Count() > 0 {
		cErrs.Append(userRes.ClientErrors...)
		return nil, nil
	}
	if !userRes.HasData {
		cErrs.Append(*ft.NewNotFoundError(domain.AttemptFieldUsername))
		return nil, nil
	}
	user := userRes.Data
	return &attemptPrincipal{
		Id:       *user.GetId(),
		Name:     *user.GetDisplayName(),
		Username: *user.GetEmail(),
	}, nil
}
