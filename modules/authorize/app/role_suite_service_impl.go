package app

import (
	"context"
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/event"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewRoleSuiteServiceImpl(roleSuiteRepo it.RoleSuiteRepository, eventBus event.EventBus) it.RoleSuiteService {
	return &RoleSuiteServiceImpl{
		roleSuiteRepo: roleSuiteRepo,
		eventBus:      eventBus,
	}
}

type RoleSuiteServiceImpl struct {
	roleSuiteRepo it.RoleSuiteRepository
	eventBus      event.EventBus
}

func (this *RoleSuiteServiceImpl) CreateRoleSuite(ctx context.Context, cmd it.CreateRoleSuiteCommand) (result *it.CreateRoleSuiteResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "create role suite"); e != nil {
			err = e
		}
	}()

	roleSuite := cmd.ToRoleSuite()
	this.setRoleSuiteDefaults(roleSuite)

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = roleSuite.Validate(false)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.sanitizeRoleSuite(roleSuite)
			return this.assertRoleSuiteUnique(ctx, roleSuite, vErrs)
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			if len(roleSuite.Roles) > 0 {
				this.validateRoles(roleSuite.Roles, vErrs)
			}
			return nil
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateRoleSuiteResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	roleSuite, err = this.roleSuiteRepo.Create(ctx, *roleSuite)
	fault.PanicOnErr(err)

	return &it.CreateRoleSuiteResult{
		Data:    roleSuite,
		HasData: roleSuite != nil,
	}, nil
}

func (this *RoleSuiteServiceImpl) GetRoleSuiteById(ctx context.Context, query it.GetRoleSuiteByIdQuery) (result *it.GetRoleSuiteByIdResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get role suite by id"); e != nil {
			err = e
		}
	}()

	var dbRoleSuite *domain.RoleSuite
	vErrs := fault.NewValidationErrors()
	dbRoleSuite, err = this.assertRoleSuiteExistsById(ctx, query.Id, &vErrs)
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetRoleSuiteByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetRoleSuiteByIdResult{
		Data:    dbRoleSuite,
		HasData: dbRoleSuite != nil,
	}, nil
}

func (this *RoleSuiteServiceImpl) SearchRoleSuites(ctx context.Context, query it.SearchRoleSuitesCommand) (result *it.SearchRoleSuitesResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "search role suites"); e != nil {
			err = e
		}
	}()

	query.SetDefaults()
	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.roleSuiteRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchRoleSuitesResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	roleSuites, err := this.roleSuiteRepo.Search(ctx, it.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
	})
	fault.PanicOnErr(err)

	return &it.SearchRoleSuitesResult{
		Data:    roleSuites,
		HasData: roleSuites.Items != nil,
	}, nil
}

func (this *RoleSuiteServiceImpl) GetRoleSuitesBySubject(ctx context.Context, query it.GetRoleSuitesBySubjectQuery) (result *it.GetRoleSuitesBySubjectResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get role suites by subject"); e != nil {
			err = e
		}
	}()

	var roleSuites []domain.RoleSuite
	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			roleSuites, err = this.roleSuiteRepo.FindAllBySubject(ctx, query)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetRoleSuitesBySubjectResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetRoleSuitesBySubjectResult{
		Data:    roleSuites,
		HasData: roleSuites != nil,
	}, nil
}

func (this *RoleSuiteServiceImpl) assertRoleSuiteExistsById(ctx context.Context, id model.Id, vErrs *fault.ValidationErrors) (dbRoleSuite *domain.RoleSuite, err error) {
	dbRoleSuite, err = this.roleSuiteRepo.FindById(ctx, it.FindByIdParam{Id: id})
	if dbRoleSuite == nil {
		vErrs.AppendNotFound("id", "roleSuite")
	}
	return
}

func (this *RoleSuiteServiceImpl) assertRoleSuiteUnique(ctx context.Context, roleSuite *domain.RoleSuite, vErrs *fault.ValidationErrors) error {
	dbRoleSuite, err := this.roleSuiteRepo.FindByName(ctx, it.FindByNameParam{Name: *roleSuite.Name})
	fault.PanicOnErr(err)

	if dbRoleSuite != nil {
		vErrs.AppendAlreadyExists("name", "role suite name")
	}

	return nil
}

func (this *RoleSuiteServiceImpl) sanitizeRoleSuite(roleSuite *domain.RoleSuite) {
	if roleSuite.Description != nil {
		roleSuite.Description = util.ToPtr(defense.SanitizePlainText(*roleSuite.Description, true))
	}
}

func (this *RoleSuiteServiceImpl) setRoleSuiteDefaults(roleSuite *domain.RoleSuite) {
	roleSuite.SetDefaults()
}

func (this *RoleSuiteServiceImpl) validateRoles(roles []domain.Role, vErrs *fault.ValidationErrors) {
	seenIds := make(map[model.Id]int)

	for i, role := range roles {
		if role.Id == nil {
			vErrs.Append(fmt.Sprintf("roles[%d]", i), "role id cannot be nil")
			continue
		}

		if firstIndex, exists := seenIds[*role.Id]; exists {
			vErrs.Append(fmt.Sprintf("roles[%d]", i), fmt.Sprintf("duplicate role id found at index %d", firstIndex))
			continue
		}

		seenIds[*role.Id] = i
	}

	if vErrs.Count() > 0 {
		return
	}
}
