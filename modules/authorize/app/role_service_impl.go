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
	itEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	itAssign "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

func NewRoleServiceImpl(
	roleRepo itRole.RoleRepository,
	entitlementRepo itEntitlement.EntitlementRepository,
	assignmentRepo itAssign.EntitlementAssignmentRepository,
	eventBus event.EventBus,
) itRole.RoleService {
	return &RoleServiceImpl{
		roleRepo:        roleRepo,
		entitlementRepo: entitlementRepo,
		assignmentRepo:  assignmentRepo,
		eventBus:        eventBus,
	}
}

type RoleServiceImpl struct {
	roleRepo        itRole.RoleRepository
	entitlementRepo itEntitlement.EntitlementRepository
	assignmentRepo  itAssign.EntitlementAssignmentRepository
	eventBus        event.EventBus
}

func (this *RoleServiceImpl) CreateRole(ctx context.Context, cmd itRole.CreateRoleCommand) (result *itRole.CreateRoleResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "create role"); e != nil {
			err = e
		}
	}()

	role := cmd.ToRole()
	this.setRoleDefaults(role)

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = role.Validate(false)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.sanitizeRole(role)
			return this.assertRoleUnique(ctx, role, vErrs)
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.validateEntitlements(ctx, cmd.EntitlementIds, vErrs)
			return nil
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itRole.CreateRoleResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	var createdRole *domain.Role
	if len(cmd.EntitlementIds) > 0 {
		createdRole, err = this.roleRepo.CreateWithEntitlements(ctx, *role, cmd.EntitlementIds)
	} else {
		createdRole, err = this.roleRepo.Create(ctx, *role)
	}
	fault.PanicOnErr(err)

	return &itRole.CreateRoleResult{
		Data:    createdRole,
		HasData: createdRole != nil,
	}, nil
}

func (this *RoleServiceImpl) UpdateRole(ctx context.Context, cmd itRole.UpdateRoleCommand) (update *itRole.UpdateRoleResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "update role"); e != nil {
			err = e
		}
	}()

	role := cmd.ToRole()
	var dbRole *domain.Role

	var oldEntitlementIds []model.Id
	var addEntitlementIds, removeEntitlementIds []model.Id

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = role.Validate(true)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbRole, err = this.assertRoleExistsById(ctx, *role.Id, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertCorrectEtag(*role.Etag, *dbRole.Etag, vErrs)
			this.sanitizeRole(role)

			if *role.Name == *dbRole.Name {
				vErrs.AppendAlreadyExists("role_name", "role name")
			}
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			oldEntitlementIds, err = this.getEntitlementByIds(ctx, role, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.validateEntitlements(ctx, cmd.EntitlementIds, vErrs)
			addEntitlementIds, removeEntitlementIds = this.diffEntitlementIds(oldEntitlementIds, cmd.EntitlementIds)
			return nil
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itRole.UpdateRoleResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := role.Etag
	role.Etag = model.NewEtag()
	role, err = this.roleRepo.UpdateWithEntitlements(ctx, *role, *prevEtag, addEntitlementIds, removeEntitlementIds)
	fault.PanicOnErr(err)

	return &itRole.UpdateRoleResult{
		Data:    role,
		HasData: role != nil,
	}, err
}

func (this *RoleServiceImpl) DeleteRole(ctx context.Context) (err error) {
	/*
		1. Check workflow
			a. Validate req
			b. Check existing role
		2. Delete handle (All must use transaction)
			a. Config role_id on permission histories
			b. Set null target_role_id on grant/revoke
			c. Delete row which have role_id on role_user
			d. Delete role
				i. Delete through role_rolesuite
				ii. Delete role on role table
		3. Return
	*/
	return nil
}

func (this *RoleServiceImpl) GetRoleById(ctx context.Context, query itRole.GetRoleByIdQuery) (result *itRole.GetRoleByIdResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get role by id"); e != nil {
			err = e
		}
	}()

	var dbRole *domain.Role
	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbRole, err = this.assertRoleExistsById(ctx, query.Id, vErrs)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itRole.GetRoleByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlementIds, err := this.getEntitlementByIds(ctx, dbRole, &vErrs)
	fault.PanicOnErr(err)
	if vErrs.Count() > 0 {
		return &itRole.GetRoleByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if len(entitlementIds) > 0 {
		entitlements, err := this.getEntitlements(ctx, entitlementIds)
		fault.PanicOnErr(err)
		dbRole.Entitlements = entitlements
	}

	return &itRole.GetRoleByIdResult{
		Data:    dbRole,
		HasData: dbRole != nil,
	}, nil
}

