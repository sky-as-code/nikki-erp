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
	itAttrVal "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

func NewAttributeValueService(prodSvc itProduct.ProductService) itAttrVal.AttributeValueService {
	return prodSvc.(itAttrVal.AttributeValueService)
}

func (this *ProductServiceImpl) CreateAttributeValue(ctx corectx.Context, cmd itAttrVal.CreateAttributeValueCommand) (*itAttrVal.CreateAttributeValueResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.AttributeValue, *domain.AttributeValue]{
		Action:         "create attribute value",
		BaseRepoGetter: this.attrValueRepo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, attributeValue *domain.AttributeValue, vErrs *ft.ClientErrors) error {
			// Check if attribute exists
			attributeId := attributeValue.GetAttributeId()
			if attributeId == nil {
				return nil
			}

			attributeResult, err := this.GetAttribute(ctx, itAttr.GetAttributeQuery{Id: *attributeId})
			if err != nil {
				return err
			}
			if !attributeResult.HasData {
				vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldAttributeId, "attribute.not_found", "attribute does not exist"))
				return nil
			}

			// Validate data type compatibility
			attribute := attributeResult.Data
			dataType := attribute.GetDataType()
			if dataType == nil {
				return nil
			}

			fieldName, _ := attributeValue.GetValue()
			expectedField := domain.ExpectedValueFieldForDataType(*dataType)
			if fieldName == "" {
				vErrs.Append(*ft.NewBusinessViolation(expectedField, "attribute_value.value_required", "value is required for this attribute type"))
			} else if fieldName != expectedField {
				vErrs.Append(*ft.NewBusinessViolation(fieldName, "attribute_value.value_mismatch", "value type does not match attribute data type"))
			}

			// Check for duplicate attribute value
			existingValue, err := this.findByAttributeAndValue(ctx, attributeValue)
			if err != nil {
				return err
			}
			if existingValue != nil {
				vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldAttributeId, "attribute_value.duplicate", "attribute value already exists"))
			}

			return nil
		},
	})
}

func (this *ProductServiceImpl) UpdateAttributeValue(ctx corectx.Context, cmd itAttrVal.UpdateAttributeValueCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.AttributeValue, *domain.AttributeValue]{
		Action:       "update attribute value",
		DbRepoGetter: this.attrValueRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, attributeValue *domain.AttributeValue, foundAttributeValue *domain.AttributeValue, vErrs *ft.ClientErrors) error {
			// Check if attribute exists
			attributeId := attributeValue.GetAttributeId()
			if attributeId != nil {
				attributeResult, err := this.GetAttribute(ctx, itAttr.GetAttributeQuery{Id: *attributeId})
				if err != nil {
					return err
				}
				if !attributeResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldAttributeId, "attribute.not_found", "attribute does not exist"))
					return nil
				}

				// Validate data type compatibility
				attribute := attributeResult.Data
				dataType := attribute.GetDataType()
				if dataType == nil {
					return nil
				}

				fieldName, _ := attributeValue.GetValue()
				if fieldName != "" {
					expectedField := domain.ExpectedValueFieldForDataType(*dataType)
					if fieldName != expectedField {
						vErrs.Append(*ft.NewBusinessViolation(fieldName, "attribute_value.value_mismatch", "value type does not match attribute data type"))
					}
				}

				// Check for duplicate attribute value (excluding current one)
				existingValue, err := this.findByAttributeAndValue(ctx, attributeValue)
				if err != nil {
					return err
				}
				if existingValue != nil {
					currentId := attributeValue.GetId()
					existingId := existingValue.GetId()
					if currentId != nil && existingId != nil && *currentId != *existingId {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldAttributeId, "attribute_value.duplicate", "attribute value already exists"))
					}
				}
			}

			return nil
		},
	})
}

