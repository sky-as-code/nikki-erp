package app

import (
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
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

func (this *RoleServiceImpl) CreateRole(ctx crud.Context, cmd itRole.CreateRoleCommand) (result *itRole.CreateRoleResult, err error) {
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

func (this *RoleServiceImpl) UpdateRole(ctx crud.Context, cmd itRole.UpdateRoleCommand) (update *itRole.UpdateRoleResult, err error) {
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
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.sanitizeRole(role)
			return this.assertRoleUniqueForUpdate(ctx, role, vErrs)
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			oldEntitlementIds, err = this.getEntitlementIdsByRoleId(ctx, role)
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
	role, err = this.roleRepo.UpdateTx(ctx, *role, *prevEtag, addEntitlementIds, removeEntitlementIds)
	fault.PanicOnErr(err)

	return &itRole.UpdateRoleResult{
		Data:    role,
		HasData: role != nil,
	}, err
}

func (this *RoleServiceImpl) DeleteRoleHard(ctx crud.Context, cmd itRole.DeleteRoleHardCommand) (result *itRole.DeleteRoleHardResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "delete hard role"); e != nil {
			err = e
		}
	}()

	var dbRole *domain.Role
	var assignmentIds []model.Id

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbRole, err = this.assertRoleExistsById(ctx, cmd.Id, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			assignmentIds, err = this.getAssignmentIdsByRoleId(ctx, dbRole)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itRole.DeleteRoleHardResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	deletedCount, err := this.roleRepo.DeleteHardTx(ctx, itRole.DeleteRoleHardParam{Id: cmd.Id, Name: *dbRole.Name})
	fault.PanicOnErr(err)

	err = this.deleteAssignments(ctx, assignmentIds)
	fault.PanicOnErr(err)

	if deletedCount == 0 {
		vErrs.AppendNotFound("role_id", "role")
		return &itRole.DeleteRoleHardResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount), nil
}

func (this *RoleServiceImpl) GetRoleById(ctx crud.Context, query itRole.GetRoleByIdQuery) (result *itRole.GetRoleByIdResult, err error) {
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

	entitlementIds, err := this.getEntitlementIdsByRoleId(ctx, dbRole)
	fault.PanicOnErr(err)

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

func (this *RoleServiceImpl) SearchRoles(ctx crud.Context, query itRole.SearchRolesQuery) (result *itRole.SearchRolesResult, err error) {
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

	for i := range roles.Items {
		entitlementIds, err := this.getEntitlementIdsByRoleId(ctx, &roles.Items[i])
		fault.PanicOnErr(err)

		if len(entitlementIds) > 0 {
			entitlements, err := this.getEntitlements(ctx, entitlementIds)
			fault.PanicOnErr(err)
			roles.Items[i].Entitlements = entitlements
		}
	}

	return &itRole.SearchRolesResult{
		Data:    roles,
		HasData: roles.Items != nil,
	}, nil
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

func (this *RoleServiceImpl) assertRoleExistsById(ctx crud.Context, id model.Id, vErrs *fault.ValidationErrors) (dbRole *domain.Role, err error) {
	dbRole, err = this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: id})
	if dbRole == nil {
		vErrs.AppendNotFound("role_id", "role")
	}
	return
}

func (this *RoleServiceImpl) assertRoleUnique(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	dbRole, err := this.roleRepo.FindByName(ctx, itRole.FindByNameParam{Name: *role.Name})
	fault.PanicOnErr(err)

	if dbRole != nil {
		vErrs.AppendAlreadyExists("role_name", "role name")
	}

	return nil
}

func (this *RoleServiceImpl) assertRoleUniqueForUpdate(ctx crud.Context, role *domain.Role, vErrs *fault.ValidationErrors) error {
	dbRole, err := this.roleRepo.FindByName(ctx, itRole.FindByNameParam{Name: *role.Name})
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

func (this *RoleServiceImpl) validateEntitlements(ctx crud.Context, entitlementIds []model.Id, vErrs *fault.ValidationErrors) {
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

func (this *RoleServiceImpl) getAssignmentsByRoleId(ctx crud.Context, role *domain.Role) ([]*domain.EntitlementAssignment, error) {
	assignmentsRes, err := this.assignmentRepo.FindAllBySubject(
		ctx,
		itAssign.GetAllEntitlementAssignmentBySubjectQuery{
			SubjectType: *domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiRole.String()),
			SubjectRef:  *role.Id,
		})
	fault.PanicOnErr(err)

	return assignmentsRes, nil
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

func (this *RoleServiceImpl) deleteAssignments(ctx crud.Context, assignmentIds []model.Id) error {
	for _, assignmentId := range assignmentIds {
		deletedCount, err := this.assignmentRepo.DeleteHardTx(ctx, itAssign.DeleteEntitlementAssignmentByIdQuery{Id: assignmentId})
		fault.PanicOnErr(err)

		if deletedCount == 0 {
			return fmt.Errorf("failed to delete assignment with ID %s", assignmentId)
		}
	}

	return nil
}
