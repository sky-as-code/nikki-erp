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
	itHier "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewHierarchyServiceImpl(
	hierarchyRepo2 itHier.HierarchyRepository,
	userRepo2 itUser.UserRepository,
	orgSvc itOrg.OrganizationService,
	cqrsBus cqrs.CqrsBus,
) itHier.HierarchyService {
	return &HierarchyServiceImpl{
		cqrsBus:        cqrsBus,
		orgSvc:         orgSvc,
		hierarchyRepo2: hierarchyRepo2,
		userRepo2:      userRepo2,
	}
}

type HierarchyServiceImpl struct {
	cqrsBus        cqrs.CqrsBus
	orgSvc         itOrg.OrganizationService
	hierarchyRepo2 itHier.HierarchyRepository
	userRepo2      itUser.UserRepository
}

func (this *HierarchyServiceImpl) CreateHierarchyLevel(
	ctx corectx.Context, cmd itHier.CreateHierarchyLevelCommand,
) (*itHier.CreateHierarchyLevelResult, error) {
	return corecrud.Create(ctx, dyn.CreateParam[domain.HierarchyLevel, *domain.HierarchyLevel]{
		Action:         "create hierarchy level",
		BaseRepoGetter: this.hierarchyRepo2,
		Data:           cmd,
	})
}

func (this *HierarchyServiceImpl) DeleteHierarchyLevel(
	ctx corectx.Context, cmd itHier.DeleteHierarchyLevelCommand,
) (*itHier.DeleteHierarchyLevelResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete hierarchy level",
		DbRepoGetter: this.hierarchyRepo2,
		Cmd:          dyn.DeleteOneQuery(cmd),
	})
}

func (this *HierarchyServiceImpl) GetHierarchyLevel(
	ctx corectx.Context, query itHier.GetHierarchyLevelQuery,
) (*itHier.GetHierarchyLevelResult, error) {
	return corecrud.GetOne[domain.HierarchyLevel](ctx, corecrud.GetOneParam{
		Action:       "get hierarchy level",
		DbRepoGetter: this.hierarchyRepo2,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *HierarchyServiceImpl) SearchHierarchyLevels(
	ctx corectx.Context, query itHier.SearchHierarchyLevelsQuery,
) (*itHier.SearchHierarchyLevelsResult, error) {
	return corecrud.Search[domain.HierarchyLevel](ctx, corecrud.SearchParam{
		Action:       "search hierarchy levels",
		DbRepoGetter: this.hierarchyRepo2,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *HierarchyServiceImpl) HierarchyLevelExists(
	ctx corectx.Context, query itHier.HierarchyLevelExistsQuery,
) (*itHier.HierarchyLevelExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if hierarchy level exists",
		DbRepoGetter: this.hierarchyRepo2,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *HierarchyServiceImpl) UpdateHierarchyLevel(
	ctx corectx.Context, cmd itHier.UpdateHierarchyLevelCommand,
) (*itHier.UpdateHierarchyLevelResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Group, *domain.Group]{
		Action:       "update group",
		DbRepoGetter: this.hierarchyRepo2,
		Data:         cmd,
	})
}

func (this *HierarchyServiceImpl) ManageHierarchyLevelUsers(
	ctx corectx.Context, cmd itHier.ManageHierarchyLevelUsersCommand,
) (result *itHier.ManageHierarchyLevelUsersResult, err error) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:         "manage hierarchy level users",
		DbRepoGetter:   this.hierarchyRepo2,
		DestSchemaName: domain.UserSchemaName,
		Associations:   hierarchyUserAssocs(cmd.HierarchyId, cmd.Add.ToSlice()),
		Desociations:   hierarchyUserAssocs(cmd.HierarchyId, cmd.Remove.ToSlice()),
	})
}

func hierarchyUserAssocs(hierarchyId model.Id, userIds []model.Id) []dyn.RepoM2mAssociation {
	out := make([]dyn.RepoM2mAssociation, 0, len(userIds))
	for _, uid := range userIds {
		out = append(out, dyn.RepoM2mAssociation{
			SrcKeys:  dmodel.DynamicFields{basemodel.FieldId: hierarchyId},
			DestKeys: dmodel.DynamicFields{basemodel.FieldId: uid},
		})
	}
	return out
}
