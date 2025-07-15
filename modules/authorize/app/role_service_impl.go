package app

import (
	"context"
	"fmt"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	entitlementIt "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	entitlementAssignIt "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	roleIt "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
)

func NewRoleServiceImpl(
	roleRepo roleIt.RoleRepository,
	eventBus event.EventBus,
	cqrsBus cqrs.CqrsBus,
) roleIt.RoleService {
	return &RoleServiceImpl{
		roleRepo: roleRepo,
		eventBus: eventBus,
		cqrsBus:  cqrsBus,
	}
}

type RoleServiceImpl struct {
	roleRepo roleIt.RoleRepository
	eventBus event.EventBus
	cqrsBus  cqrs.CqrsBus
}

func (this *RoleServiceImpl) CreateRole(ctx context.Context, cmd roleIt.CreateRoleCommand) (result *roleIt.CreateRoleResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create role"); e != nil {
			err = e
		}
	}()

	role := cmd.ToRole()
	err = role.SetDefaults()
	ft.PanicOnErr(err)

	vErrs := role.Validate(false)
	this.assertRoleUnique(ctx, role, &vErrs)
	if len(role.Entitlements) > 0 {
		this.validateEntitlements(ctx, role.Entitlements, &vErrs)
	}

	if vErrs.Count() > 0 {
		return &roleIt.CreateRoleResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	var createdRole *domain.Role
	if len(role.Entitlements) > 0 {
		entitlementIds := make([]model.Id, 0, len(role.Entitlements))
		for _, ent := range role.Entitlements {
			if ent.Id != nil {
				entitlementIds = append(entitlementIds, *ent.Id)
			}
		}

		if len(entitlementIds) > 0 {
			createdRole, err = this.roleRepo.CreateWithEntitlements(ctx, *role, entitlementIds)
		} else {
			createdRole, err = this.roleRepo.Create(ctx, *role)
		}
	} else {
		createdRole, err = this.roleRepo.Create(ctx, *role)
	}

	ft.PanicOnErr(err)

	return &roleIt.CreateRoleResult{Data: createdRole}, nil
}

func (s *RoleServiceImpl) assertRoleUnique(ctx context.Context, role *domain.Role, errors *ft.ValidationErrors) {
	if errors.Has("name") {
		return
	}

	dbRole, err := s.roleRepo.FindByName(ctx, roleIt.FindByNameParam{Name: *role.Name})
	ft.PanicOnErr(err)

	if dbRole != nil {
		errors.Append("name", "name already exists")
	}
}

func (this *RoleServiceImpl) validateEntitlements(ctx context.Context, entitlementIds []*domain.Entitlement, errors *ft.ValidationErrors) {
	if len(entitlementIds) == 0 {
		return
	}

	for i, ent := range entitlementIds {
		if ent.Id == nil {
			errors.Append(fmt.Sprintf("entitlements[%d]", i), "entitlement id cannot be nil")
			continue
		}

		existsRes := entitlementIt.EntitlementExistsResult{}
		err := this.cqrsBus.Request(ctx, entitlementIt.EntitlementExistsCommand{Id: *ent.Id}, &existsRes)
		ft.PanicOnErr(err)

		if existsRes.ClientError != nil {
			errors.MergeClientError(existsRes.ClientError)
			continue
		}

		if !existsRes.Data {
			errors.Append(fmt.Sprintf("entitlements[%d]", i), "entitlement not found")
		}
	}
}

func (this *RoleServiceImpl) GetRoleById(ctx context.Context, query roleIt.GetRoleByIdQuery) (result *roleIt.GetRoleByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get role by id"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &roleIt.GetRoleByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	role, err := this.roleRepo.FindById(ctx, query)
	ft.PanicOnErr(err)

	if role == nil {
		vErrs.Append("id", "role not found")
		return &roleIt.GetRoleByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlementIds, err := this.getEntitlementByIds(ctx, role, &vErrs)
	ft.PanicOnErr(err)
	if vErrs.Count() > 0 {
		return &roleIt.GetRoleByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if len(entitlementIds) == 0 {
		return &roleIt.GetRoleByIdResult{
			Data: role,
		}, nil
	}

	entitlements, err := this.getEntitlements(ctx, entitlementIds)
	ft.PanicOnErr(err)

	role.Entitlements = entitlements

	return &roleIt.GetRoleByIdResult{
		Data: role,
	}, nil
}

func (this *RoleServiceImpl) getEntitlementByIds(ctx context.Context, role *domain.Role, vErrs *ft.ValidationErrors) ([]model.Id, error) {
	assignments := entitlementAssignIt.GetAllEntitlementAssignmentBySubjectQuery{
		SubjectType: domain.EntitlementAssignmentSubjectTypeNikkiRole.String(),
		SubjectRef:  *role.Id,
	}
	assignmentsRes := entitlementAssignIt.GetAllEntitlementAssignmentBySubjectResult{}
	err := this.cqrsBus.Request(ctx, assignments, &assignmentsRes)
	ft.PanicOnErr(err)

	if assignmentsRes.ClientError != nil {
		vErrs.MergeClientError(assignmentsRes.ClientError)
		return nil, err
	}

	// Extract unique entitlement IDs from assignments
	entitlementIdSet := make(map[model.Id]bool)
	uniqueEntitlementIds := make([]model.Id, 0)

	for _, assignment := range assignmentsRes.Data {
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

func (this *RoleServiceImpl) getEntitlements(ctx context.Context, entitlementIds []model.Id) ([]*domain.Entitlement, error) {
	entitlements := entitlementIt.GetAllEntitlementByIdsQuery{
		Ids: entitlementIds,
	}
	entitlementsRes := entitlementIt.GetAllEntitlementByIdsResult{}
	err := this.cqrsBus.Request(ctx, entitlements, &entitlementsRes)
	ft.PanicOnErr(err)

	if entitlementsRes.ClientError != nil {
		return nil, entitlementsRes.ClientError
	}

	return entitlementsRes.Data, nil
}

func (this *RoleServiceImpl) SearchRoles(ctx context.Context, query roleIt.SearchRolesQuery) (result *roleIt.SearchRolesResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list roles"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.roleRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &roleIt.SearchRolesResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	roles, err := this.roleRepo.Search(ctx, roleIt.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
	})
	ft.PanicOnErr(err)

	// Populate entitlements for each role
	for _, role := range roles.Items {
		entitlementIds, err := this.getEntitlementByIds(ctx, role, &vErrsModel)
		ft.PanicOnErr(err)

		if vErrsModel.Count() > 0 {
			return &roleIt.SearchRolesResult{
				ClientError: vErrsModel.ToClientError(),
			}, nil
		}

		if len(entitlementIds) > 0 {
			entitlements, err := this.getEntitlements(ctx, entitlementIds)
			ft.PanicOnErr(err)
			role.Entitlements = entitlements
		}
	}

	return &roleIt.SearchRolesResult{
		Data: roles,
	}, nil
}
