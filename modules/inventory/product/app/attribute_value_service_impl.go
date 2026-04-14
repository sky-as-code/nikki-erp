package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

			valueText := attributeValue.GetValueText()
			valueDecimal := attributeValue.GetValueDecimal()
			valueInteger := attributeValue.GetValueInteger()
			valueBool := attributeValue.GetValueBool()
			valueRef := attributeValue.GetValueRef()

			switch *dataType {
			case domain.AttributeDataTypeText:
				if valueText == nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_required", "text value is required for text attribute"))
				}
				if valueDecimal != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueDecimal, "attribute_value.value_mismatch", "decimal value should be empty for text attribute"))
				}
				if valueInteger != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueInteger, "attribute_value.value_mismatch", "integer value should be empty for text attribute"))
				}
				if valueBool != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for text attribute"))
				}
				if valueRef != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for text attribute"))
				}

			case domain.AttributeDataTypeDecimal:
				if valueDecimal != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueDecimal, "attribute_value.value_mismatch", "decimal value should be empty for text attribute"))
				}
				if valueInteger != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueInteger, "attribute_value.value_mismatch", "integer value should be empty for text attribute"))
				}
				if valueText != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_mismatch", "text value should be empty for number attribute"))
				}
				if valueBool != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for number attribute"))
				}
				if valueRef != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for number attribute"))
				}

			case domain.AttributeDataTypeBoolean:
				if valueBool == nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_required", "boolean value is required for boolean attribute"))
				}
				if valueText != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_mismatch", "text value should be empty for boolean attribute"))
				}
				if valueDecimal != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueDecimal, "attribute_value.value_mismatch", "decimal value should be empty for text attribute"))
				}
				if valueInteger != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueInteger, "attribute_value.value_mismatch", "integer value should be empty for text attribute"))
				}
				if valueRef != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for boolean attribute"))
				}

			case domain.AttributeDataTypeReference:
				if valueRef == nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_required", "reference value is required for reference attribute"))
				}
				if valueText != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_mismatch", "text value should be empty for reference attribute"))
				}
				if valueDecimal != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueDecimal, "attribute_value.value_mismatch", "decimal value should be empty for text attribute"))
				}
				if valueInteger != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueInteger, "attribute_value.value_mismatch", "integer value should be empty for text attribute"))
				}
				if valueBool != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for reference attribute"))
				}
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

				valueText := attributeValue.GetValueText()
				valueDecimal := attributeValue.GetValueDecimal()
				valueInteger := attributeValue.GetValueInteger()
				valueBool := attributeValue.GetValueBool()
				valueRef := attributeValue.GetValueRef()

				switch *dataType {
				case domain.AttributeDataTypeText:
					if valueDecimal != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueDecimal, "attribute_value.value_mismatch", "decimal value should be empty for text attribute"))
					}
					if valueInteger != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueInteger, "attribute_value.value_mismatch", "integer value should be empty for text attribute"))
					}
					if valueBool != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for text attribute"))
					}
					if valueRef != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for text attribute"))
					}

				case domain.AttributeDataTypeDecimal:
					if valueText != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_mismatch", "text value should be empty for number attribute"))
					}
					if valueBool != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for number attribute"))
					}
					if valueRef != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for number attribute"))
					}

				case domain.AttributeDataTypeInteger:
					if valueText != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_mismatch", "text value should be empty for number attribute"))
					}
					if valueBool != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for number attribute"))
					}
					if valueRef != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for number attribute"))
					}

				case domain.AttributeDataTypeBoolean:
					if valueText != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_mismatch", "text value should be empty for boolean attribute"))
					}
					if valueDecimal != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueDecimal, "attribute_value.value_mismatch", "decimal value should be empty for text attribute"))
					}
					if valueInteger != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueInteger, "attribute_value.value_mismatch", "integer value should be empty for text attribute"))
					}
					if valueRef != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for boolean attribute"))
					}

				case domain.AttributeDataTypeReference:
					if valueText != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_mismatch", "text value should be empty for reference attribute"))
					}
					if valueDecimal != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueDecimal, "attribute_value.value_mismatch", "decimal value should be empty for text attribute"))
					}
					if valueInteger != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueInteger, "attribute_value.value_mismatch", "integer value should be empty for text attribute"))
					}
					if valueBool != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for reference attribute"))
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

	// Build search conditions based on attribute ID
	graph := dmodel.NewSearchGraph().
		NewCondition(domain.AttrValFieldAttributeId, dmodel.Equals, *attributeId)

	// Add value-specific conditions based on which value field is set
	valueText := attributeValue.GetValueText()
	valueDecimal := attributeValue.GetValueDecimal()
	valueInteger := attributeValue.GetValueInteger()
	valueBool := attributeValue.GetValueBool()
	valueRef := attributeValue.GetValueRef()

	if valueText != nil {
		graph.NewCondition(domain.AttrValFieldValueText, dmodel.Equals, *valueText)
	} else if valueDecimal != nil {
		graph.NewCondition(domain.AttrValFieldValueDecimal, dmodel.Equals, *valueDecimal)
	} else if valueInteger != nil {
		graph.NewCondition(domain.AttrValFieldValueInteger, dmodel.Equals, *valueInteger)
	} else if valueBool != nil {
		graph.NewCondition(domain.AttrValFieldValueBool, dmodel.Equals, *valueBool)
	} else if valueRef != nil {
		graph.NewCondition(domain.AttrValFieldValueRef, dmodel.Equals, *valueRef)
	} else {
		// No value set, cannot search
		return nil, nil
	}

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
