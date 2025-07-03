package app

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationServiceImpl(orgRepo itOrg.OrganizationRepository, cqrsBus cqrs.CqrsBus) itOrg.OrganizationService {
	return &OrganizationServiceImpl{
		cqrsBus: cqrsBus,
		orgRepo: orgRepo,
	}
}

type OrganizationServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	orgRepo itOrg.OrganizationRepository
}

func (this *OrganizationServiceImpl) CreateOrganization(ctx context.Context, cmd itOrg.CreateOrganizationCommand) (result *itOrg.CreateOrganizationResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create organization"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	if vErrs.Count() > 0 {
		return &itOrg.CreateOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	orgDb, err := this.findOrganizationBySlug(ctx, cmd.Slug)
	ft.PanicOnErr(err)

	if orgDb != nil {
		vErrs.Append("slug", "organization already exists")
		return &itOrg.CreateOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	org := cmd.ToOrganization()
	this.setOrganizationDefaults(org)

	org, err = this.orgRepo.Create(ctx, *org)
	ft.PanicOnErr(err)

	return &itOrg.CreateOrganizationResult{
		Data: org,
	}, nil
}

func (this *OrganizationServiceImpl) setOrganizationDefaults(org *domain.Organization) {
	err := org.SetDefaults()
	ft.PanicOnErr(err)
}

func (this *OrganizationServiceImpl) UpdateOrganization(ctx context.Context, cmd itOrg.UpdateOrganizationCommand) (result *itOrg.UpdateOrganizationResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update organization"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	if vErrs.Count() > 0 {
		return &itOrg.UpdateOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	orgDb, err := this.findOrganizationBySlug(ctx, cmd.Slug)
	ft.PanicOnErr(err)

	if orgDb == nil {
		vErrs.Append("slug", "organization not found")
		return &itOrg.UpdateOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	org := cmd.ToOrganization()
	if *org.Etag != *orgDb.Etag {
		vErrs.Append("etag", "etag mismatch")
		return &itOrg.UpdateOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	this.setOrganizationDefaults(org)
	org.Id = orgDb.Id
	org, err = this.orgRepo.Update(ctx, *org)
	ft.PanicOnErr(err)

	return &itOrg.UpdateOrganizationResult{
		Data: org,
	}, nil
}

func (this *OrganizationServiceImpl) DeleteOrganization(ctx context.Context, cmd itOrg.DeleteOrganizationCommand) (result *itOrg.DeleteOrganizationResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete organization"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	if vErrs.Count() > 0 {
		return &itOrg.DeleteOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	orgDb, err := this.findOrganizationBySlug(ctx, cmd.Slug)
	ft.PanicOnErr(err)

	if orgDb == nil {
		vErrs.Append("slug", "organization not found")
		return &itOrg.DeleteOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	err = this.orgRepo.DeleteHard(ctx, model.Id(*orgDb.Id))
	ft.PanicOnErr(err)

	return &itOrg.DeleteOrganizationResult{
		Data: &itOrg.DeleteOrganizationResultData{
			DeletedAt: time.Now(),
		},
	}, nil
}

func (this *OrganizationServiceImpl) GetOrganizationBySlug(ctx context.Context, query itOrg.GetOrganizationBySlugQuery) (result *itOrg.GetOrganizationBySlugResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get organization by slug"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &itOrg.GetOrganizationBySlugResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	orgDb, err := this.orgRepo.FindBySlug(ctx, query)
	ft.PanicOnErr(err)

	if orgDb == nil {
		vErrs.Append("slug", "organization not found")
		return &itOrg.GetOrganizationBySlugResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &itOrg.GetOrganizationBySlugResult{
		Data: orgDb,
	}, nil
}

func (this *OrganizationServiceImpl) SearchOrganizations(ctx context.Context, query itOrg.SearchOrganizationsQuery) (result *itOrg.SearchOrganizationsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to search organizations"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.orgRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &itOrg.SearchOrganizationsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	orgs, err := this.orgRepo.Search(ctx, predicate, order, query)
	ft.PanicOnErr(err)

	return &itOrg.SearchOrganizationsResult{
		Data: orgs,
	}, nil
}

func (this *OrganizationServiceImpl) findOrganizationBySlug(ctx context.Context, slug model.Slug) (*domain.Organization, error) {
	org, err := this.orgRepo.FindBySlug(ctx,
		itOrg.GetOrganizationBySlugQuery{
			Slug:           slug,
			IncludeDeleted: false,
		})
	if err != nil {
		return nil, err
	}

	return org, nil
}
