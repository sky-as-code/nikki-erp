package services

import (
	"time"

	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/iam/app/methods"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/login"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
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
	UserSvc     itUser.UserDomainService
	CqrsBus     cqrs.CqrsBus
}

func NewAttemptDomainServiceImpl(param NewAttemptServiceParam) it.AttemptDomainService {
	return &AttemptDomainServiceImpl{
		cqrsBus:             param.CqrsBus,
		attemptRepo:         param.AttemptRepo,
		attemptDurationSecs: param.ConfigSvc.GetInt(c.LoginAttemptDurationSecs),
		userSvc:             param.UserSvc,
	}
}

type AttemptDomainServiceImpl struct {
	cqrsBus     cqrs.CqrsBus
	attemptRepo it.AttemptRepository
	userSvc     itUser.UserDomainService

	attemptDurationSecs int
}

func (this *AttemptDomainServiceImpl) CreateLoginAttempt(
	ctx corectx.Context, cmd it.CreateLoginAttemptCommand,
) (*it.CreateLoginAttemptResult, error) {
	var principal *attemptPrincipal

	resAttempt, err := corecrud.Create(ctx, corecrud.CreateParam[models.LoginAttempt, *models.LoginAttempt]{
		Action:         "create login attempt",
		BaseRepoGetter: this.attemptRepo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, attempt *models.LoginAttempt, cErrsTotal *ft.ClientErrors) error {
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
							ft.ErrorKey("no_available_methods", "iam"),
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
		AfterValidationSuccess: func(ctx corectx.Context, attempt *models.LoginAttempt) (*models.LoginAttempt, error) {
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

func (this *AttemptDomainServiceImpl) UpdateLoginAttempt(
	ctx corectx.Context, cmd it.UpdateLoginAttemptCommand,
) (*it.UpdateLoginAttemptResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.LoginAttempt, *models.LoginAttempt]{
		Action:       "update login attempt",
		DbRepoGetter: this.attemptRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, attempt *models.LoginAttempt, foundAttempt *models.LoginAttempt, cErrsTotal *ft.ClientErrors) error {
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

func (this *AttemptDomainServiceImpl) GetAttempt(ctx corectx.Context, query it.GetAttemptQuery) (*it.GetAttemptResult, error) {
	return corecrud.GetOne[models.LoginAttempt](ctx, corecrud.GetOneParam{
		Action:       "get login attempt",
		DbRepoGetter: this.attemptRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *AttemptDomainServiceImpl) assertNewStatusValid(
	dbAttempt *models.LoginAttempt, newStatus *models.AttemptStatus, clientErrs *ft.ClientErrors,
) {
	if newStatus == nil {
		return
	}
	st := dbAttempt.GetStatus()
	if st != nil && *st != models.AttemptStatusPending {
		clientErrs.Append(*ft.NewValidationError(
			"status", ft.ErrorKey("err_attempt_already_settled", "iam"), "attempt already settled",
		))
	}
}

func (this *AttemptDomainServiceImpl) assertNewMethodValid(
	dbAttempt *models.LoginAttempt, newMethod *string, clientErrs *ft.ClientErrors,
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
			"current_method", ft.ErrorKey("err_not_applicable_login_method", "iam"), "not applicable login method",
		))
	}
}

type attemptPrincipal struct {
	Id       model.Id
	Name     string
	Username string
}

func (this *AttemptDomainServiceImpl) assertPrincipalExists(
	ctx corectx.Context, attempt *models.LoginAttempt, clientErrs *ft.ClientErrors,
) (subject *attemptPrincipal, err error) {
	switch attempt.MustGetPrincipalType() {
	case models.PrincipalTypeNikkiUser:
		subject, err = this.assertUserExists(ctx, attempt.MustGetUsername(), clientErrs)
	}
	// case models.SubjectTypeCustomer:
	// 	subject, err = this.assertCustomerExists(ctx, username, vErrs)
	// }
	if err != nil {
		return nil, err
	}
	return subject, nil
}

func (this *AttemptDomainServiceImpl) assertUserExists(
	ctx corectx.Context, username string, cErrs *ft.ClientErrors,
) (*attemptPrincipal, error) {
	query := itUser.GetUserQuery{
		Email: &username,
	}
	userRes, err := this.userSvc.GetUser(ctx, query)
	if err != nil {
		return nil, err
	}
	if userRes.ClientErrors.Count() > 0 {
		cErrs.Append(userRes.ClientErrors...)
		return nil, nil
	}
	if !userRes.HasData {
		cErrs.Append(*ft.NewNotFoundError(models.AttemptFieldUsername))
		return nil, nil
	}
	user := userRes.Data
	return &attemptPrincipal{
		Id:       *user.GetId(),
		Name:     *user.GetDisplayName(),
		Username: *user.GetEmail(),
	}, nil
}
