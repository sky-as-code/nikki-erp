package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
)

func NewAttributeGroupServiceImpl(
	attributeGroupRepo itAttributeGroup.AttributeGroupRepository,
) itAttributeGroup.AttributeGroupService {
	return &AttributeGroupServiceImpl{
		attributeGroupRepo: attributeGroupRepo,
	}
}

type AttributeGroupServiceImpl struct {
	attributeGroupRepo itAttributeGroup.AttributeGroupRepository
}

// Create

func (s *AttributeGroupServiceImpl) CreateAttributeGroup(ctx crud.Context, cmd itAttributeGroup.CreateAttributeGroupCommand) (*itAttributeGroup.CreateAttributeGroupResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.AttributeGroup, itAttributeGroup.CreateAttributeGroupCommand, itAttributeGroup.CreateAttributeGroupResult]{
		Action:     "create attribute group",
		Command:    cmd,
		RepoCreate: s.attributeGroupRepo.Create,
		Sanitize:   s.sanitizeAttributeGroup,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeGroup.CreateAttributeGroupResult {
			return &itAttributeGroup.CreateAttributeGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.AttributeGroup) *itAttributeGroup.CreateAttributeGroupResult {
			return &itAttributeGroup.CreateAttributeGroupResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (s *AttributeGroupServiceImpl) UpdateAttributeGroup(ctx crud.Context, cmd itAttributeGroup.UpdateAttributeGroupCommand) (*itAttributeGroup.UpdateAttributeGroupResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.AttributeGroup, itAttributeGroup.UpdateAttributeGroupCommand, itAttributeGroup.UpdateAttributeGroupResult]{
		Action:       "update attribute group",
		Command:      cmd,
		AssertExists: s.assertAttributeGroupIdExists,
		RepoUpdate:   s.attributeGroupRepo.Update,
		Sanitize:     s.sanitizeAttributeGroup,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeGroup.UpdateAttributeGroupResult {
			return &itAttributeGroup.UpdateAttributeGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.AttributeGroup) *itAttributeGroup.UpdateAttributeGroupResult {
			return &itAttributeGroup.UpdateAttributeGroupResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *AttributeGroupServiceImpl) DeleteAttributeGroup(ctx crud.Context, cmd itAttributeGroup.DeleteAttributeGroupCommand) (*itAttributeGroup.DeleteAttributeGroupResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.AttributeGroup, itAttributeGroup.DeleteAttributeGroupCommand, itAttributeGroup.DeleteAttributeGroupResult]{
		Action:       "delete attribute group",
		Command:      cmd,
		AssertExists: s.assertAttributeGroupIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.AttributeGroup) (int, error) {
			return s.attributeGroupRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeGroup.DeleteAttributeGroupResult {
			return &itAttributeGroup.DeleteAttributeGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *domain.AttributeGroup, deletedCount int) *itAttributeGroup.DeleteAttributeGroupResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *AttributeGroupServiceImpl) GetAttributeGroupById(ctx crud.Context, query itAttributeGroup.GetAttributeGroupByIdQuery) (*itAttributeGroup.GetAttributeGroupByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.AttributeGroup, itAttributeGroup.GetAttributeGroupByIdQuery, itAttributeGroup.GetAttributeGroupByIdResult]{
		Action: "get attribute group by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itAttributeGroup.GetAttributeGroupByIdQuery, vErrs *ft.ValidationErrors) (*domain.AttributeGroup, error) {
			dbAttributeGroup, err := s.attributeGroupRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbAttributeGroup == nil {
				vErrs.AppendNotFound("id", "attribute group id")
			}
			return dbAttributeGroup, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeGroup.GetAttributeGroupByIdResult {
			return &itAttributeGroup.GetAttributeGroupByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.AttributeGroup) *itAttributeGroup.GetAttributeGroupByIdResult {
			return &itAttributeGroup.GetAttributeGroupByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (s *AttributeGroupServiceImpl) SearchAttributeGroups(ctx crud.Context, query itAttributeGroup.SearchAttributeGroupsQuery) (*itAttributeGroup.SearchAttributeGroupsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.AttributeGroup, itAttributeGroup.SearchAttributeGroupsQuery, itAttributeGroup.SearchAttributeGroupsResult]{
		Action: "search attribute groups",
		Query:  query,
		SetQueryDefaults: func(q *itAttributeGroup.SearchAttributeGroupsQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return s.attributeGroupRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query itAttributeGroup.SearchAttributeGroupsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.AttributeGroup], error) {
			return s.attributeGroupRepo.Search(ctx, itAttributeGroup.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
				ProductId: query.ProductId,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeGroup.SearchAttributeGroupsResult {
			return &itAttributeGroup.SearchAttributeGroupsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[domain.AttributeGroup]) *itAttributeGroup.SearchAttributeGroupsResult {
			return &itAttributeGroup.SearchAttributeGroupsResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *AttributeGroupServiceImpl) sanitizeAttributeGroup(_ *domain.AttributeGroup) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *AttributeGroupServiceImpl) assertAttributeGroupIdExists(ctx crud.Context, attributeGroup *domain.AttributeGroup, vErrs *ft.ValidationErrors) (*domain.AttributeGroup, error) {
	dbAttributeGroup, err := s.attributeGroupRepo.FindById(ctx, itAttributeGroup.FindByIdParam{
		Id: *attributeGroup.Id,
	})
	if err != nil {
		return nil, err
	}

	if dbAttributeGroup == nil {
		vErrs.Append("id", "attribute group not found")
		return nil, nil
	}

	return dbAttributeGroup, nil
}
