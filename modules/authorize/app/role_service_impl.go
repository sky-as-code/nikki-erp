package app

// import (
// 	"fmt"

// 	"github.com/sky-as-code/nikki-erp/common/defense"
// 	"github.com/sky-as-code/nikki-erp/common/fault"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/common/orm"
// 	"github.com/sky-as-code/nikki-erp/common/util"
// 	"github.com/sky-as-code/nikki-erp/common/validator"
// 	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
// 	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
// 	"github.com/sky-as-code/nikki-erp/modules/core/crud"
// 	itHierarchy "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
// 	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	itAction "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/action"
// 	itEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement"
// 	itAssign "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement_assignment"
// 	"github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/external"
// 	itGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/grant_request"
// 	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/resource"
// 	itRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/revoke_request"
// 	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/role"
// )

// func NewRoleServiceImpl(
// 	assignmentService itAssign.EntitlementAssignmentService,
// 	cqrsBus cqrs.CqrsBus,
// 	entitlementService itEntitlement.EntitlementService,
// 	grantRequestService itGrantRequest.GrantRequestService,
// 	revokeRequestService itRevokeRequest.RevokeRequestService,
// 	orgExtSvc external.OrganizationExtService,
// 	roleRepo itRole.RoleRepository,
// ) itRole.RoleService {
// 	return &RoleServiceImpl{
// 		assignmentService:    assignmentService,
// 		cqrsBus:              cqrsBus,
// 		entitlementService:   entitlementService,
// 		grantRequestService:  grantRequestService,
// 		revokeRequestService: revokeRequestService,
// 		orgExtSvc:            orgExtSvc,
// 		roleRepo:             roleRepo,
// 	}
// }

// type RoleServiceImpl struct {
// 	assignmentService    itAssign.EntitlementAssignmentService
// 	cqrsBus              cqrs.CqrsBus
// 	entitlementService   itEntitlement.EntitlementService
// 	grantRequestService  itGrantRequest.GrantRequestService
// 	revokeRequestService itRevokeRequest.RevokeRequestService
// 	roleRepo             itRole.RoleRepository
// 	orgExtSvc            external.OrganizationExtService
// }

// func (this *RoleServiceImpl) AddEntitlements(ctx crud.Context, cmd itRole.AddEntitlementsCommand) (result *itRole.AddEntitlementsResult, err error) {
// 	tx, err := this.roleRepo.BeginTransaction(ctx)
// 	fault.PanicOnErr(err)

// 	ctx.SetDbTranx(tx)

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 			return
// 		}

// 		if result != nil && result.ClientError != nil {
// 			tx.Rollback()
// 			return
// 		}

// 		tx.Commit()
// 	}()

