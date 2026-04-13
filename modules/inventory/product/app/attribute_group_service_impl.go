package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

func NewAttributeGroupServiceImpl(
	repo it.AttributeGroupRepository,
	cqrsBus cqrs.CqrsBus,
) it.AttributeGroupService {
	return &AttributeGroupServiceImpl{
		repo:    repo,
		cqrsBus: cqrsBus,
	}
}

type AttributeGroupServiceImpl struct {
	repo       it.AttributeGroupRepository
	productSvc itProduct.ProductService
	cqrsBus    cqrs.CqrsBus
}

// SetProductService wires ProductService to break circular dependency
func (s *AttributeGroupServiceImpl) SetProductService(productSvc itProduct.ProductService) {
	s.productSvc = productSvc
}

func (s *AttributeGroupServiceImpl) CreateAttributeGroup(ctx corectx.Context, cmd it.CreateAttributeGroupCommand) (*it.CreateAttributeGroupResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.AttributeGroup, *domain.AttributeGroup]{
		Action:         "create attribute group",
		BaseRepoGetter: s.repo,
		Data:           cmd,
		BeforeValidation: func(ctx corectx.Context, attributeGroup *domain.AttributeGroup, _ *ft.ClientErrors) (*domain.AttributeGroup, error) {
			// Set the next index if not provided
			if attributeGroup.GetIndex() == nil {
				nextIndex, err := s.getNextIndex(ctx, attributeGroup.GetProductId())
				if err != nil {
					return nil, err
				}
				nextIndexInt64 := int64(nextIndex)
				attributeGroup.SetIndex(&nextIndexInt64)
			}
			return attributeGroup, nil
		},
		ValidateExtra: func(ctx corectx.Context, attributeGroup *domain.AttributeGroup, vErrs *ft.ClientErrors) error {
			// Check if product exists
			productId := attributeGroup.GetProductId()
			if productId != nil && s.productSvc != nil {
				productIdStr := string(*productId)
				productResult, err := s.productSvc.GetProduct(ctx, itProduct.GetProductQuery{Id: &productIdStr})
				if err != nil {
					return err
				}
				if !productResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrGrpFieldProductId, "product.not_found", "product does not exist"))
				}
			}
			return nil
		},
	})
}

func (s *AttributeGroupServiceImpl) UpdateAttributeGroup(ctx corectx.Context, cmd it.UpdateAttributeGroupCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.AttributeGroup, *domain.AttributeGroup]{
		Action:       "update attribute group",
		DbRepoGetter: s.repo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, attributeGroup *domain.AttributeGroup, foundAttributeGroup *domain.AttributeGroup, vErrs *ft.ClientErrors) error {
			// Check if product exists (if product ID is being changed)
			productId := attributeGroup.GetProductId()
			if productId != nil && s.productSvc != nil {
				productIdStr := string(*productId)
				productResult, err := s.productSvc.GetProduct(ctx, itProduct.GetProductQuery{Id: &productIdStr})
				if err != nil {
					return err
				}
				if !productResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrGrpFieldProductId, "product.not_found", "product does not exist"))
				}
			}
			return nil
		},
	})
}

func (s *AttributeGroupServiceImpl) DeleteAttributeGroup(ctx corectx.Context, cmd it.DeleteAttributeGroupCommand) (*it.DeleteAttributeGroupResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete attribute group",
		DbRepoGetter: s.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (s *AttributeGroupServiceImpl) GetAttributeGroup(ctx corectx.Context, query it.GetAttributeGroupQuery) (*it.GetAttributeGroupResult, error) {
	var q dyn.GetOneQuery
	if query.Id != nil {
		q.Id = *query.Id
	}
	q.Columns = query.Columns
	return corecrud.GetOne[domain.AttributeGroup](ctx, corecrud.GetOneParam{
		Action:       "get attribute group",
		DbRepoGetter: s.repo,
		Query:        q,
	})
}

func (s *AttributeGroupServiceImpl) SearchAttributeGroups(ctx corectx.Context, query it.SearchAttributeGroupsQuery) (*it.SearchAttributeGroupsResult, error) {
	return corecrud.Search[domain.AttributeGroup](ctx, corecrud.SearchParam{
		Action:       "search attribute groups",
		DbRepoGetter: s.repo,
		Query:        dyn.SearchQuery(query),
	})
}

func (s *AttributeGroupServiceImpl) AttributeGroupExists(ctx corectx.Context, query it.AttributeGroupExistsQuery) (*it.AttributeGroupExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "attribute group exists",
		DbRepoGetter: s.repo,
		Query:        dyn.ExistsQuery(query),
	})
}

// Helper methods
// ---------------------------------------------------------------------------------------------------------------------------------------------

// getNextIndex returns the next available index for a product's attribute groups
func (s *AttributeGroupServiceImpl) getNextIndex(ctx corectx.Context, productId *model.Id) (int, error) {
	if productId == nil {
		return 0, nil
	}

	// Search for all attribute groups with the given product ID to find max index
	graph := dmodel.NewSearchGraph().NewCondition(domain.AttrGrpFieldProductId, dmodel.Equals, *productId)
	searchResult, err := s.repo.Search(ctx, dyn.RepoSearchParam{
		Graph:   graph,
		Columns: []string{domain.AttrGrpFieldIndex},
		Page:    0,
		Size:    1000, // Get enough to find the max
	})

	if err != nil {
		return 0, err
	}

	if !searchResult.HasData || len(searchResult.Data.Items) == 0 {
		return 0, nil
	}

	// Find the maximum index
	maxIndex := int64(0)
	for _, item := range searchResult.Data.Items {
		index := item.GetIndex()
		if index != nil && *index > maxIndex {
			maxIndex = *index
		}
	}

	return int(maxIndex + 1), nil
}
