package app

import (
	"encoding/json"

	"github.com/shopspring/decimal"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func NewVariantServiceImpl(
	repo it.VariantRepository,
	productSvc itProduct.ProductService,
	attributeValueSvc itAttributeValue.AttributeValueService,
	cqrsBus cqrs.CqrsBus,
) it.VariantService {
	return &VariantServiceImpl{
		repo:              repo,
		productSvc:        productSvc,
		attributeValueSvc: attributeValueSvc,
		cqrsBus:           cqrsBus,
	}
}

type VariantServiceImpl struct {
	repo              it.VariantRepository
	productSvc        itProduct.ProductService
	attributeSvc      itAttribute.AttributeService
	attributeValueSvc itAttributeValue.AttributeValueService
	cqrsBus           cqrs.CqrsBus
}

// SetAttributeService wires AttributeService to break circular dependency
func (s *VariantServiceImpl) SetAttributeService(attributeSvc itAttribute.AttributeService) {
	s.attributeSvc = attributeSvc
}

func (s *VariantServiceImpl) CreateVariant(ctx corectx.Context, cmd it.CreateVariantCommand) (*it.CreateVariantResult, error) {
	result, err := corecrud.Create(ctx, corecrud.CreateParam[domain.Variant, *domain.Variant]{
		Action:         "create variant",
		BaseRepoGetter: s.repo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, variant *domain.Variant, vErrs *ft.ClientErrors) error {
			// Check if product exists
			productId := variant.GetProductId()
			if productId != nil {
				productIdStr := string(*productId)
				productResult, err := s.productSvc.GetProduct(ctx, itProduct.GetProductQuery{Id: &productIdStr})
				if err != nil {
					return err
				}
				if !productResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation(domain.VarFieldProductId, "product.not_found", "product does not exist"))
				}
			}

			// Validate attributes if present
			return s.validateAndProcessAttributes(ctx, variant, vErrs, true)
		},
	})
	if err != nil || result == nil || result.ClientErrors != nil {
		return result, err
	}
	return result, nil
}

func (s *VariantServiceImpl) UpdateVariant(ctx corectx.Context, cmd it.UpdateVariantCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Variant, *domain.Variant]{
		Action:       "update variant",
		DbRepoGetter: s.repo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, variant *domain.Variant, foundVariant *domain.Variant, vErrs *ft.ClientErrors) error {
			// Check if product exists (if product ID is being changed)
			productId := variant.GetProductId()
			if productId != nil {
				productIdStr := string(*productId)
				productResult, err := s.productSvc.GetProduct(ctx, itProduct.GetProductQuery{Id: &productIdStr})
				if err != nil {
					return err
				}
				if !productResult.HasData {
					vErrs.Append(*ft.NewBusinessViolation(domain.VarFieldProductId, "product.not_found", "product does not exist"))
				}
			}

			// Validate attributes if present
			return s.validateAndProcessAttributes(ctx, variant, vErrs, false)
		},
	})
}

func (s *VariantServiceImpl) DeleteVariant(ctx corectx.Context, cmd it.DeleteVariantCommand) (*it.DeleteVariantResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete variant",
		DbRepoGetter: s.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (s *VariantServiceImpl) GetVariant(ctx corectx.Context, query it.GetVariantQuery) (*it.GetVariantResult, error) {
	var q dyn.GetOneQuery
	if query.Id != nil {
		q.Id = *query.Id
	}
	q.Columns = query.Columns
	return corecrud.GetOne[domain.Variant](ctx, corecrud.GetOneParam{
		Action:       "get variant",
		DbRepoGetter: s.repo,
		Query:        q,
	})
}

func (s *VariantServiceImpl) SearchVariants(ctx corectx.Context, query it.SearchVariantsQuery) (*it.SearchVariantsResult, error) {
	return corecrud.Search[domain.Variant](ctx, corecrud.SearchParam{
		Action:       "search variants",
		DbRepoGetter: s.repo,
		Query:        dyn.SearchQuery(query),
	})
}

func (s *VariantServiceImpl) VariantExists(ctx corectx.Context, query it.VariantExistsQuery) (*it.VariantExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "variant exists",
		DbRepoGetter: s.repo,
		Query:        dyn.ExistsQuery(query),
	})
}

// Helper methods
// ---------------------------------------------------------------------------------------------------------------------------------------------

// validateAndProcessAttributes validates and processes the attributes map for a variant
func (s *VariantServiceImpl) validateAndProcessAttributes(ctx corectx.Context, variant *domain.Variant, vErrs *ft.ClientErrors, isCreate bool) error {
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
		attribute, err := s.findAttributeByCodeName(ctx, *productId, codeName)
		if err != nil {
			return err
		}

		if attribute == nil {
			vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "attribute.not_found", "attribute with code name "+codeName+" not found"))
			continue
		}

		// If we're creating and have a variant ID, create the attribute value
		if isCreate && variantId != nil {
			if err := s.createAttributeValueForVariant(ctx, attribute, value, *productId, *variantId, codeName, vErrs); err != nil {
				return err
			}
		}
	}

	return nil
}

// findAttributeByCodeName finds an attribute by code name and product ID
func (s *VariantServiceImpl) findAttributeByCodeName(ctx corectx.Context, productId model.Id, codeName string) (*domain.Attribute, error) {
	if s.attributeSvc == nil {
		return nil, nil
	}

	graph := dmodel.NewSearchGraph().
		NewCondition(domain.AttrFieldProductId, dmodel.Equals, productId).
		NewCondition(domain.AttrFieldCodeName, dmodel.Equals, codeName)

	searchResult, err := s.attributeSvc.SearchAttributes(ctx, itAttribute.SearchAttributesQuery(dyn.SearchQuery{
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
func (s *VariantServiceImpl) createAttributeValueForVariant(
	ctx corectx.Context,
	attribute *domain.Attribute,
	value any,
	productId model.Id,
	variantId model.Id,
	codeName string,
	vErrs *ft.ClientErrors,
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
	case domain.AttributeDataTypeNumber:
		// Handle number type
		valueNumber, ok := value.(float64)
		if !ok {
			vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "invalid_value_type", "value must be a number"))
			return nil
		}

		// Convert to decimal
		dec := decimal.NewFromFloat(valueNumber)
		attrValue.SetValueNumber(&dec)

	case domain.AttributeDataTypeBoolean:
		// Handle boolean type
		valueBool, ok := value.(bool)
		if !ok {
			vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "invalid_value_type", "value must be a boolean"))
			return nil
		}
		attrValue.SetValueBool(&valueBool)

	case domain.AttributeDataTypeText:
		// Handle text type (LangJson)
		bytes, err := json.Marshal(value)
		if err != nil {
			vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "invalid_json_value", "value must be valid JSON"))
			return nil
		}

		var langJson model.LangJson
		err = json.Unmarshal(bytes, &langJson)
		if err != nil {
			vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "invalid_language_structure", "value must be a valid language JSON structure"))
			return nil
		}
		attrValue.SetValueText(&langJson)

	case domain.AttributeDataTypeReference:
		// Handle reference type
		valueRef, ok := value.(string)
		if !ok {
			vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "invalid_value_type", "value must be a string reference"))
			return nil
		}
		attrValue.SetValueRef(&valueRef)

	default:
		vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "unsupported_data_type", "unsupported attribute data type"))
		return nil
	}

	// Create the attribute value
	_, err := s.attributeValueSvc.CreateAttributeValue(ctx, itAttributeValue.CreateAttributeValueCommand{
		AttributeValue: attrValue,
	})
	if err != nil {
		return err
	}

	return nil
}
