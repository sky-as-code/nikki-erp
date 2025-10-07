package app

import (
	"fmt"

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
	itEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	itAssign "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
	itRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/revoke_request"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

func NewRoleServiceImpl(
	assignmentService itAssign.EntitlementAssignmentService,
	cqrsBus cqrs.CqrsBus,
	entitlementRepo itEntitlement.EntitlementRepository,
	grantRequestService itGrantRequest.GrantRequestService,
	revokeRequestService itRevokeRequest.RevokeRequestService,
	roleRepo itRole.RoleRepository,
) itRole.RoleService {
	return &RoleServiceImpl{
		assignmentService:    assignmentService,
		cqrsBus:              cqrsBus,
		entitlementRepo:      entitlementRepo,
		grantRequestService:  grantRequestService,
		revokeRequestService: revokeRequestService,
		roleRepo:             roleRepo,
	}
}

type RoleServiceImpl struct {
	assignmentService    itAssign.EntitlementAssignmentService
	cqrsBus              cqrs.CqrsBus
	entitlementRepo      itEntitlement.EntitlementRepository
	grantRequestService  itGrantRequest.GrantRequestService
	revokeRequestService itRevokeRequest.RevokeRequestService
	roleRepo             itRole.RoleRepository
}

func (this *RoleServiceImpl) CreateRole(ctx crud.Context, cmd itRole.CreateRoleCommand) (*itRole.CreateRoleResult, error) {
	return crud.Create(ctx, crud.CreateParam[*domain.Role, itRole.CreateRoleCommand, itRole.CreateRoleResult]{
		Action:              "create role",
		Command:             cmd,
		AssertBusinessRules: this.assertBusinessRuleCreateRole,
		RepoCreate:          this.roleRepo.Create,
		SetDefault:          this.setRoleDefaults,
		Sanitize:            this.sanitizeRole,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.CreateRoleResult {
			return &itRole.CreateRoleResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Role) *itRole.CreateRoleResult {
			return &itRole.CreateRoleResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})
}

func (this *RoleServiceImpl) UpdateRole(ctx crud.Context, cmd itRole.UpdateRoleCommand) (*itRole.UpdateRoleResult, error) {
	return crud.Update(ctx, crud.UpdateParam[*domain.Role, itRole.UpdateRoleCommand, itRole.UpdateRoleResult]{
		Action:              "update role",
		Command:             cmd,
		AssertExists:        this.assertRoleExistsById,
		AssertBusinessRules: this.assertBusinessRuleUpdateRole,
		RepoUpdate:          this.roleRepo.Update,
		Sanitize:            this.sanitizeRole,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.UpdateRoleResult {
			return &itRole.UpdateRoleResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Role) *itRole.UpdateRoleResult {
			return &itRole.UpdateRoleResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})
}

func (this *RoleServiceImpl) DeleteRoleHard(ctx crud.Context, cmd itRole.DeleteRoleHardCommand) (result *itRole.DeleteRoleHardResult, err error) {
	tx, err := this.entitlementRepo.BeginTransaction(ctx)
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

	result, err = crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Role, itRole.DeleteRoleHardCommand, itRole.DeleteRoleHardResult]{
		Action:              "delete Role",
		Command:             cmd,
		AssertExists:        this.assertRoleExistsById,
		AssertBusinessRules: this.assertBusinessRuleDeleteRole,
		RepoDelete: func(ctx crud.Context, model *domain.Role) (int, error) {
			return this.roleRepo.DeleteHard(ctx, itRole.DeleteRoleHardCommand{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.DeleteRoleHardResult {
			return &itRole.DeleteRoleHardResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Role, deletedCount int) *itRole.DeleteRoleHardResult {
			return crud.NewSuccessDeletionResult(*model.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *RoleServiceImpl) GetRoleById(ctx crud.Context, query itRole.GetRoleByIdQuery) (result *itRole.GetRoleByIdResult, err error) {
	return crud.GetOne(ctx, crud.GetOneParam[*domain.Role, itRole.GetRoleByIdQuery, itRole.GetRoleByIdResult]{
		Action:      "get role by Id",
		Query:       query,
		RepoFindOne: this.getRoleByIdFull,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.GetRoleByIdResult {
			return &itRole.GetRoleByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Role) *itRole.GetRoleByIdResult {
			return &itRole.GetRoleByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})
}

func (this *RoleServiceImpl) SearchRoles(ctx crud.Context, query itRole.SearchRolesQuery) (result *itRole.SearchRolesResult, err error) {
	result, err = crud.Search(ctx, crud.SearchParam[domain.Role, itRole.SearchRolesQuery, itRole.SearchRolesResult]{
		Action: "search roles",
		Query:  query,
		SetQueryDefaults: func(query *itRole.SearchRolesQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.roleRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itRole.SearchRolesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Role], error) {
			result, err := this.roleRepo.Search(ctx, itRole.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
			fault.PanicOnErr(err)

			err = this.populateRoleDetails(ctx, result.Items)
			fault.PanicOnErr(err)

			return result, err
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.SearchRolesResult {
			return &itRole.SearchRolesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.Role]) *itRole.SearchRolesResult {
			return &itRole.SearchRolesResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *RoleServiceImpl) GetRolesBySubject(ctx crud.Context, query itRole.GetRolesBySubjectQuery) (result *itRole.GetRolesBySubjectResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get role by subject"); e != nil {
			err = e
		}
	}()

	var roles []domain.Role

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			roles, err = this.roleRepo.FindAllBySubject(ctx, query)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itRole.GetRolesBySubjectResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &itRole.GetRolesBySubjectResult{
		Data:    roles,
		HasData: roles != nil,
	}, nil
}

func (this *RoleServiceImpl) assertRoleExistsById(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) (dbRole *domain.Role, err error) {
	dbRole, err = this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: *role.Id})
	if dbRole == nil {
		vErrs.AppendNotFound("role_id", "role")
	}

	return
}

func (this *RoleServiceImpl) assertRoleUnique(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	dbRole, err := this.roleRepo.FindByName(
		ctx,
		itRole.FindByNameParam{
			Name:  *role.Name,
			OrgId: role.OrgId,
		})
	fault.PanicOnErr(err)

	if dbRole != nil {
		vErrs.AppendAlreadyExists("role_name", "role name")
	}

	return nil
}

func (this *RoleServiceImpl) assertRoleNameUniqueForUpdate(ctx crud.Context, role *domain.Role, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
	dbRole, err := this.roleRepo.FindByName(
		ctx,
		itRole.FindByNameParam{
			Name:  *role.Name,
			OrgId: dbRole.OrgId,
		},
	)
	fault.PanicOnErr(err)

	if dbRole != nil && *dbRole.Id != *role.Id {
		vErrs.AppendAlreadyExists("role_name", "role name")
	}

	return nil
}

func (this *RoleServiceImpl) sanitizeRole(role *domain.Role) {
	if role.Name != nil {
		*role.Name = defense.SanitizePlainText(*role.Name, true)
	} else {
		role.Name = util.ToPtr("")
	}

	if role.Description != nil {
		role.Description = util.ToPtr(defense.SanitizePlainText(*role.Description, true))
	}
}

func (this *RoleServiceImpl) setRoleDefaults(role *domain.Role) {
	role.SetDefaults()
}

func (this *RoleServiceImpl) getAssignmentsByRoleId(ctx crud.Context, role *domain.Role) ([]domain.EntitlementAssignment, error) {
	assignmentsRes, err := this.assignmentService.FindAllBySubject(
		ctx,
		itAssign.GetAllEntitlementAssignmentBySubjectQuery{
			SubjectType: *domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiRole.String()),
			SubjectRef:  *role.Id,
		})
	fault.PanicOnErr(err)

	return assignmentsRes.Data, nil
}

func (this *RoleServiceImpl) getEntitlementIdsByRoleId(ctx crud.Context, role *domain.Role) ([]model.Id, error) {
	assignmentsRes, err := this.getAssignmentsByRoleId(ctx, role)

	if len(assignmentsRes) == 0 {
		return nil, err
	}

	// Extract unique assignments IDs
	entitlementIdSet := make(map[model.Id]bool)
	uniqueEntitlementIds := make([]model.Id, 0)

	for _, assignment := range assignmentsRes {
		if assignment.EntitlementId != nil {
			entId := *assignment.EntitlementId
			if !entitlementIdSet[entId] {
				entitlementIdSet[entId] = true
				uniqueEntitlementIds = append(uniqueEntitlementIds, entId)
			}
		}
	}

	return uniqueEntitlementIds, nil
}

func (this *RoleServiceImpl) getAssignmentIdsByRoleId(ctx crud.Context, role *domain.Role) ([]model.Id, error) {
	assignmentsRes, err := this.getAssignmentsByRoleId(ctx, role)

	if len(assignmentsRes) == 0 {
		return nil, err
	}

	// Extract unique assignments IDs
	assignmentIdSet := make(map[model.Id]bool)
	uniqueAssignmentIds := make([]model.Id, 0)

	for _, assignment := range assignmentsRes {
		if assignment.Id != nil {
			entId := *assignment.Id
			if !assignmentIdSet[entId] {
				assignmentIdSet[entId] = true
				uniqueAssignmentIds = append(uniqueAssignmentIds, entId)
			}
		}
	}

	return uniqueAssignmentIds, nil
}

func (this *RoleServiceImpl) getEntitlements(ctx crud.Context, entitlementIds []model.Id) ([]domain.Entitlement, error) {
	entitlementsRes, err := this.entitlementRepo.FindAllByIds(ctx, itEntitlement.GetAllEntitlementByIdsQuery{Ids: entitlementIds})
	fault.PanicOnErr(err)

	return entitlementsRes, nil
}

func (this *RoleServiceImpl) deleteAssignments(ctx crud.Context, assignmentIds []model.Id) error {
	for _, assignmentId := range assignmentIds {
		deletedCount, err := this.assignmentService.DeleteHardAssignment(ctx, itAssign.DeleteEntitlementAssignmentByIdCommand{Id: assignmentId})
		fault.PanicOnErr(err)

		if deletedCount.ClientError != nil {
			return fmt.Errorf("failed to delete assignment with ID %s", assignmentId)
		}
	}

	return nil
}

func (this *RoleServiceImpl) assertBusinessRuleCreateRole(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	err := this.assertOrgExists(ctx, role, vErrs)
	fault.PanicOnErr(err)

	err = this.assertRoleUnique(ctx, role, vErrs)
	fault.PanicOnErr(err)

	return nil
}

func (this *RoleServiceImpl) assertOrgExists(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	if role.OrgId == nil {
		return nil
	}

	existCmd := &itOrg.ExistsOrgByIdCommand{
		Id: *role.OrgId,
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

func (this *RoleServiceImpl) assertBusinessRuleUpdateRole(ctx crud.Context, role *domain.Role, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
	err := this.assertRoleNameUniqueForUpdate(ctx, role, dbRole, vErrs)
	fault.PanicOnErr(err)

	return nil
}
func (this *RoleServiceImpl) roleIdDeleted(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	updateGrantRequest, err := this.grantRequestService.TargetIsDeleted(
		ctx,
		itGrantRequest.TargetIsDeletedCommand{
			TargetType: domain.GrantRequestTargetTypeRole,
			TargetRef:  *role.Id,
			TargetName: *role.Name,
		},
	)
	fault.PanicOnErr(err)

	if updateGrantRequest.ClientError != nil {
		vErrs.MergeClientError(updateGrantRequest.ClientError)
		return nil
	}
	if !updateGrantRequest.Data {
		vErrs.Append("role_id", "can not delete role with grant requests")
		return nil
	}

	updateRevokeRequest, err := this.revokeRequestService.TargetIsDeleted(
		ctx,
		itRevokeRequest.TargetIsDeletedCommand{
			TargetType: domain.GrantRequestTargetTypeRole,
			TargetRef:  *role.Id,
			TargetName: *role.Name,
		},
	)
	fault.PanicOnErr(err)

	if updateRevokeRequest.ClientError != nil {
		vErrs.MergeClientError(updateRevokeRequest.ClientError)
		return nil
	}
	if !updateRevokeRequest.Data {
		vErrs.Append("role_id", "can not delete role with revoke requests")
		return nil
	}

	return nil
}

func (this *RoleServiceImpl) assertBusinessRuleDeleteRole(ctx crud.Context, cmd itRole.DeleteRoleHardCommand, role *domain.Role, vErrs *fault.ValidationErrors) error {
	var assignmentIds []model.Id

	assignmentIds, err := this.getAssignmentIdsByRoleId(ctx, role)
	fault.PanicOnErr(err)

	if len(assignmentIds) > 0 {
		err = this.deleteAssignments(ctx, assignmentIds)
		fault.PanicOnErr(err)
	}

	err = this.roleIdDeleted(ctx, role, vErrs)
	fault.PanicOnErr(err)

	return nil
}

func (this *RoleServiceImpl) getRoleByIdFull(ctx crud.Context, query itRole.GetRoleByIdQuery, vErrs *fault.ValidationErrors) (dbRole *domain.Role, err error) {
	dbRole, err = this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: query.Id})
	fault.PanicOnErr(err)

	if dbRole == nil {
		vErrs.AppendNotFound("role_id", "role")
	}

	entitlementIds, err := this.getEntitlementIdsByRoleId(ctx, dbRole)
	fault.PanicOnErr(err)

	if len(entitlementIds) > 0 {
		entitlements, err := this.getEntitlements(ctx, entitlementIds)
		fault.PanicOnErr(err)
		dbRole.Entitlements = entitlements
	}

	return
}

func (this *RoleServiceImpl) populateRoleDetails(ctx crud.Context, dbRoles []domain.Role) error {
	for i := range dbRoles {
		entitlementIds, err := this.getEntitlementIdsByRoleId(ctx, &dbRoles[i])
		fault.PanicOnErr(err)

		if len(entitlementIds) > 0 {
			entitlements, err := this.getEntitlements(ctx, entitlementIds)
			fault.PanicOnErr(err)
			dbRoles[i].Entitlements = entitlements
		}
	}

	return nil
}
