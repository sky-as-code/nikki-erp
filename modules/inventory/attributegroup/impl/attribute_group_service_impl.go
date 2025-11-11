package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attributegroup/interfaces"
)

func NewAttributeGroupServiceImpl(
	attributeGroupRepo it.AttributeGroupRepository,
) it.AttributeGroupService {
	return &AttributeGroupServiceImpl{
		attributeGroupRepo: attributeGroupRepo,
	}
}

type AttributeGroupServiceImpl struct {
	attributeGroupRepo it.AttributeGroupRepository
}

// Create

func (s *AttributeGroupServiceImpl) CreateAttributeGroup(ctx crud.Context, cmd it.CreateAttributeGroupCommand) (*it.CreateAttributeGroupResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*it.AttributeGroup, it.CreateAttributeGroupCommand, it.CreateAttributeGroupResult]{
		Action:     "create attribute group",
		Command:    cmd,
		RepoCreate: s.attributeGroupRepo.Create,
		Sanitize:   s.sanitizeAttributeGroup,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateAttributeGroupResult {
			return &it.CreateAttributeGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.AttributeGroup) *it.CreateAttributeGroupResult {
			return &it.CreateAttributeGroupResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (s *AttributeGroupServiceImpl) UpdateAttributeGroup(ctx crud.Context, cmd it.UpdateAttributeGroupCommand) (*it.UpdateAttributeGroupResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*it.AttributeGroup, it.UpdateAttributeGroupCommand, it.UpdateAttributeGroupResult]{
		Action:       "update attribute group",
		Command:      cmd,
		AssertExists: s.assertAttributeGroupIdExists,
		RepoUpdate:   s.attributeGroupRepo.Update,
		Sanitize:     s.sanitizeAttributeGroup,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateAttributeGroupResult {
			return &it.UpdateAttributeGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.AttributeGroup) *it.UpdateAttributeGroupResult {
			return &it.UpdateAttributeGroupResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *AttributeGroupServiceImpl) DeleteAttributeGroup(ctx crud.Context, cmd it.DeleteAttributeGroupCommand) (*it.DeleteAttributeGroupResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*it.AttributeGroup, it.DeleteAttributeGroupCommand, it.DeleteAttributeGroupResult]{
		Action:       "delete attribute group",
		Command:      cmd,
		AssertExists: s.assertAttributeGroupIdExists,
		RepoDelete: func(ctx crud.Context, model *it.AttributeGroup) (int, error) {
			return s.attributeGroupRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteAttributeGroupResult {
			return &it.DeleteAttributeGroupResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *it.AttributeGroup, deletedCount int) *it.DeleteAttributeGroupResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *AttributeGroupServiceImpl) GetAttributeGroupById(ctx crud.Context, query it.GetAttributeGroupByIdQuery) (*it.GetAttributeGroupByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*it.AttributeGroup, it.GetAttributeGroupByIdQuery, it.GetAttributeGroupByIdResult]{
		Action: "get attribute group by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetAttributeGroupByIdQuery, vErrs *ft.ValidationErrors) (*it.AttributeGroup, error) {
			dbAttributeGroup, err := s.attributeGroupRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbAttributeGroup == nil {
				vErrs.AppendNotFound("id", "attribute group id")
			}
			return dbAttributeGroup, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetAttributeGroupByIdResult {
			return &it.GetAttributeGroupByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.AttributeGroup) *it.GetAttributeGroupByIdResult {
			return &it.GetAttributeGroupByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (s *AttributeGroupServiceImpl) SearchAttributeGroups(ctx crud.Context, query it.SearchAttributeGroupsQuery) (*it.SearchAttributeGroupsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[it.AttributeGroup, it.SearchAttributeGroupsQuery, it.SearchAttributeGroupsResult]{
		Action: "search attribute groups",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchAttributeGroupsQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return s.attributeGroupRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query it.SearchAttributeGroupsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[it.AttributeGroup], error) {
			return s.attributeGroupRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
				ProductId: query.ProductId,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchAttributeGroupsResult {
			return &it.SearchAttributeGroupsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[it.AttributeGroup]) *it.SearchAttributeGroupsResult {
			return &it.SearchAttributeGroupsResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *AttributeGroupServiceImpl) sanitizeAttributeGroup(_ *it.AttributeGroup) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *AttributeGroupServiceImpl) assertAttributeGroupIdExists(ctx crud.Context, attributeGroup *it.AttributeGroup, vErrs *ft.ValidationErrors) (*it.AttributeGroup, error) {
	dbAttributeGroup, err := s.attributeGroupRepo.FindById(ctx, it.FindByIdParam{
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
