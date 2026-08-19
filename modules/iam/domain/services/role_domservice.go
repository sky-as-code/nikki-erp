package services

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
	itOrgz "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

func NewRoleDomainServiceImpl(
	roleRepo itRole.RoleRepository,
	entitlementRepo itEnt.EntitlementRepository,
	orgRepo itOrgz.OrganizationRepository,
	orgUnitRepo itOrg.OrgUnitRepository,
	orgUnitSvc itOrg.OrgUnitDomainService,
	permRepo itPerm.PermissionRepository,
	cqrsBus cqrs.CqrsBus,
) itRole.RoleDomainService {
	return &RoleDomainServiceImpl{
		cqrsBus:         cqrsBus,
		roleRepo:        roleRepo,
		entitlementRepo: entitlementRepo,
		orgRepo:         orgRepo,
		orgUnitRepo:     orgUnitRepo,
		orgUnitSvc:      orgUnitSvc,
		permRepo:        permRepo,
	}
}

type RoleDomainServiceImpl struct {
	cqrsBus         cqrs.CqrsBus
	roleRepo        itRole.RoleRepository
	entitlementRepo itEnt.EntitlementRepository
	// orgRepo and orgUnitRepo exist only to resolve entitlement scope display names in
	// DescribeRoles; role CRUD does not touch them.
	orgRepo     itOrgz.OrganizationRepository
	orgUnitRepo itOrg.OrgUnitRepository
	orgUnitSvc  itOrg.OrgUnitDomainService
	permRepo    itPerm.PermissionRepository
}

func (this *RoleDomainServiceImpl) CreateRole(
	ctx corectx.Context, cmd itRole.CreateRoleCommand, options ...corecrud.ServiceCreateOptions[*domain.Role],
) (*itRole.CreateRoleResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceCreateOptions[*domain.Role]{})
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Role, *domain.Role]{
		Action:                 "create role",
		BaseRepoGetter:         this.roleRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *RoleDomainServiceImpl) DeleteRole(
	ctx corectx.Context, cmd itRole.DeleteRoleCommand, options ...corecrud.ServiceDeleteOptions,
) (*itRole.DeleteRoleResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceDeleteOptions{})
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:                 "delete role",
		DbRepoGetter:           this.roleRepo,
		Cmd:                    dyn.DeleteOneCommand(cmd),
		AfterValidationSuccess: opts.AfterValidationSuccess,
		ValidateExtra: func(ctx corectx.Context, keyFields dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
			resRole, err := this.roleRepo.GetOne(ctx, dyn.RepoGetOneParam{
				Filter: keyFields,
				Fields: []string{domain.RoleFieldIsPrivate},
			})
			if err != nil {
				return err
			}
			if resRole.ClientErrors.Count() > 0 {
				return errors.Wrap(resRole.ClientErrors.ToError(), "failed to get role before deletion")
			}
			if !resRole.HasData {
				return ft.NewAnonymousNotFoundError()
			}
			if *resRole.Data.IsPrivate() {
				return ft.NewAnonymousBusinessViolation(
					ft.ErrorKey("iam", "err_private_role_deletion_not_allowed"),
					"private role deletion is not allowed.",
				)
			}
			return nil
		},
	})
}

