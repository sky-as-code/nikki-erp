package app

import (
	"context"
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
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
		if e := ft.RecoverPanicFailedTo(recover(), "add or remove users"); e != nil {
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

	var dbGroup *domain.Group
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbGroup, err = this.assertGroupExists(ctx, cmd.GroupId, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertCorrectEtag(cmd.Etag, *dbGroup.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertUserIdsExist(ctx, vErrs, "add", cmd.Add)
		}).
		End()

	if vErrs.Count() > 0 {
		return &itGrp.AddRemoveUsersResult{
			ClientError: vErrs.ToClientError(),
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
			Id:        cmd.GroupId,
			Etag:      cmd.Etag,
			UpdatedAt: time.Now(),
		},
		HasData: true,
	}, nil
}

func (this *GroupServiceImpl) assertUserIdsExist(ctx context.Context, valErrs *ft.ValidationErrors, field string, userIds []string) error {
	if len(userIds) == 0 {
		return nil
	}

	existCmd := &itUser.UserExistsMultiCommand{
		Ids: userIds,
	}
	existRes := itUser.UserExistsMultiResult{}
	err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
	if err != nil {
		return err
	}

	if existRes.ClientError != nil {
		valErrs.MergeClientError(existRes.ClientError)
		return nil
	}

	if len(existRes.Data.NotExisting) > 0 {
		valErrs.Append(field, "not existing users: "+strings.Join(existRes.Data.NotExisting, ", "))
	}
	return nil
}

func (this *GroupServiceImpl) CreateGroup(ctx context.Context, cmd itGrp.CreateGroupCommand) (result *itGrp.CreateGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create group"); e != nil {
			err = e
		}
	}()

	group := cmd.ToGroup()
	group.SetDefaults()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = group.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeGroup(group)
			return this.assertUniqueGroupName(ctx, group, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrp.CreateGroupResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdGroup, err := this.groupRepo.Create(ctx, *group)
	ft.PanicOnErr(err)

	return &itGrp.CreateGroupResult{
		Data:    createdGroup,
		HasData: createdGroup != nil,
	}, err
}

func (this *GroupServiceImpl) sanitizeGroup(group *domain.Group) {
	if group.Name != nil {
		cleanedName := strings.TrimSpace(*group.Name)
		cleanedName = defense.SanitizePlainText(cleanedName)
		group.Name = &cleanedName
	}
}

func (this *GroupServiceImpl) UpdateGroup(ctx context.Context, cmd itGrp.UpdateGroupCommand) (result *itGrp.UpdateGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "update group"); e != nil {
			err = e
		}
	}()

	group := cmd.ToGroup()

	var dbGroup *domain.Group
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = group.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbGroup, err = this.assertGroupExists(ctx, *group.Id, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertCorrectEtag(*group.Etag, *dbGroup.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			// Sanitize after we've made sure this is the correct group
			this.sanitizeGroup(group)
			return this.assertUniqueGroupName(ctx, group, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrp.UpdateGroupResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := group.Etag
	group.Etag = model.NewEtag()
	groupWithOrg, err := this.groupRepo.Update(ctx, *group, *prevEtag)
	ft.PanicOnErr(err)

	return &itGrp.UpdateGroupResult{
		Data:    groupWithOrg,
		HasData: groupWithOrg != nil,
	}, err
}

func (this *GroupServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *ft.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *GroupServiceImpl) assertGroupExists(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (dbGroup *domain.Group, err error) {
	dbGroup, err = this.groupRepo.FindById(ctx, itGrp.FindByIdParam{Id: id})
	if dbGroup == nil {
		vErrs.AppendNotFound("id", "group id")
	}
	return
}

func (this *GroupServiceImpl) assertUniqueGroupName(ctx context.Context, group *domain.Group, vErrs *ft.ValidationErrors) error {
	dbGroup, err := this.groupRepo.FindByName(ctx, itGrp.FindByNameParam{
		Name: *group.Name,
	})
	if err != nil {
		return err
	}

	if dbGroup != nil {
		vErrs.AppendAlreadyExists("name", "group name")
	}
	return nil
}

func (this *GroupServiceImpl) DeleteGroup(ctx context.Context, cmd itGrp.DeleteGroupCommand) (result *itGrp.DeleteGroupResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "delete group"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()

	if vErrs.Count() > 0 {
		return &itGrp.DeleteGroupResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	deletedCount, err := this.groupRepo.DeleteHard(ctx, cmd)
	ft.PanicOnErr(err)
	if deletedCount == 0 {
		vErrs.AppendNotFound("id", "group id")
		return &itGrp.DeleteGroupResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount), nil
}

func (this *GroupServiceImpl) GetGroupById(ctx context.Context, query itGrp.GetGroupByIdQuery) (result *itGrp.GetGroupByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get group by Id"); e != nil {
			err = e
		}
	}()

	var dbGroup *domain.Group
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbGroup, err = this.assertGroupExists(ctx, query.Id, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrp.GetGroupByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &itGrp.GetGroupByIdResult{
		Data:    dbGroup,
		HasData: dbGroup != nil,
	}, nil
}

func (thisSvc *GroupServiceImpl) SearchGroups(ctx context.Context, query itGrp.SearchGroupsQuery) (result *itGrp.SearchGroupsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "list groups"); e != nil {
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
		Data:    groups,
		HasData: groups.Items != nil,
	}, nil
}

func (this *GroupServiceImpl) Exist(ctx context.Context, cmd itGrp.GroupExistsCommand) (result *itGrp.GroupExistsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "check if group exists"); e != nil {
			err = e
		}
	}()

	exists, err := this.groupRepo.Exists(ctx, cmd)
	ft.PanicOnErr(err)

	return &itGrp.GroupExistsResult{
		Data:    exists,
		HasData: true,
	}, nil
}
