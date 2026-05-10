package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func NewVariantService(prodSvc itProduct.ProductService) it.VariantService {
	return prodSvc.(it.VariantService)
}

func (this *ProductServiceImpl) CreateVariant(ctx corectx.Context, cmd it.CreateVariantCommand) (*it.CreateVariantResult, error) {
	var attrValueIds []model.Id

	result, err := corecrud.Create(ctx, corecrud.CreateParam[domain.Variant, *domain.Variant]{
		Action:         "create variant",
		BaseRepoGetter: this.variantRepo,
		Data:           cmd,
		BeforeValidation: func(ctx corectx.Context, variant *domain.Variant, vErrs *ft.ClientErrors) (*domain.Variant, error) {
			attributes := extractAttributesFromFieldData(variant)
			ids, err := this.findOrCreateAttributeValues(ctx, variant, attributes, vErrs)
			if err != nil {
				return variant, err
			}
			attrValueIds = ids
			return variant, nil
		},
		ValidateExtra: func(ctx corectx.Context, variant *domain.Variant, vErrs *ft.ClientErrors) error {
			return this.validateVariantProduct(ctx, variant, vErrs)
		},
	})
	if err != nil || result == nil || result.ClientErrors != nil {
		return result, err
	}

	// Link attribute values to the variant
	if len(attrValueIds) > 0 {
		variantId := result.Data.GetId()
		if variantId != nil {
			if err := this.LinkAttributeValuesToVariant(ctx, *variantId, attrValueIds); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (this *ProductServiceImpl) UpdateVariant(ctx corectx.Context, cmd it.UpdateVariantCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	var attrValueIds []model.Id
	var variantId *model.Id
	var attributesProvided bool
	var savedAttributes map[string]any

	result, err := corecrud.Update(ctx, corecrud.UpdateParam[domain.Variant, *domain.Variant]{
		Action:       "update variant",
		DbRepoGetter: this.variantRepo,
		Data:         cmd,
		BeforeValidation: func(ctx corectx.Context, v *domain.Variant, vErrs *ft.ClientErrors) (*domain.Variant, error) {
			// Extract attributes before schema.Validate strips it (attributes is not a schema field)
			attrs := extractAttributesFromFieldData(v)
			if attrs != nil {
				attributesProvided = true
				savedAttributes = attrs
			}
			return v, nil
		},
		ValidateExtra: func(ctx corectx.Context, v *domain.Variant, foundVariant *domain.Variant, vErrs *ft.ClientErrors) error {
			// Check if product exists (if product ID is being changed)
			if err := this.validateVariantProduct(ctx, v, vErrs); err != nil {
				return err
			}

			if !attributesProvided {
				return nil
			}

			// Get variant ID from foundVariant (reliable DB record)
			variantId = foundVariant.GetId()

			// Use product_id from request body, or fall back to the DB record's product_id
			productId := v.GetProductId()
			if productId == nil {
				productId = foundVariant.GetProductId()
				if productId != nil {
					v.SetProductId(productId)
				}
			}

			ids, err := this.findOrCreateAttributeValues(ctx, v, savedAttributes, vErrs)
			if err != nil {
				return err
			}
			attrValueIds = ids

			return nil
		},
	})
	if err != nil || result == nil || result.ClientErrors != nil {
		return result, err
	}

	// Replace attribute values for the variant only when attributes were explicitly provided
	if attributesProvided && variantId != nil {
		if err := this.ReplaceAttributeValuesForVariant(ctx, *variantId, attrValueIds); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (this *ProductServiceImpl) DeleteVariant(ctx corectx.Context, cmd it.DeleteVariantCommand) (*it.DeleteVariantResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete variant",
		DbRepoGetter: this.variantRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ProductServiceImpl) GetVariant(ctx corectx.Context, query it.GetVariantQuery) (*it.GetVariantResult, error) {
	result, err := corecrud.GetOne[domain.Variant](ctx, corecrud.GetOneParam{
		Action:       "get variant",
		DbRepoGetter: this.variantRepo,
		Query:        dyn.GetOneQuery(query),
	})
	if err != nil || result == nil || !result.HasData {
		return result, err
	}

	// Populate attributes field
	if err := this.populateVariantAttributes(ctx, &result.Data); err != nil {
		return nil, err
	}

	return result, nil
}

func (this *ProductServiceImpl) SearchVariants(ctx corectx.Context, query it.SearchVariantsQuery) (*it.SearchVariantsResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &it.SearchVariantsResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*it.SearchVariantsQuery))

	cond := dmodel.NewCondition(domain.VarFieldProductId, dmodel.Equals, query.ProductId)
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
	result, err := corecrud.Search[domain.Variant](ctx, corecrud.SearchParam{
		Action:       "search variants",
		DbRepoGetter: this.variantRepo,
		Query: dyn.SearchQuery{
			Fields: query.Columns,
			Graph:  graph,
			Page:   query.Page,
			Size:   query.Size,
		},
	})
	if err != nil || result == nil || !result.HasData {
		return result, err
	}

	// Populate attributes field for each variant
	for i := range result.Data.Items {
		if err := this.populateVariantAttributes(ctx, &result.Data.Items[i]); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (this *ProductServiceImpl) VariantExists(ctx corectx.Context, query it.VariantExistsQuery) (*it.VariantExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "variant exists",
		DbRepoGetter: this.variantRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

// validateVariantProduct validates that the product exists
func (this *ProductServiceImpl) validateVariantProduct(ctx corectx.Context, variant *domain.Variant, vErrs *ft.ClientErrors) error {
	productId := variant.GetProductId()
	if productId == nil {
		return nil
	}

	productResult, err := this.GetProduct(ctx, itProduct.GetProductQuery{Id: *productId})
	if err != nil {
		return err
	}

	if !productResult.HasData {
		vErrs.Append(*ft.NewBusinessViolation(domain.VarFieldProductId, "product.not_found", "product does not exist"))
	}

	return nil
}

// findOrCreateAttributeValues finds or creates attribute values for the given attributes map
// Returns a list of AttributeValue IDs that should be linked to the variant
func (this *ProductServiceImpl) findOrCreateAttributeValues(
	ctx corectx.Context,
	variant *domain.Variant,
	attributes map[string]any,
	vErrs *ft.ClientErrors,
) ([]model.Id, error) {
	if attributes == nil {
		return nil, nil
	}

	productId := variant.GetProductId()
	if productId == nil {
		return nil, nil
	}

	var attrValueIds []model.Id

	// Process each attribute in the map
	for codeName, value := range attributes {
		// Find the attribute by code name
		attribute, err := this.findAttributeByCodeName(ctx, *productId, codeName)
		if err != nil {
			return nil, err
		}

		if attribute == nil {
			vErrs.Append(*ft.NewBusinessViolation("attributes."+codeName, "attribute.not_found", "attribute with code name "+codeName+" not found"))
			continue
		}

		// Find or create the attribute value
		attrValueId, err := this.FindOrCreateAttributeValue(ctx, attribute, value, codeName, vErrs)
		if err != nil {
			return nil, err
		}

		if attrValueId != nil {
			attrValueIds = append(attrValueIds, *attrValueId)
		}
	}

	return attrValueIds, nil
}

// LinkAttributeValuesToVariant links attribute values to a variant using Many-to-Many relationship (public method)
func (this *ProductServiceImpl) LinkAttributeValuesToVariant(
	ctx corectx.Context,
	variantId model.Id,
	attrValueIds []model.Id,
) error {
	if len(attrValueIds) == 0 {
		return nil
	}

	repo := this.variantRepo.GetBaseRepo()

	// Convert to Set
	associatedIds := make(map[model.Id]struct{})
	for _, id := range attrValueIds {
		associatedIds[id] = struct{}{}
	}

	_, err := repo.ManageM2m(ctx, dyn.RepoManageM2mParam{
		DestSchemaName:     domain.AttributeValueSchemaName,
		SrcId:              variantId,
		SrcIdFieldForError: domain.VarFieldId,
		SrcEdgeName:        domain.VarEdgeAttributeValues,
		AssociatedIds:      associatedIds,
		DisassociatedIds:   make(map[model.Id]struct{}),
	})

	return err
}

// ReplaceAttributeValuesForVariant replaces all attribute values for a variant (public method)
func (this *ProductServiceImpl) ReplaceAttributeValuesForVariant(
	ctx corectx.Context,
	variantId model.Id,
	newAttrValueIds []model.Id,
) error {
	repo := this.variantRepo.GetBaseRepo()

	// Get current attribute value IDs via the AttributeValue service
	currentSlice, err := this.GetAttributeValueIdsByVariantId(ctx, variantId)
	if err != nil {
		return err
	}

	// Build sets
	currentIds := make(map[model.Id]struct{}, len(currentSlice))
	for _, id := range currentSlice {
		currentIds[id] = struct{}{}
	}

	newIds := make(map[model.Id]struct{}, len(newAttrValueIds))
	for _, id := range newAttrValueIds {
		newIds[id] = struct{}{}
	}

	// Compute diff
	associatedIds := make(map[model.Id]struct{})
	for id := range newIds {
		if _, exists := currentIds[id]; !exists {
			associatedIds[id] = struct{}{}
		}
	}

	disassociatedIds := make(map[model.Id]struct{})
	for id := range currentIds {
		if _, exists := newIds[id]; !exists {
			disassociatedIds[id] = struct{}{}
		}
	}

	if len(associatedIds) == 0 && len(disassociatedIds) == 0 {
		return nil
	}

	_, err = repo.ManageM2m(ctx, dyn.RepoManageM2mParam{
		DestSchemaName:     domain.AttributeValueSchemaName,
		SrcId:              variantId,
		SrcIdFieldForError: domain.VarFieldId,
		SrcEdgeName:        domain.VarEdgeAttributeValues,
		AssociatedIds:      associatedIds,
		DisassociatedIds:   disassociatedIds,
	})

	return err
}

// findAttributeByCodeName finds an attribute by code name and product ID
func (this *ProductServiceImpl) findAttributeByCodeName(ctx corectx.Context, productId model.Id, codeName string) (*domain.Attribute, error) {
	graph := dmodel.NewSearchGraph().
		NewCondition(domain.AttrFieldCodeName, dmodel.Equals, codeName)

	searchResult, err := this.SearchAttributes(ctx, itAttribute.SearchAttributesQuery{
		ProductId: productId,
		Graph:     graph,
		Page:      0,
		Size:      1,
	})

	if err != nil {
		return nil, err
	}

	if !searchResult.HasData || len(searchResult.Data.Items) == 0 {
		return nil, nil
	}

	return &searchResult.Data.Items[0], nil
}

func extractAttributesFromFieldData(variant *domain.Variant) map[string]any {
	raw := variant.GetFieldData().GetAny("attributes")
	if raw == nil {
		return nil
	}
	attrs, _ := raw.(map[string]any)
	return attrs
}

func (this *ProductServiceImpl) populateVariantAttributes(ctx corectx.Context, variant *domain.Variant) error {
	variantId := variant.GetId()
	if variantId == nil {
		return nil
	}

	// Get attribute value IDs via the AttributeValue service
	attrValueIds, err := this.GetAttributeValueIdsByVariantId(ctx, *variantId)
	if err != nil {
		return err
	}

	if len(attrValueIds) == 0 {
		variant.SetAttributes(make(map[string]any))
		return nil
	}

	attrValGraph := dmodel.NewSearchGraph().NewCondition(domain.AttrValFieldId, dmodel.In, anySlice(attrValueIds)...)
	attrValResult, err := this.attrValueRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: attrValGraph,
		Page:  0,
		Size:  100,
	})
	if err != nil {
		return err
	}

	if !attrValResult.HasData || len(attrValResult.Data.Items) == 0 {
		variant.SetAttributes(make(map[string]any))
		return nil
	}

	// Collect attribute IDs and build value map
	attributeIds := make([]model.Id, 0)
	attrValueMap := make(map[model.Id]any) // attribute_id -> value

	for _, attrVal := range attrValResult.Data.Items {
		attrId := attrVal.GetAttributeId()
		if attrId == nil {
			continue
		}

		// Get the actual value from the attribute value
		_, value := attrVal.GetValue()
		if value != nil {
			attributeIds = append(attributeIds, *attrId)
			attrValueMap[*attrId] = value
		}
	}

	if len(attributeIds) == 0 {
		variant.SetAttributes(make(map[string]any))
		return nil
	}

	// Query attributes to get code names
	attrGraph := dmodel.NewSearchGraph().NewCondition(domain.AttrFieldId, dmodel.In, anySlice(attributeIds)...)
	attrResult, err := this.attrRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:  attrGraph,
		Fields: []string{domain.AttrFieldId, domain.AttrFieldCodeName},
		Page:   0,
		Size:   100,
	})
	if err != nil {
		return err
	}

	if !attrResult.HasData || len(attrResult.Data.Items) == 0 {
		variant.SetAttributes(make(map[string]any))
		return nil
	}

	// Build the final attributes map: code_name -> value
	attributesMap := make(map[string]any)
	for _, attr := range attrResult.Data.Items {
		attrId := attr.GetId()
		codeName := attr.GetCodeName()
		if attrId != nil && codeName != nil {
			if value, ok := attrValueMap[*attrId]; ok {
				attributesMap[*codeName] = value
			}
		}
	}

	variant.SetAttributes(attributesMap)
	return nil
}

// anySlice converts a slice of model.Id to a slice of any
func anySlice[T any](items []T) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}
