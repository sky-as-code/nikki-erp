package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationServiceImpl(
	orgRepo2 it.OrganizationRepository,
	cqrsBus cqrs.CqrsBus,
) it.OrganizationService {
	return &OrganizationServiceImpl{cqrsBus: cqrsBus, orgRepo2: orgRepo2}
}

type OrganizationServiceImpl struct {
	cqrsBus  cqrs.CqrsBus
	orgRepo2 it.OrganizationRepository
}

func (this *OrganizationServiceImpl) CreateOrg(
	ctx corectx.Context, cmd it.CreateOrgCommand,
) (*it.CreateOrgResult, error) {
	return corecrud.Create(ctx, dyn.CreateParam[domain.Organization, *domain.Organization]{
		Action:         "create organization",
		BaseRepoGetter: this.orgRepo2,
		Data:           cmd,
	})
}

func (this *OrganizationServiceImpl) DeleteOrg(
	ctx corectx.Context, cmd it.DeleteOrgCommand,
) (*it.DeleteOrgResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete organization",
		DbRepoGetter: this.orgRepo2,
		Cmd:          dyn.DeleteOneQuery(cmd),
	})
}

func (this *OrganizationServiceImpl) GetOrg(ctx corectx.Context, query it.GetOrgQuery) (*it.GetOrgResult, error) {
	return this.getOrgWithArchived(ctx, query, nil)
}

func (this *OrganizationServiceImpl) GetActiveOrg(ctx corectx.Context, query it.GetOrgQuery) (*it.GetOrgResult, error) {
	return this.getOrgWithArchived(ctx, query, util.ToPtr(true))
}

func (this *OrganizationServiceImpl) getOrgWithArchived(ctx corectx.Context, query it.GetOrgQuery, isArchived *bool) (*it.GetOrgResult, error) {
	if query.Id == nil && query.Slug == nil {
		return nil, ft.NewExclusiveFieldsError([]string{basemodel.FieldId, domain.OrgFieldSlug})
	}
	statusNode := dmodel.NewSearchNode()
	if isArchived != nil {
		statusNode.NewCondition(basemodel.FieldIsArchived, dmodel.Equals, *isArchived)
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*statusNode,
		*dmodel.NewSearchNode().Or(
			*dmodel.NewSearchNode().NewCondition(basemodel.FieldId, dmodel.Equals, query.Id),
			*dmodel.NewSearchNode().NewCondition(domain.OrgFieldSlug, dmodel.Equals, query.Slug),
		),
	)
	graph.Or()
	searchquery := it.SearchOrgsQuery{
		Columns: query.Columns,
		Graph:   graph,
		Page:    0,
		Size:    1,
	}

	searchRes, err := this.SearchOrgs(ctx, searchquery)
	if err != nil {
		return nil, err
	}
	result := &it.GetOrgResult{
		ClientErrors: searchRes.ClientErrors,
		HasData:      searchRes.HasData,
	}

	if searchRes.HasData {
		result.Data = searchRes.Data.Items[0]
	}

	return result, nil
}

func getOrgSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"identity.get_org_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dmodel.DefineField().
					Name(basemodel.FieldColumns).
					DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType())).
				Field(dmodel.DefineField().
					Name(basemodel.FieldId).
					DataType(dmodel.FieldDataTypeUlid()),
				// Note: Not Required()
				).
				Field(dmodel.DefineField().
					Name(domain.OrgFieldSlug).
					DataType(dmodel.FieldDataTypeEmail()))
		},
	)
}

func (this *OrganizationServiceImpl) OrgExists(
	ctx corectx.Context, query it.OrgExistsQuery,
) (*it.OrgExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if organizations exist",
		DbRepoGetter: this.orgRepo2,
		Query:        dyn.ExistsQuery(query),
	})
}

func orgUserAssocs(orgId model.Id, userIds []model.Id) []dyn.RepoM2mAssociation {
	out := make([]dyn.RepoM2mAssociation, 0, len(userIds))
	for _, uid := range userIds {
		u := uid
		out = append(out, dyn.RepoM2mAssociation{
			SrcKeys:  dmodel.DynamicFields{basemodel.FieldId: orgId},
			DestKeys: dmodel.DynamicFields{basemodel.FieldId: u},
		})
	}
	return out
}

func (this *OrganizationServiceImpl) ManageOrgUsers(ctx corectx.Context, cmd it.ManageOrgUsersCommand) (
	result *it.ManageOrgUsersResult, err error,
) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:         "manage organization users",
		DbRepoGetter:   this.orgRepo2,
		DestSchemaName: domain.UserSchemaName,
		Associations:   orgUserAssocs(cmd.OrgId, cmd.Add.ToSlice()),
		Desociations:   orgUserAssocs(cmd.OrgId, cmd.Remove.ToSlice()),
	})
}

func (this *OrganizationServiceImpl) SearchOrgs(
	ctx corectx.Context, query it.SearchOrgsQuery,
) (*it.SearchOrgsResult, error) {
	return corecrud.Search[domain.Organization](ctx, corecrud.SearchParam{
		Action:       "search organizations",
		DbRepoGetter: this.orgRepo2,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *OrganizationServiceImpl) SetOrgIsArchived(ctx corectx.Context, cmd it.SetOrgIsArchivedCommand) (*it.SetOrgIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.orgRepo2, dyn.SetIsArchivedCommand(cmd))
}

func (this *OrganizationServiceImpl) UpdateOrg(
	ctx corectx.Context, cmd it.UpdateOrgCommand,
) (*it.UpdateOrgResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Organization, *domain.Organization]{
		Action:       "update organization",
		DbRepoGetter: this.orgRepo2,
		Data:         cmd,
	})
}
