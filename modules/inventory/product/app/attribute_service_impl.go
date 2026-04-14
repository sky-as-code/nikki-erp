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
			// Check if product exists
			productId := attribute.GetProductId()
			if productId != nil {
				productResult, err := this.GetProduct(ctx, itProduct.GetProductQuery{Id: *productId})
				if err != nil {
					return err
				}
				if !productResult.HasData {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldProductId, "product.not_found", "product does not exist"))
				}
			}

			// Check if codeName is unique for this product
			codeName := attribute.GetCodeName()
			if codeName != nil && productId != nil {
				existingAttr, err := this.findByCodeName(ctx, *productId, *codeName)
				if err != nil {
					return err
				}
				if existingAttr != nil {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldCodeName, "attribute.code_name_exists", "attribute code name already exists for this product"))
				}
			}

			// Validate enum logic
			isEnum := attribute.GetIsEnum()
			enumValue := attribute.GetEnumValue()
			dataType := attribute.GetDataType()

			if isEnum == nil || !*isEnum {
				// If not enum, enumValue should be nil or empty
				if len(enumValue) > 0 {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldEnumValue, "attribute.enum_value_not_allowed", "enum value should be empty when isEnum is false"))
				}
				return nil
			}

			// If isEnum is true, validate enumValue based on dataType
			if dataType == nil {
				return nil
			}

			if *dataType == domain.AttributeDataTypeText {
				if len(enumValue) == 0 {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldEnumValue, "attribute.enum_value_required", "enum value is required when isEnum is true"))
				}
			} else if *dataType == domain.AttributeDataTypeDecimal {
				if len(enumValue) == 0 {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldEnumValue, "attribute.enum_value_required", "enum value is required when isEnum is true"))
				}
			} else if *dataType != domain.AttributeDataTypeBoolean && *dataType != domain.AttributeDataTypeReference {
				vErrs.Append(*ft.NewValidationError(domain.AttrFieldDataType, "attribute.invalid_enum_type", "only text and number data types are allowed for enum attribute"))
			}

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
			// Check if product exists (if product ID is being changed)
			productId := attribute.GetProductId()
			if productId != nil {
				productResult, err := this.GetProduct(ctx, itProduct.GetProductQuery{Id: *productId})
				if err != nil {
					return err
				}
				if !productResult.HasData {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldProductId, "product.not_found", "product does not exist"))
				}
			}

			// Check if codeName is unique for this product (excluding current attribute)
			codeName := attribute.GetCodeName()
			attrId := attribute.GetId()
			if codeName != nil && productId != nil {
				existingAttr, err := this.findByCodeName(ctx, *productId, *codeName)
				if err != nil {
					return err
				}
				// If found, check if it's not the current attribute
				if existingAttr != nil {
					existingId := existingAttr.GetId()
					if existingId != nil && attrId != nil && *existingId != *attrId {
						vErrs.Append(*ft.NewValidationError(domain.AttrFieldCodeName, "attribute.code_name_exists", "attribute code name already exists for this product"))
					}
				}
			}

			// Validate enum logic (same as create)
			isEnum := attribute.GetIsEnum()
			enumValue := attribute.GetEnumValue()
			dataType := attribute.GetDataType()

			if isEnum == nil || !*isEnum {
				if len(enumValue) > 0 {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldEnumValue, "attribute.enum_value_not_allowed", "enum value should be empty when isEnum is false"))
				}
				return nil
			}

			if dataType == nil {
				return nil
			}

			if *dataType == domain.AttributeDataTypeText {
				if len(enumValue) == 0 {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldEnumValue, "attribute.enum_value_required", "enum value is required when isEnum is true"))
				}
			} else if *dataType == domain.AttributeDataTypeDecimal {
				if len(enumValue) == 0 {
					vErrs.Append(*ft.NewValidationError(domain.AttrFieldEnumValue, "attribute.enum_value_required", "enum value is required when isEnum is true"))
				}
			} else if *dataType != domain.AttributeDataTypeBoolean && *dataType != domain.AttributeDataTypeReference {
				vErrs.Append(*ft.NewValidationError(domain.AttrFieldDataType, "attribute.invalid_enum_type", "only text and number data types are allowed for enum attribute"))
			}

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
	return corecrud.GetOne[domain.Attribute](ctx, corecrud.GetOneParam{
		Action:       "get attribute",
		DbRepoGetter: this.attrRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ProductServiceImpl) SearchAttributes(ctx corectx.Context, query itAttr.SearchAttributesQuery) (*itAttr.SearchAttributesResult, error) {
	return corecrud.Search[domain.Attribute](ctx, corecrud.SearchParam{
		Action:       "search attributes",
		DbRepoGetter: this.attrRepo,
		Query:        dyn.SearchQuery(query),
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
		Graph:   graph,
		Columns: []string{domain.AttrFieldSortIndex},
		Page:    0,
		Size:    1000, // Get enough to find the max
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

// findByCodeName finds an attribute by code name and product ID
func (this *ProductServiceImpl) findByCodeName(ctx corectx.Context, productId model.Id, codeName string) (*domain.Attribute, error) {
	graph := dmodel.NewSearchGraph().
		NewCondition(domain.AttrFieldProductId, dmodel.Equals, productId).
		NewCondition(domain.AttrFieldCodeName, dmodel.Equals, codeName)
	searchResult, err := this.attrRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})

	if err != nil {
		return nil, err
	}

	if !searchResult.HasData || len(searchResult.Data.Items) == 0 {
		return nil, nil
	}

	attr := &searchResult.Data.Items[0]
	return attr, nil
}
