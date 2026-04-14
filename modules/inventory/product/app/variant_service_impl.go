package app

import (
	"encoding/json"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func NewVariantService(prodSvc itProduct.ProductService) it.VariantService {
	return prodSvc.(it.VariantService)
}

func (this *ProductServiceImpl) CreateVariant(ctx corectx.Context, cmd it.CreateVariantCommand) (*it.CreateVariantResult, error) {
	result, err := corecrud.Create(ctx, corecrud.CreateParam[domain.Variant, *domain.Variant]{
		Action:         "create variant",
		BaseRepoGetter: this.variantRepo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, variant *domain.Variant, vErrs *ft.ClientErrors) error {
			// Check if product exists
			productId := variant.GetProductId()
			if productId != nil {
				productResult, err := this.GetProduct(ctx, itProduct.GetProductQuery{Id: *productId})
				if err != nil {
					return err
				}
				if !productResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation(domain.VarFieldProductId, "product.not_found", "product does not exist"))
				}
			}

			// Validate attributes if present
			return this.validateAndProcessAttributes(ctx, variant, vErrs, true)
		},
	})
	if err != nil || result == nil || result.ClientErrors != nil {
		return result, err
	}
	return result, nil
}

func (this *ProductServiceImpl) UpdateVariant(ctx corectx.Context, cmd it.UpdateVariantCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Variant, *domain.Variant]{
		Action:       "update variant",
		DbRepoGetter: this.variantRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, variant *domain.Variant, foundVariant *domain.Variant, vErrs *ft.ClientErrors) error {
			// Check if product exists (if product ID is being changed)
			productId := variant.GetProductId()
			if productId != nil {
				productResult, err := this.GetProduct(ctx, itProduct.GetProductQuery{Id: *productId})
				if err != nil {
					return err
				}
				if !productResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation(domain.VarFieldProductId, "product.not_found", "product does not exist"))
				}
			}

			// Validate attributes if present
			return this.validateAndProcessAttributes(ctx, variant, vErrs, false)
		},
	})
}