func (this *RoleServiceImpl) SearchRoles(ctx context.Context, query itRole.SearchRolesQuery) (result *itRole.SearchRolesResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "search roles"); e != nil {
			err = e
		}
	}()

	query.SetDefaults()
	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.roleRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &itRole.SearchRolesResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	roles, err := this.roleRepo.Search(ctx, itRole.SearchParam{
		Predicate:        predicate,
		Order:            order,
		Page:             *query.Page,
		Size:             *query.Size,
		WithEntitlements: false,
	})
	fault.PanicOnErr(err)

	for _, role := range roles.Items {
		entitlementIds, err := this.getEntitlementByIds(ctx, &role, &vErrsModel)
		fault.PanicOnErr(err)

		if vErrsModel.Count() > 0 {
			return &itRole.SearchRolesResult{
				ClientError: vErrsModel.ToClientError(),
			}, nil
		}

		if len(entitlementIds) > 0 {
			entitlements, err := this.getEntitlements(ctx, entitlementIds)
			fault.PanicOnErr(err)
			role.Entitlements = entitlements
		}
	}

	return &itRole.SearchRolesResult{
		Data:    roles,
		HasData: roles.Items != nil,
	}, nil
}

func (this *RoleServiceImpl) GetRolesBySubject(ctx context.Context, query itRole.GetRolesBySubjectQuery) (result *itRole.GetRolesBySubjectResult, err error) {
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

func (this *RoleServiceImpl) assertRoleExistsById(ctx context.Context, id model.Id, vErrs *fault.ValidationErrors) (dbRole *domain.Role, err error) {
	dbRole, err = this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: id})
	if dbRole == nil {
		vErrs.AppendNotFound("role_id", "role")
	}
	return
}

func (this *RoleServiceImpl) assertRoleUnique(ctx context.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	dbRole, err := this.roleRepo.FindByName(ctx, itRole.FindByNameParam{Name: *role.Name})
	fault.PanicOnErr(err)

	if dbRole != nil {
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

func (this *RoleServiceImpl) validateEntitlements(ctx context.Context, entitlementIds []model.Id, vErrs *fault.ValidationErrors) {
	if len(entitlementIds) == 0 {
		return
	}

	// Check for duplicate entitlement IDs and null IDs first
	seenIds := make(map[model.Id]int)
	validEntitlementIds := make([]model.Id, 0)

	for i, entId := range entitlementIds {
		if firstIndex, exists := seenIds[entId]; exists {
			vErrs.Append(fmt.Sprintf("entitlements[%d]", i), fmt.Sprintf("duplicate entitlement id found at index %d", firstIndex))
			continue
		}

		seenIds[entId] = i
		validEntitlementIds = append(validEntitlementIds, entId)
	}

	if vErrs.Count() > 0 {
		return
	}

	for _, entId := range validEntitlementIds {
		existsRes, err := this.entitlementRepo.Exists(ctx, itEntitlement.FindByIdParam{Id: entId})
		fault.PanicOnErr(err)

		if !existsRes {
			originalIndex := seenIds[entId]
			vErrs.Append(fmt.Sprintf("entitlements[%d]", originalIndex), "entitlement not found")
		}
	}
}

func (this *RoleServiceImpl) getEntitlementByIds(ctx context.Context, role *domain.Role, vErrs *fault.ValidationErrors) ([]model.Id, error) {
	assignmentsRes, err := this.assignmentRepo.FindAllBySubject(
		ctx,
		itAssign.GetAllEntitlementAssignmentBySubjectQuery{
			SubjectType: *domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiRole.String()),
			SubjectRef:  *role.Id,
		})
	fault.PanicOnErr(err)

	if len(assignmentsRes) == 0 {
		return nil, err
	}

	// Extract unique entitlement IDs from assignments
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

func (this *RoleServiceImpl) getEntitlements(ctx context.Context, entitlementIds []model.Id) ([]domain.Entitlement, error) {
	entitlementsRes, err := this.entitlementRepo.FindAllByIds(ctx, itEntitlement.GetAllEntitlementByIdsQuery{Ids: entitlementIds})
	fault.PanicOnErr(err)

	if entitlementsRes != nil {
		return nil, err
	}

	return entitlementsRes, nil
}

func (this *RoleServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *fault.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *RoleServiceImpl) diffEntitlementIds(oldIds, newIds []model.Id) (added, removed []model.Id) {
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
