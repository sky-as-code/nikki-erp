package app

import (
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itHier "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewHierarchyServiceImpl(hierarchyRepo itHier.HierarchyRepository, orgSvc itOrg.OrganizationService, cqrsBus cqrs.CqrsBus) itHier.HierarchyService {
	return &HierarchyServiceImpl{
		cqrsBus:       cqrsBus,
		hierarchyRepo: hierarchyRepo,
		orgSvc:        orgSvc,
	}
}

type HierarchyServiceImpl struct {
	cqrsBus       cqrs.CqrsBus
	orgSvc        itOrg.OrganizationService
	hierarchyRepo itHier.HierarchyRepository
}

func (this *HierarchyServiceImpl) AddRemoveUsers(ctx crud.Context, cmd itHier.AddRemoveUsersCommand) (result *itHier.AddRemoveUsersResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to add or remove users"); e != nil {
			err = e
		}
	}()

	if len(cmd.Add) == 0 && len(cmd.Remove) == 0 {
		return &itHier.AddRemoveUsersResult{
			ClientError: &ft.ClientError{
				Code:    "invalid_request",
				Details: "no users to add or remove",
			},
		}, nil
	}

	var hierarchy *domain.HierarchyLevel

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertCorrectHierarchyLevel(ctx, cmd.HierarchyId, cmd.Etag, cmd.ScopeRef, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertUserIdsExist(ctx, vErrs, "add", cmd.Add, hierarchy.OrgId)
		}).
		End()

	if vErrs.Count() > 0 {
		return &itHier.AddRemoveUsersResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	cmd.Etag = *model.NewEtag()
	clientErr, err := this.hierarchyRepo.AddRemoveUsers(ctx, cmd)
	ft.PanicOnErr(err)

	// TODO: The hierarchy level or users may have been deleted by another process
	if clientErr != nil {
		return &itHier.AddRemoveUsersResult{
			ClientError: clientErr,
		}, nil
	}

	return &itHier.AddRemoveUsersResult{
		HasData: true,
		Data: &itHier.AddRemoveUsersResultData{
			Id:        cmd.HierarchyId,
			Etag:      cmd.Etag,
			UpdatedAt: time.Now(),
		},
	}, nil
}

func (this *HierarchyServiceImpl) CreateHierarchyLevel(ctx crud.Context, cmd itHier.CreateHierarchyLevelCommand) (*itHier.CreateHierarchyLevelResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.HierarchyLevel, itHier.CreateHierarchyLevelCommand, itHier.CreateHierarchyLevelResult]{
		Action:              "create hierarchy level",
		Command:             cmd,
		AssertBusinessRules: this.assertCreateRules,
		RepoCreate:          this.hierarchyRepo.Create,
		SetDefault:          this.setGroupDefaults,
		Sanitize:            this.sanitizeHierarchyLevel,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itHier.CreateHierarchyLevelResult {
			return &itHier.CreateHierarchyLevelResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.HierarchyLevel) *itHier.CreateHierarchyLevelResult {
			return &itHier.CreateHierarchyLevelResult{
				HasData: true,
				Data:    model,
			}
		},
	})

	return result, err
}

