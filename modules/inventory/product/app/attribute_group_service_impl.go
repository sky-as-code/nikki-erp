package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

func NewAttributeGroupServiceImpl(
	attributeGroupRepo itAttributeGroup.AttributeGroupRepository,
) itAttributeGroup.AttributeGroupService {
	return &AttributeGroupServiceImpl{
		attributeGroupRepo: attributeGroupRepo,
		productSvc:         nil,
	}
}

type AttributeGroupServiceImpl struct {
	attributeGroupRepo itAttributeGroup.AttributeGroupRepository
	productSvc         itProduct.ProductService
}

func (s *AttributeGroupServiceImpl) SetProductService(productSvc itProduct.ProductService) {
	s.productSvc = productSvc
}

// Create

func (s *AttributeGroupServiceImpl) CreateAttributeGroup(ctx crud.Context, cmd itAttributeGroup.CreateAttributeGroupCommand) (*itAttributeGroup.CreateAttributeGroupResult, error) {

	attributeGroup := cmd.ToDomainModel()
	s.SetDefaults(ctx, attributeGroup)

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			s.assertCreateAttributeGroup(ctx, attributeGroup, vErrs)
			return nil
		}).
		End()

	if vErrs.Count() > 0 {
		return &itAttributeGroup.CreateAttributeGroupResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbAttribute, err := s.attributeGroupRepo.Create(ctx, attributeGroup)
	if err != nil {
		return nil, err
	}

	return &itAttributeGroup.CreateAttributeGroupResult{
		HasData: true,
		Data:    dbAttribute,
	}, nil
}

// Update

func (s *AttributeGroupServiceImpl) UpdateAttributeGroup(ctx crud.Context, cmd itAttributeGroup.UpdateAttributeGroupCommand) (*itAttributeGroup.UpdateAttributeGroupResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.AttributeGroup, itAttributeGroup.UpdateAttributeGroupCommand, itAttributeGroup.UpdateAttributeGroupResult]{
		Action:       "update attribute group",
		Command:      cmd,
		AssertExists: s.assertAttributeGroupId,
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
		AssertExists: s.assertAttributeGroupId,
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

func (s *AttributeGroupServiceImpl) SetDefaults(ctx crud.Context, attributeGroup *domain.AttributeGroup) {
	attributeGroup.SetDefaults()

	nextIndex, err := s.attributeGroupRepo.GetNextIndex(ctx, *attributeGroup.ProductId)
	ft.PanicOnErr(err)

	attributeGroup.Index = &nextIndex
}

func (s *AttributeGroupServiceImpl) assertCreateAttributeGroup(ctx crud.Context, attributeGroup *domain.AttributeGroup, vErrs *ft.ValidationErrors) error {
	product, err := s.productSvc.GetProductById(ctx, itProduct.GetProductByIdQuery{
		Id: *attributeGroup.ProductId,
	})
	ft.PanicOnErr(err)

	if product.Data == nil {
		vErrs.Append("id", "product does not exist")
		return nil
	}

	return nil
}

func (s *AttributeGroupServiceImpl) sanitizeAttributeGroup(_ *domain.AttributeGroup) {
}

func (s *AttributeGroupServiceImpl) assertAttributeGroupId(ctx crud.Context, attributeGroup *domain.AttributeGroup, vErrs *ft.ValidationErrors) (*domain.AttributeGroup, error) {
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
