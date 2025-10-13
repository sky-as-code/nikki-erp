package app

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
	itRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/revoke_request"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewRoleSuiteServiceImpl(
	cqrsBus cqrs.CqrsBus,
	roleSuiteRepo it.RoleSuiteRepository,
	roleService itRole.RoleService,
	grantRequestService itGrantRequest.GrantRequestService,
	revokeRequestService itRevokeRequest.RevokeRequestService,
) it.RoleSuiteService {
	return &RoleSuiteServiceImpl{
		cqrsBus:              cqrsBus,
		grantRequestService:  grantRequestService,
		revokeRequestService: revokeRequestService,
		roleSuiteRepo:        roleSuiteRepo,
		roleService:          roleService,
	}
}

type RoleSuiteServiceImpl struct {
	cqrsBus              cqrs.CqrsBus
	grantRequestService  itGrantRequest.GrantRequestService
	revokeRequestService itRevokeRequest.RevokeRequestService
	roleSuiteRepo        it.RoleSuiteRepository
	roleService          itRole.RoleService
}

func (this *RoleSuiteServiceImpl) CreateRoleSuite(ctx crud.Context, cmd it.CreateRoleSuiteCommand) (result *it.CreateRoleSuiteResult, err error) {
	result, err = crud.Create(ctx, crud.CreateParam[*domain.RoleSuite, it.CreateRoleSuiteCommand, it.CreateRoleSuiteResult]{
		Action:  "create suite",
		Command: cmd,
		AssertBusinessRules: func(ctx crud.Context, roleSuite *domain.RoleSuite, vErrs *fault.ValidationErrors) error {
			err := this.assertOrgExists(ctx, roleSuite, vErrs)
			fault.PanicOnErr(err)

			err = this.assertRoleSuiteUnique(ctx, roleSuite, vErrs)
			fault.PanicOnErr(err)

			this.validateRoles(ctx, cmd.RoleIds, roleSuite.OrgId, vErrs)
			return nil
		},
		RepoCreate: func(ctx crud.Context, model *domain.RoleSuite) (*domain.RoleSuite, error) {
			return this.roleSuiteRepo.Create(ctx, *model, cmd.RoleIds)
		},
		SetDefault: this.setRoleSuiteDefaults,
		Sanitize:   this.sanitizeRoleSuite,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.CreateRoleSuiteResult {
			return &it.CreateRoleSuiteResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.RoleSuite) *it.CreateRoleSuiteResult {
			return &it.CreateRoleSuiteResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *RoleSuiteServiceImpl) UpdateRoleSuite(ctx crud.Context, cmd it.UpdateRoleSuiteCommand) (result *it.UpdateRoleSuiteResult, err error) {
	var addRoleIds, removeRoleIds []model.Id

	result, err = crud.Update(ctx, crud.UpdateParam[*domain.RoleSuite, it.UpdateRoleSuiteCommand, it.UpdateRoleSuiteResult]{
		Action:       "update suite",
		Command:      cmd,
		AssertExists: this.assertRoleSuiteExistsById,
		AssertBusinessRules: func(ctx crud.Context, roleSuite *domain.RoleSuite, modelFromDb *domain.RoleSuite, vErrs *fault.ValidationErrors) error {
			err := this.assertOrgExists(ctx, roleSuite, vErrs)
			fault.PanicOnErr(err)

			err = this.assertRoleSuiteUniqueForUpdate(ctx, roleSuite, modelFromDb.OrgId, vErrs)
			fault.PanicOnErr(err)

			this.validateRoles(ctx, cmd.RoleIds, modelFromDb.OrgId, vErrs)

			oldRoleIds := this.getRoleIdsByDomain(modelFromDb)
			addRoleIds, removeRoleIds = this.diffRoleIds(oldRoleIds, cmd.RoleIds)

			return nil
		},
		RepoUpdate: func(ctx crud.Context, roleSuite *domain.RoleSuite, prevEtag model.Etag) (*domain.RoleSuite, error) {
			return this.roleSuiteRepo.Update(ctx, *roleSuite, prevEtag, addRoleIds, removeRoleIds)
		},
		Sanitize: this.sanitizeRoleSuite,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.UpdateRoleSuiteResult {
			return &it.UpdateRoleSuiteResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.RoleSuite) *it.UpdateRoleSuiteResult {
			return &it.UpdateRoleSuiteResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *RoleSuiteServiceImpl) DeleteHardRoleSuite(ctx crud.Context, cmd it.DeleteRoleSuiteCommand) (result *it.DeleteRoleSuiteResult, err error) {
	tx, err := this.roleSuiteRepo.BeginTransaction(ctx)
	fault.PanicOnErr(err)

	ctx.SetDbTranx(tx)

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if result != nil && result.ClientError != nil {
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	result, err = crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.RoleSuite, it.DeleteRoleSuiteCommand, it.DeleteRoleSuiteResult]{
		Action:              "delete suite",
		Command:             cmd,
		AssertExists:        this.assertRoleSuiteExistsById,
		AssertBusinessRules: this.assertBusinessRuleDeleteRoleSuite,
		RepoDelete: func(ctx crud.Context, model *domain.RoleSuite) (int, error) {
			return this.roleSuiteRepo.DeleteHard(ctx, it.DeleteRoleSuiteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.DeleteRoleSuiteResult {
			return &it.DeleteRoleSuiteResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.RoleSuite, deletedCount int) *it.DeleteRoleSuiteResult {
			return crud.NewSuccessDeletionResult(*model.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *RoleSuiteServiceImpl) GetRoleSuiteById(ctx crud.Context, query it.GetRoleSuiteByIdQuery) (result *it.GetRoleSuiteByIdResult, err error) {
	result, err = crud.GetOne(ctx, crud.GetOneParam[*domain.RoleSuite, it.GetRoleSuiteByIdQuery, it.GetRoleSuiteByIdResult]{
		Action:      "get suite by Id",
		Query:       query,
		RepoFindOne: this.getRoleSuiteByIdFull,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.GetRoleSuiteByIdResult {
			return &it.GetRoleSuiteByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.RoleSuite) *it.GetRoleSuiteByIdResult {
			return &it.GetRoleSuiteByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *RoleSuiteServiceImpl) SearchRoleSuites(ctx crud.Context, query it.SearchRoleSuitesCommand) (result *it.SearchRoleSuitesResult, err error) {
	result, err = crud.Search(ctx, crud.SearchParam[domain.RoleSuite, it.SearchRoleSuitesCommand, it.SearchRoleSuitesResult]{
		Action: "search suites",
		Query:  query,
		SetQueryDefaults: func(query *it.SearchRoleSuitesCommand) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.roleSuiteRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query it.SearchRoleSuitesCommand, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.RoleSuite], error) {
			return this.roleSuiteRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.SearchRoleSuitesResult {
			return &it.SearchRoleSuitesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.RoleSuite]) *it.SearchRoleSuitesResult {
			return &it.SearchRoleSuitesResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *RoleSuiteServiceImpl) GetRoleSuitesBySubject(ctx crud.Context, query it.GetRoleSuitesBySubjectQuery) (result *it.GetRoleSuitesBySubjectResult, err error) {
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

func (this *RoleSuiteServiceImpl) assertRoleSuiteExistsById(ctx crud.Context, roleSuite *domain.RoleSuite, vErrs *fault.ValidationErrors) (dbRoleSuite *domain.RoleSuite, err error) {
	dbRoleSuite, err = this.roleSuiteRepo.FindById(ctx, it.FindByIdParam{Id: *roleSuite.Id})
	fault.PanicOnErr(err)

	if dbRoleSuite == nil {
		vErrs.AppendNotFound("role_suite_id", "role suite")
	}
	return
}

func (this *RoleSuiteServiceImpl) getRoleSuiteByIdFull(ctx crud.Context, query it.GetRoleSuiteByIdQuery, vErrs *fault.ValidationErrors) (dbRoleSuite *domain.RoleSuite, err error) {
	dbRoleSuite, err = this.roleSuiteRepo.FindById(ctx, it.FindByIdParam{Id: query.Id})
	fault.PanicOnErr(err)

	if dbRoleSuite == nil {
		vErrs.AppendNotFound("role_suite_id", "role suite")
		return
	}

	return
}

func (this *RoleSuiteServiceImpl) assertRoleSuiteUnique(ctx crud.Context, roleSuite *domain.RoleSuite, vErrs *fault.ValidationErrors) error {
	dbRoleSuite, err := this.roleSuiteRepo.FindByName(
		ctx,
		it.FindByNameParam{
			Name:  *roleSuite.Name,
			OrgId: roleSuite.OrgId,
		},
	)
	fault.PanicOnErr(err)

	if dbRoleSuite != nil {
		vErrs.AppendAlreadyExists("role_suite_name", "role suite name")
	}

	return nil
}

func (this *RoleSuiteServiceImpl) assertRoleSuiteUniqueForUpdate(ctx crud.Context, roleSuite *domain.RoleSuite, orgId *model.Id, vErrs *fault.ValidationErrors) error {
	if roleSuite.Name == nil {
		return nil
	}

	dbRoleSuite, err := this.roleSuiteRepo.FindByName(
		ctx,
		it.FindByNameParam{
			Name:  *roleSuite.Name,
			OrgId: orgId,
		},
	)
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

func (this *RoleSuiteServiceImpl) validateRoles(ctx crud.Context, roleIds []model.Id, suiteOrgId *model.Id, vErrs *fault.ValidationErrors) {
	if len(roleIds) == 0 {
		return
	}

	seenIds := make(map[model.Id]int)
	for i, roleId := range roleIds {
		if _, exists := seenIds[roleId]; exists {
			vErrs.AppendNotAllow("role_id", roleId)
			continue
		}
		seenIds[roleId] = i
	}

	if vErrs.Count() > 0 {
		return
	}

	for _, roleId := range roleIds {
		role, err := this.roleService.GetRoleById(ctx, itRole.GetRoleByIdQuery{Id: roleId})
		fault.PanicOnErr(err)

		if role.ClientError != nil {
			vErrs.MergeClientError(role.ClientError)
			continue
		}

		if role == nil {
			vErrs.AppendNotFound("role_id", roleId)
			continue
		}

		if role.Data.OrgId != nil {
			if suiteOrgId == nil {
				vErrs.AppendNotAllow("role_id", roleId)
			} else if *role.Data.OrgId != *suiteOrgId {
				vErrs.AppendNotAllow("role_id", roleId)
			}
		}
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

func (this *RoleSuiteServiceImpl) assertOrgExists(ctx crud.Context, roleSuite *domain.RoleSuite, vErrs *fault.ValidationErrors) error {
	if roleSuite.OrgId == nil {
		return nil
	}

	existCmd := &itOrg.ExistsOrgByIdCommand{
		Id: *roleSuite.OrgId,
	}
	existRes := itOrg.ExistsOrgByIdResult{}
	err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
	fault.PanicOnErr(err)

	if existRes.ClientError != nil {
		vErrs.MergeClientError(existRes.ClientError)
		return nil
	}

	if !existRes.Data {
		vErrs.Append("orgId", "not existing organization")
	}
	return nil
}

func (this *RoleSuiteServiceImpl) assertBusinessRuleDeleteRoleSuite(ctx crud.Context, cmd it.DeleteRoleSuiteCommand, roleSuiteDB *domain.RoleSuite, vErrs *fault.ValidationErrors) error {
	updateGrantRequest, err := this.grantRequestService.TargetIsDeleted(
		ctx,
		itGrantRequest.TargetIsDeletedCommand{
			TargetType: domain.GrantRequestTargetTypeSuite,
			TargetRef:  *roleSuiteDB.Id,
			TargetName: *roleSuiteDB.Name,
		},
	)
	fault.PanicOnErr(err)

	if updateGrantRequest.ClientError != nil {
		vErrs.MergeClientError(updateGrantRequest.ClientError)
		return nil
	}
	if !updateGrantRequest.Data {
		vErrs.Append("role_suite_id", "can not delete role suite with grant requests")
		return nil
	}

	updateRevokeRequest, err := this.revokeRequestService.TargetIsDeleted(
		ctx,
		itRevokeRequest.TargetIsDeletedCommand{
			TargetType: domain.GrantRequestTargetTypeSuite,
			TargetRef:  *roleSuiteDB.Id,
			TargetName: *roleSuiteDB.Name,
		},
	)
	fault.PanicOnErr(err)

	if updateRevokeRequest.ClientError != nil {
		vErrs.MergeClientError(updateRevokeRequest.ClientError)
		return nil
	}
	if !updateRevokeRequest.Data {
		vErrs.Append("role_suite_id", "can not delete role suite with revoke requests")
		return nil
	}

	return nil
}
