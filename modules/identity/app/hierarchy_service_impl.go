package app

import (
	"context"
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itHier "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewHierarchyServiceImpl(hierarchyRepo itHier.HierarchyRepository, cqrsBus cqrs.CqrsBus) itHier.HierarchyService {
	return &HierarchyServiceImpl{
		cqrsBus:       cqrsBus,
		hierarchyRepo: hierarchyRepo,
	}
}

type HierarchyServiceImpl struct {
	cqrsBus       cqrs.CqrsBus
	hierarchyRepo itHier.HierarchyRepository
}

func (this *HierarchyServiceImpl) AddRemoveUsers(ctx context.Context, cmd itHier.AddRemoveUsersCommand) (result *itHier.AddRemoveUsersResult, err error) {
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

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertCorrectHierarchyLevel(ctx, cmd.HierarchyId, cmd.Etag, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertUserIdsExist(ctx, vErrs, "add", cmd.Add)
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
		Data: &itHier.AddRemoveUsersResultData{
			Id:        cmd.HierarchyId,
			Etag:      cmd.Etag,
			UpdatedAt: time.Now(),
		},
	}, nil
}

func (this *HierarchyServiceImpl) assertUserIdsExist(ctx context.Context, valErrs *ft.ValidationErrors, field string, userIds []string) error {
	if len(userIds) == 0 {
		return nil
	}

	existCmd := &itUser.UserExistsMultiCommand{
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

func (this *HierarchyServiceImpl) CreateHierarchyLevel(ctx context.Context, cmd itHier.CreateHierarchyLevelCommand) (result *itHier.CreateHierarchyLevelResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create hierarchy level"); e != nil {
			err = e
		}
	}()

	hierarchyLevel := cmd.ToHierarchyLevel()
	hierarchyLevel.SetDefaults()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = hierarchyLevel.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeHierarchyLevel(hierarchyLevel)
			return this.assertUniqueHierarchyLevelName(ctx, hierarchyLevel, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itHier.CreateHierarchyLevelResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdHierarchyLevel, err := this.hierarchyRepo.Create(ctx, *hierarchyLevel)
	ft.PanicOnErr(err)

	return &itHier.CreateHierarchyLevelResult{
		Data:    createdHierarchyLevel,
		HasData: createdHierarchyLevel != nil,
	}, nil
}

func (this *HierarchyServiceImpl) sanitizeHierarchyLevel(hierarchyLevel *domain.HierarchyLevel) {
	if hierarchyLevel.Name != nil {
		cleanedName := strings.TrimSpace(*hierarchyLevel.Name)
		cleanedName = defense.SanitizePlainText(cleanedName)
		hierarchyLevel.Name = &cleanedName
	}
}

func (this *HierarchyServiceImpl) UpdateHierarchyLevel(ctx context.Context, cmd itHier.UpdateHierarchyLevelCommand) (result *itHier.UpdateHierarchyLevelResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update hierarchy level"); e != nil {
			err = e
		}
	}()

	hierarchyLevel := cmd.ToHierarchyLevel()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = hierarchyLevel.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertCorrectHierarchyLevel(ctx, *hierarchyLevel.Id, *hierarchyLevel.Etag, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			// Sanitize after we've made sure this is the correct hierarchy level
			this.sanitizeHierarchyLevel(hierarchyLevel)
			return this.assertUniqueHierarchyLevelName(ctx, hierarchyLevel, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itHier.UpdateHierarchyLevelResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := hierarchyLevel.Etag
	hierarchyLevel.Etag = model.NewEtag()
	hierarchyLevelWithOrg, err := this.hierarchyRepo.Update(ctx, *hierarchyLevel, *prevEtag)
	ft.PanicOnErr(err)

	return &itHier.UpdateHierarchyLevelResult{
		Data:    hierarchyLevelWithOrg,
		HasData: hierarchyLevelWithOrg != nil,
	}, err
}

func (this *HierarchyServiceImpl) DeleteHierarchyLevel(ctx context.Context, cmd itHier.DeleteHierarchyLevelCommand) (result *itHier.DeleteHierarchyLevelResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete hierarchy level"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	_, err = this.assertHierarchyLevelIdExists(ctx, cmd.Id, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itHier.DeleteHierarchyLevelResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	deletedCount, err := this.hierarchyRepo.DeleteHard(ctx, cmd.Id)
	ft.PanicOnErr(err)
	if deletedCount == 0 {
		vErrs.AppendNotFound("id", "hierarchy")
		return &itHier.DeleteHierarchyLevelResult{
			ClientError: vErrs.ToClientError(),
		}, nil

	}

	return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount), nil
}

func (this *HierarchyServiceImpl) GetHierarchyLevelById(ctx context.Context, query itHier.GetHierarchyLevelByIdQuery) (result *itHier.GetHierarchyLevelByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get hierarchy level"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	dbHierarchyLevel, err := this.assertHierarchyLevelIdExists(ctx, query.Id, &vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itHier.GetHierarchyLevelByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &itHier.GetHierarchyLevelByIdResult{
		Data:    dbHierarchyLevel,
		HasData: dbHierarchyLevel != nil,
	}, nil
}

func (thisSvc *HierarchyServiceImpl) SearchHierarchyLevels(ctx context.Context, query itHier.SearchHierarchyLevelsQuery) (result *itHier.SearchHierarchyLevelsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list hierarchy levels"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := thisSvc.hierarchyRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &itHier.SearchHierarchyLevelsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	hierarchyLevels, err := thisSvc.hierarchyRepo.Search(ctx, itHier.SearchParam{
		Predicate:      predicate,
		Order:          order,
		Page:           *query.Page,
		Size:           *query.Size,
		WithOrg:        query.WithOrg,
		WithParent:     query.WithParent,
		WithChildren:   query.WithChildren,
		IncludeDeleted: query.IncludeDeleted,
	})
	ft.PanicOnErr(err)

	return &itHier.SearchHierarchyLevelsResult{
		Data:    hierarchyLevels,
		HasData: hierarchyLevels != nil,
	}, nil
}

func (this *HierarchyServiceImpl) assertCorrectHierarchyLevel(ctx context.Context, id model.Id, etag model.Etag, vErrs *ft.ValidationErrors) error {
	dbHierarchyLevel, err := this.assertHierarchyLevelIdExists(ctx, id, vErrs)
	if err != nil {
		return err
	}

	if dbHierarchyLevel != nil && *dbHierarchyLevel.Etag != etag {
		vErrs.Append("etag", "hierarchy level has been modified by another user")
		return nil
	}

	return nil
}

func (this *HierarchyServiceImpl) assertHierarchyLevelIdExists(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (*domain.HierarchyLevel, error) {
	dbHierarchyLevel, err := this.hierarchyRepo.FindById(ctx, itHier.FindByIdParam{
		Id: id,
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

func (this *HierarchyServiceImpl) assertUniqueHierarchyLevelName(ctx context.Context, hierarchyLevel *domain.HierarchyLevel, vErrs *ft.ValidationErrors) error {
	dbHierarchyLevel, err := this.hierarchyRepo.FindByName(ctx, itHier.FindByNameParam{
		Name: *hierarchyLevel.Name,
	})
	if err != nil {
		return err
	}

	if dbHierarchyLevel != nil {
		vErrs.Append("name", "hierarchy level name already exists")
	}
	return nil
}