// 	result, err = crud.Update(ctx, crud.UpdateParam[*domain.Role, itRole.AddEntitlementsCommand, itRole.AddEntitlementsResult]{
// 		Action:       "add entitlements",
// 		Command:      cmd,
// 		AssertExists: this.assertRoleExistsById,
// 		AssertBusinessRules: func(ctx crud.Context, role *domain.Role, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
// 			return this.assertBusinessRuleAddEntitlements(ctx, cmd, dbRole, vErrs)
// 		},
// 		RepoUpdate: func(ctx crud.Context, role *domain.Role, prevEtag model.Etag) (*domain.Role, error) {
// 			return this.roleRepo.Update(ctx, role, prevEtag)
// 		},
// 		Sanitize: this.sanitizeRole,
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.AddEntitlementsResult {
// 			return &itRole.AddEntitlementsResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(model *domain.Role) *itRole.AddEntitlementsResult {
// 			return &itRole.AddEntitlementsResult{
// 				Data:    model,
// 				HasData: model != nil,
// 			}
// 		},
// 	})

// 	return result, err
// }

// func (this *RoleServiceImpl) RemoveEntitlements(ctx crud.Context, cmd itRole.RemoveEntitlementsCommand) (result *itRole.RemoveEntitlementsResult, err error) {
// 	tx, err := this.roleRepo.BeginTransaction(ctx)
// 	fault.PanicOnErr(err)

// 	ctx.SetDbTranx(tx)

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 			return
// 		}

// 		if result != nil && result.ClientError != nil {
// 			tx.Rollback()
// 			return
// 		}

// 		tx.Commit()
// 	}()

// 	result, err = crud.Update(ctx, crud.UpdateParam[*domain.Role, itRole.RemoveEntitlementsCommand, itRole.RemoveEntitlementsResult]{
// 		Action:       "remove entitlements",
// 		Command:      cmd,
// 		AssertExists: this.assertRoleExistsById,
// 		AssertBusinessRules: func(ctx crud.Context, role *domain.Role, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
// 			return this.assertBusinessRuleRemoveEntitlements(ctx, cmd, dbRole, vErrs)
// 		},
// 		RepoUpdate: func(ctx crud.Context, role *domain.Role, prevEtag model.Etag) (*domain.Role, error) {
// 			return this.roleRepo.Update(ctx, role, prevEtag)
// 		},
// 		Sanitize: this.sanitizeRole,
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.RemoveEntitlementsResult {
// 			return &itRole.RemoveEntitlementsResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(model *domain.Role) *itRole.RemoveEntitlementsResult {
// 			return &itRole.RemoveEntitlementsResult{
// 				Data:    model,
// 				HasData: model != nil,
// 			}
// 		},
// 	})

// 	return result, err
// }

// func (this *RoleServiceImpl) CreateRole(ctx crud.Context, cmd itRole.CreateRoleCommand) (*itRole.CreateRoleResult, error) {
// 	return crud.Create(ctx, crud.CreateParam[*domain.Role, itRole.CreateRoleCommand, itRole.CreateRoleResult]{
// 		Action:              "create role",
// 		Command:             cmd,
// 		AssertBusinessRules: this.assertBusinessRuleCreateRole,
// 		RepoCreate:          this.roleRepo.Create,
// 		SetDefault:          this.setRoleDefaults,
// 		Sanitize:            this.sanitizeRole,
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.CreateRoleResult {
// 			return &itRole.CreateRoleResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(model *domain.Role) *itRole.CreateRoleResult {
// 			return &itRole.CreateRoleResult{
// 				Data:    model,
// 				HasData: model != nil,
// 			}
// 		},
// 	})
// }

// func (this *RoleServiceImpl) UpdateRole(ctx crud.Context, cmd itRole.UpdateRoleCommand) (*itRole.UpdateRoleResult, error) {
// 	return crud.Update(ctx, crud.UpdateParam[*domain.Role, itRole.UpdateRoleCommand, itRole.UpdateRoleResult]{
// 		Action:              "update role",
// 		Command:             cmd,
// 		AssertExists:        this.assertRoleExistsById,
// 		AssertBusinessRules: this.assertBusinessRuleUpdateRole,
// 		RepoUpdate:          this.roleRepo.Update,
// 		Sanitize:            this.sanitizeRole,
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.UpdateRoleResult {
// 			return &itRole.UpdateRoleResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(model *domain.Role) *itRole.UpdateRoleResult {
// 			return &itRole.UpdateRoleResult{
// 				Data:    model,
// 				HasData: model != nil,
// 			}
// 		},
// 	})
// }

// func (this *RoleServiceImpl) DeleteRoleHard(ctx crud.Context, cmd itRole.DeleteRoleHardCommand) (result *itRole.DeleteRoleHardResult, err error) {
// 	tx, err := this.roleRepo.BeginTransaction(ctx)
// 	fault.PanicOnErr(err)

// 	ctx.SetDbTranx(tx)

// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 			return
// 		}

// 		if result != nil && result.ClientError != nil {
// 			tx.Rollback()
// 			return
// 		}

// 		tx.Commit()
// 	}()

// 	result, err = crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Role, itRole.DeleteRoleHardCommand, itRole.DeleteRoleHardResult]{
// 		Action:              "delete role",
// 		Command:             cmd,
// 		AssertExists:        this.assertRoleExistsById,
// 		AssertBusinessRules: this.assertBusinessRuleDeleteRole,
// 		RepoDelete: func(ctx crud.Context, dm *domain.Role) (int, error) {
// 			return this.roleRepo.DeleteHard(ctx, itRole.DeleteRoleHardCommand{Id: *dm.GetId()})
// 		},
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.DeleteRoleHardResult {
// 			return &itRole.DeleteRoleHardResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(dm *domain.Role, deletedCount int) *itRole.DeleteRoleHardResult {
// 			return crud.NewSuccessDeletionResult(*dm.GetId(), &deletedCount)
// 		},
// 	})

// 	return result, err
// }

// func (this *RoleServiceImpl) GetRoleById(ctx crud.Context, query itRole.GetRoleByIdQuery) (result *itRole.GetRoleByIdResult, err error) {
// 	return crud.GetOne(ctx, crud.GetOneParam[*domain.Role, itRole.GetRoleByIdQuery, itRole.GetRoleByIdResult]{
// 		Action:      "get role by Id",
// 		Query:       query,
// 		RepoFindOne: this.getRoleByIdFull,
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.GetRoleByIdResult {
// 			return &itRole.GetRoleByIdResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(model *domain.Role) *itRole.GetRoleByIdResult {
// 			return &itRole.GetRoleByIdResult{
// 				Data:    model,
// 				HasData: model != nil,
// 			}
// 		},
// 	})
// }

// func (this *RoleServiceImpl) SearchRoles(ctx crud.Context, query itRole.SearchRolesQuery) (result *itRole.SearchRolesResult, err error) {
// 	result, err = crud.Search(ctx, crud.SearchParam[domain.Role, itRole.SearchRolesQuery, itRole.SearchRolesResult]{
// 		Action: "search roles",
// 		Query:  query,
// 		SetQueryDefaults: func(query *itRole.SearchRolesQuery) {
// 			query.SetDefaults()
// 		},
// 		ParseSearchGraph: this.roleRepo.ParseSearchGraph,
// 		RepoSearch: func(ctx crud.Context, query itRole.SearchRolesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Role], error) {
// 			result, err := this.roleRepo.Search(ctx, itRole.SearchParam{
// 				GraphCriteria: query.Graph,
// 				Predicate:     predicate,
// 				Order:         order,
// 				Page:          *query.Page,
// 				Size:          *query.Size,
// 			})
// 			fault.PanicOnErr(err)

// 			err = this.populateRoleDetails(ctx, result.Items)
// 			fault.PanicOnErr(err)

// 			return result, err
// 		},
// 		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRole.SearchRolesResult {
// 			return &itRole.SearchRolesResult{
// 				ClientError: vErrs.ToClientError(),
// 			}
// 		},
// 		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.Role]) *itRole.SearchRolesResult {
// 			return &itRole.SearchRolesResult{
// 				Data:    pagedResult,
// 				HasData: pagedResult.Items != nil,
// 			}
// 		},
// 	})

// 	return result, err
// }

// func (this *RoleServiceImpl) GetRolesBySubject(ctx crud.Context, query itRole.GetRolesBySubjectQuery) (result *itRole.GetRolesBySubjectResult, err error) {
// 	defer func() {
// 		if e := fault.RecoverPanicFailedTo(recover(), "get role by subject"); e != nil {
// 			err = e
// 		}
// 	}()

// 	var roles []domain.Role

// 	flow := validator.StartValidationFlow()
// 	vErrs, err := flow.
// 		Step(func(vErrs *fault.ValidationErrors) error {
// 			*vErrs = query.Validate()
// 			return nil
// 		}).
// 		Step(func(vErrs *fault.ValidationErrors) error {
// 			roles, err = this.roleRepo.FindAllBySubject(ctx, query)
// 			return err
// 		}).
// 		End()
// 	fault.PanicOnErr(err)

// 	if vErrs.Count() > 0 {
// 		return &itRole.GetRolesBySubjectResult{
// 			ClientError: vErrs.ToClientError(),
// 		}, nil
// 	}

// 	return &itRole.GetRolesBySubjectResult{
// 		Data:    roles,
// 		HasData: roles != nil,
// 	}, nil
// }

// func (this *RoleServiceImpl) assertRoleExistsById(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) (dbRole *domain.Role, err error) {
// 	dbRole, err = this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: *role.GetId()})
// 	if dbRole == nil {
// 		vErrs.AppendNotFound("role_id", "role")
// 	} else {
// 		if role.GetName() == nil {
// 			role.SetName(dbRole.GetName())
// 		}
// 	}

// 	return
// }

// func (this *RoleServiceImpl) assertRoleUnique(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
// 	if role.GetName() == nil {
// 		return nil
// 	}

// 	dbRole, err := this.roleRepo.FindByName(
// 		ctx,
// 		itRole.FindByNameParam{
// 			Name:  *role.GetName(),
// 			OrgId: role.GetOrgId(),
// 		})
// 	fault.PanicOnErr(err)

// 	if dbRole != nil {
// 		vErrs.AppendAlreadyExists("role_name", "role name")
// 	}

// 	return nil
// }

// func (this *RoleServiceImpl) assertRoleNameUniqueForUpdate(ctx crud.Context, role *domain.Role, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
// 	if role.GetName() == nil {
// 		return nil
// 	}

// 	other, err := this.roleRepo.FindByName(
// 		ctx,
// 		itRole.FindByNameParam{
// 			Name:  *role.GetName(),
// 			OrgId: dbRole.GetOrgId(),
// 		},
// 	)
// 	fault.PanicOnErr(err)

// 	if other != nil && other.GetId() != nil && role.GetId() != nil && *other.GetId() != *role.GetId() {
// 		vErrs.AppendAlreadyExists("role_name", "role name")
// 	}

// 	return nil
// }

// func (this *RoleServiceImpl) sanitizeRole(role *domain.Role) {
// 	if nm := role.GetName(); nm != nil {
// 		san := defense.SanitizePlainText(*nm, true)
// 		role.SetName(&san)
// 	} else {
// 		role.SetName(util.ToPtr(""))
// 	}

// 	if d := role.GetDescription(); d != nil {
// 		san := defense.SanitizePlainText(*d, true)
// 		role.SetDescription(&san)
// 	}
// }

// func (this *RoleServiceImpl) setRoleDefaults(role *domain.Role) {
// 	role.SetDefaults()
// }

// func (this *RoleServiceImpl) getAssignmentsByRoleId(ctx crud.Context, role *domain.Role) ([]domain.EntitlementGrant, error) {
// 	assignmentsRes, err := this.assignmentService.FindAllBySubject(
// 		ctx,
// 		itAssign.GetAllEntitlementAssignmentBySubjectQuery{
// 			SubjectType: domain.EntitlementAssignmentSubjectTypeNikkiRole,
// 			SubjectRef:  string(*role.GetId()),
// 		})
// 	fault.PanicOnErr(err)

// 	return assignmentsRes.Data, nil
// }

// func (this *RoleServiceImpl) getEntitlementIdsByRoleId(ctx crud.Context, role *domain.Role) ([]model.Id, error) {
// 	assignmentsRes, err := this.getAssignmentsByRoleId(ctx, role)

// 	if len(assignmentsRes) == 0 {
// 		return nil, err
// 	}

// 	// Extract unique assignments IDs
// 	entitlementIdSet := make(map[model.Id]bool)
// 	uniqueEntitlementIds := make([]model.Id, 0)

// 	for _, assignment := range assignmentsRes {
// 		if assignment.GetEntitlementId() != nil {
// 			entId := *assignment.GetEntitlementId()
// 			if !entitlementIdSet[entId] {
// 				entitlementIdSet[entId] = true
// 				uniqueEntitlementIds = append(uniqueEntitlementIds, entId)
// 			}
// 		}
// 	}

// 	return uniqueEntitlementIds, nil
// }

// func (this *RoleServiceImpl) getAssignmentIdsByRoleId(ctx crud.Context, role *domain.Role) ([]model.Id, error) {
// 	assignmentsRes, err := this.getAssignmentsByRoleId(ctx, role)

// 	if len(assignmentsRes) == 0 {
// 		return nil, err
// 	}

// 	// Extract unique assignments IDs
// 	assignmentIdSet := make(map[model.Id]bool)
// 	uniqueAssignmentIds := make([]model.Id, 0)

// 	for _, assignment := range assignmentsRes {
// 		if assignment.GetId() != nil {
// 			entId := *assignment.GetId()
// 			if !assignmentIdSet[entId] {
// 				assignmentIdSet[entId] = true
// 				uniqueAssignmentIds = append(uniqueAssignmentIds, entId)
// 			}
// 		}
// 	}

// 	return uniqueAssignmentIds, nil
// }

// func (this *RoleServiceImpl) getEntitlements(ctx crud.Context, entitlementIds []model.Id) ([]domain.Entitlement, error) {
// 	entitlementsRes, err := this.entitlementService.GetAllEntitlementByIds(ctx, itEntitlement.GetAllEntitlementByIdsQuery{Ids: entitlementIds})
// 	fault.PanicOnErr(err)

// 	if entitlementsRes.ClientError != nil {
// 		return nil, entitlementsRes.ClientError
// 	}

// 	return entitlementsRes.Data, nil
// }

// func (this *RoleServiceImpl) buildEntitlementsFromAssignments(ctx crud.Context, assignments []domain.EntitlementGrant) ([]domain.Entitlement, error) {
// 	if len(assignments) == 0 {
// 		return nil, nil
// 	}

// 	entitlementIdSet := make(map[model.Id]bool)
// 	entitlementIds := make([]model.Id, 0)
// 	for _, assignment := range assignments {
// 		if assignment.GetEntitlementId() != nil {
// 			entId := *assignment.GetEntitlementId()
// 			if !entitlementIdSet[entId] {
// 				entitlementIdSet[entId] = true
// 				entitlementIds = append(entitlementIds, entId)
// 			}
// 		}
// 	}

// 	entitlements, err := this.getEntitlements(ctx, entitlementIds)
// 	if err != nil {
// 		return nil, err
// 	}

// 	entitlementMap := make(map[model.Id]*domain.Entitlement)
// 	for i := range entitlements {
// 		ent := &entitlements[i]
// 		if ent.GetId() != nil {
// 			entitlementMap[*ent.GetId()] = ent
// 		}
// 	}

// 	result := make([]domain.Entitlement, 0, len(assignments))
// 	for _, assignment := range assignments {
// 		if assignment.GetEntitlementId() == nil {
// 			continue
// 		}

// 		ent := entitlementMap[*assignment.GetEntitlementId()]
// 		if ent == nil {
// 			continue
// 		}

// 		entWithScope := *ent
// 		if assignment.GetScopeRef() != nil {
// 			entWithScope.ScopeRef = assignment.GetScopeRef()
// 		}

// 		result = append(result, entWithScope)
// 	}

// 	return result, nil
// }

// func (this *RoleServiceImpl) deleteAssignments(ctx crud.Context, assignmentIds []model.Id) error {
// 	for _, assignmentId := range assignmentIds {
// 		deletedCount, err := this.assignmentService.DeleteHardAssignment(ctx, itAssign.DeleteEntitlementAssignmentByIdCommand{Id: assignmentId})
// 		fault.PanicOnErr(err)

// 		if deletedCount.ClientError != nil {
// 			return deletedCount.ClientError
// 		}
// 	}

// 	return nil
// }

// func (this *RoleServiceImpl) assertBusinessRuleCreateRole(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
// 	err := this.assertOrgExists(ctx, role, vErrs)
// 	fault.PanicOnErr(err)

// 	err = this.assertRoleUnique(ctx, role, vErrs)
// 	fault.PanicOnErr(err)

// 	return nil
// }

// func (this *RoleServiceImpl) assertOrgExists(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
// 	if role.GetOrgId() == nil {
// 		return nil
// 	}

// 	existCmd := &itOrg.OrgExistsQuery{
// 		Ids: []model.Id{*role.GetOrgId()},
// 	}
// 	existRes := itOrg.OrgExistsResult{}
// 	err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
// 	fault.PanicOnErr(err)

// 	if len(existRes.ClientErrors) > 0 {
// 		for i := range existRes.ClientErrors {
// 			e := existRes.ClientErrors[i]
// 			vErrs.Append(e.Field, e.Message)
// 		}
// 		return nil
// 	}

// 	if !existRes.Data.Exists(*role.GetOrgId()) {
// 		vErrs.Append("orgId", "not existing organization")
// 	}
// 	return nil
// }

// func (this *RoleServiceImpl) assertBusinessRuleUpdateRole(ctx crud.Context, role *domain.Role, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
// 	err := this.assertRoleNameUniqueForUpdate(ctx, role, dbRole, vErrs)
// 	fault.PanicOnErr(err)

// 	return nil
// }

// func (this *RoleServiceImpl) roleIsDeleted(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
// 	updateGrantRequest, err := this.grantRequestService.TargetIsDeleted(
// 		ctx,
// 		itGrantRequest.TargetIsDeletedCommand{
// 			TargetType: domain.GrantRequestTargetTypeRole,
// 			TargetRef:  *role.GetId(),
// 			TargetName: *role.GetName(),
// 		},
// 	)
// 	fault.PanicOnErr(err)

// 	if updateGrantRequest.ClientError != nil {
// 		vErrs.MergeClientError(updateGrantRequest.ClientError)
// 		return nil
// 	}
// 	if !updateGrantRequest.Data {
// 		vErrs.Append("role_id", "can not delete role with grant requests")
// 		return nil
// 	}

// 	updateRevokeRequest, err := this.revokeRequestService.TargetIsDeleted(
// 		ctx,
// 		itRevokeRequest.TargetIsDeletedCommand{
// 			TargetType: domain.GrantRequestTargetTypeRole,
// 			TargetRef:  *role.GetId(),
// 			TargetName: *role.GetName(),
// 		},
// 	)
// 	fault.PanicOnErr(err)

// 	if updateRevokeRequest.ClientError != nil {
// 		vErrs.MergeClientError(updateRevokeRequest.ClientError)
// 		return nil
// 	}
// 	if !updateRevokeRequest.Data {
// 		vErrs.Append("role_id", "can not delete role with revoke requests")
// 		return nil
// 	}

// 	return nil
// }

// func (this *RoleServiceImpl) assertBusinessRuleDeleteRole(ctx crud.Context, cmd itRole.DeleteRoleHardCommand, role *domain.Role, vErrs *fault.ValidationErrors) error {
// 	var assignmentIds []model.Id

// 	assignmentIds, err := this.getAssignmentIdsByRoleId(ctx, role)
// 	fault.PanicOnErr(err)

// 	if len(assignmentIds) > 0 {
// 		err = this.deleteAssignments(ctx, assignmentIds)
// 		fault.PanicOnErr(err)
// 	}

// 	err = this.roleIsDeleted(ctx, role, vErrs)
// 	fault.PanicOnErr(err)

// 	return nil
// }

// func (this *RoleServiceImpl) getRoleByIdFull(ctx crud.Context, query itRole.GetRoleByIdQuery, vErrs *fault.ValidationErrors) (dbRole *domain.Role, err error) {
// 	dbRole, err = this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: query.Id})
// 	fault.PanicOnErr(err)

// 	if dbRole == nil {
// 		vErrs.AppendNotFound("role_id", "role")
// 		return
// 	}

// 	assignments, err := this.getAssignmentsByRoleId(ctx, dbRole)
// 	fault.PanicOnErr(err)

// 	if len(assignments) > 0 {
// 		entitlements, err := this.buildEntitlementsFromAssignments(ctx, assignments)
// 		fault.PanicOnErr(err)

// 		dbRole.Entitlements = entitlements
// 	}

// 	// Populate organization
// 	if dbRole.GetOrgId() != nil {
// 		org, err := this.getOrganizationById(ctx, *dbRole.GetOrgId())
// 		fault.PanicOnErr(err)
// 		if org != nil {
// 			dbRole.Organization = org
// 			dbRole.OrgName = org.GetDisplayName()
// 		}
// 	}

// 	return
// }

// func (this *RoleServiceImpl) populateRoleDetails(ctx crud.Context, dbRoles []domain.Role) error {
// 	// Collect unique orgIds for batch fetching
// 	orgIdSet := make(map[model.Id]bool)
// 	for i := range dbRoles {
// 		if dbRoles[i].GetOrgId() != nil {
// 			orgIdSet[*dbRoles[i].GetOrgId()] = true
// 		}
// 	}

// 	// Batch fetch organizations
// 	orgMap := make(map[model.Id]*external.Organization)
// 	for orgId := range orgIdSet {
// 		org, err := this.getOrganizationById(ctx, orgId)
// 		fault.PanicOnErr(err)
// 		if org != nil {
// 			orgMap[orgId] = org
// 		}
// 	}

// 	// Populate entitlements and organizations
// 	for i := range dbRoles {
// 		assignments, err := this.getAssignmentsByRoleId(ctx, &dbRoles[i])
// 		fault.PanicOnErr(err)

// 		if len(assignments) > 0 {
// 			entitlements, err := this.buildEntitlementsFromAssignments(ctx, assignments)
// 			fault.PanicOnErr(err)

// 			dbRoles[i].Entitlements = entitlements
// 		}

// 		if dbRoles[i].GetOrgId() != nil {
// 			if org, exists := orgMap[*dbRoles[i].GetOrgId()]; exists {
// 				dbRoles[i].Organization = org
// 				dbRoles[i].OrgName = org.GetDisplayName()
// 			}
// 		}
// 	}

// 	return nil
// }

// func (this *RoleServiceImpl) getOrganizationById(ctx crud.Context, orgId model.Id) (*external.Organization, error) {
// 	query := external.GetOrgQuery{
// 		Id: &orgId,
// 	}
// 	coreCtx := ctx.(corectx.Context)
// 	orgRes, err := this.orgExtSvc.GetOrg(coreCtx, query)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if orgRes.ClientErrors.Count() > 0 || !orgRes.HasData {
// 		return nil, nil
// 	}

// 	return &orgRes.Data, nil
// }

// // makeEntitlementKey create composite key
// func (this *RoleServiceImpl) makeEntitlementKey(entitlementId model.Id, scopeRef *model.Id) string {
// 	if scopeRef == nil {
// 		return string(entitlementId) + ":_"
// 	}
// 	return string(entitlementId) + ":" + string(*scopeRef)
// }

// func assignmentScopeRefAsId(scopeRef *string) *model.Id {
// 	if scopeRef == nil {
// 		return nil
// 	}
// 	id := model.Id(*scopeRef)
// 	return &id
// }

// func (this *RoleServiceImpl) validateUniqueEntitlementInputs(inputs []itRole.EntitlementInput, vErrs *fault.ValidationErrors) ([]itRole.EntitlementInput, bool) {
// 	seen := make(map[string]bool)
// 	uniqueInputs := make([]itRole.EntitlementInput, 0, len(inputs))

// 	for _, input := range inputs {
// 		key := this.makeEntitlementKey(input.EntitlementId, input.ScopeRef)
// 		if seen[key] {
// 			vErrs.AppendNotAllowed("entitlementId", input.EntitlementId)
// 			return nil, false
// 		}
// 		seen[key] = true
// 		uniqueInputs = append(uniqueInputs, input)
// 	}

// 	return uniqueInputs, true
// }

// func (this *RoleServiceImpl) fetchAndMapEntitlements(ctx crud.Context, inputs []itRole.EntitlementInput) (map[model.Id]*domain.Entitlement, error) {
// 	entitlementIds := make([]model.Id, len(inputs))
// 	for i, input := range inputs {
// 		entitlementIds[i] = input.EntitlementId
// 	}

// 	entitlements, err := this.getEntitlements(ctx, entitlementIds)
// 	if err != nil {
// 		return nil, err
// 	}

// 	entitlementMap := make(map[model.Id]*domain.Entitlement)
// 	for i := range entitlements {
// 		e := entitlements[i]
// 		if e.GetId() != nil {
// 			entitlementMap[*e.GetId()] = &e
// 		}
// 	}

// 	return entitlementMap, nil
// }

// func (this *RoleServiceImpl) getExistingAssignmentKeys(ctx crud.Context, roleId model.Id) (map[string]bool, error) {
// 	assignmentsRes, err := this.assignmentService.FindAllBySubject(ctx, itAssign.GetAllEntitlementAssignmentBySubjectQuery{
// 		SubjectType: domain.EntitlementAssignmentSubjectTypeNikkiRole,
// 		SubjectRef:  roleId,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	existingAssignments := make(map[string]bool)
// 	for _, a := range assignmentsRes.Data {
// 		if a.GetEntitlementId() == nil {
// 			continue
// 		}
// 		key := this.makeEntitlementKey(*a.GetEntitlementId(), assignmentScopeRefAsId(a.GetScopeRef()))
// 		existingAssignments[key] = true
// 	}

// 	return existingAssignments, nil
// }

// func (this *RoleServiceImpl) buildResolvedExpression(roleId model.Id, scopeRef *model.Id, actionExpr string) string {
// 	if scopeRef != nil {
// 		return fmt.Sprintf("%s:%s:%s", roleId, *scopeRef, actionExpr)
// 	}

// 	return fmt.Sprintf("%s:*:%s", roleId, actionExpr)
// }

// func (this *RoleServiceImpl) actionAndResourceNames(ctx crud.Context, ent *domain.Entitlement) (actionName, resourceName *string) {
// 	if aid := ent.GetActionId(); aid != nil {
// 		ar := itAction.GetActionByIdResult{}
// 		err := this.cqrsBus.Request(ctx, itAction.GetActionByIdQuery{Id: *aid}, &ar)
// 		fault.PanicOnErr(err)
// 		if ar.Data != nil {
// 			actionName = ar.Data.GetName()
// 		}
// 	}
// 	if rid := ent.GetResourceId(); rid != nil {
// 		rr := itResource.GetResourceByIdResult{}
// 		err := this.cqrsBus.Request(ctx, itResource.GetResourceByIdQuery{Id: *rid}, &rr)
// 		fault.PanicOnErr(err)
// 		if rr.Data != nil {
// 			resourceName = rr.Data.GetName()
// 		}
// 	}
// 	return actionName, resourceName
// }

// func (this *RoleServiceImpl) createNewEntitlementAssignment(ctx crud.Context, input itRole.EntitlementInput, entitlement *domain.Entitlement, roleId model.Id, vErrs *fault.ValidationErrors) error {
// 	actionName, resourceName := this.actionAndResourceNames(ctx, entitlement)

// 	actionExpr := ""
// 	if entitlement.GetActionExpr() != nil {
// 		actionExpr = *entitlement.GetActionExpr()
// 	}
// 	resolvedExpr := this.buildResolvedExpression(roleId, input.ScopeRef, actionExpr)

// 	creation, err := this.assignmentService.CreateEntitlementAssignment(ctx, itAssign.CreateEntitlementAssignmentCommand{
// 		SubjectType:   domain.EntitlementAssignmentSubjectTypeNikkiRole,
// 		SubjectRef:    roleId,
// 		EntitlementId: input.EntitlementId,
// 		ScopeRef:      input.ScopeRef,
// 		ActionName:    actionName,
// 		ResourceName:  resourceName,
// 		ResolvedExpr:  resolvedExpr,
// 	})
// 	fault.PanicOnErr(err)

// 	if creation.ClientError != nil {
// 		vErrs.MergeClientError(creation.ClientError)
// 		return nil
// 	}

// 	return nil
// }

// func (this *RoleServiceImpl) assertBusinessRuleAddEntitlements(ctx crud.Context, cmd itRole.AddEntitlementsCommand, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
// 	if len(cmd.EntitlementInputs) == 0 {
// 		return nil
// 	}

// 	uniqueInputs, isValid := this.validateUniqueEntitlementInputs(cmd.EntitlementInputs, vErrs)
// 	if !isValid {
// 		return nil
// 	}

// 	entitlementMap, err := this.fetchAndMapEntitlements(ctx, uniqueInputs)
// 	fault.PanicOnErr(err)

// 	for _, input := range uniqueInputs {
// 		entitlement := entitlementMap[input.EntitlementId]
// 		if entitlement == nil {
// 			vErrs.AppendNotFound("entitlementId", input.EntitlementId)
// 			return nil
// 		}
// 	}

// 	existingAssignments, err := this.getExistingAssignmentKeys(ctx, *dbRole.GetId())
// 	fault.PanicOnErr(err)

// 	for _, input := range uniqueInputs {
// 		key := this.makeEntitlementKey(input.EntitlementId, input.ScopeRef)

// 		if existingAssignments[key] {
// 			vErrs.AppendNotAllowed("entitlementId", input.EntitlementId)
// 			return nil
// 		}

// 		entitlement := entitlementMap[input.EntitlementId]
// 		if entitlement == nil {
// 			vErrs.AppendNotFound("entitlementId", input.EntitlementId)
// 			return nil
// 		}

// 		err = this.assertRoleOrgIsolationForEntitlement(ctx, dbRole, entitlement, input, vErrs)
// 		fault.PanicOnErr(err)

// 		if vErrs.Count() > 0 {
// 			return nil
// 		}

// 		err = this.createNewEntitlementAssignment(ctx, input, entitlement, *dbRole.GetId(), vErrs)
// 		fault.PanicOnErr(err)
// 	}

// 	return nil
// }

// func (this *RoleServiceImpl) assertRoleOrgIsolationForEntitlement(
// 	ctx crud.Context,
// 	dbRole *domain.Role,
// 	entitlement *domain.Entitlement,
// 	input itRole.EntitlementInput,
// 	vErrs *fault.ValidationErrors,
// ) error {
// 	if dbRole.GetOrgId() == nil {
// 		return nil
// 	}

// 	rid := entitlement.GetResourceId()
// 	if rid == nil {
// 		return nil
// 	}
// 	rr := itResource.GetResourceByIdResult{}
// 	err := this.cqrsBus.Request(ctx, itResource.GetResourceByIdQuery{Id: *rid}, &rr)
// 	fault.PanicOnErr(err)
// 	if rr.Data == nil || rr.Data.GetScopeType() == nil {
// 		return nil
// 	}
// 	scopeType := *rr.Data.GetScopeType()

// 	if scopeType == domain.ResourceMinScopeOrg {
// 		if input.ScopeRef == nil {
// 			return nil
// 		}

// 		if *input.ScopeRef != *dbRole.GetOrgId() {
// 			vErrs.AppendNotAllowed("scopeRef", "scopeRef must match role's orgId for org-specific role")
// 			return nil
// 		}
// 		return nil
// 	}

// 	if scopeType == domain.ResourceMinScopeHierarchy {
// 		if input.ScopeRef == nil {
// 			return nil
// 		}

// 		sid := *input.ScopeRef
// 		hierarchyQuery := &itHierarchy.GetHierarchyLevelQuery{
// 			Id: sid,
// 		}
// 		hierarchyRes := itHierarchy.GetHierarchyLevelResult{}
// 		err = this.cqrsBus.Request(ctx, *hierarchyQuery, &hierarchyRes)
// 		fault.PanicOnErr(err)

// 		if len(hierarchyRes.ClientErrors) > 0 {
// 			for i := range hierarchyRes.ClientErrors {
// 				e := hierarchyRes.ClientErrors[i]
// 				vErrs.Append(e.Field, e.Message)
// 			}
// 			return nil
// 		}

// 		if !hierarchyRes.HasData {
// 			vErrs.AppendNotFound("scopeRef", "hierarchy level not found")
// 			return nil
// 		}

// 		hOrgId := hierarchyRes.Data.GetOrgId()
// 		if hOrgId == nil || *hOrgId != *dbRole.GetOrgId() {
// 			vErrs.AppendNotAllowed("scopeRef", "hierarchy level must belong to role's organization")
// 			return nil
// 		}

// 		return nil
// 	}

// 	return nil
// }

// func (this *RoleServiceImpl) assertBusinessRuleRemoveEntitlements(ctx crud.Context, cmd itRole.RemoveEntitlementsCommand, dbRole *domain.Role, vErrs *fault.ValidationErrors) error {
// 	if len(cmd.EntitlementInputs) == 0 {
// 		return nil
// 	}

// 	uniqueInputs, isValid := this.validateUniqueEntitlementInputs(cmd.EntitlementInputs, vErrs)
// 	if !isValid {
// 		return nil
// 	}

// 	assignments, err := this.getAssignmentsByRoleId(ctx, dbRole)
// 	fault.PanicOnErr(err)

// 	assignmentMap := make(map[string]*domain.EntitlementGrant)
// 	for i := range assignments {
// 		assignment := &assignments[i]
// 		if assignment.GetEntitlementId() == nil {
// 			continue
// 		}
// 		key := this.makeEntitlementKey(*assignment.GetEntitlementId(), assignmentScopeRefAsId(assignment.GetScopeRef()))
// 		assignmentMap[key] = assignment
// 	}

// 	for _, input := range uniqueInputs {
// 		key := this.makeEntitlementKey(input.EntitlementId, input.ScopeRef)

// 		assignment, exists := assignmentMap[key]
// 		if !exists {
// 			vErrs.AppendNotFound("entitlementId", input.EntitlementId)
// 			return nil
// 		}

// 		_, err = this.assignmentService.DeleteHardAssignment(ctx, itAssign.DeleteEntitlementAssignmentByIdCommand{
// 			Id: *assignment.GetId(),
// 		})
// 		fault.PanicOnErr(err)
// 	}

// 	return nil
// }
