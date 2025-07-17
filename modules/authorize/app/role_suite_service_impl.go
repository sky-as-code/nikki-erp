package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
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
		if e := ft.RecoverPanic(recover(), "failed to create role suite"); e != nil {
			err = e
		}
	}()

	roleSuite := cmd.ToRoleSuite()
	this.setRoleSuiteDefaults(ctx, roleSuite)
	roleSuite.SetCreatedAt(time.Now())

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = roleSuite.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeRoleSuite(roleSuite)
			return this.assertRoleSuiteUnique(ctx, roleSuite, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if len(roleSuite.Roles) > 0 {
				this.validateRoles(ctx, roleSuite.Roles, vErrs)
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateRoleSuiteResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	roleSuite, err = this.roleSuiteRepo.Create(ctx, *roleSuite)
	ft.PanicOnErr(err)

	return &it.CreateRoleSuiteResult{
		Data:    roleSuite,
		HasData: roleSuite != nil,
	}, nil
}

func (this *RoleSuiteServiceImpl) GetRoleSuiteById(ctx context.Context, query it.GetRoleSuiteByIdQuery) (result *it.GetRoleSuiteByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get role suite by id"); e != nil {
			err = e
		}
	}()

	var dbRoleSuite *domain.RoleSuite
	vErrs := ft.NewValidationErrors()
	dbRoleSuite, err = this.assertRoleSuiteExistsById(ctx, query.Id, &vErrs)
	ft.PanicOnErr(err)

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
		if e := ft.RecoverPanic(recover(), "failed to list role suites"); e != nil {
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
	ft.PanicOnErr(err)

	return &it.SearchRoleSuitesResult{
		Data:    roleSuites,
		HasData: roleSuites.Items != nil,
	}, nil
}

func (this *RoleSuiteServiceImpl) GetRoleSuitesBySubject(ctx context.Context, query it.GetRoleSuitesBySubjectQuery) (result *it.GetRoleSuitesBySubjectResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get role suites by subject"); e != nil {
			err = e
		}
	}()

	var roleSuites []domain.RoleSuite
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			// *vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			roleSuites, err = this.roleSuiteRepo.FindAllBySubject(ctx, query)
			return err
		}).
		End()
	ft.PanicOnErr(err)

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

func (this *RoleSuiteServiceImpl) assertRoleSuiteExistsById(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (dbRoleSuite *domain.RoleSuite, err error) {
	dbRoleSuite, err = this.roleSuiteRepo.FindById(ctx, it.FindByIdParam{Id: id})
	if dbRoleSuite == nil {
		vErrs.AppendIdNotFound("roleSuite")
	}
	return
}

func (this *RoleSuiteServiceImpl) assertRoleSuiteUnique(ctx context.Context, roleSuite *domain.RoleSuite, vErrs *ft.ValidationErrors) error {
	if vErrs.Has("name") {
		return nil
	}

	dbRoleSuite, err := this.roleSuiteRepo.FindByName(ctx, it.FindByNameParam{Name: *roleSuite.Name})
	ft.PanicOnErr(err)

	if dbRoleSuite != nil {
		vErrs.Append("name", "name already exists")
	}

	return nil
}

func (this *RoleSuiteServiceImpl) sanitizeRoleSuite(roleSuite *domain.RoleSuite) {
	if roleSuite.Description != nil {
		cleanedDescription := strings.TrimSpace(*roleSuite.Description)
		cleanedDescription = defense.SanitizePlainText(cleanedDescription)
		roleSuite.Description = &cleanedDescription
	}
}

func (this *RoleSuiteServiceImpl) setRoleSuiteDefaults(ctx context.Context, roleSuite *domain.RoleSuite) {
	roleSuite.SetDefaults()
}

func (this *RoleSuiteServiceImpl) validateRoles(ctx context.Context, roles []*domain.Role, vErrs *ft.ValidationErrors) {
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