func (this *RoleDomainServiceImpl) GetRole(
	ctx corectx.Context, query itRole.GetRoleQuery,
) (*dyn.OpResult[domain.Role], error) {
	return corecrud.GetOne[domain.Role](ctx, corecrud.GetOneParam{
		Action:       "get role",
		DbRepoGetter: this.roleRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *RoleDomainServiceImpl) ManageRoleEntitlements(
	ctx corectx.Context, cmd itRole.ManageRoleEntitlementsCommand,
) (*itRole.ManageRoleEntitlementsResult, error) {
	if !cmd.Add.IsEmpty() {
		cErrs, err := this.validateAddsBelongToRoleOrg(ctx, cmd.RoleId, cmd.Add)
		if err != nil {
			return nil, err
		}
		if len(cErrs) > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
		}
	}
	result, err := corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage role entitlements",
		DbRepoGetter:       this.roleRepo,
		DestSchemaName:     domain.EntitlementSchemaName,
		SrcId:              cmd.RoleId,
		SrcIdFieldForError: "role_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
	})
	if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
		return result, err
	}
	// What this role grants has changed, so everyone holding it needs recomputing.
	if err := this.permRepo.RebuildUserPermissionsForRole(ctx, cmd.RoleId); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *RoleDomainServiceImpl) validateAddsBelongToRoleOrg(
	ctx corectx.Context, roleID model.Id, add datastructure.Set[model.Id],
) (ft.ClientErrors, error) {
	var out ft.ClientErrors
	addIDs := add.ToSlice()
	roleRes, err := corecrud.GetOne[domain.Role](ctx, corecrud.GetOneParam{
		Action:       "validate role for entitlement adds",
		DbRepoGetter: this.roleRepo,
		Query: dyn.GetOneQuery{
			Id:     roleID,
			Fields: []string{domain.RoleFieldOrgId},
		},
	})
	if err != nil {
		return nil, err
	}
	if roleRes.ClientErrors.Count() > 0 {
		return roleRes.ClientErrors, nil
	}
	if !roleRes.HasData {
		out = append(out, *ft.NewBusinessViolation("add", "role_not_found", "role not found"))
		return out, nil
	}
	roleOrgID := roleRes.Data.GetFieldData().GetModelId(domain.RoleFieldOrgId)
	searchRes, err := this.entitlementRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:  entitlementIdsSearchGraph(addIDs),
		Fields: []string{basemodel.FieldId, domain.EntitlementFieldOrgUnitId},
		Page:   0,
		Size:   len(addIDs),
	})
	if err != nil {
		return nil, err
	}
	if searchRes.ClientErrors.Count() > 0 {
		return searchRes.ClientErrors, nil
	}
	found := make(map[model.Id]domain.Entitlement, len(searchRes.Data.Items))
	for _, ent := range searchRes.Data.Items {
		idPtr := ent.GetFieldData().GetModelId(basemodel.FieldId)
		if idPtr != nil {
			found[*idPtr] = ent
		}
	}
	for _, wantID := range addIDs {
		ent, ok := found[wantID]
		if !ok {
			out = append(out, *ft.NewBusinessViolation(
				"add", "entitlement_not_found", "entitlement not found",
				map[string]any{"id": string(wantID)},
			))
			continue
		}
		ouID := ent.GetFieldData().GetModelId(domain.EntitlementFieldOrgUnitId)
		if ouID == nil || *ouID == "" {
			continue
		}
		if roleOrgID == nil || *roleOrgID == "" {
			out = append(out, *ft.NewBusinessViolation(
				"add", "role_org_required",
				"role org_id is required to assign entitlements with org_unit_id",
				map[string]any{"org_unit_id": string(*ouID)},
			))
			continue
		}
		ouRes, ouErr := this.orgUnitSvc.GetOrgUnit(ctx, itOrg.GetOrgUnitQuery{
			Id:     *ouID,
			Fields: []string{domain.OrgUnitFieldPath},
		})
		if ouErr != nil {
			return nil, ouErr
		}
		if ouRes.ClientErrors.Count() > 0 {
			for i := range ouRes.ClientErrors {
				item := ouRes.ClientErrors[i]
				item.Field = "add"
				out = append(out, item)
			}
			continue
		}
		if !ouRes.HasData {
			out = append(out, *ft.NewBusinessViolation(
				"add", "org_unit_not_found", "org unit not found",
				map[string]any{"org_unit_id": string(*ouID)},
			))
			continue
		}
		path := ouRes.Data.GetPath()
		if len(path) == 0 || path[0] != string(*roleOrgID) {
			out = append(out, *ft.NewBusinessViolation(
				"add", "entitlement_org_mismatch",
				`Entitlement's org_unit_id {{org_unit_id}} must belong to the role's org_id {{org_id}}`,
				map[string]any{"org_unit_id": string(*ouID), "org_id": string(*roleOrgID)},
			))
		}
	}
	return out, nil
}

func entitlementIdsSearchGraph(ids []model.Id) *dmodel.SearchGraph {
	ops := make([]any, len(ids))
	for i := range ids {
		ops[i] = ids[i]
	}
	graph := dmodel.NewSearchGraph()
	graph.Condition(dmodel.NewCondition(basemodel.FieldId, dmodel.In, ops...))
	return graph
}

