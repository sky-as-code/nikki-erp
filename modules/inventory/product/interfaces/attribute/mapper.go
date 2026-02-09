package attribute

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func EntToAttribute(entAttribute *ent.Attribute) *domain.Attribute {
	attribute := &domain.Attribute{}
	model.MustCopy(entAttribute, attribute)

	// Handle relations if loaded
	// if entAttribute.Edges.AttributeGroup != nil {
	// 	// Handle attribute group relation if needed
	// }

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
	attribute.Id = &this.Id
	return attribute
}
