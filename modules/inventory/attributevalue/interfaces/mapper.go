package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
)

func EntToAttributeValue(entAttributeValue *ent.AttributeValue) *AttributeValue {
	attributeValue := &AttributeValue{}
	model.MustCopy(entAttributeValue, attributeValue)

	// Handle relations if loaded
	// if entAttributeValue.Edges.Attribute != nil {
	// 	// Handle attribute relation if needed
	// }

	return attributeValue
}

func EntToAttributeValues(entAttributeValues []*ent.AttributeValue) []AttributeValue {
	if entAttributeValues == nil {
		return nil
	}
	return array.Map(entAttributeValues, func(entAttributeValue *ent.AttributeValue) AttributeValue {
		return *EntToAttributeValue(entAttributeValue)
	})
}

func (cmd CreateAttributeValueCommand) ToDomainModel() *AttributeValue {
	attributeValue := &AttributeValue{}
	model.MustCopy(cmd, attributeValue)
	return attributeValue
}

func (cmd UpdateAttributeValueCommand) ToDomainModel() *AttributeValue {
	attributeValue := &AttributeValue{}
	model.MustCopy(cmd, attributeValue)
	return attributeValue
}

func (this DeleteAttributeValueCommand) ToDomainModel() *AttributeValue {
	attributeValue := &AttributeValue{}
	attributeValue.Id = &this.Id
	return attributeValue
}
