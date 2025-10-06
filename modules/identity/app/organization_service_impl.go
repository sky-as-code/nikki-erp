package app

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	itEnum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationServiceImpl(
	enumSvc itEnum.EnumService,
	orgRepo itOrg.OrganizationRepository,
) itOrg.OrganizationService {
	return &OrganizationServiceImpl{
		enumSvc: enumSvc,
		orgRepo: orgRepo,
	}
}

type OrganizationServiceImpl struct {
	enumSvc itEnum.EnumService
	orgRepo itOrg.OrganizationRepository
}

func (this *OrganizationServiceImpl) CreateOrganization(ctx crud.Context, cmd itOrg.CreateOrganizationCommand) (*itOrg.CreateOrganizationResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.Organization, itOrg.CreateOrganizationCommand, itOrg.CreateOrganizationResult]{
		Action:              "create organization",
		Command:             cmd,
		AssertBusinessRules: this.assertCreateRules,
		RepoCreate:          this.orgRepo.Create,
		SetDefault:          this.setOrgDefaults,
		Sanitize:            this.sanitizeOrg,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itOrg.CreateOrganizationResult {
			return &itOrg.CreateOrganizationResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Organization) *itOrg.CreateOrganizationResult {
			return &itOrg.CreateOrganizationResult{
				HasData: true,
				Data:    model,
			}
		},
	})

	return result, err

	// defer func() {
	// 	if e := ft.RecoverPanicFailedTo(recover(), "create organization"); e != nil {
	// 		err = e
	// 	}
	// }()

	// org := cmd.ToOrganization()
	// this.setOrgDefaults(ctx, org)

	// flow := val.StartValidationFlow()
	// vErrs, err := flow.
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		*vErrs = org.Validate(false)
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		this.sanitizeOrg(org)
	// 		return this.assertOrgUnique(ctx, org.Slug, vErrs)
	// 	}).
	// 	End()
	// ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &itOrg.CreateOrganizationResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// org, err = this.orgRepo.Create(ctx, *org)
	// ft.PanicOnErr(err)

	// return &itOrg.CreateOrganizationResult{
	// 	Data:    org,
	// 	HasData: org != nil,
	// }, nil
}

func (this *OrganizationServiceImpl) UpdateOrganization(ctx crud.Context, cmd itOrg.UpdateOrganizationCommand) (result *itOrg.UpdateOrganizationResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "update organization"); e != nil {
			err = e
		}
	}()

	updatedOrg := cmd.ToDomainModel()

	var dbOrg *domain.Organization
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = updatedOrg.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeOrg(updatedOrg)
			dbOrg, err = this.assertCorrectOrg(ctx, updatedOrg, vErrs)
			return err
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
	updatedOrg, err = this.orgRepo.Update(ctx, updatedOrg, prevEtag)
	ft.PanicOnErr(err)

	return &itOrg.UpdateOrganizationResult{
		Data:    updatedOrg,
		HasData: updatedOrg != nil,
	}, nil
}

func (this *OrganizationServiceImpl) DeleteOrganization(ctx crud.Context, cmd itOrg.DeleteOrganizationCommand) (*itOrg.DeleteOrganizationResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Organization, itOrg.DeleteOrganizationCommand, itOrg.DeleteOrganizationResult]{
		Action:       "delete organization",
		Command:      cmd,
		AssertExists: this.assertOrgExists,
		RepoDelete: func(ctx crud.Context, model *domain.Organization) (int, error) {
			return this.orgRepo.DeleteHard(ctx, itOrg.DeleteParam{Slug: *model.Slug})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itOrg.DeleteOrganizationResult {
			return &itOrg.DeleteOrganizationResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Organization, deletedCount int) *itOrg.DeleteOrganizationResult {
			var del *int
			if deletedCount > 0 {
				del = &deletedCount
			}
			return &itOrg.DeleteOrganizationResult{
				Data: &itOrg.DeleteOrganizationResultData{
					Slug:         *model.Slug,
					DeletedAt:    time.Now(),
					DeletedCount: del,
				},
				HasData: true,
			}
		},
	})

	return result, err
}

func (this *OrganizationServiceImpl) GetOrganizationBySlug(ctx crud.Context, query itOrg.GetOrganizationBySlugQuery) (*itOrg.GetOrganizationBySlugResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Organization, itOrg.GetOrganizationBySlugQuery, itOrg.GetOrganizationBySlugResult]{
		Action:      "get organization by slug",
		Query:       query,
		RepoFindOne: this.getOrgBySlugFull,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itOrg.GetOrganizationBySlugResult {
			return &itOrg.GetOrganizationBySlugResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Organization) *itOrg.GetOrganizationBySlugResult {
			return &itOrg.GetOrganizationBySlugResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *OrganizationServiceImpl) GetOrganizationById(ctx crud.Context, query itOrg.GetOrganizationByIdQuery) (*itOrg.GetOrganizationByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Organization, itOrg.GetOrganizationByIdQuery, itOrg.GetOrganizationByIdResult]{
		Action:      "get organization by id",
		Query:       query,
		RepoFindOne: this.getOrgByIdFull,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itOrg.GetOrganizationByIdResult {
			return &itOrg.GetOrganizationByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Organization) *itOrg.GetOrganizationByIdResult {
			return &itOrg.GetOrganizationByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})
	return result, err
}