func (this *HierarchyServiceImpl) UpdateHierarchyLevel(ctx crud.Context, cmd itHier.UpdateHierarchyLevelCommand) (*itHier.UpdateHierarchyLevelResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.HierarchyLevel, itHier.UpdateHierarchyLevelCommand, itHier.UpdateHierarchyLevelResult]{
		Action:              "update hierarchy level",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateRules,
		AssertExists:        this.assertHierarchyLevelByDomain,
		RepoUpdate:          this.hierarchyRepo.Update,
		Sanitize:            this.sanitizeHierarchyLevel,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itHier.UpdateHierarchyLevelResult {
			return &itHier.UpdateHierarchyLevelResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.HierarchyLevel) *itHier.UpdateHierarchyLevelResult {
			return &itHier.UpdateHierarchyLevelResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *HierarchyServiceImpl) DeleteHierarchyLevel(ctx crud.Context, cmd itHier.DeleteHierarchyLevelCommand) (*itHier.DeleteHierarchyLevelResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.HierarchyLevel, itHier.DeleteHierarchyLevelCommand, itHier.DeleteHierarchyLevelResult]{
		Action:       "delete hierarchy level",
		Command:      cmd,
		AssertExists: this.assertHierarchyLevelByDomain,
		AssertBusinessRules: func(ctx crud.Context, cmd itHier.DeleteHierarchyLevelCommand, dbHierarchy *domain.HierarchyLevel, vErrs *ft.ValidationErrors) error {
			this.assertScopeRefMatchesOrg(cmd.ScopeRef, dbHierarchy.OrgId, vErrs)
			return nil
		},
		RepoDelete: func(ctx crud.Context, model *domain.HierarchyLevel) (int, error) {
			return this.hierarchyRepo.DeleteHard(ctx, itHier.DeleteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itHier.DeleteHierarchyLevelResult {
			return &itHier.DeleteHierarchyLevelResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.HierarchyLevel, deletedCount int) *itHier.DeleteHierarchyLevelResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *HierarchyServiceImpl) GetHierarchyLevelById(ctx crud.Context, query itHier.GetHierarchyLevelByIdQuery) (*itHier.GetHierarchyLevelByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.HierarchyLevel, itHier.GetHierarchyLevelByIdQuery, itHier.GetHierarchyLevelByIdResult]{
		Action:      "get hierarchy level by Id",
		Query:       query,
		RepoFindOne: this.getHierarchyLevelByIdFull,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itHier.GetHierarchyLevelByIdResult {
			return &itHier.GetHierarchyLevelByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.HierarchyLevel) *itHier.GetHierarchyLevelByIdResult {
			return &itHier.GetHierarchyLevelByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *HierarchyServiceImpl) SearchHierarchyLevels(ctx crud.Context, query itHier.SearchHierarchyLevelsQuery) (*itHier.SearchHierarchyLevelsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.HierarchyLevel, itHier.SearchHierarchyLevelsQuery, itHier.SearchHierarchyLevelsResult]{
		Action: "search hierarchy levels",
		Query:  query,
		SetQueryDefaults: func(query *itHier.SearchHierarchyLevelsQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.hierarchyRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itHier.SearchHierarchyLevelsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.HierarchyLevel], error) {
			return this.hierarchyRepo.Search(ctx, itHier.SearchParam{
				Predicate:      predicate,
				Order:          order,
				Page:           *query.Page,
				Size:           *query.Size,
				IncludeDeleted: query.IncludeDeleted,
				WithOrg:        query.WithOrg,
				WithParent:     query.WithParent,
				WithChildren:   query.WithChildren,
				OrgId:          query.ScopeRef,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itHier.SearchHierarchyLevelsResult {
			return &itHier.SearchHierarchyLevelsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.HierarchyLevel]) *itHier.SearchHierarchyLevelsResult {
			return &itHier.SearchHierarchyLevelsResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *HierarchyServiceImpl) ExistsHierarchyById(ctx crud.Context, cmd itHier.ExistsHierarchyLevelByIdQuery) (result *itHier.ExistsHierarchyLevelByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "check if hierarchy level exists"); e != nil {
			err = e
		}
	}()

	exists, err := this.hierarchyRepo.Exists(ctx, itHier.ExistsHierarchyLevelByIdQuery{Id: cmd.Id})
	fault.PanicOnErr(err)

	return &itHier.ExistsHierarchyLevelByIdResult{
		Data: exists,
	}, nil
}

// assert methods
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *HierarchyServiceImpl) assertCreateRules(ctx crud.Context, hierarchyLevel *domain.HierarchyLevel, vErrs *ft.ValidationErrors) error {
	this.assertScopeRefMatchesOrg(hierarchyLevel.ScopeRef, hierarchyLevel.OrgId, vErrs)

	org, err := this.orgSvc.GetOrganizationById(ctx, itOrg.GetOrganizationByIdQuery{
		Id: *hierarchyLevel.OrgId,
	})
	if err != nil {
		return err
	}

	if org.ClientError != nil {
		vErrs.MergeClientError(org.ClientError)
		return nil
	}

	return this.assertUniqueHierarchyLevelName(ctx, hierarchyLevel, vErrs)
}

func (this *HierarchyServiceImpl) assertUpdateRules(ctx crud.Context, hierarchyLevel *domain.HierarchyLevel, dbHierarchy *domain.HierarchyLevel, vErrs *ft.ValidationErrors) error {
	orgId := hierarchyLevel.OrgId
	if orgId == nil && dbHierarchy != nil {
		orgId = dbHierarchy.OrgId
	}
	this.assertScopeRefMatchesOrg(hierarchyLevel.ScopeRef, orgId, vErrs)

	if hierarchyLevel.OrgId != nil {
		org, err := this.orgSvc.GetOrganizationById(ctx, itOrg.GetOrganizationByIdQuery{
			Id: *hierarchyLevel.OrgId,
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
	}

	if hierarchyLevel.ParentId != nil {
		dbHier, err := this.assertHierarchyLevelId(ctx, *hierarchyLevel.ParentId, hierarchyLevel.ScopeRef, vErrs)
		if err != nil {
			return err
		}

		if dbHier == nil {
			vErrs.Append("id", "parent ID hierarchy level not found")
			return nil
		}
	}

	return this.assertUniqueHierarchyLevelName(ctx, hierarchyLevel, vErrs)
}

//---------------------------------------------------------------------------------------------------------------------------------------------//

func (this *HierarchyServiceImpl) assertUserIdsExist(ctx crud.Context, valErrs *ft.ValidationErrors, field string, userIds []string, orgId *model.Id) error {
	if len(userIds) == 0 || orgId == nil {
		return nil
	}

	existCmd := &itUser.UserExistsMultiQuery{
		Ids:   userIds,
		OrgId: orgId,
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

func (this *HierarchyServiceImpl) sanitizeHierarchyLevel(hierarchyLevel *domain.HierarchyLevel) {
	if hierarchyLevel.Name != nil {
		cleanedName := strings.TrimSpace(*hierarchyLevel.Name)
		cleanedName = defense.SanitizePlainText(cleanedName)
		hierarchyLevel.Name = &cleanedName
	}
}

func (this *HierarchyServiceImpl) assertCorrectHierarchyLevel(ctx crud.Context, id model.Id, etag model.Etag, scopeRef *model.Id, vErrs *ft.ValidationErrors) error {
	dbHierarchyLevel, err := this.assertHierarchyLevelId(ctx, id, scopeRef, vErrs)
	fault.PanicOnErr(err)

	if dbHierarchyLevel != nil && *dbHierarchyLevel.Etag != etag {
		vErrs.Append("etag", "invalid etag")
		return nil
	}

	return nil
}

func (this *HierarchyServiceImpl) setGroupDefaults(hierarchyLevel *domain.HierarchyLevel) {
	hierarchyLevel.SetDefaults()
}

func (this *HierarchyServiceImpl) assertHierarchyLevelByDomain(ctx crud.Context, model *domain.HierarchyLevel, vErrs *ft.ValidationErrors) (*domain.HierarchyLevel, error) {
	if model.Id == nil {
		vErrs.Append("id", "id is required")
		return nil, nil
	}
	return this.assertHierarchyLevelId(ctx, *model.Id, model.ScopeRef, vErrs)
}

func (this *HierarchyServiceImpl) assertHierarchyLevelId(ctx crud.Context, id model.Id, scopeRef *model.Id, vErrs *ft.ValidationErrors) (*domain.HierarchyLevel, error) {
	dbHierarchyLevel, err := this.hierarchyRepo.FindById(ctx, itHier.FindByIdParam{
		Id:       id,
		ScopeRef: scopeRef,
	})
	if err != nil {
		return nil, err
	}

	if dbHierarchyLevel == nil {
		vErrs.Append("id", "hierarchy level not found")
		return nil, nil
	}
	return dbHierarchyLevel, nil
}

func (this *HierarchyServiceImpl) assertUniqueHierarchyLevelName(ctx crud.Context, hierarchyLevel *domain.HierarchyLevel, vErrs *ft.ValidationErrors) error {
	dbHierarchyLevel, err := this.hierarchyRepo.FindByName(ctx, itHier.FindByNameParam{
		Name: *hierarchyLevel.Name,
	})
	if err != nil {
		return err
	}

	if dbHierarchyLevel != nil && *dbHierarchyLevel.Id != *hierarchyLevel.Id {
		vErrs.Append("name", "hierarchy level name already exists")
	}
	return nil
}

func (this *HierarchyServiceImpl) getHierarchyLevelByIdFull(ctx crud.Context, query itHier.GetHierarchyLevelByIdQuery, vErrs *ft.ValidationErrors) (*domain.HierarchyLevel, error) {
	dbHier, err := this.hierarchyRepo.FindById(ctx, itHier.FindByIdParam{
		Id:             query.Id,
		WithChildren:   query.WithChildren,
		IncludeDeleted: query.IncludeDeleted,
		ScopeRef:       query.ScopeRef,
	})
	if err != nil {
		return nil, err
	}

	if dbHier == nil {
		vErrs.Append("id", "hierarchy level not found")
		return nil, nil
	}
	return dbHier, nil
}

func (this *HierarchyServiceImpl) assertScopeRefMatchesOrg(scopeRef *model.Id, orgId *model.Id, vErrs *ft.ValidationErrors) {
	if scopeRef == nil {
		return
	}
	if orgId == nil || *scopeRef != *orgId {
		vErrs.AppendNotAllowed("scopeRef", "scopeRef must match hierarchy's orgId")
	}
}