func (this *RoleDomainServiceImpl) RoleExists(
	ctx corectx.Context, query itRole.RoleExistsQuery,
) (*itRole.RoleExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if role exists",
		DbRepoGetter: this.roleRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *RoleDomainServiceImpl) SearchRoles(
	ctx corectx.Context, query itRole.SearchRolesQuery, options ...corecrud.ServiceSearchOptions,
) (*itRole.SearchRolesResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	return corecrud.Search[domain.Role](ctx, corecrud.SearchParam{
		Action:                 "search roles",
		DbRepoGetter:           this.roleRepo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *RoleDomainServiceImpl) SearchUserRoles(
	ctx corectx.Context, query itRole.SearchUserRolesQuery, options ...corecrud.ServiceSearchOptions,
) (result *itRole.SearchUserRolesResult, err error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "search user roles"); e != nil {
			err = e
		}
	}()
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itRole.SearchUserRolesResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itRole.SearchUserRolesQuery))

	return corecrud.Search[domain.Role](ctx, corecrud.SearchParam{
		Action:                 "search user roles",
		DbRepoGetter:           this.roleRepo,
		AfterValidationSuccess: opts.AfterValidationSuccess,
		Query: dyn.SearchQuery{
			Fields:          assignedRoleFields(query.Fields),
			Graph:           assignmentGraph(domain.RoleEdgeAssignedUsers, query.UserId, query.Graph),
			Page:            query.Page,
			Size:            query.Size,
			IncludeArchived: query.IncludeArchived,
		},
	})
}

func (this *RoleDomainServiceImpl) SearchGroupRoles(
	ctx corectx.Context, query itRole.SearchGroupRolesQuery, options ...corecrud.ServiceSearchOptions,
) (result *itRole.SearchGroupRolesResult, err error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "search group roles"); e != nil {
			err = e
		}
	}()
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itRole.SearchGroupRolesResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itRole.SearchGroupRolesQuery))

	return corecrud.Search[domain.Role](ctx, corecrud.SearchParam{
		Action:                 "search group roles",
		DbRepoGetter:           this.roleRepo,
		AfterValidationSuccess: opts.AfterValidationSuccess,
		Query: dyn.SearchQuery{
			Fields:          assignedRoleFields(query.Fields),
			Graph:           assignmentGraph(domain.RoleEdgeAssignedGroups, query.GroupId, query.Graph),
			Page:            query.Page,
			Size:            query.Size,
			IncludeArchived: query.IncludeArchived,
		},
	})
}

// assignedRoleFields defaults an assigned-role listing to the id and name the assignment UI
// needs, while still honouring an explicit field list from the caller.
func assignedRoleFields(fields []string) []string {
	if len(fields) > 0 {
		return fields
	}
	return []string{basemodel.FieldId, domain.RoleFieldName}
}

// assignmentGraph ANDs "assigned to this principal" onto the caller's own graph. `linked` is
// the operator for membership on a many edge: role<->user and role<->group are both
// many-to-many through a junction table, so no column equality can express it.
func assignmentGraph(edge string, principalId model.Id, caller *dmodel.SearchGraph) *dmodel.SearchGraph {
	cond := dmodel.NewCondition(edge, dmodel.Linked, principalId)
	graph := dmodel.NewSearchGraph()
	if caller == nil {
		graph.Condition(cond)
		return graph
	}
	graph.And(*dmodel.NewSearchNode().Condition(cond), *caller.ToSearchNode())
	return graph
}

func (this *RoleDomainServiceImpl) SetRoleIsArchived(
	ctx corectx.Context, cmd itRole.SetRoleIsArchivedCommand,
) (*itRole.SetRoleIsArchivedResult, error) {
	result, err := corecrud.SetIsArchived(ctx, this.roleRepo, dyn.SetIsArchivedCommand(cmd))
	if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
		return result, err
	}
	// Archiving a role revokes it. Nothing cascades - the rebuild function filters
	// archived roles out - so without this the role keeps answering for everyone
	// who already holds it.
	if err := this.permRepo.RebuildUserPermissionsForRole(ctx, cmd.Id); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *RoleDomainServiceImpl) UpdateRole(
	ctx corectx.Context, cmd itRole.UpdateRoleCommand, options ...corecrud.ServiceUpdateOptions[*domain.Role],
) (*itRole.UpdateRoleResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceUpdateOptions[*domain.Role]{})
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Role, *domain.Role]{
		Action:                 "update role",
		DbRepoGetter:           this.roleRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}
