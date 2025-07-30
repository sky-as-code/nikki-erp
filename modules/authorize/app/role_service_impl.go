package app

import (
	"context"
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/event"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	itAssign "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

func NewRoleServiceImpl(
	roleRepo itRole.RoleRepository,
	eventBus event.EventBus,
	cqrsBus cqrs.CqrsBus,
) itRole.RoleService {
	return &RoleServiceImpl{
		roleRepo: roleRepo,
		eventBus: eventBus,
		cqrsBus:  cqrsBus,
	}
}

type RoleServiceImpl struct {
	roleRepo itRole.RoleRepository
	eventBus event.EventBus
	cqrsBus  cqrs.CqrsBus
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
			if len(role.Entitlements) > 0 {
				this.validateEntitlements(ctx, role.Entitlements, vErrs)
			}
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
	fault.PanicOnErr(err)

	return &itRole.CreateRoleResult{
		Data:    createdRole,
		HasData: createdRole != nil,
	}, nil
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
		WithEntitlements: false, // We'll load entitlements manually
	})
	fault.PanicOnErr(err)

	// Populate entitlements for each role
	for _, role := range roles.Items {
		entitlementIds, err := this.getEntitlementByIds(ctx, role, &vErrsModel)
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

	var roles []*domain.Role

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
		vErrs.AppendNotFound("id", "role")
	}
	return
}

func (this *RoleServiceImpl) assertRoleUnique(ctx context.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	dbRole, err := this.roleRepo.FindByName(ctx, itRole.FindByNameParam{Name: *role.Name})
	fault.PanicOnErr(err)

	if dbRole != nil {
		vErrs.AppendAlreadyExists("name", "role name")
	}

	return nil
}

func (this *RoleServiceImpl) sanitizeRole(role *domain.Role) {
	if role.Description != nil {
		role.Description = util.ToPtr(defense.SanitizePlainText(*role.Description, true))
	}
}

func (this *RoleServiceImpl) setRoleDefaults(role *domain.Role) {
	role.SetDefaults()
}

func (this *RoleServiceImpl) validateEntitlements(ctx context.Context, entitlements []*domain.Entitlement, vErrs *fault.ValidationErrors) {
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
		existsRes := itEntitlement.EntitlementExistsResult{}
		err := this.cqrsBus.Request(ctx, itEntitlement.EntitlementExistsCommand{Id: *ent.Id}, &existsRes)
		fault.PanicOnErr(err)

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

func (this *RoleServiceImpl) getEntitlementByIds(ctx context.Context, role *domain.Role, vErrs *fault.ValidationErrors) ([]model.Id, error) {
	assignments := itAssign.GetAllEntitlementAssignmentBySubjectQuery{
		SubjectType: *domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiRole.String()),
		SubjectRef:  *role.Id,
	}
	assignmentsRes := itAssign.GetAllEntitlementAssignmentBySubjectResult{}
	err := this.cqrsBus.Request(ctx, assignments, &assignmentsRes)
	fault.PanicOnErr(err)

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
	entitlements := itEntitlement.GetAllEntitlementByIdsQuery{
		Ids: entitlementIds,
	}
	entitlementsRes := itEntitlement.GetAllEntitlementByIdsResult{}
	err := this.cqrsBus.Request(ctx, entitlements, &entitlementsRes)
	fault.PanicOnErr(err)

	if entitlementsRes.ClientError != nil {
		return nil, entitlementsRes.ClientError
	}

	return entitlementsRes.Data, nil
}
