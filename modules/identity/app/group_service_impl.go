package app

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"

	// util "github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

func NewGroupServiceImpl(groupRepo it.GroupRepository) it.GroupService {
	return &GroupServiceImpl{
		groupRepo: groupRepo,
	}
}

type GroupServiceImpl struct {
	groupRepo it.GroupRepository
}

func (this *GroupServiceImpl) CreateGroup(ctx context.Context, cmd it.CreateGroupCommand) (result *it.CreateGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create group"); e != nil {
			err = e
		}
	}()

	group := cmd.ToGroup()
	this.setGroupDefaults(group)

	valErr := group.Validate(false)
	if valErr.Count() > 0 {
		return &it.CreateGroupResult{
			ClientError: valErr.ToClientError(),
		}, nil
	}

	dbGroup, err := this.groupRepo.FindByName(ctx, *group.Name)
	ft.PanicOnErr(err)

	if dbGroup != nil {
		vErrs := ft.NewValidationErrors()
		vErrs.Append("name", "group name already exists")
		return &it.CreateGroupResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdGroup, err := this.groupRepo.Create(ctx, *group)
	ft.PanicOnErr(err)

	return &it.CreateGroupResult{Data: createdGroup}, err
}

func (this *GroupServiceImpl) setGroupDefaults(group *domain.Group) {
	id, err := model.NewId()
	ft.PanicOnErr(err)
	group.Id = id
	group.Etag = model.NewEtag()
}

func (this *GroupServiceImpl) UpdateGroup(ctx context.Context, cmd it.UpdateGroupCommand) (result *it.UpdateGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update group"); e != nil {
			err = e
		}
	}()

	group := cmd.ToGroup()

	valErr := group.Validate(true)
	if valErr.Count() > 0 {
		return &it.UpdateGroupResult{
			ClientError: valErr.ToClientError(),
		}, nil
	}

	dbGroup, err := this.groupRepo.FindById(ctx, it.FindByIdParam{
		Id: cmd.Id,
	})
	ft.PanicOnErr(err)

	if dbGroup == nil {
		vErrors := ft.NewValidationErrors()
		vErrors.Append("id", "group not found")

		return &it.UpdateGroupResult{
			ClientError: vErrors.ToClientError(),
		}, nil

	}

	if dbGroup.Etag.String() != group.Etag.String() {
		vErrors := ft.NewValidationErrors()
		vErrors.Append("etag", "group has been modified by another user")
		return &it.UpdateGroupResult{
			ClientError: vErrors.ToClientError(),
		}, nil
	}

	dbGroup, err = this.groupRepo.FindByName(ctx, *group.Name)
	ft.PanicOnErr(err)

	if dbGroup != nil {
		vErrors := ft.NewValidationErrors()
		vErrors.Append("name", "group with this name already exists")
		return &it.UpdateGroupResult{
			ClientError: vErrors.ToClientError(),
		}, nil
	}

	group.Etag = model.NewEtag()
	groupWithOrg, err := this.groupRepo.Update(ctx, *group)
	ft.PanicOnErr(err)

	return &it.UpdateGroupResult{Data: groupWithOrg}, err
}

func (this *GroupServiceImpl) DeleteGroup(ctx context.Context, cmd it.DeleteGroupCommand) (result *it.DeleteGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete group"); e != nil {
			err = e
		}
	}()

	dbGroup, err := this.groupRepo.FindById(ctx, it.FindByIdParam{
		Id: cmd.Id,
	})
	ft.PanicOnErr(err)

	if dbGroup == nil {
		vErrors := ft.NewValidationErrors()
		vErrors.Append("id", "group not found")
		return &it.DeleteGroupResult{
			ClientError: vErrors.ToClientError(),
		}, nil
	}

	err = this.groupRepo.Delete(ctx, model.Id(cmd.Id))
	ft.PanicOnErr(err)

	return &it.DeleteGroupResult{
		Data: it.DeleteGroupResultData{
			DeletedAt: time.Now(),
		},
	}, nil
}

func (this *GroupServiceImpl) GetGroupById(ctx context.Context, query it.GetGroupByIdQuery) (result *it.GetGroupByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get group"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetGroupByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbGroup, err := this.groupRepo.FindById(ctx, query)
	ft.PanicOnErr(err)

	if dbGroup == nil {
		vErrs.Append("id", "group not found")
		return &it.GetGroupByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetGroupByIdResult{
		Data: dbGroup,
	}, nil

}
