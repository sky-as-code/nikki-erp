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
	entitlementService itEntitlement.EntitlementService,
	grantRequestService itGrantRequest.GrantRequestService,
	revokeRequestService itRevokeRequest.RevokeRequestService,
	roleRepo itRole.RoleRepository,
) itRole.RoleService {
	return &RoleServiceImpl{
		assignmentService:    assignmentService,
		cqrsBus:              cqrsBus,
		entitlementService:   entitlementService,
		grantRequestService:  grantRequestService,
		revokeRequestService: revokeRequestService,
		roleRepo:             roleRepo,
	}
}

type RoleServiceImpl struct {
	assignmentService    itAssign.EntitlementAssignmentService
	cqrsBus              cqrs.CqrsBus
	entitlementService   itEntitlement.EntitlementService
	grantRequestService  itGrantRequest.GrantRequestService
	revokeRequestService itRevokeRequest.RevokeRequestService
	roleRepo             itRole.RoleRepository
}

func (this *RoleServiceImpl) AddEntitlements(ctx crud.Context, cmd itRole.AddEntitlementsCommand) (result *itRole.AddEntitlementsResult, err error) {
	tx, err := this.roleRepo.BeginTransaction(ctx)
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

	result, err = crud.Update(ctx, crud.UpdateParam[*domain.Role, itRole.AddEntitlementsCommand, itRole.AddEntitlementsResult]{
		Action:       "add entitlements",
		Command:      cmd,
		AssertExists: this.assertRoleExistsById,
		AssertBusinessRules: func(ctx crud.Context, role *domain.Role, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
			return this.assertBusinessRuleAddEntitlements(ctx, cmd, dbRole, vErrs)
		},
		RepoUpdate: func(ctx crud.Context, role *domain.Role, prevEtag model.Etag) (*domain.Role, error) {
			return this.roleRepo.Update(ctx, role, prevEtag)
		},
		Sanitize: this.sanitizeRole,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.AddEntitlementsResult {
			return &itRole.AddEntitlementsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Role) *itRole.AddEntitlementsResult {
			return &itRole.AddEntitlementsResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *RoleServiceImpl) RemoveEntitlements(ctx crud.Context, cmd itRole.RemoveEntitlementsCommand) (result *itRole.RemoveEntitlementsResult, err error) {
	tx, err := this.roleRepo.BeginTransaction(ctx)
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

	result, err = crud.Update(ctx, crud.UpdateParam[*domain.Role, itRole.RemoveEntitlementsCommand, itRole.RemoveEntitlementsResult]{
		Action:       "remove entitlements",
		Command:      cmd,
		AssertExists: this.assertRoleExistsById,
		AssertBusinessRules: func(ctx crud.Context, role *domain.Role, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
			return this.assertBusinessRuleRemoveEntitlements(ctx, cmd, dbRole, vErrs)
		},
		RepoUpdate: func(ctx crud.Context, role *domain.Role, prevEtag model.Etag) (*domain.Role, error) {
			return this.roleRepo.Update(ctx, role, prevEtag)
		},
		Sanitize: this.sanitizeRole,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.RemoveEntitlementsResult {
			return &itRole.RemoveEntitlementsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Role) *itRole.RemoveEntitlementsResult {
			return &itRole.RemoveEntitlementsResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
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
	tx, err := this.roleRepo.BeginTransaction(ctx)
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
		Action:              "delete role",
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
	} else {
		if role.Name == nil {
			role.Name = dbRole.Name
		}
	}

	return
}

func (this *RoleServiceImpl) assertRoleUnique(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	if role.Name == nil {
		return nil
	}

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
	if role.Name == nil {
		return nil
	}

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
			SubjectType: domain.EntitlementAssignmentSubjectTypeNikkiRole,
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
	entitlementsRes, err := this.entitlementService.GetAllEntitlementByIds(ctx, itEntitlement.GetAllEntitlementByIdsQuery{Ids: entitlementIds})
	fault.PanicOnErr(err)

	if entitlementsRes.ClientError != nil {
		return nil, entitlementsRes.ClientError
	}

	return entitlementsRes.Data, nil
}

func (this *RoleServiceImpl) deleteAssignments(ctx crud.Context, assignmentIds []model.Id) error {
	for _, assignmentId := range assignmentIds {
		deletedCount, err := this.assignmentService.DeleteHardAssignment(ctx, itAssign.DeleteEntitlementAssignmentByIdCommand{Id: assignmentId})
		fault.PanicOnErr(err)

		if deletedCount.ClientError != nil {
			return deletedCount.ClientError
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

func (this *RoleServiceImpl) roleIsDeleted(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
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

	err = this.roleIsDeleted(ctx, role, vErrs)
	fault.PanicOnErr(err)

	return nil
}

func (this *RoleServiceImpl) getRoleByIdFull(ctx crud.Context, query itRole.GetRoleByIdQuery, vErrs *fault.ValidationErrors) (dbRole *domain.Role, err error) {
	dbRole, err = this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: query.Id})
	fault.PanicOnErr(err)

	if dbRole == nil {
		vErrs.AppendNotFound("role_id", "role")
		return
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

// makeEntitlementKey create composite key
func (this *RoleServiceImpl) makeEntitlementKey(entitlementId model.Id, scopeRef *model.Id) string {
	if scopeRef == nil {
		return string(entitlementId) + ":_"
	}
	return string(entitlementId) + ":" + string(*scopeRef)
}

func (this *RoleServiceImpl) validateUniqueEntitlementInputs(inputs []itRole.EntitlementInput, vErrs *fault.ValidationErrors) ([]itRole.EntitlementInput, bool) {
	seen := make(map[string]bool)
	uniqueInputs := make([]itRole.EntitlementInput, 0, len(inputs))

	for _, input := range inputs {
		key := this.makeEntitlementKey(input.EntitlementId, input.ScopeRef)
		if seen[key] {
			vErrs.AppendNotAllow("entitlementId", input.EntitlementId)
			return nil, false
		}
		seen[key] = true
		uniqueInputs = append(uniqueInputs, input)
	}

	return uniqueInputs, true
}

func (this *RoleServiceImpl) fetchAndMapEntitlements(ctx crud.Context, inputs []itRole.EntitlementInput) (map[model.Id]*domain.Entitlement, error) {
	entitlementIds := make([]model.Id, len(inputs))
	for i, input := range inputs {
		entitlementIds[i] = input.EntitlementId
	}

	entitlements, err := this.getEntitlements(ctx, entitlementIds)
	if err != nil {
		return nil, err
	}

	entitlementMap := make(map[model.Id]*domain.Entitlement)
	for i := range entitlements {
		e := entitlements[i]
		entitlementMap[*e.Id] = &e
	}

	return entitlementMap, nil
}

// validateEntitlementOrgConsistency validates organization-level consistency between entitlements and role
// This prevents cross-org violations and ensures proper tenant isolation
//
// Valid cases:
//   - Case 1: ent.OrgId=nil + role.OrgId=nil (both domain-level) ✅
//   - Case 2: ent.OrgId=nil + role.OrgId!=nil (global ent for org role) ✅
//   - Case 5: ent.OrgId==role.OrgId (same org) ✅
//
// Invalid cases:
//   - Case 3: ent.OrgId!=nil + role.OrgId=nil (org ent in domain role) ❌
//   - Case 4: ent.OrgId!=role.OrgId (different orgs) ❌
func (this *RoleServiceImpl) validateEntitlementOrgConsistency(inputs []itRole.EntitlementInput, entitlementMap map[model.Id]*domain.Entitlement, dbRole *domain.Role, vErrs *fault.ValidationErrors) bool {
	for _, input := range inputs {
		entitlement := entitlementMap[input.EntitlementId]
		if entitlement == nil {
			vErrs.AppendNotFound("entitlementId", input.EntitlementId)
			return false
		}

		// Case 3: Reject org-specific entitlement in domain-level role
		// This violates tenant isolation as domain roles can be assigned to users in any org
		if entitlement.OrgId != nil && dbRole.OrgId == nil {
			vErrs.Append("entitlementId", "cannot add org-specific entitlement to domain-level role")
			return false
		}

		// Case 4: Reject cross-org assignment
		// Entitlement from OrgA cannot be added to role in OrgB
		if entitlement.OrgId != nil && dbRole.OrgId != nil && *entitlement.OrgId != *dbRole.OrgId {
			vErrs.Append("entitlementId", "entitlement belongs to different organization than role")
			return false
		}

		// Cases 1, 2, 5 are valid and will pass through
		// Note: Case 2 (global ent + org role) is allowed, but scopeRef validation
		// will be handled by EntitlementAssignmentService to ensure proper scope isolation
	}

	return true
}

func (this *RoleServiceImpl) getExistingAssignmentKeys(ctx crud.Context, roleId model.Id) (map[string]bool, error) {
	assignmentsRes, err := this.assignmentService.FindAllBySubject(ctx, itAssign.GetAllEntitlementAssignmentBySubjectQuery{
		SubjectType: domain.EntitlementAssignmentSubjectTypeNikkiRole,
		SubjectRef:  roleId,
	})
	if err != nil {
		return nil, err
	}

	existingAssignments := make(map[string]bool)
	for _, a := range assignmentsRes.Data {
		key := this.makeEntitlementKey(*a.EntitlementId, a.ScopeRef)
		existingAssignments[key] = true
	}

	return existingAssignments, nil
}

func (this *RoleServiceImpl) buildResolvedExpression(roleId model.Id, scopeRef *model.Id, resourceName string, actionName string) string {
	if scopeRef != nil {
		return fmt.Sprintf("%s:%s:%s:%s", roleId, *scopeRef, resourceName, actionName)
	}
	return fmt.Sprintf("%s::%s:%s", roleId, resourceName, actionName)
}

func (this *RoleServiceImpl) createNewEntitlementAssignment(ctx crud.Context, input itRole.EntitlementInput, entitlement *domain.Entitlement, roleId model.Id, vErrs *fault.ValidationErrors) error {
	actionName := ""
	resourceName := ""

	if entitlement.Action != nil && entitlement.Action.Name != nil {
		actionName = *entitlement.Action.Name
	}
	if entitlement.Resource != nil && entitlement.Resource.Name != nil {
		resourceName = *entitlement.Resource.Name
	}

	resolvedExpr := this.buildResolvedExpression(roleId, input.ScopeRef, resourceName, actionName)

	creation, err := this.assignmentService.CreateEntitlementAssignment(ctx, itAssign.CreateEntitlementAssignmentCommand{
		SubjectType:   domain.EntitlementAssignmentSubjectTypeNikkiRole,
		SubjectRef:    roleId,
		EntitlementId: input.EntitlementId,
		ScopeRef:      input.ScopeRef,
		ActionName:    &actionName,
		ResourceName:  &resourceName,
		ResolvedExpr:  resolvedExpr,
	})
	fault.PanicOnErr(err)

	if creation.ClientError != nil {
		vErrs.MergeClientError(creation.ClientError)
		return nil
	}

	return nil
}

// assertBusinessRuleAddEntitlements validates and creates entitlement assignments for a role
//
// This function orchestrates the entire process of adding entitlements to a role with the following steps:
//  1. Validate unique inputs (no duplicates in the request)
//  2. Fetch and map entitlements by ID
//  3. Validate organization-level consistency (cross-org prevention)
//  4. Get existing assignments to avoid duplicates
//  5. Create new assignments for valid, non-existing entitlements
//
// Cross-Org Validation Strategy:
// ┌─────────────────────────────────────────────────────────────────────────────┐
// │ Layer 1: RoleService (this function)                                        │
// │ - Validates org-level consistency between entitlements and roles            │
// │ - Prevents Case 3: org-specific entitlement in domain-level role            │
// │ - Prevents Case 4: cross-org assignments (OrgA ent → OrgB role)             │
// │ - Allows Case 1, 2, 5, 6 to pass through                                    │
// └─────────────────────────────────────────────────────────────────────────────┘
//
//	↓
//
// ┌─────────────────────────────────────────────────────────────────────────────┐
// │ Layer 2: EntitlementAssignmentService                                       │
// │ - Validates scopeRef against resource's ScopeType                           │
// │ - Ensures scope stays within proper boundaries (org/hierarchy/domain)       │
// │ - Handles Case 2/6: global ent + org role with proper scope isolation       │
// └─────────────────────────────────────────────────────────────────────────────┘
//
// Valid Cases:
//
//	✅ Case 1: ent.OrgId=nil + role.OrgId=nil (both domain-level)
//	✅ Case 2: ent.OrgId=nil + role.OrgId!=nil (global ent for org role, scope validated in Layer 2)
//	✅ Case 5: ent.OrgId==role.OrgId (same org boundary)
//
// Invalid Cases (rejected in this layer):
//
//	❌ Case 3: ent.OrgId!=nil + role.OrgId=nil (org ent in domain role → isolation violation)
//	❌ Case 4: ent.OrgId!=role.OrgId (cross-org → tenant isolation violation)
func (this *RoleServiceImpl) assertBusinessRuleAddEntitlements(ctx crud.Context, cmd itRole.AddEntitlementsCommand, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
	if len(cmd.EntitlementInputs) == 0 {
		return nil
	}

	uniqueInputs, isValid := this.validateUniqueEntitlementInputs(cmd.EntitlementInputs, vErrs)
	if !isValid {
		return nil
	}

	entitlementMap, err := this.fetchAndMapEntitlements(ctx, uniqueInputs)
	fault.PanicOnErr(err)

	if !this.validateEntitlementOrgConsistency(uniqueInputs, entitlementMap, dbRole, vErrs) {
		return nil
	}

	existingAssignments, err := this.getExistingAssignmentKeys(ctx, *dbRole.Id)
	fault.PanicOnErr(err)

	for _, input := range uniqueInputs {
		key := this.makeEntitlementKey(input.EntitlementId, input.ScopeRef)

		if existingAssignments[key] {
			vErrs.AppendNotAllow("entitlementId", input.EntitlementId)
			return nil
		}

		entitlement := entitlementMap[input.EntitlementId]
		if entitlement == nil {
			vErrs.AppendNotFound("entitlementId", input.EntitlementId)
			return nil
		}

		err = this.createNewEntitlementAssignment(ctx, input, entitlement, *dbRole.Id, vErrs)
		fault.PanicOnErr(err)

		if vErrs.Count() > 0 {
			return nil
		}
	}

	return nil
}

func (this *RoleServiceImpl) assertBusinessRuleRemoveEntitlements(ctx crud.Context, cmd itRole.RemoveEntitlementsCommand, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
	return nil
}
