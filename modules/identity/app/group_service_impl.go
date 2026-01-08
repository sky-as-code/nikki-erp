package app

import (
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewGroupServiceImpl(groupRepo itGrp.GroupRepository, orgSvc itOrg.OrganizationService, cqrsBus cqrs.CqrsBus) itGrp.GroupService {
	return &GroupServiceImpl{
		cqrsBus:   cqrsBus,
		groupRepo: groupRepo,
		orgSvc:    orgSvc,
	}
}

type GroupServiceImpl struct {
	cqrsBus   cqrs.CqrsBus
	orgSvc    itOrg.OrganizationService
	groupRepo itGrp.GroupRepository
}

func (this *GroupServiceImpl) AddRemoveUsers(ctx crud.Context, cmd itGrp.AddRemoveUsersCommand) (result *itGrp.AddRemoveUsersResult, err error) {
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
			dbGroup, err = this.assertGroupByID(ctx, cmd.GroupId, vErrs)
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

func (this *GroupServiceImpl) CreateGroup(ctx crud.Context, cmd itGrp.CreateGroupCommand) (*itGrp.CreateGroupResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.Group, itGrp.CreateGroupCommand, itGrp.CreateGroupResult]{
		Action:              "create group",
		Command:             cmd,
		AssertBusinessRules: this.assertCreateRules,
		RepoCreate:          this.groupRepo.Create,
		SetDefault:          this.setGroupDefaults,
		Sanitize:            this.sanitizeGroup,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itGrp.CreateGroupResult {
			return &itGrp.CreateGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Group) *itGrp.CreateGroupResult {
			return &itGrp.CreateGroupResult{
				HasData: true,
				Data:    model,
			}
		},
	})

	return result, err
}

