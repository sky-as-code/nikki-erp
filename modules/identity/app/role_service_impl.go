package app

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itEnt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/entitlement"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itRole "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/role"
)

func NewRoleServiceImpl(
	roleRepo itRole.RoleRepository,
	entitlementRepo itEnt.EntitlementRepository,
	orgUnitSvc itOrg.OrgUnitService,
	cqrsBus cqrs.CqrsBus,
) itRole.RoleService {
	return &RoleServiceImpl{
		cqrsBus:         cqrsBus,
		roleRepo:        roleRepo,
		entitlementRepo: entitlementRepo,
		orgUnitSvc:      orgUnitSvc,
	}
}

type RoleServiceImpl struct {
	cqrsBus         cqrs.CqrsBus
	roleRepo        itRole.RoleRepository
	entitlementRepo itEnt.EntitlementRepository
	orgUnitSvc      itOrg.OrgUnitService
}

func (this *RoleServiceImpl) CreateRole(
	ctx corectx.Context, cmd itRole.CreateRoleCommand,
) (*itRole.CreateRoleResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Role, *domain.Role]{
		Action:         "create role",
		BaseRepoGetter: this.roleRepo,
		Data:           cmd,
	})
}

func (this *RoleServiceImpl) CreatePrivateRole(
	ctx corectx.Context, cmd itRole.CreatePrivateRoleCommand,
) (*itRole.CreateRoleResult, error) {
	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &itRole.CreateRoleResult{ClientErrors: cErrs}, nil
	}
	cmd = *(sanitized.(*itRole.CreatePrivateRoleCommand))

	var newRole *domain.Role
	ownerId := string(cmd.OwnerId)
	if cmd.OwnerType == "user" {
		newRole = domain.NewRoleFrom(dmodel.DynamicFields{
			domain.RoleFieldName:          fmt.Sprintf("Private role for user %s", ownerId),
			domain.RoleFieldIsPrivate:     true,
			domain.RoleFieldOwnerUserId:   ownerId,
			domain.RoleFieldIsRequestable: false, // Important!
		})
	} else {
		newRole = domain.NewRoleFrom(dmodel.DynamicFields{
			domain.RoleFieldName:          fmt.Sprintf("Private role for group %s", ownerId),
			domain.RoleFieldIsPrivate:     true,
			domain.RoleFieldOwnerGroupId:  ownerId,
			domain.RoleFieldIsRequestable: false, // Important!
		})
	}
	createCmd := itRole.CreateRoleCommand{Role: *newRole}
	createRes, err := this.CreateRole(ctx, createCmd)
	if err != nil {
		return nil, err
	}
	if createRes.ClientErrors.Count() > 0 {
		return createRes, nil
	}
	return createRes, nil
}

func (this *RoleServiceImpl) DeleteRole(
	ctx corectx.Context, cmd itRole.DeleteRoleCommand,
) (*itRole.DeleteRoleResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete role",
		DbRepoGetter: this.roleRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
		ValidateExtra: func(ctx corectx.Context, keyFields dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
			resRole, err := this.roleRepo.GetOne(ctx, dyn.RepoGetOneParam{
				Filter:  keyFields,
				Columns: []string{domain.RoleFieldIsPrivate},
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
					ft.ErrorKey("authorize", "err_private_role_deletion_not_allowed"),
					"private role deletion is not allowed, it's automatically removed when its owner is deleted.",
				)
			}
			return nil
		},
	})
}

func (this *RoleServiceImpl) DeletePrivateRole(
	ctx corectx.Context, cmd itRole.DeletePrivateRoleCommand,
) (*itRole.DeleteRoleResult, error) {
	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &itRole.DeleteRoleResult{ClientErrors: cErrs}, nil
	}
	cmd = *(sanitized.(*itRole.DeletePrivateRoleCommand))

	searchRes, err := this.roleRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:   privateRoleOwnerSearchGraph(cmd.OwnerId),
		Columns: []string{basemodel.FieldId},
		Page:    0,
		Size:    1,
	})
	if err != nil {
		return nil, err
	}
	if searchRes.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: searchRes.ClientErrors}, nil
	}
	if !searchRes.HasData {
		return &dyn.OpResult[dyn.MutateResultData]{HasData: false}, nil
	}

	delId := searchRes.Data.Items[0].GetId()

	// TODO: Build query DELETE FROM (SELECT * FROM roles WHERE dedicated_group_id = $1 OR dedicated_user_id = $1 LIMIT 1)
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete private role",
		DbRepoGetter: this.roleRepo,
		Cmd:          dyn.DeleteOneCommand{Id: *delId},
	})
}

