package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttr "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

func NewAttributeService(prodSvc itProduct.ProductService) itAttr.AttributeService {
	return prodSvc.(itAttr.AttributeService)
}

func (this *ProductServiceImpl) CreateAttribute(ctx corectx.Context, cmd itAttr.CreateAttributeCommand) (*itAttr.CreateAttributeResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Attribute, *domain.Attribute]{
		Action:         "create attribute",
		BaseRepoGetter: this.attrRepo,
		Data:           cmd,
		BeforeValidation: func(ctx corectx.Context, attribute *domain.Attribute, _ *ft.ClientErrors) (*domain.Attribute, error) {
			if attribute.GetSortIndex() == nil {
				nextIndex, err := this.getNextSortIndex(ctx, attribute.GetProductId())
				if err != nil {
					return nil, err
				}
				nextIndexInt64 := int64(nextIndex)
				attribute.SetSortIndex(&nextIndexInt64)
			}
			return attribute, nil
		},
		ValidateExtra: func(ctx corectx.Context, attribute *domain.Attribute, vErrs *ft.ClientErrors) error {
			// Validate product and enum values
			if err := this.validateAttributeProduct(ctx, attribute, vErrs, true); err != nil {
				return err
			}
			this.validateAttributeEnumValues(attribute, vErrs)

			return nil
		},
	})
}

func (this *ProductServiceImpl) UpdateAttribute(ctx corectx.Context, cmd itAttr.UpdateAttributeCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Attribute, *domain.Attribute]{
		Action:       "update attribute",
		DbRepoGetter: this.attrRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, attribute *domain.Attribute, foundAttribute *domain.Attribute, vErrs *ft.ClientErrors) error {
			// Validate product and enum values
			if err := this.validateAttributeProduct(ctx, attribute, vErrs, false); err != nil {
				return err
			}
			this.validateAttributeEnumValues(attribute, vErrs)

			return nil
		},
	})
}

func (this *ProductServiceImpl) DeleteAttribute(ctx corectx.Context, cmd itAttr.DeleteAttributeCommand) (*itAttr.DeleteAttributeResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete attribute",
		DbRepoGetter: this.attrRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ProductServiceImpl) GetAttribute(ctx corectx.Context, query itAttr.GetAttributeQuery) (*itAttr.GetAttributeResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itAttr.GetAttributeResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itAttr.GetAttributeQuery))

	graph := dmodel.NewSearchGraph()
	graph.And(
		*dmodel.NewSearchNode().Condition(dmodel.NewCondition(domain.AttrFieldProductId, dmodel.Equals, query.ProductId)),
		*dmodel.NewSearchNode().Condition(dmodel.NewCondition("id", dmodel.Equals, query.Id)),
	)
	searchResult, err := this.attrRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:  graph,
		Fields: query.Columns,
		Page:   0,
		Size:   1,
	})
	if err != nil {
		return nil, err
	}
	if searchResult.ClientErrors.Count() > 0 {
		return &itAttr.GetAttributeResult{ClientErrors: searchResult.ClientErrors}, nil
	}

	var result itAttr.GetAttributeResult
	result.HasData = searchResult.HasData
	if searchResult.HasData {
		result.Data = searchResult.Data.Items[0]
	}
	return &result, nil
}

func (this *ProductServiceImpl) SearchAttributes(ctx corectx.Context, query itAttr.SearchAttributesQuery) (*itAttr.SearchAttributesResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itAttr.SearchAttributesResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itAttr.SearchAttributesQuery))

	cond := dmodel.NewCondition(domain.AttrFieldProductId, dmodel.Equals, query.ProductId)
	graph := dmodel.NewSearchGraph()
	if query.Graph != nil {
		node := query.Graph.ToSearchNode()
		graph.And(
			*dmodel.NewSearchNode().Condition(cond),
			*node,
		)
	} else {
		graph.Condition(cond)
	}
	return corecrud.Search[domain.Attribute](ctx, corecrud.SearchParam{
		Action:       "search attributes",
		DbRepoGetter: this.attrRepo,
		Query: dyn.SearchQuery{
			Fields: query.Columns,
			Graph:  graph,
			Page:   query.Page,
			Size:   query.Size,
		},
	})
}

