package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationDomainServiceImpl(
	orgRepo it.OrganizationRepository,
	cqrsBus cqrs.CqrsBus,
) it.OrganizationDomainService {
	return &OrganizationDomainServiceImpl{cqrsBus: cqrsBus, orgRepo: orgRepo}
}

type OrganizationDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	orgRepo it.OrganizationRepository
}

func (this *OrganizationDomainServiceImpl) CreateOrg(
	ctx corectx.Context, cmd it.CreateOrgCommand, options ...corecrud.ServiceCreateOptions[*domain.Organization],
) (*it.CreateOrgResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceCreateOptions[*domain.Organization]{})
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Organization, *domain.Organization]{
		Action:                 "create organization",
		BaseRepoGetter:         this.orgRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *OrganizationDomainServiceImpl) DeleteOrg(
	ctx corectx.Context, cmd it.DeleteOrgCommand, options ...corecrud.ServiceDeleteOptions,
) (*it.DeleteOrgResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceDeleteOptions{})
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:                 "delete organization",
		DbRepoGetter:           this.orgRepo,
		Cmd:                    dyn.DeleteOneCommand(cmd),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *OrganizationDomainServiceImpl) GetOrg(ctx corectx.Context, query it.GetOrgQuery) (*dyn.OpResult[domain.Organization], error) {
	return this.getOrgWithArchived(ctx, query, nil)
}

func (this *OrganizationDomainServiceImpl) GetActiveOrg(ctx corectx.Context, query it.GetOrgQuery) (*dyn.OpResult[domain.Organization], error) {
	return this.getOrgWithArchived(ctx, query, util.ToPtr(true))
}

func (this *OrganizationDomainServiceImpl) getOrgWithArchived(ctx corectx.Context, query it.GetOrgQuery, isArchived *bool) (*dyn.OpResult[domain.Organization], error) {
	sanitizedFields, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[domain.Organization]{ClientErrors: cErrs}, nil
	}
	query = *(sanitizedFields.(*it.GetOrgQuery))

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
		Fields: query.Fields,
		Graph:  graph,
		Page:   0,
		Size:   1,
	}

	searchRes, err := this.SearchOrgs(ctx, searchquery)
	if err != nil {
		return nil, err
	}
	result := &dyn.OpResult[domain.Organization]{
		ClientErrors: searchRes.ClientErrors,
		HasData:      searchRes.HasData,
	}

	if searchRes.HasData {
		result.Data = searchRes.Data.Items[0]
	}

	return result, nil
}

func (this *OrganizationDomainServiceImpl) OrgExists(
	ctx corectx.Context, query it.OrgExistsQuery,
) (*it.OrgExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if organizations exist",
		DbRepoGetter: this.orgRepo,
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

func (this *OrganizationDomainServiceImpl) ManageOrgUsers(ctx corectx.Context, cmd it.ManageOrgUsersCommand) (
	result *it.ManageOrgUsersResult, err error,
) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage organization users",
		DbRepoGetter:       this.orgRepo,
		DestSchemaName:     domain.UserSchemaName,
		SrcId:              cmd.OrgId,
		SrcIdFieldForError: "org_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
		BeforeInsert: func(ctx corectx.Context, dbRecords []dmodel.DynamicFields) error {
			ulidType := dmodel.FieldDataTypeUlid()
			for _, rec := range dbRecords {
				rec[basemodel.FieldId] = *ulidType.DefaultValue().Get()
			}
			return nil
		},
	})
}

func (this *OrganizationDomainServiceImpl) SearchOrgs(
	ctx corectx.Context, query it.SearchOrgsQuery, options ...corecrud.ServiceSearchOptions,
) (*it.SearchOrgsResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	return corecrud.Search[domain.Organization](ctx, corecrud.SearchParam{
		Action:                 "search organizations",
		DbRepoGetter:           this.orgRepo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *OrganizationDomainServiceImpl) SetOrgIsArchived(ctx corectx.Context, cmd it.SetOrgIsArchivedCommand) (*it.SetOrgIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.orgRepo, dyn.SetIsArchivedCommand(cmd))
}

func (this *OrganizationDomainServiceImpl) UpdateOrg(
	ctx corectx.Context, cmd it.UpdateOrgCommand, options ...corecrud.ServiceUpdateOptions[*domain.Organization],
) (*it.UpdateOrgResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceUpdateOptions[*domain.Organization]{})
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Organization, *domain.Organization]{
		Action:                 "update organization",
		DbRepoGetter:           this.orgRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}