func privateRoleOwnerSearchGraph(ownerId model.Id) *dmodel.SearchGraph {
	oid := string(ownerId)
	userNode := dmodel.NewSearchNode().NewCondition(domain.RoleFieldOwnerUserId, dmodel.Equals, oid)
	groupNode := dmodel.NewSearchNode().NewCondition(domain.RoleFieldOwnerGroupId, dmodel.Equals, oid)
	nodeIsPrivate := dmodel.NewSearchNode().NewCondition(domain.RoleFieldIsPrivate, dmodel.Equals, true)
	nodeOwner := dmodel.NewSearchNode().Or(*userNode, *groupNode)
	graph := dmodel.NewSearchGraph()
	graph.And(*nodeIsPrivate, *nodeOwner)
	return graph
}

func (this *RoleServiceImpl) GetRole(
	ctx corectx.Context, query itRole.GetRoleQuery,
) (*itRole.GetRoleResult, error) {
	return corecrud.GetOne[domain.Role](ctx, corecrud.GetOneParam{
		Action:       "get role",
		DbRepoGetter: this.roleRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *RoleServiceImpl) ManageRoleEntitlements(
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
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage role entitlements",
		DbRepoGetter:       this.roleRepo,
		DestSchemaName:     domain.EntitlementSchemaName,
		SrcId:              cmd.RoleId,
		SrcIdFieldForError: "role_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
	})
}

func (this *RoleServiceImpl) validateAddsBelongToRoleOrg(
	ctx corectx.Context, roleID model.Id, add datastructure.Set[model.Id],
) (ft.ClientErrors, error) {
	var out ft.ClientErrors
	addIDs := add.ToSlice()
	roleRes, err := corecrud.GetOne[domain.Role](ctx, corecrud.GetOneParam{
		Action:       "validate role for entitlement adds",
		DbRepoGetter: this.roleRepo,
		Query: dyn.GetOneQuery{
			Id:      roleID,
			Columns: []string{domain.RoleFieldOrgId},
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
		Graph:   entitlementIdsSearchGraph(addIDs),
		Columns: []string{basemodel.FieldId, domain.EntitlementFieldOrgUnitId},
		Page:    0,
		Size:    len(addIDs),
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
			Id:      *ouID,
			Columns: []string{domain.OrgUnitFieldPath},
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
				`Entitlement's org_unit_id {{.org_unit_id}} must belong to the role's org_id {{.org_id}}`,
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

// func (this *RoleServiceImpl) ManageRoleAssignments(
// 	ctx corectx.Context, cmd itRole.ManageRoleEntitlementsCommand,
// ) (*itRole.ManageRoleEntitlementsResult, error) {
// 	sanitizedCmd, cErrs := corecrud.Validate(cmd)
// 	if cErrs.Count() > 0 {
// 		return &itRole.ManageRoleEntitlementsResult{ClientErrors: cErrs}, nil
// 	}

// 	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
// 		Action:             "manage role assignments",
// 		DbRepoGetter:       this.roleRepo,
// 		DestSchemaName:     domain.RoleAssignmentSchemaName,
// 		SrcId:              cmd.RoleId,
// 		SrcIdFieldForError: "role_id",
// 		AssociatedIds:      cmd.Add,
// 		DisassociatedIds:   cmd.Remove,
// 	})
// }

func (this *RoleServiceImpl) RoleExists(
	ctx corectx.Context, query itRole.RoleExistsQuery,
) (*itRole.RoleExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if role exists",
		DbRepoGetter: this.roleRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *RoleServiceImpl) SearchRoles(
	ctx corectx.Context, query itRole.SearchRolesQuery,
) (*itRole.SearchRolesResult, error) {
	return corecrud.Search[domain.Role](ctx, corecrud.SearchParam{
		Action:       "search roles",
		DbRepoGetter: this.roleRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *RoleServiceImpl) SetRoleIsArchived(
	ctx corectx.Context, cmd itRole.SetRoleIsArchivedCommand,
) (*itRole.SetRoleIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.roleRepo, dyn.SetIsArchivedCommand(cmd))
}

func (this *RoleServiceImpl) UpdateRole(
	ctx corectx.Context, cmd itRole.UpdateRoleCommand,
) (*itRole.UpdateRoleResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Role, *domain.Role]{
		Action:       "update role",
		DbRepoGetter: this.roleRepo,
		Data:         cmd,
	})
}
