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
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
	"go.bryk.io/pkg/errors"
)

func NewOrganizationDomainServiceImpl(
	orgRepo it.OrganizationRepository,
	cqrsBus cqrs.CqrsBus,
	settingsSvc itExt.OrgSettingsInitExtService,
) it.OrganizationDomainService {
	return &OrganizationDomainServiceImpl{cqrsBus: cqrsBus, orgRepo: orgRepo, settingsSvc: settingsSvc}
}

type OrganizationDomainServiceImpl struct {
	cqrsBus     cqrs.CqrsBus
	orgRepo     it.OrganizationRepository
	settingsSvc itExt.OrgSettingsInitExtService
}

// CreateOrg creates the organization and seeds its settings in one transaction.
//
// The seeding shares the transaction deliberately: an organization that exists without its settings
// rows would render an empty settings page until someone noticed, and there is no later point that
// would repair it. Either both land or neither does.
//
// The call goes out through a port rather than into the settings module directly — settings may not
// import iam, so the dependency runs this way round only.
func (this *OrganizationDomainServiceImpl) CreateOrg(
	ctx corectx.Context, cmd it.CreateOrgCommand, options ...corecrud.ServiceCreateOptions[*domain.Organization],
) (*it.CreateOrgResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceCreateOptions[*domain.Organization]{})

	tranx, err := this.orgRepo.GetBaseRepo().BeginTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "create organization")
	}
	defer tranx.Rollback()

	// The transaction goes on a clone: setting it on the caller's context would leave a committed
	// transaction visible to whatever runs next.
	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	result, err := corecrud.Create(tranxCtx, corecrud.CreateParam[domain.Organization, *domain.Organization]{
		Action:                 "create organization",
		BaseRepoGetter:         this.orgRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
	if err != nil {
		return nil, err
	}
	// A rejected create has written nothing, so there is nothing to seed and nothing to commit.
	if result.ClientErrors.Count() > 0 || !result.HasData {
		return result, nil
	}

	if err := this.initOrgSettings(tranxCtx, &result.Data); err != nil {
		return nil, err
	}
	return result, errors.Wrap(tranx.Commit(), "create organization")
}

// initOrgSettings copies the tenant's settings onto the newly created organization.
func (this *OrganizationDomainServiceImpl) initOrgSettings(
	ctx corectx.Context, org *domain.Organization,
) error {
	orgId := org.GetId()
	if orgId == nil {
		return errors.New("create organization: the created organization has no id")
	}

	initResult, err := this.settingsSvc.InitOrgSettings(ctx, itExt.InitOwnerSettingsCommand{
		OwnerId: *orgId,
	})
	if err != nil {
		return errors.Wrap(err, "create organization")
	}
	// A rejected seeding is a defect in this call, not something the caller creating an
	// organization can correct, so it fails the whole create rather than being reported as a
	// validation error against a field they submitted.
	if initResult.ClientErrors.Count() > 0 {
		return errors.Wrap(initResult.ClientErrors.ToError(), "create organization")
	}
	return nil
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
