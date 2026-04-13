package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
)

func NewAttributeValueServiceImpl(
	repo it.AttributeValueRepository,
	attributeSvc itAttribute.AttributeService,
	cqrsBus cqrs.CqrsBus,
) it.AttributeValueService {
	return &AttributeValueServiceImpl{
		repo:         repo,
		attributeSvc: attributeSvc,
		cqrsBus:      cqrsBus,
	}
}

type AttributeValueServiceImpl struct {
	repo         it.AttributeValueRepository
	attributeSvc itAttribute.AttributeService
	cqrsBus      cqrs.CqrsBus
}

func (s *AttributeValueServiceImpl) CreateAttributeValue(ctx corectx.Context, cmd it.CreateAttributeValueCommand) (*it.CreateAttributeValueResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.AttributeValue, *domain.AttributeValue]{
		Action:         "create attribute value",
		BaseRepoGetter: s.repo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, attributeValue *domain.AttributeValue, vErrs *ft.ClientErrors) error {
			// Check if attribute exists
			attributeId := attributeValue.GetAttributeId()
			if attributeId == nil {
				return nil
			}

			attributeIdStr := string(*attributeId)
			attributeResult, err := s.attributeSvc.GetAttribute(ctx, itAttribute.GetAttributeQuery{Id: &attributeIdStr})
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
			valueNumber := attributeValue.GetValueNumber()
			valueBool := attributeValue.GetValueBool()
			valueRef := attributeValue.GetValueRef()

			switch *dataType {
			case domain.AttributeDataTypeText:
				if valueText == nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_required", "text value is required for text attribute"))
				}
				if valueNumber != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueNumber, "attribute_value.value_mismatch", "number value should be empty for text attribute"))
				}
				if valueBool != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for text attribute"))
				}
				if valueRef != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for text attribute"))
				}

			case domain.AttributeDataTypeNumber:
				if valueNumber == nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueNumber, "attribute_value.value_required", "number value is required for number attribute"))
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
				if valueNumber != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueNumber, "attribute_value.value_mismatch", "number value should be empty for boolean attribute"))
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
				if valueNumber != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueNumber, "attribute_value.value_mismatch", "number value should be empty for reference attribute"))
				}
				if valueBool != nil {
					vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for reference attribute"))
				}
			}

			// Check for duplicate attribute value
			existingValue, err := s.findByAttributeAndValue(ctx, attributeValue)
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

func (s *AttributeValueServiceImpl) UpdateAttributeValue(ctx corectx.Context, cmd it.UpdateAttributeValueCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.AttributeValue, *domain.AttributeValue]{
		Action:       "update attribute value",
		DbRepoGetter: s.repo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, attributeValue *domain.AttributeValue, foundAttributeValue *domain.AttributeValue, vErrs *ft.ClientErrors) error {
			// Check if attribute exists
			attributeId := attributeValue.GetAttributeId()
			if attributeId != nil {
				attributeIdStr := string(*attributeId)
				attributeResult, err := s.attributeSvc.GetAttribute(ctx, itAttribute.GetAttributeQuery{Id: &attributeIdStr})
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
				valueNumber := attributeValue.GetValueNumber()
				valueBool := attributeValue.GetValueBool()
				valueRef := attributeValue.GetValueRef()

				switch *dataType {
				case domain.AttributeDataTypeText:
					if valueNumber != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueNumber, "attribute_value.value_mismatch", "number value should be empty for text attribute"))
					}
					if valueBool != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for text attribute"))
					}
					if valueRef != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for text attribute"))
					}

				case domain.AttributeDataTypeNumber:
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
					if valueNumber != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueNumber, "attribute_value.value_mismatch", "number value should be empty for boolean attribute"))
					}
					if valueRef != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueRef, "attribute_value.value_mismatch", "reference value should be empty for boolean attribute"))
					}

				case domain.AttributeDataTypeReference:
					if valueText != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueText, "attribute_value.value_mismatch", "text value should be empty for reference attribute"))
					}
					if valueNumber != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueNumber, "attribute_value.value_mismatch", "number value should be empty for reference attribute"))
					}
					if valueBool != nil {
						vErrs.Append(*ft.NewBusinessViolation(domain.AttrValFieldValueBool, "attribute_value.value_mismatch", "boolean value should be empty for reference attribute"))
					}
				}

				// Check for duplicate attribute value (excluding current one)
				existingValue, err := s.findByAttributeAndValue(ctx, attributeValue)
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

func (s *AttributeValueServiceImpl) DeleteAttributeValue(ctx corectx.Context, cmd it.DeleteAttributeValueCommand) (*it.DeleteAttributeValueResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete attribute value",
		DbRepoGetter: s.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (s *AttributeValueServiceImpl) GetAttributeValue(ctx corectx.Context, query it.GetAttributeValueQuery) (*it.GetAttributeValueResult, error) {
	var q dyn.GetOneQuery
	if query.Id != nil {
		q.Id = *query.Id
	}
	q.Columns = query.Columns
	return corecrud.GetOne[domain.AttributeValue](ctx, corecrud.GetOneParam{
		Action:       "get attribute value",
		DbRepoGetter: s.repo,
		Query:        q,
	})
}

func (s *AttributeValueServiceImpl) SearchAttributeValues(ctx corectx.Context, query it.SearchAttributeValuesQuery) (*it.SearchAttributeValuesResult, error) {
	return corecrud.Search[domain.AttributeValue](ctx, corecrud.SearchParam{
		Action:       "search attribute values",
		DbRepoGetter: s.repo,
		Query:        dyn.SearchQuery(query),
	})
}

func (s *AttributeValueServiceImpl) AttributeValueExists(ctx corectx.Context, query it.AttributeValueExistsQuery) (*it.AttributeValueExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "attribute value exists",
		DbRepoGetter: s.repo,
		Query:        dyn.ExistsQuery(query),
	})
}

// Helper methods
// ---------------------------------------------------------------------------------------------------------------------------------------------

// findByAttributeAndValue finds an attribute value by attribute ID and value
func (s *AttributeValueServiceImpl) findByAttributeAndValue(ctx corectx.Context, attributeValue *domain.AttributeValue) (*domain.AttributeValue, error) {
	attributeId := attributeValue.GetAttributeId()
	if attributeId == nil {
		return nil, nil
	}

	// Build search conditions based on attribute ID
	graph := dmodel.NewSearchGraph().
		NewCondition(domain.AttrValFieldAttributeId, dmodel.Equals, *attributeId)

	// Add value-specific conditions based on which value field is set
	valueText := attributeValue.GetValueText()
	valueNumber := attributeValue.GetValueNumber()
	valueBool := attributeValue.GetValueBool()
	valueRef := attributeValue.GetValueRef()

	if valueText != nil {
		graph.NewCondition(domain.AttrValFieldValueText, dmodel.Equals, *valueText)
	} else if valueNumber != nil {
		graph.NewCondition(domain.AttrValFieldValueNumber, dmodel.Equals, *valueNumber)
	} else if valueBool != nil {
		graph.NewCondition(domain.AttrValFieldValueBool, dmodel.Equals, *valueBool)
	} else if valueRef != nil {
		graph.NewCondition(domain.AttrValFieldValueRef, dmodel.Equals, *valueRef)
	} else {
		// No value set, cannot search
		return nil, nil
	}

	searchResult, err := s.repo.Search(ctx, dyn.RepoSearchParam{
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