func (this *ProductServiceImpl) DeleteVariant(ctx corectx.Context, cmd it.DeleteVariantCommand) (*it.DeleteVariantResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete variant",
		DbRepoGetter: this.variantRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ProductServiceImpl) GetVariant(ctx corectx.Context, query it.GetVariantQuery) (*it.GetVariantResult, error) {
	return corecrud.GetOne[domain.Variant](ctx, corecrud.GetOneParam{
		Action:       "get variant",
		DbRepoGetter: this.variantRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ProductServiceImpl) SearchVariants(ctx corectx.Context, query it.SearchVariantsQuery) (*it.SearchVariantsResult, error) {
	return corecrud.Search[domain.Variant](ctx, corecrud.SearchParam{
		Action:       "search variants",
		DbRepoGetter: this.variantRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *ProductServiceImpl) VariantExists(ctx corectx.Context, query it.VariantExistsQuery) (*it.VariantExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "variant exists",
		DbRepoGetter: this.variantRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

// validateAndProcessAttributes validates and processes the attributes map for a variant
func (this *ProductServiceImpl) validateAndProcessAttributes(ctx corectx.Context, variant *domain.Variant, vErrs *ft.ClientErrors, isCreate bool) error {
	// Check if there's an attributes field in the dynamic fields
	attributes := variant.GetFieldData().GetAny("attributes")
	if attributes == nil {
		return nil
	}

	attrMap, ok := attributes.(map[string]any)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation("attributes", "invalid_format", "attributes must be a map"))
		return nil
	}

	productId := variant.GetProductId()
	if productId == nil {
		return nil
	}

	variantId := variant.GetId()

	// Process each attribute in the map
	for codeName, value := range attrMap {
		// Find the attribute by code name
		attribute, err := this.findAttributeByCodeName(ctx, *productId, codeName)
		if err != nil {
			return err
		}

		if attribute == nil {
			vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "attribute.not_found", "attribute with code name "+codeName+" not found"))
			continue
		}

		// If we're creating and have a variant ID, create the attribute value
		if isCreate && variantId != nil {
			if err := this.createAttributeValueForVariant(ctx, attribute, value, codeName, vErrs); err != nil {
				return err
			}
		}
	}

	return nil
}

// findAttributeByCodeName finds an attribute by code name and product ID
func (this *ProductServiceImpl) findAttributeByCodeName(ctx corectx.Context, productId model.Id, codeName string) (*domain.Attribute, error) {
	graph := dmodel.NewSearchGraph().
		NewCondition(domain.AttrFieldProductId, dmodel.Equals, productId).
		NewCondition(domain.AttrFieldCodeName, dmodel.Equals, codeName)

	searchResult, err := this.SearchAttributes(ctx, itAttribute.SearchAttributesQuery(dyn.SearchQuery{
		Graph: graph,
		Page:  0,
		Size:  1,
	}))

	if err != nil {
		return nil, err
	}

	if !searchResult.HasData || len(searchResult.Data.Items) == 0 {
		return nil, nil
	}

	return &searchResult.Data.Items[0], nil
}

// createAttributeValueForVariant creates an attribute value for a variant based on the attribute data type
func (this *ProductServiceImpl) createAttributeValueForVariant(
	ctx corectx.Context, attribute *domain.Attribute, value any,
	codeName string, vErrs *ft.ClientErrors,
) error {
	dataType := attribute.GetDataType()
	if dataType == nil {
		return nil
	}

	attributeId := attribute.GetId()
	if attributeId == nil {
		return nil
	}

	var attrValue domain.AttributeValue
	attrValue.SetAttributeId(attributeId)

	switch *dataType {
	case domain.AttributeDataTypeDecimal:
		valueDecimal, ok := value.(string)
		if !ok {
			vErrs.Append(*ft.NewValidationError("attributes."+codeName, "invalid_value_type", "value must be a number"))
			return nil
		}

		attrValue.SetValueDecimal(&valueDecimal)

	case domain.AttributeDataTypeInteger:
		valueInt, ok := value.(int64)
		if !ok {
			vErrs.Append(*ft.NewValidationError("attributes."+codeName, "invalid_value_type", "value must be an integer"))
			return nil
		}

		attrValue.SetValueInteger(&valueInt)

	case domain.AttributeDataTypeBoolean:
		// Handle boolean type
		valueBool, ok := value.(bool)
		if !ok {
			vErrs.Append(*ft.NewValidationError("attributes."+codeName, "invalid_value_type", "value must be a boolean"))
			return nil
		}
		attrValue.SetValueBool(&valueBool)

	case domain.AttributeDataTypeText:
		// Handle text type (LangJson)
		bytes, err := json.Marshal(value)
		if err != nil {
			vErrs.Append(*ft.NewValidationError("attributes."+codeName, "invalid_json_value", "value must be valid JSON"))
			return nil
		}

		var langJson model.LangJson
		err = json.Unmarshal(bytes, &langJson)
		if err != nil {
			vErrs.Append(*ft.NewValidationError("attributes."+codeName, "invalid_language_structure", "value must be a valid language JSON structure"))
			return nil
		}
		attrValue.SetValueText(&langJson)

	case domain.AttributeDataTypeReference:
		// Handle reference type
		valueRef, ok := value.(string)
		if !ok {
			vErrs.Append(*ft.NewValidationError("attributes."+codeName, "invalid_value_type", "value must be a string reference"))
			return nil
		}
		attrValue.SetValueRef(&valueRef)

	default:
		vErrs.Append(*ft.NewValidationError("attributes."+codeName, "unsupported_data_type", "unsupported attribute data type"))
		return nil
	}

	// Create the attribute value
	_, err := this.CreateAttributeValue(ctx, itAttributeValue.CreateAttributeValueCommand{
		AttributeValue: attrValue,
	})
	if err != nil {
		return err
	}

	return nil
}
