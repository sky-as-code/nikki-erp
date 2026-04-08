package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itRole "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/role"
	"go.bryk.io/pkg/errors"
)

func NewGroupServiceImpl(
	groupRepo2 itGrp.GroupRepository,
	roleSvc itRole.RoleService,
	cqrsBus cqrs.CqrsBus,
) itGrp.GroupService {
	return &GroupServiceImpl{
		cqrsBus:    cqrsBus,
		groupRepo2: groupRepo2,
		roleSvc:    roleSvc,
	}
}

type GroupServiceImpl struct {
	cqrsBus    cqrs.CqrsBus
	groupRepo2 itGrp.GroupRepository
	roleSvc    itRole.RoleService
}

func (this *GroupServiceImpl) CreateGroup(
	ctx corectx.Context, cmd itGrp.CreateGroupCommand,
) (*itGrp.CreateGroupResult, error) {
	return corecrud.ExecInTranx(ctx, this.groupRepo2, func(tranxCtx corectx.Context) (*itGrp.CreateGroupResult, error) {
		result, err := corecrud.Create(tranxCtx, corecrud.CreateParam[domain.Group, *domain.Group]{
			Action:         "create group",
			BaseRepoGetter: this.groupRepo2,
			Data:           cmd,
		})
		if err != nil {
			return nil, err
		}
		if result.ClientErrors.Count() > 0 {
			return result, nil
		}
		return this.createPrivateRole(tranxCtx, result)
	})
}

func (this *GroupServiceImpl) createPrivateRole(tranxCtx corectx.Context, grpResult *itGrp.CreateGroupResult) (*itGrp.CreateGroupResult, error) {
	oid := string(*grpResult.Data.GetId())
	roleRes, rErr := this.roleSvc.CreatePrivateRole(tranxCtx, itRole.CreatePrivateRoleCommand{OwnerId: oid})
	if rErr != nil {
		return nil, rErr
	}
	if roleRes.ClientErrors.Count() > 0 {
		return nil, errors.Errorf("create private role: %v", roleRes.ClientErrors)
	}
	return grpResult, nil
}

func (this *GroupServiceImpl) DeleteGroup(ctx corectx.Context, cmd itGrp.DeleteGroupCommand) (
	*itGrp.DeleteGroupResult, error,
) {
	return corecrud.ExecInTranx(ctx, this.groupRepo2, func(tranxCtx corectx.Context) (*itGrp.DeleteGroupResult, error) {
		privRes, pErr := this.roleSvc.DeletePrivateRole(tranxCtx, itRole.DeletePrivateRoleCommand{OwnerId: cmd.Id})
		if pErr != nil {
			return nil, pErr
		}
		if privRes.ClientErrors.Count() > 0 {
			return nil, errors.Errorf("delete private role: %v", privRes.ClientErrors)
		}
		return corecrud.DeleteOne(tranxCtx, corecrud.DeleteOneParam{
			Action:       "delete group",
			DbRepoGetter: this.groupRepo2,
			Cmd:          dyn.DeleteOneCommand{Id: cmd.Id},
		})
	})
}

func (this *GroupServiceImpl) GroupExists(ctx corectx.Context, query itGrp.GroupExistsQuery) (
	*itGrp.GroupExistsResult, error,
) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if group exists",
		DbRepoGetter: this.groupRepo2,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *GroupServiceImpl) GetGroup(
	ctx corectx.Context, query itGrp.GetGroupQuery,
) (*itGrp.GetGroupResult, error) {
	return corecrud.GetOne[domain.Group](ctx, corecrud.GetOneParam{
		Action:       "get group",
		DbRepoGetter: this.groupRepo2,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *GroupServiceImpl) ManageGroupUsers(
	ctx corectx.Context, cmd itGrp.ManageGroupUsersCommand,
) (result *itGrp.ManageGroupUsersResult, err error) {

	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage group users",
		DbRepoGetter:       this.groupRepo2,
		DestSchemaName:     domain.UserSchemaName,
		SrcId:              cmd.GroupId,
		SrcIdFieldForError: "group_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
	})
}

func (this *GroupServiceImpl) SearchGroups(ctx corectx.Context, query itGrp.SearchGroupsQuery) (
	*itGrp.SearchGroupsResult, error,
) {
	return corecrud.Search[domain.Group](ctx, corecrud.SearchParam{
		Action:       "search groups",
		DbRepoGetter: this.groupRepo2,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *GroupServiceImpl) UpdateGroup(
	ctx corectx.Context, cmd itGrp.UpdateGroupCommand,
) (*itGrp.UpdateGroupResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Group, *domain.Group]{
		Action:       "update group",
		DbRepoGetter: this.groupRepo2,
		Data:         cmd,
	})
}