func (this *OrganizationServiceImpl) SearchOrganizations(ctx crud.Context, query itOrg.SearchOrganizationsQuery) (*itOrg.SearchOrganizationsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Organization, itOrg.SearchOrganizationsQuery, itOrg.SearchOrganizationsResult]{
		Action: "search organizations",
		Query:  query,
		SetQueryDefaults: func(query *itOrg.SearchOrganizationsQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.orgRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itOrg.SearchOrganizationsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Organization], error) {
			return this.orgRepo.Search(ctx, itOrg.SearchParam{
				Predicate:      predicate,
				Order:          order,
				Page:           *query.Page,
				Size:           *query.Size,
				IncludeDeleted: query.IncludeDeleted,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itOrg.SearchOrganizationsResult {
			return &itOrg.SearchOrganizationsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.Organization]) *itOrg.SearchOrganizationsResult {
			return &itOrg.SearchOrganizationsResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

// assert methods
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *OrganizationServiceImpl) assertCreateRules(ctx crud.Context, org *domain.Organization, vErrs *ft.ValidationErrors) error {
	dbOrg, err := this.orgRepo.FindBySlug(ctx, itOrg.GetOrganizationBySlugQuery{
		Slug: *org.Slug,
	})
	if err != nil {
		return err
	}

	if dbOrg != nil {
		vErrs.Append("slug", "organization slug exists")
		return nil
	}

	return nil
}

//---------------------------------------------------------------------------------------------------------------------------------------------//

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

func (this *OrganizationServiceImpl) setOrgDefaults(org *domain.Organization) {
	org.SetDefaults()
	org.Status = util.ToPtr(domain.OrgStatusActive)
}

// func (this *OrganizationServiceImpl) assertOrgUnique(ctx crud.Context, newSlug *model.Slug, vErrs *ft.ValidationErrors) error {
// 	if newSlug == nil {
// 		return nil
// 	}
// 	dbOrg, err := this.orgRepo.FindBySlug(ctx, itOrg.GetOrganizationBySlugQuery{
// 		Slug:           *newSlug,
// 		IncludeDeleted: true,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	if dbOrg != nil {
// 		vErrs.AppendAlreadyExists("slug", "organization slug")
// 	}
// 	return nil
// }

func (this *OrganizationServiceImpl) assertCorrectOrg(ctx crud.Context, org *domain.Organization, vErrs *ft.ValidationErrors) (*domain.Organization, error) {
	dbOrg, err := this.assertOrgExists(ctx, org, vErrs)
	if err != nil {
		return nil, err
	}

	if dbOrg == nil {
		vErrs.Append("slug", "organization slug not found")
		return nil, nil
	}

	if org.Etag == nil || *dbOrg.Etag != *org.Etag {
		vErrs.Append("etag", "invalid etag")
		return nil, nil
	}

	return dbOrg, nil
}

func (this *OrganizationServiceImpl) assertOrgExists(ctx crud.Context, org *domain.Organization, vErrs *ft.ValidationErrors) (dbOrg *domain.Organization, err error) {
	dbOrg, err = this.orgRepo.FindBySlug(ctx, itOrg.GetOrganizationBySlugQuery{
		Slug:           *org.Slug,
		IncludeDeleted: true,
	})
	if err != nil {
		return nil, err
	}

	if dbOrg == nil {
		vErrs.Append("slug", "organization slug not found")
		return nil, nil
	}

	return dbOrg, nil
}

func (this *OrganizationServiceImpl) getOrgBySlugFull(ctx crud.Context, query itOrg.GetOrganizationBySlugQuery, vErrs *ft.ValidationErrors) (dbOrg *domain.Organization, err error) {
	dbOrg, err = this.orgRepo.FindBySlug(ctx, query)
	if err != nil {
		return nil, err
	}

	if dbOrg == nil {
		vErrs.Append("slug", "organization slug not found")
	}
	return dbOrg, err
}

func (this *OrganizationServiceImpl) getOrgByIdFull(ctx crud.Context, query itOrg.GetOrganizationByIdQuery, vErrs *ft.ValidationErrors) (dbOrg *domain.Organization, err error) {
	dbOrg, err = this.orgRepo.FindById(ctx, query.Id)
	if err != nil {
		return nil, err
	}

	if dbOrg == nil {
		vErrs.Append("id", "organization not found")
	}

	return dbOrg, err
}

func (this *OrganizationServiceImpl) ExistsOrgById(ctx crud.Context, cmd itOrg.ExistsOrgByIdCommand) (result *itOrg.ExistsOrgByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "check if organization exists"); e != nil {
			err = e
		}
	}()

	exists, err := this.orgRepo.Exists(ctx, cmd.Id)
	ft.PanicOnErr(err)

	return &itOrg.ExistsOrgByIdResult{
		Data:    exists,
		HasData: true,
	}, nil
}