func (this *ProductServiceImpl) AttributeExists(ctx corectx.Context, query itAttr.AttributeExistsQuery) (*itAttr.AttributeExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "attribute exists",
		DbRepoGetter: this.attrRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

// getNextSortIndex returns the next available sort index for a product's attributes
func (this *ProductServiceImpl) getNextSortIndex(ctx corectx.Context, productId *model.Id) (int, error) {
	if productId == nil {
		return 0, nil
	}

	// Search for all attributes with the given product ID to find max sort index
	graph := dmodel.NewSearchGraph().NewCondition(domain.AttrFieldProductId, dmodel.Equals, *productId)
	searchResult, err := this.attrRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:  graph,
		Fields: []string{domain.AttrFieldSortIndex},
		Page:   0,
		Size:   1000, // Get enough to find the max
	})

	if err != nil {
		return 0, err
	}

	if !searchResult.HasData || len(searchResult.Data.Items) == 0 {
		return 0, nil
	}

	// Find the maximum sort index
	maxSortIndex := int64(0)
	for _, item := range searchResult.Data.Items {
		sortIndex := item.GetSortIndex()
		if sortIndex != nil && *sortIndex > maxSortIndex {
			maxSortIndex = *sortIndex
		}
	}

	return int(maxSortIndex + 1), nil
}

// validateAttributeProduct validates that the product exists and is not archived (for create)
func (this *ProductServiceImpl) validateAttributeProduct(ctx corectx.Context, attribute *domain.Attribute, vErrs *ft.ClientErrors, checkArchived bool) error {
	productId := attribute.GetProductId()
	if productId == nil {
		return nil
	}

	productResult, err := this.GetProduct(ctx, itProduct.GetProductQuery{Id: *productId})
	if err != nil {
		return err
	}
	if productResult == nil || !productResult.HasData {
		vErrs.Append(*ft.NewBusinessViolation(domain.AttrFieldProductId, "product.not_found", "product does not exist"))
		return nil
	}

	if checkArchived && productResult.Data.IsArchived() != nil && *productResult.Data.IsArchived() {
		vErrs.Append(*ft.NewValidationError(domain.AttrFieldProductId, "product.archived", "cannot add attribute to archived product"))
	}

	return nil
}

// validateAttributeEnumValues validates that enum values match the data type
func (this *ProductServiceImpl) validateAttributeEnumValues(attribute *domain.Attribute, vErrs *ft.ClientErrors) {
	isEnum := attribute.GetIsEnum()
	dataType := attribute.GetDataType()

	// Only validate if is_enum is true
	if isEnum == nil || !*isEnum {
		return
	}

	if dataType == nil {
		return
	}

	switch *dataType {
	case domain.AttributeDataTypeNumber:
		// If data type is number and is enum, enum_value_number must not be nil
		enumValueNumber := attribute.GetEnumValueNumber()
		if len(enumValueNumber) == 0 {
			vErrs.Append(*ft.NewValidationError(
				domain.AttrFieldEnumValueNumber,
				"attribute.enum_value_number_required",
				"enum_value_number is required when data_type is number and is_enum is true",
			))
		}

	case domain.AttributeDataTypeText:
		// If data type is text and is enum, enum_value_text must not be nil
		enumValueText := attribute.GetEnumValueText()
		if len(enumValueText) == 0 {
			vErrs.Append(*ft.NewValidationError(
				domain.AttrFieldEnumValueText,
				"attribute.enum_value_text_required",
				"enum_value_text is required when data_type is text and is_enum is true",
			))
		}
	}
}
