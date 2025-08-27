package app

import (
	"context"
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/event"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewRoleSuiteServiceImpl(roleSuiteRepo it.RoleSuiteRepository, roleRepo itRole.RoleRepository, eventBus event.EventBus) it.RoleSuiteService {
	return &RoleSuiteServiceImpl{
		roleSuiteRepo: roleSuiteRepo,
		roleRepo:      roleRepo,
		eventBus:      eventBus,
	}
}

type RoleSuiteServiceImpl struct {
	roleSuiteRepo it.RoleSuiteRepository
	roleRepo      itRole.RoleRepository
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
			this.validateRoles(ctx, cmd.RoleIds, vErrs)
			return nil
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateRoleSuiteResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	roleSuite, err = this.roleSuiteRepo.Create(ctx, *roleSuite, cmd.RoleIds)
	fault.PanicOnErr(err)

	return &it.CreateRoleSuiteResult{
		Data:    roleSuite,
		HasData: roleSuite != nil,
	}, nil
}

func (this *RoleSuiteServiceImpl) UpdateRoleSuite(ctx context.Context, cmd it.UpdateRoleSuiteCommand) (update *it.UpdateRoleSuiteResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "update role suite"); e != nil {
			err = e
		}
	}()

	roleSuite := cmd.ToRoleSuite()
	var dbRoleSuite *domain.RoleSuite

	var addRoleIds, removeRoleIds []model.Id

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = roleSuite.Validate(true)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbRoleSuite, err = this.assertRoleSuiteExistsById(ctx, cmd.Id, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertCorrectEtag(*roleSuite.Etag, *dbRoleSuite.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.sanitizeRoleSuite(roleSuite)
			return this.assertRoleSuiteUniqueForUpdate(ctx, roleSuite, vErrs)
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.validateRoles(ctx, cmd.RoleIds, vErrs)
			addRoleIds, removeRoleIds = this.diffRoleIds(this.getRoleIdsByDomain(dbRoleSuite), cmd.RoleIds)
			return nil
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateRoleSuiteResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := roleSuite.Etag
	roleSuite.Etag = model.NewEtag()
	roleSuite, err = this.roleSuiteRepo.UpdateTx(ctx, *roleSuite, *prevEtag, addRoleIds, removeRoleIds)
	fault.PanicOnErr(err)

	return &it.UpdateRoleSuiteResult{
		Data:    roleSuite,
		HasData: roleSuite != nil,
	}, err
}

func (this *RoleSuiteServiceImpl) DeleteHardRoleSuite(ctx context.Context, cmd it.DeleteRoleSuiteCommand) (result *it.DeleteRoleSuiteResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "delete role suite"); e != nil {
			err = e
		}
	}()

	var dbRoleSuite *domain.RoleSuite

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbRoleSuite, err = this.assertRoleSuiteExistsById(ctx, cmd.Id, vErrs)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.DeleteRoleSuiteResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	deletedCount, err := this.roleSuiteRepo.DeleteHardTx(ctx, it.DeleteRoleSuiteParam{
		Id:   cmd.Id,
		Name: *dbRoleSuite.Name,
	})
	fault.PanicOnErr(err)

	if deletedCount == 0 {
		vErrs.AppendNotFound("role_suite_id", "role suite")
		return &it.DeleteRoleSuiteResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount), nil
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
	fault.PanicOnErr(err)

	if dbRoleSuite == nil {
		vErrs.AppendNotFound("role_suite_id", "role suite")
	}
	return
}

func (this *RoleSuiteServiceImpl) assertRoleSuiteUnique(ctx context.Context, roleSuite *domain.RoleSuite, vErrs *fault.ValidationErrors) error {
	dbRoleSuite, err := this.roleSuiteRepo.FindByName(ctx, it.FindByNameParam{Name: *roleSuite.Name})
	fault.PanicOnErr(err)

	if dbRoleSuite != nil {
		vErrs.AppendAlreadyExists("role_suite_name", "role suite name")
	}

	return nil
}

func (this *RoleSuiteServiceImpl) assertRoleSuiteUniqueForUpdate(ctx context.Context, roleSuite *domain.RoleSuite, vErrs *fault.ValidationErrors) error {
	dbRoleSuite, err := this.roleSuiteRepo.FindByName(ctx, it.FindByNameParam{Name: *roleSuite.Name})
	fault.PanicOnErr(err)

	if dbRoleSuite != nil && *dbRoleSuite.Id != *roleSuite.Id {
		vErrs.AppendAlreadyExists("role_suite_name", "role suite name")
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

func (this *RoleSuiteServiceImpl) validateRoles(ctx context.Context, roleIds []model.Id, vErrs *fault.ValidationErrors) {
	if len(roleIds) == 0 {
		return
	}

	seenIds := make(map[model.Id]int)

	for i, roleId := range roleIds {
		if firstIndex, exists := seenIds[roleId]; exists {
			vErrs.Append(fmt.Sprintf("roles[%d]", i), fmt.Sprintf("duplicate role id found at index %d", firstIndex))
			continue
		}

		seenIds[roleId] = i
	}

	if vErrs.Count() > 0 {
		return
	}

	for i, roleId := range roleIds {
		existingRole, err := this.roleRepo.Exist(ctx, itRole.ExistRoleParam{Id: roleId})
		fault.PanicOnErr(err)

		if !existingRole {
			vErrs.Append(fmt.Sprintf("roles[%d]", i), fmt.Sprintf("role with id '%s' does not exist", roleId))
		}
	}
}

func (this *RoleSuiteServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *fault.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *RoleSuiteServiceImpl) getRoleIdsByDomain(roleSuite *domain.RoleSuite) []model.Id {
	roleIds := make([]model.Id, len(roleSuite.Roles))
	for i, role := range roleSuite.Roles {
		roleIds[i] = *role.Id
	}

	return roleIds
}

func (this *RoleSuiteServiceImpl) diffRoleIds(oldIds, newIds []model.Id) (added, removed []model.Id) {
	oldMap := make(map[model.Id]bool)
	newMap := make(map[model.Id]bool)

	for _, id := range oldIds {
		oldMap[id] = true
	}
	for _, id := range newIds {
		newMap[id] = true
	}

	for _, id := range newIds {
		if !oldMap[id] {
			added = append(added, id)
		}
	}
	for _, id := range oldIds {
		if !newMap[id] {
			removed = append(removed, id)
		}
	}

	return
}