func (this *GroupServiceImpl) UpdateGroup(ctx crud.Context, cmd itGrp.UpdateGroupCommand) (*itGrp.UpdateGroupResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Group, itGrp.UpdateGroupCommand, itGrp.UpdateGroupResult]{
		Action:              "update group",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateRules,
		AssertExists:        this.assertGroupByDomain,
		RepoUpdate:          this.groupRepo.Update,
		Sanitize:            this.sanitizeGroup,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itGrp.UpdateGroupResult {
			return &itGrp.UpdateGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Group) *itGrp.UpdateGroupResult {
			return &itGrp.UpdateGroupResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err

}

func (this *GroupServiceImpl) GetGroupById(ctx crud.Context, query itGrp.GetGroupByIdQuery) (*itGrp.GetGroupByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Group, itGrp.GetGroupByIdQuery, itGrp.GetGroupByIdResult]{
		Action:      "get group by Id",
		Query:       query,
		RepoFindOne: this.getGroupByIdFull,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itGrp.GetGroupByIdResult {
			return &itGrp.GetGroupByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Group) *itGrp.GetGroupByIdResult {
			return &itGrp.GetGroupByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *GroupServiceImpl) DeleteGroup(ctx crud.Context, cmd itGrp.DeleteGroupCommand) (*itGrp.DeleteGroupResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Group, itGrp.DeleteGroupCommand, itGrp.DeleteGroupResult]{
		Action:       "delete group",
		Command:      cmd,
		AssertExists: this.assertGroupByDomain,
		RepoDelete: func(ctx crud.Context, model *domain.Group) (int, error) {
			return this.groupRepo.DeleteHard(ctx, itGrp.DeleteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itGrp.DeleteGroupResult {
			return &itGrp.DeleteGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Group, deletedCount int) *itGrp.DeleteGroupResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *GroupServiceImpl) SearchGroups(ctx crud.Context, query itGrp.SearchGroupsQuery) (*itGrp.SearchGroupsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Group, itGrp.SearchGroupsQuery, itGrp.SearchGroupsResult]{
		Action: "search groups",
		Query:  query,
		SetQueryDefaults: func(query *itGrp.SearchGroupsQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.groupRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itGrp.SearchGroupsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Group], error) {
			return this.groupRepo.Search(ctx, itGrp.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
				WithOrg:   query.WithOrg,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itGrp.SearchGroupsResult {
			return &itGrp.SearchGroupsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.Group]) *itGrp.SearchGroupsResult {
			return &itGrp.SearchGroupsResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *GroupServiceImpl) Exist(ctx crud.Context, cmd itGrp.GroupExistsCommand) (result *itGrp.GroupExistsResult, err error) {
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

// assert methods
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *GroupServiceImpl) assertCreateRules(ctx crud.Context, group *domain.Group, vErrs *ft.ValidationErrors) error {
	if group.OrgId == nil {
		return nil
	}

	org, err := this.orgSvc.GetOrganizationById(ctx, itOrg.GetOrganizationByIdQuery{
		Id:             *group.OrgId,
		IncludeDeleted: false,
	})
	if err != nil {
		return err
	}

	if org.ClientError != nil {
		vErrs.MergeClientError(org.ClientError)
		return nil
	}

	if org.Data == nil {
		vErrs.Append("orgId", "organization not found")
		return nil
	}

	return this.assertUniqueGroupName(ctx, group, vErrs)
}

func (this *GroupServiceImpl) assertUpdateRules(ctx crud.Context, group *domain.Group, _ *domain.Group, vErrs *ft.ValidationErrors) error {
	_, err := this.assertGroupByID(ctx, *group.Id, vErrs)
	if err != nil {
		return err
	}

	return this.assertUniqueGroupName(ctx, group, vErrs)
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *GroupServiceImpl) sanitizeGroup(group *domain.Group) {
	if group.Name != nil {
		cleanedName := strings.TrimSpace(*group.Name)
		cleanedName = defense.SanitizePlainText(cleanedName)
		group.Name = &cleanedName
	}
}

func (this *GroupServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *ft.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *GroupServiceImpl) assertGroupByDomain(ctx crud.Context, group *domain.Group, vErrs *ft.ValidationErrors) (dbGroup *domain.Group, err error) {
	dbGroup, err = this.assertGroupByID(ctx, *group.Id, vErrs)
	if err != nil {
		return nil, err
	}

	return dbGroup, err
}

func (this *GroupServiceImpl) assertGroupByID(ctx crud.Context, id model.Id, vErrs *ft.ValidationErrors) (dbGroup *domain.Group, err error) {
	dbGroup, err = this.groupRepo.FindById(ctx, itGrp.GetGroupByIdQuery{Id: id})
	if err != nil {
		return nil, err
	}

	if dbGroup == nil {
		vErrs.Append("id", "group not found")
	}
	return dbGroup, err
}

func (this *GroupServiceImpl) assertUniqueGroupName(ctx crud.Context, group *domain.Group, vErrs *ft.ValidationErrors) error {
	dbGroup, err := this.groupRepo.FindByName(ctx, itGrp.FindByNameParam{
		Name: *group.Name,
	})
	if err != nil {
		return err
	}

	if dbGroup != nil && *dbGroup.Id != *group.Id {
		vErrs.AppendAlreadyExists("name", "group name")
	}
	return nil
}

func (this *GroupServiceImpl) setGroupDefaults(group *domain.Group) {
	group.SetDefaults()
}

func (this *GroupServiceImpl) assertUserIdsExist(ctx crud.Context, valErrs *ft.ValidationErrors, field string, userIds []string) error {
	if len(userIds) == 0 {
		return nil
	}

	existCmd := &itUser.UserExistsMultiQuery{
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

func (this *GroupServiceImpl) getGroupByIdFull(ctx crud.Context, query itGrp.GetGroupByIdQuery, vErrs *ft.ValidationErrors) (dbGroup *domain.Group, err error) {
	dbGroup, err = this.groupRepo.FindById(ctx, query)
	if dbGroup == nil {
		vErrs.Append("id", "group not found")
	}
	return
}
