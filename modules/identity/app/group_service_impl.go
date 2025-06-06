package app

import (
	"context"
	"time"

	"go.bryk.io/pkg/errors"

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
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to create group")
		}
	}()

	group := cmd.ToGroup()
	this.setGroupDefaults(group)

	valErr := group.Validate(false)
	if valErr.Count() > 0 {
		return &it.CreateGroupResult{
			ClientError: ft.WrapValidationErrors(valErr),
		}, nil
	}

	groupFromDb, err := this.groupRepo.FindByName(ctx, group.Name)
	ft.PanicOnErr(err)

	if groupFromDb != nil {
		vErrors := ft.NewValidationErrors()
		vErrors.Append(ft.ValidationErrorItem{
			Field: "name",
			Error: "group name already exists",
		})
		return &it.CreateGroupResult{
			ClientError: ft.WrapValidationErrors(vErrors),
		}, nil
	}

	groupWithOrg, err := this.groupRepo.Create(ctx, *group)
	ft.PanicOnErr(err)

	return &it.CreateGroupResult{Data: groupWithOrg}, err
}

func (this *GroupServiceImpl) setGroupDefaults(group *domain.Group) {
	id, err := model.NewId()
	ft.PanicOnErr(err)
	group.Id = id
	group.Etag = model.NewEtag()
}

func (this *GroupServiceImpl) UpdateGroup(ctx context.Context, cmd it.UpdateGroupCommand) (result *it.UpdateGroupResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to update group")
		}
	}()

	group := cmd.ToGroup()

	valErr := group.Validate(true)
	if valErr.Count() > 0 {
		return &it.UpdateGroupResult{
			ClientError: ft.WrapValidationErrors(valErr),
		}, nil
	}

	groupFromDb, err := this.groupRepo.FindById(ctx, model.Id(cmd.Id), false)
	ft.PanicOnErr(err)

	if groupFromDb == nil {
		vErrors := ft.NewValidationErrors()
		vErrors.Append(ft.ValidationErrorItem{
			Field: "id",
			Error: "group not found",
		})

		return &it.UpdateGroupResult{
			ClientError: ft.WrapValidationErrors(vErrors),
		}, nil

	}

	if *groupFromDb.Group.Etag != *group.Etag {
		vErrors := ft.NewValidationErrors()
		vErrors.Append(ft.ValidationErrorItem{
			Field: "etag",
			Error: "group has been modified by another user",
		})
		return &it.UpdateGroupResult{
			ClientError: ft.WrapValidationErrors(vErrors),
		}, nil
	}

	groupFromDb, err = this.groupRepo.FindByName(ctx, group.Name)
	ft.PanicOnErr(err)

	if groupFromDb != nil {
		vErrors := ft.NewValidationErrors()
		vErrors.Append(ft.ValidationErrorItem{
			Field: "name",
			Error: "group with this name already exists",
		})
		return &it.UpdateGroupResult{
			ClientError: ft.WrapValidationErrors(vErrors),
		}, nil
	}

	group.Etag = model.NewEtag()
	groupWithOrg, err := this.groupRepo.Update(ctx, *group)
	ft.PanicOnErr(err)

	return &it.UpdateGroupResult{Data: groupWithOrg}, err
}

func (this *GroupServiceImpl) DeleteGroup(ctx context.Context, Id string, deletedBy string) (result *it.DeleteGroupResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to delete group")
		}
	}()

	group, err := this.groupRepo.FindById(ctx, model.Id(Id), false)
	ft.PanicOnErr(err)

	if group == nil {
		vErrors := ft.NewValidationErrors()
		vErrors.Append(ft.ValidationErrorItem{
			Field: "id",
			Error: "group not found",
		})
		return &it.DeleteGroupResult{
			ClientError: ft.WrapValidationErrors(vErrors),
		}, nil
	}

	err = this.groupRepo.Delete(ctx, model.Id(Id))
	ft.PanicOnErr(err)

	return &it.DeleteGroupResult{
		Data: it.DeleteGroupResultData{
			DeletedAt: time.Now(),
		},
	}, nil
}

func (this *GroupServiceImpl) GetGroupByID(ctx context.Context, Id string, withOrg bool) (result *it.GetGroupByIdResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to get group")
		}
	}()

	vErrors := ft.NewValidationErrors()
	if Id == "" {
		vErrors.Append(ft.ValidationErrorItem{
			Field: "id",
			Error: "group ID cannot be empty",
		})
		return &it.GetGroupByIdResult{
			ClientError: ft.WrapValidationErrors(vErrors),
		}, nil
	}

	group, err := this.groupRepo.FindById(ctx, model.Id(Id), withOrg)
	ft.PanicOnErr(err)

	if group == nil {
		vErrors.Append(ft.ValidationErrorItem{
			Field: "id",
			Error: "group not found",
		})
		return &it.GetGroupByIdResult{
			ClientError: ft.WrapValidationErrors(vErrors),
		}, nil
	}

	return &it.GetGroupByIdResult{
		Data: group,
	}, nil

}