func (this *ProductServiceImpl) DeleteAttributeValue(ctx corectx.Context, cmd itAttrVal.DeleteAttributeValueCommand) (*itAttrVal.DeleteAttributeValueResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete attribute value",
		DbRepoGetter: this.attrValueRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ProductServiceImpl) GetAttributeValue(ctx corectx.Context, query itAttrVal.GetAttributeValueQuery) (*itAttrVal.GetAttributeValueResult, error) {
	return corecrud.GetOne[domain.AttributeValue](ctx, corecrud.GetOneParam{
		Action:       "get attribute value",
		DbRepoGetter: this.attrValueRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ProductServiceImpl) SearchAttributeValues(ctx corectx.Context, query itAttrVal.SearchAttributeValuesQuery) (*itAttrVal.SearchAttributeValuesResult, error) {
	return corecrud.Search[domain.AttributeValue](ctx, corecrud.SearchParam{
		Action:       "search attribute values",
		DbRepoGetter: this.attrValueRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *ProductServiceImpl) AttributeValueExists(ctx corectx.Context, query itAttrVal.AttributeValueExistsQuery) (*itAttrVal.AttributeValueExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "attribute value exists",
		DbRepoGetter: this.attrValueRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

// findByAttributeAndValue finds an attribute value by attribute ID and value
func (this *ProductServiceImpl) findByAttributeAndValue(ctx corectx.Context, attributeValue *domain.AttributeValue) (*domain.AttributeValue, error) {
	attributeId := attributeValue.GetAttributeId()
	if attributeId == nil {
		return nil, nil
	}

	fieldName, value := attributeValue.GetValue()
	if fieldName == "" {
		return nil, nil
	}

	graph := dmodel.NewSearchGraph().
		NewCondition(domain.AttrValFieldAttributeId, dmodel.Equals, *attributeId).
		NewCondition(fieldName, dmodel.Equals, value)

	searchResult, err := this.attrValueRepo.Search(ctx, dyn.RepoSearchParam{
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

	return &searchResult.Data.Items[0], nil
}

// BuildAttributeValue builds an AttributeValue domain object from the raw value and data type
func (this *ProductServiceImpl) BuildAttributeValue(
	attributeId model.Id,
	dataType domain.AttributeDataType,
	value any,
	codeName string,
	vErrs *ft.ClientErrors,
) *domain.AttributeValue {
	attrValue := domain.NewAttributeValue()
	attrValue.SetAttributeId(&attributeId)

	if err := attrValue.SetValueFromRaw(dataType, value); err != nil {
		vErrs.Append(*ft.NewValidationError("attributes."+codeName, "invalid_value_type", err.Error()))
		return nil
	}

	return attrValue
}

// FindOrCreateAttributeValue finds an existing attribute value or creates a new one
// Returns the AttributeValue ID
func (this *ProductServiceImpl) FindOrCreateAttributeValue(
	ctx corectx.Context,
	attribute *domain.Attribute,
	value any,
	codeName string,
	vErrs *ft.ClientErrors,
) (*model.Id, error) {
	dataType := attribute.GetDataType()
	if dataType == nil {
		return nil, nil
	}

	attributeId := attribute.GetId()
	if attributeId == nil {
		return nil, nil
	}

	// Build the attribute value based on data type
	attrValue := this.BuildAttributeValue(*attributeId, *dataType, value, codeName, vErrs)
	if attrValue == nil {
		return nil, nil
	}

	// Try to find existing attribute value
	existingValue, err := this.findByAttributeAndValue(ctx, attrValue)
	if err != nil {
		return nil, err
	}

	// If found, return existing ID
	if existingValue != nil {
		return existingValue.GetId(), nil
	}

	// Create new attribute value
	createResult, err := this.CreateAttributeValue(ctx, itAttrVal.CreateAttributeValueCommand{
		AttributeValue: *attrValue,
	})
	if err != nil {
		return nil, err
	}

	if createResult.ClientErrors != nil && createResult.ClientErrors.Count() > 0 {
		vErrs.Append(createResult.ClientErrors...)
		return nil, nil
	}

	return createResult.Data.GetId(), nil
}

// GetAttributeValueIdsByVariantId returns all AttributeValue IDs linked to the given variant
func (this *ProductServiceImpl) GetAttributeValueIdsByVariantId(ctx corectx.Context, variantId model.Id) ([]model.Id, error) {
	return this.attrValueRepo.GetIdsByVariantId(ctx, variantId)
}
