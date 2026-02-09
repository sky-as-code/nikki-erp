package attributevalue

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func EntToAttributeValue(entAttributeValue *ent.AttributeValue) *domain.AttributeValue {
	attributeValue := &domain.AttributeValue{}
	model.MustCopy(entAttributeValue, attributeValue)

	// Handle relations if loaded
	// if entAttributeValue.Edges.Attribute != nil {
	// 	// Handle attribute relation if needed
	// }

	return attributeValue
}

func EntToAttributeValues(entAttributeValues []*ent.AttributeValue) []domain.AttributeValue {
	if entAttributeValues == nil {
		return nil
	}
	return array.Map(entAttributeValues, func(entAttributeValue *ent.AttributeValue) domain.AttributeValue {
		return *EntToAttributeValue(entAttributeValue)
	})
}

func (cmd CreateAttributeValueCommand) ToDomainModel() *domain.AttributeValue {
	attributeValue := &domain.AttributeValue{}
	model.MustCopy(cmd, attributeValue)
	return attributeValue
}

func (cmd UpdateAttributeValueCommand) ToDomainModel() *domain.AttributeValue {
	attributeValue := &domain.AttributeValue{}
	model.MustCopy(cmd, attributeValue)
	return attributeValue
}

func (this DeleteAttributeValueCommand) ToDomainModel() *domain.AttributeValue {
	attributeValue := &domain.AttributeValue{}
	attributeValue.Id = &this.Id
	return attributeValue
}
