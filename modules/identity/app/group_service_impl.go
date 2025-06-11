package app

import (
	"context"
	"strings"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"

	// util "github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewGroupServiceImpl(groupRepo itGrp.GroupRepository, cqrsBus cqrs.CqrsBus) itGrp.GroupService {
	return &GroupServiceImpl{
		cqrsBus:   cqrsBus,
		groupRepo: groupRepo,
	}
}

type GroupServiceImpl struct {
	cqrsBus   cqrs.CqrsBus
	groupRepo itGrp.GroupRepository
}

func (this *GroupServiceImpl) AddRemoveUsers(ctx context.Context, cmd itGrp.AddRemoveUsersCommand) (result *itGrp.AddRemoveUsersResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to add or remove users"); e != nil {
			err = e
		}
	}()

	if len(cmd.Add) == 0 && len(cmd.Remove) == 0 {
		return &itGrp.AddRemoveUsersResult{
			ClientError: &ft.ClientError{
				Code:    "invalid_request",
				Details: "no users to add or remove",
			},
		}, nil
	}

	valErrs := cmd.Validate()
	dbGroup := this.assertGroupIdExists(ctx, valErrs, cmd.GroupId)
	this.assertSameGroupEtag(valErrs, dbGroup, cmd.Etag)
	this.assertUserIdsExist(ctx, valErrs, "add", cmd.Add)
	this.assertUserIdsExist(ctx, valErrs, "remove", cmd.Remove)

	if valErrs.Count() > 0 {
		return &itGrp.AddRemoveUsersResult{
			ClientError: valErrs.ToClientError(),
		}, nil
	}

	cmd.Etag = *model.NewEtag()
	clientErr, err := this.groupRepo.AddRemoveUsers(ctx, cmd)
	ft.PanicOnErr(err)

	// TODO: The group or users may have been deleted by another process
	if clientErr != nil {
		return &itGrp.AddRemoveUsersResult{
			ClientError: clientErr,
		}, nil
	}

	return &itGrp.AddRemoveUsersResult{
		Data: &itGrp.AddRemoveUsersResultData{
			UpdatedAt: time.Now(),
		},
	}, nil
}

func (this *GroupServiceImpl) assertUserIdsExist(ctx context.Context, valErrs ft.ValidationErrors, field string, userIds []string) {
	if valErrs.Count() > 0 || len(userIds) == 0 {
		return
	}

	existCmd := &itUser.UserExistsMultiCommand{
		Ids: userIds,
	}
	existRes := itUser.UserExistsMultiResult{}
	err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
	ft.PanicOnErr(err)

	if existRes.ClientError != nil {
		valErrs.MergeClientError(existRes.ClientError)
		return
	}

	if len(existRes.Data.NotExisting) > 0 {
		valErrs.Append(field, "not existing users: "+strings.Join(existRes.Data.NotExisting, ", "))
	}
}

func (this *GroupServiceImpl) CreateGroup(ctx context.Context, cmd itGrp.CreateGroupCommand) (result *itGrp.CreateGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create group"); e != nil {
			err = e
		}
	}()

	group := cmd.ToGroup()
	this.setGroupDefaults(group)

	valErr := group.Validate(false)
	this.assertUniqueGroupName(ctx, valErr, *group.Name)

	if valErr.Count() > 0 {
		return &itGrp.CreateGroupResult{
			ClientError: valErr.ToClientError(),
		}, nil
	}

	createdGroup, err := this.groupRepo.Create(ctx, *group)
	ft.PanicOnErr(err)

	return &itGrp.CreateGroupResult{Data: createdGroup}, err
}

func (this *GroupServiceImpl) setGroupDefaults(group *domain.Group) {
	id, err := model.NewId()
	ft.PanicOnErr(err)
	group.Id = id
	group.Etag = model.NewEtag()
}

func (this *GroupServiceImpl) UpdateGroup(ctx context.Context, cmd itGrp.UpdateGroupCommand) (result *itGrp.UpdateGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update group"); e != nil {
			err = e
		}
	}()

	group := cmd.ToGroup()
	valErrs := group.Validate(true)
	dbGroup := this.assertGroupIdExists(ctx, valErrs, cmd.Id)
	this.assertSameGroupEtag(valErrs, dbGroup, *group.Etag)
	this.assertUniqueGroupName(ctx, valErrs, *group.Name)

	if valErrs.Count() > 0 {
		return &itGrp.UpdateGroupResult{
			ClientError: valErrs.ToClientError(),
		}, nil
	}

	group.Etag = model.NewEtag()
	groupWithOrg, err := this.groupRepo.Update(ctx, *group)
	ft.PanicOnErr(err)

	return &itGrp.UpdateGroupResult{Data: groupWithOrg}, err
}

func (this *GroupServiceImpl) DeleteGroup(ctx context.Context, cmd itGrp.DeleteGroupCommand) (result *itGrp.DeleteGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete group"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	this.assertGroupIdExists(ctx, vErrs, cmd.Id)

	if vErrs.Count() > 0 {
		return &itGrp.DeleteGroupResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	err = this.groupRepo.Delete(ctx, cmd)
	ft.PanicOnErr(err)

	return &itGrp.DeleteGroupResult{
		Data: itGrp.DeleteGroupResultData{
			DeletedAt: time.Now(),
		},
	}, nil
}

func (this *GroupServiceImpl) GetGroupById(ctx context.Context, query itGrp.GetGroupByIdQuery) (result *itGrp.GetGroupByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get group"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	dbGroup := this.assertGroupIdExists(ctx, vErrs, query.Id)

	if vErrs.Count() > 0 {
		return &itGrp.GetGroupByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &itGrp.GetGroupByIdResult{
		Data: dbGroup,
	}, nil
}

func (thisSvc *GroupServiceImpl) SearchGroups(ctx context.Context, query itGrp.SearchGroupsQuery) (result *itGrp.SearchGroupsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list groups"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := thisSvc.groupRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &itGrp.SearchGroupsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	groups, err := thisSvc.groupRepo.Search(ctx, itGrp.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
		WithOrg:   query.WithOrg,
	})
	ft.PanicOnErr(err)

	return &itGrp.SearchGroupsResult{
		Data: groups,
	}, nil
}

func (this *GroupServiceImpl) assertGroupIdExists(ctx context.Context, valErrs ft.ValidationErrors, id string) *domain.Group {
	if valErrs.Count() > 0 {
		return nil
	}

	dbGroup, err := this.groupRepo.FindById(ctx, itGrp.FindByIdParam{
		Id: id,
	})
	ft.PanicOnErr(err)

	if dbGroup == nil {
		valErrs.Append("id", "group not found")
		return nil
	}

	return dbGroup
}

func (this *GroupServiceImpl) assertSameGroupEtag(valErrs ft.ValidationErrors, dbGroup *domain.Group, newEtag string) {
	if valErrs.Count() > 0 {
		return
	}

	if *dbGroup.Etag != newEtag {
		valErrs.Append("etag", "group has been modified by another user")
	}
}

func (this *GroupServiceImpl) assertUniqueGroupName(ctx context.Context, valErrs ft.ValidationErrors, name string) {
	if valErrs.Count() > 0 {
		return
	}

	dbGroup, err := this.groupRepo.FindByName(ctx, itGrp.FindByNameParam{
		Name: name,
	})
	ft.PanicOnErr(err)

	if dbGroup != nil {
		valErrs.Append("name", "group name already exists")
	}
}
