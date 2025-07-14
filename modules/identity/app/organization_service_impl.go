package app

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationServiceImpl(
	enumSvc enum.EnumService,
	orgRepo itOrg.OrganizationRepository,
) itOrg.OrganizationService {
	return &OrganizationServiceImpl{
		enumSvc: enumSvc,
		orgRepo: orgRepo,
	}
}

type OrganizationServiceImpl struct {
	enumSvc enum.EnumService
	orgRepo itOrg.OrganizationRepository
}

func (this *OrganizationServiceImpl) CreateOrganization(ctx context.Context, cmd itOrg.CreateOrganizationCommand) (result *itOrg.CreateOrganizationResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create organization"); e != nil {
			err = e
		}
	}()

	org := cmd.ToOrganization()
	this.setOrgDefaults(ctx, org)

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = org.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeOrg(org)
			return this.assertOrgUnique(ctx, org.Slug, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itOrg.CreateOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	org, err = this.orgRepo.Create(ctx, *org)
	ft.PanicOnErr(err)

	return &itOrg.CreateOrganizationResult{
		Data:    org,
		HasData: true,
	}, nil
}

func (this *OrganizationServiceImpl) UpdateOrganization(ctx context.Context, cmd itOrg.UpdateOrganizationCommand) (result *itOrg.UpdateOrganizationResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "update organization"); e != nil {
			err = e
		}
	}()

	updatedOrg := cmd.ToOrganization()

	var dbOrg *domain.Organization
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = updatedOrg.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbOrg, err = this.assertOrgExists(ctx, *updatedOrg.Slug, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertCorrectEtag(updatedOrg, dbOrg, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeOrg(updatedOrg)
			return this.assertOrgUnique(ctx, cmd.NewSlug, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itOrg.UpdateOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	updatedOrg.Id = dbOrg.Id

	prevEtag := *updatedOrg.Etag
	updatedOrg.Etag = model.NewEtag()
	updatedOrg, err = this.orgRepo.Update(ctx, *updatedOrg, prevEtag)
	ft.PanicOnErr(err)

	return &itOrg.UpdateOrganizationResult{
		Data:    updatedOrg,
		HasData: true,
	}, nil
}

func (this *OrganizationServiceImpl) assertOrgUnique(ctx context.Context, slug *model.Slug, vErrs *ft.ValidationErrors) error {
	if slug == nil {
		return nil
	}
	dbOrg, err := this.orgRepo.FindBySlug(ctx, itOrg.GetOrganizationBySlugQuery{
		Slug:           *slug,
		IncludeDeleted: true,
	})
	if err != nil {
		return err
	}

	if dbOrg != nil {
		vErrs.AppendAlreadyExists("slug", "organization slug")
	}
	return nil
}

func (this *OrganizationServiceImpl) assertCorrectEtag(updatedOrg *domain.Organization, dbOrg *domain.Organization, vErrs *ft.ValidationErrors) {
	if *updatedOrg.Etag != *dbOrg.Etag {
		vErrs.Append("etag", "etag mismatched")
	}
}

func (this *OrganizationServiceImpl) assertOrgExists(ctx context.Context, slug model.Slug, vErrs *ft.ValidationErrors) (dbOrg *domain.Organization, err error) {
	dbOrg, err = this.orgRepo.FindBySlug(ctx, itOrg.GetOrganizationBySlugQuery{
		Slug:           slug,
		IncludeDeleted: true,
	})
	if dbOrg == nil {
		vErrs.AppendNotFound("slug", "organization slug")
	}
	return
}

func (this *OrganizationServiceImpl) sanitizeOrg(org *domain.Organization) {
	if org.Address != nil {
		org.Address = util.ToPtr(defense.SanitizePlainText(*org.Address, true))
	}
	if org.DisplayName != nil {
		org.DisplayName = util.ToPtr(defense.SanitizePlainText(*org.DisplayName, true))
	}
	if org.LegalName != nil {
		org.LegalName = util.ToPtr(defense.SanitizePlainText(*org.LegalName, true))
	}
}

func (this *OrganizationServiceImpl) setOrgDefaults(ctx context.Context, org *domain.Organization) {
	org.SetDefaults()

	activeEnum, err := this.enumSvc.GetEnum(ctx, enum.GetEnumQuery{
		Value: util.ToPtr(domain.OrgStatusActive),
		Type:  util.ToPtr(domain.OrgStatusEnumType),
	})
	ft.PanicOnErr(err)
	ft.PanicOnErr(activeEnum.ClientError)

	org.Status = domain.WrapIdentStatus(activeEnum.Data)
	org.StatusId = activeEnum.Data.Id
}

func (this *OrganizationServiceImpl) DeleteOrganization(ctx context.Context, cmd itOrg.DeleteOrganizationCommand) (result *itOrg.DeleteOrganizationResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "delete organization"); e != nil {
			err = e
		}
	}()

	var dbOrg *domain.Organization
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbOrg, err = this.assertOrgExists(ctx, cmd.Slug, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itOrg.DeleteOrganizationResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	err = this.orgRepo.DeleteHard(ctx, *dbOrg.Id)
	ft.PanicOnErr(err)

	return crud.NewSuccessDeletionResult(*dbOrg.Id), nil
}

func (this *OrganizationServiceImpl) GetOrganizationBySlug(ctx context.Context, query itOrg.GetOrganizationBySlugQuery) (result *itOrg.GetOrganizationBySlugResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get organization by slug"); e != nil {
			err = e
		}
	}()

	var dbOrg *domain.Organization
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbOrg, err = this.assertOrgExists(ctx, query.Slug, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itOrg.GetOrganizationBySlugResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &itOrg.GetOrganizationBySlugResult{
		Data:    dbOrg,
		HasData: dbOrg != nil,
	}, nil
}

func (this *OrganizationServiceImpl) SearchOrganizations(ctx context.Context, query itOrg.SearchOrganizationsQuery) (result *itOrg.SearchOrganizationsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "search organizations"); e != nil {
			err = e
		}
	}()

	query.SetDefaults()
	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.orgRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &itOrg.SearchOrganizationsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	orgs, err := this.orgRepo.Search(ctx, itOrg.SearchParam{
		Predicate:      predicate,
		Order:          order,
		Page:           *query.Page,
		Size:           *query.Size,
		IncludeDeleted: query.IncludeDeleted,
	})
	ft.PanicOnErr(err)

	return &itOrg.SearchOrganizationsResult{
		Data: orgs,
	}, nil
}
