package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

func NewGroupServiceImpl(
	groupRepo2 itGrp.GroupRepository,
	cqrsBus cqrs.CqrsBus,
) itGrp.GroupService {
	return &GroupServiceImpl{
		cqrsBus:    cqrsBus,
		groupRepo2: groupRepo2,
	}
}

type GroupServiceImpl struct {
	cqrsBus    cqrs.CqrsBus
	groupRepo2 itGrp.GroupRepository
}

func (this *GroupServiceImpl) CreateGroup(
	ctx corectx.Context, cmd itGrp.CreateGroupCommand,
) (*itGrp.CreateGroupResult, error) {
	return corecrud.Create(ctx, dyn.CreateParam[domain.Group, *domain.Group]{
		Action:         "create group",
		BaseRepoGetter: this.groupRepo2,
		Data:           cmd,
	})
}

func (this *GroupServiceImpl) DeleteGroup(ctx corectx.Context, cmd itGrp.DeleteGroupCommand) (
	*itGrp.DeleteGroupResult, error,
) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete group",
		DbRepoGetter: this.groupRepo2,
		Cmd:          dyn.DeleteOneCommand{Id: cmd.Id},
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
		Action:         "manage group users",
		DbRepoGetter:   this.groupRepo2,
		DestSchemaName: domain.UserSchemaName,
		Associations:   groupUserAssocs(cmd.GroupId, cmd.Add.ToSlice()),
		Desociations:   groupUserAssocs(cmd.GroupId, cmd.Remove.ToSlice()),
	})
}

func groupUserAssocs(groupId model.Id, userIds []model.Id) []dyn.RepoM2mAssociation {
	out := make([]dyn.RepoM2mAssociation, 0, len(userIds))
	for _, uid := range userIds {
		out = append(out, dyn.RepoM2mAssociation{
			SrcKeys:  dmodel.DynamicFields{basemodel.FieldId: groupId},
			DestKeys: dmodel.DynamicFields{basemodel.FieldId: uid},
		})
	}
	return out
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
