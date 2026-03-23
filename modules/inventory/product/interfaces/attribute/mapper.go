package attribute

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func EntToAttribute(entAttribute *ent.Attribute) *domain.Attribute {
	attribute := &domain.Attribute{}
	model.MustCopy(entAttribute, attribute)

	if entAttribute.Edges.AttributeValues != nil {
		attribute.AttributeValues = array.Map(entAttribute.Edges.AttributeValues, func(entAttributeValue *ent.AttributeValue) domain.AttributeValue {
			return *itAttributeValue.EntToAttributeValue(entAttributeValue)
		})
	}

	if entAttribute.Edges.AttributeValues != nil {
		attribute.AttributeValues = array.Map(entAttribute.Edges.AttributeValues, func(entAttributeValue *ent.AttributeValue) domain.AttributeValue {
			return *itAttributeValue.EntToAttributeValue(entAttributeValue)
		})

		for _, av := range entAttribute.Edges.AttributeValues {
			if av.Edges.Variant != nil {
				attribute.Variants = array.Map(av.Edges.Variant, func(entVariant *ent.Variant) domain.Variant {
					return *itVariant.EntToVariant(entVariant)
				})
			}
		}
		// if len(variantIds) > 0 {
		// 	attribute.VariantIds = variantIds
		// }
	}
	// if entAttribute.Edges.AttributeValues

	return attribute
}

func EntToAttributes(entAttributes []*ent.Attribute) []domain.Attribute {
	if entAttributes == nil {
		return nil
	}
	return array.Map(entAttributes, func(entAttribute *ent.Attribute) domain.Attribute {
		return *EntToAttribute(entAttribute)
	})
}

func (cmd CreateAttributeCommand) ToDomainModel() *domain.Attribute {
	attribute := &domain.Attribute{}
	model.MustCopy(cmd, attribute)
	return attribute
}

func (cmd UpdateAttributeCommand) ToDomainModel() *domain.Attribute {
	attribute := &domain.Attribute{}
	model.MustCopy(cmd, attribute)
	return attribute
}

func (this DeleteAttributeCommand) ToDomainModel() *domain.Attribute {
	attribute := &domain.Attribute{}
	model.MustCopy(this, attribute)
	return attribute
}
