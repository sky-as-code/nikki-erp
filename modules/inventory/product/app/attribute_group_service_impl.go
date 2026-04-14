package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttrGrp "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

func NewAttributeGroupService(prodSvc itProduct.ProductService) itAttrGrp.AttributeGroupService {
	return prodSvc.(itAttrGrp.AttributeGroupService)
}

func (this *ProductServiceImpl) CreateAttributeGroup(ctx corectx.Context, cmd itAttrGrp.CreateAttributeGroupCommand) (*itAttrGrp.CreateAttributeGroupResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.AttributeGroup, *domain.AttributeGroup]{
		Action:         "create attribute group",
		BaseRepoGetter: this.attrGrpRepo,
		Data:           cmd,
		BeforeValidation: func(ctx corectx.Context, attributeGroup *domain.AttributeGroup, _ *ft.ClientErrors) (*domain.AttributeGroup, error) {
			// Set the next index if not provided
			if attributeGroup.GetIndex() == nil {
				nextIndex, err := this.getNextIndex(ctx, attributeGroup.GetProductId())
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
			if productId != nil {
				productResult, err := this.GetProduct(ctx, itProduct.GetProductQuery{Id: *productId})
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

func (this *ProductServiceImpl) UpdateAttributeGroup(ctx corectx.Context, cmd itAttrGrp.UpdateAttributeGroupCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.AttributeGroup, *domain.AttributeGroup]{
		Action:       "update attribute group",
		DbRepoGetter: this.attrGrpRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, attributeGroup *domain.AttributeGroup, foundAttributeGroup *domain.AttributeGroup, vErrs *ft.ClientErrors) error {
			// Check if product exists (if product ID is being changed)
			productId := attributeGroup.GetProductId()
			if productId != nil {
				productResult, err := this.GetProduct(ctx, itProduct.GetProductQuery{Id: *productId})
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

func (this *ProductServiceImpl) DeleteAttributeGroup(ctx corectx.Context, cmd itAttrGrp.DeleteAttributeGroupCommand) (*itAttrGrp.DeleteAttributeGroupResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete attribute group",
		DbRepoGetter: this.attrGrpRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ProductServiceImpl) GetAttributeGroup(ctx corectx.Context, query itAttrGrp.GetAttributeGroupQuery) (*itAttrGrp.GetAttributeGroupResult, error) {
	return corecrud.GetOne[domain.AttributeGroup](ctx, corecrud.GetOneParam{
		Action:       "get attribute group",
		DbRepoGetter: this.attrGrpRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ProductServiceImpl) SearchAttributeGroups(ctx corectx.Context, query itAttrGrp.SearchAttributeGroupsQuery) (*itAttrGrp.SearchAttributeGroupsResult, error) {
	return corecrud.Search[domain.AttributeGroup](ctx, corecrud.SearchParam{
		Action:       "search attribute groups",
		DbRepoGetter: this.attrGrpRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *ProductServiceImpl) AttributeGroupExists(ctx corectx.Context, query itAttrGrp.AttributeGroupExistsQuery) (*itAttrGrp.AttributeGroupExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "attribute group exists",
		DbRepoGetter: this.attrGrpRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

// getNextIndex returns the next available index for a product's attribute groups
func (this *ProductServiceImpl) getNextIndex(ctx corectx.Context, productId *model.Id) (int, error) {
	if productId == nil {
		return 0, nil
	}

	// Search for all attribute groups with the given product ID to find max index
	graph := dmodel.NewSearchGraph().NewCondition(domain.AttrGrpFieldProductId, dmodel.Equals, *productId)
	searchResult, err := this.attrGrpRepo.Search(ctx, dyn.RepoSearchParam{
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
