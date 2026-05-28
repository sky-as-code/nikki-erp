package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itRole "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/role"
)

func NewGroupDomainServiceImpl(
	groupRepo itGrp.GroupRepository,
	roleSvc itRole.RoleDomainService,
	cqrsBus cqrs.CqrsBus,
) itGrp.GroupDomainService {
	return &GroupDomainServiceImpl{
		cqrsBus:   cqrsBus,
		groupRepo: groupRepo,
		roleSvc:   roleSvc,
	}
}

type GroupDomainServiceImpl struct {
	cqrsBus   cqrs.CqrsBus
	groupRepo itGrp.GroupRepository
	roleSvc   itRole.RoleDomainService
}

func (this *GroupDomainServiceImpl) CreateGroup(
	ctx corectx.Context, cmd itGrp.CreateGroupCommand, options ...corecrud.ServiceCreateOptions[*domain.Group],
) (*itGrp.CreateGroupResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceCreateOptions[*domain.Group]{})
	result, err := corecrud.Create(ctx, corecrud.CreateParam[domain.Group, *domain.Group]{
		Action:                 "create group",
		BaseRepoGetter:         this.groupRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
	return result, err

}

func (this *GroupDomainServiceImpl) DeleteGroup(
	ctx corectx.Context, cmd itGrp.DeleteGroupCommand, options ...corecrud.ServiceDeleteOptions,
) (*itGrp.DeleteGroupResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceDeleteOptions{})
	result, err := corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:                 "delete group",
		DbRepoGetter:           this.groupRepo,
		Cmd:                    dyn.DeleteOneCommand{Id: cmd.Id},
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
	return result, err
}

func (this *GroupDomainServiceImpl) GroupExists(ctx corectx.Context, query itGrp.GroupExistsQuery) (
	*itGrp.GroupExistsResult, error,
) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if group exists",
		DbRepoGetter: this.groupRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *GroupDomainServiceImpl) GetGroup(
	ctx corectx.Context, query itGrp.GetGroupQuery,
) (*dyn.OpResult[domain.Group], error) {
	return corecrud.GetOne[domain.Group](ctx, corecrud.GetOneParam{
		Action:       "get group",
		DbRepoGetter: this.groupRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *GroupDomainServiceImpl) ManageGroupUsers(
	ctx corectx.Context, cmd itGrp.ManageGroupUsersCommand,
) (result *itGrp.ManageGroupUsersResult, err error) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage group users",
		DbRepoGetter:       this.groupRepo,
		DestSchemaName:     domain.UserSchemaName,
		SrcId:              cmd.GroupId,
		SrcIdFieldForError: "group_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
		BeforeInsert: func(ctx corectx.Context, dbRecords []dmodel.DynamicFields) error {
			for _, record := range dbRecords {
				relationId, err := model.NewId()
				if err != nil {
					return err
				}
				record[domain.GrpUsrRelFieldId] = *relationId
			}
			return nil
		},
	})
}

func (this *GroupDomainServiceImpl) SearchGroups(
	ctx corectx.Context, query itGrp.SearchGroupsQuery, options ...corecrud.ServiceSearchOptions,
) (
	*itGrp.SearchGroupsResult, error,
) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	return corecrud.Search[domain.Group](ctx, corecrud.SearchParam{
		Action:                 "search groups",
		DbRepoGetter:           this.groupRepo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *GroupDomainServiceImpl) UpdateGroup(
	ctx corectx.Context, cmd itGrp.UpdateGroupCommand, options ...corecrud.ServiceUpdateOptions[*domain.Group],
) (*itGrp.UpdateGroupResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceUpdateOptions[*domain.Group]{})
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Group, *domain.Group]{
		Action:                 "update group",
		DbRepoGetter:           this.groupRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}
