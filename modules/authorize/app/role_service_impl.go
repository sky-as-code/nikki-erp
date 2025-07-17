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
	this.setRoleDefaults(ctx, role)
	role.SetCreatedAt(time.Now())

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = role.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeRole(role)
			return this.assertRoleUnique(ctx, role, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if len(role.Entitlements) > 0 {
				this.validateEntitlements(ctx, role.Entitlements, vErrs)
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

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

	return &roleIt.CreateRoleResult{
		Data:    createdRole,
		HasData: createdRole != nil,
	}, nil
}

func (this *RoleServiceImpl) GetRoleById(ctx context.Context, query roleIt.GetRoleByIdQuery) (result *roleIt.GetRoleByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get role by id"); e != nil {
			err = e
		}
	}()

	var dbRole *domain.Role
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbRole, err = this.assertRoleExistsById(ctx, query.Id, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &roleIt.GetRoleByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlementIds, err := this.getEntitlementByIds(ctx, dbRole, vErrs)
	ft.PanicOnErr(err)
	if vErrs.Count() > 0 {
		return &roleIt.GetRoleByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if len(entitlementIds) > 0 {
		entitlements, err := this.getEntitlements(ctx, entitlementIds)
		ft.PanicOnErr(err)
		dbRole.Entitlements = entitlements
	}

	return &roleIt.GetRoleByIdResult{
		Data:    dbRole,
		HasData: dbRole != nil,
	}, nil
}

func (this *RoleServiceImpl) SearchRoles(ctx context.Context, query roleIt.SearchRolesQuery) (result *roleIt.SearchRolesResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list roles"); e != nil {
			err = e
		}
	}()

	query.SetDefaults()
	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.roleRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &roleIt.SearchRolesResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	roles, err := this.roleRepo.Search(ctx, roleIt.SearchParam{
		Predicate:        predicate,
		Order:            order,
		Page:             *query.Page,
		Size:             *query.Size,
		WithEntitlements: false, // We'll load entitlements manually
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
		Data:    roles,
		HasData: roles.Items != nil,
	}, nil
}

func (this *RoleServiceImpl) GetRolesBySubject(ctx context.Context, query roleIt.GetRolesBySubjectQuery) (result *roleIt.GetRolesBySubjectResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get role by subject"); e != nil {
			err = e
		}
	}()

	var roles []*domain.Role

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			// *vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			roles, err = this.roleRepo.FindAllBySubject(ctx, query)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &roleIt.GetRolesBySubjectResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &roleIt.GetRolesBySubjectResult{
		Data:    roles,
		HasData: roles != nil,
	}, nil
}

func (this *RoleServiceImpl) assertRoleExistsById(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (dbRole *domain.Role, err error) {
	dbRole, err = this.roleRepo.FindById(ctx, roleIt.FindByIdParam{Id: id})
	if dbRole == nil {
		vErrs.AppendIdNotFound("role")
	}
	return
}

func (this *RoleServiceImpl) assertRoleUnique(ctx context.Context, role *domain.Role, vErrs *ft.ValidationErrors) error {
	if vErrs.Has("name") {
		return nil
	}

	dbRole, err := this.roleRepo.FindByName(ctx, roleIt.FindByNameParam{Name: *role.Name})
	ft.PanicOnErr(err)

	if dbRole != nil {
		vErrs.Append("name", "name already exists")
	}

	return nil
}

func (this *RoleServiceImpl) sanitizeRole(role *domain.Role) {
	if role.Description != nil {
		cleanedDescription := strings.TrimSpace(*role.Description)
		cleanedDescription = defense.SanitizePlainText(cleanedDescription)
		role.Description = &cleanedDescription
	}
}

func (this *RoleServiceImpl) setRoleDefaults(ctx context.Context, role *domain.Role) {
	role.SetDefaults()
}

func (this *RoleServiceImpl) validateEntitlements(ctx context.Context, entitlements []*domain.Entitlement, vErrs *ft.ValidationErrors) {
	if len(entitlements) == 0 {
		return
	}

	// Check for duplicate entitlement IDs and null IDs first
	seenIds := make(map[model.Id]int)
	validEntitlements := make([]*domain.Entitlement, 0)

	for i, ent := range entitlements {
		if ent.Id == nil {
			vErrs.Append(fmt.Sprintf("entitlements[%d]", i), "entitlement id cannot be nil")
			continue
		}

		if firstIndex, exists := seenIds[*ent.Id]; exists {
			vErrs.Append(fmt.Sprintf("entitlements[%d]", i), fmt.Sprintf("duplicate entitlement id found at index %d", firstIndex))
			continue
		}

		seenIds[*ent.Id] = i
		validEntitlements = append(validEntitlements, ent)
	}

	if vErrs.Count() > 0 {
		return
	}

	for _, ent := range validEntitlements {
		existsRes := entitlementIt.EntitlementExistsResult{}
		err := this.cqrsBus.Request(ctx, entitlementIt.EntitlementExistsCommand{Id: *ent.Id}, &existsRes)
		ft.PanicOnErr(err)

		if existsRes.ClientError != nil {
			vErrs.MergeClientError(existsRes.ClientError)
			continue
		}

		if !existsRes.Data {
			originalIndex := seenIds[*ent.Id]
			vErrs.Append(fmt.Sprintf("entitlements[%d]", originalIndex), "entitlement not found")
		}
	}
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
