package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
)

func EntToAttribute(entAttribute *ent.Attribute) *Attribute {
	attribute := &Attribute{}
	model.MustCopy(entAttribute, attribute)

	// Handle relations if loaded
	// if entAttribute.Edges.AttributeGroup != nil {
	// 	// Handle attribute group relation if needed
	// }

	return attribute
}

func EntToAttributes(entAttributes []*ent.Attribute) []Attribute {
	if entAttributes == nil {
		return nil
	}
	return array.Map(entAttributes, func(entAttribute *ent.Attribute) Attribute {
		return *EntToAttribute(entAttribute)
	})
}

func (cmd CreateAttributeCommand) ToDomainModel() *Attribute {
	attribute := &Attribute{}
	model.MustCopy(cmd, attribute)
	return attribute
}

func (cmd UpdateAttributeCommand) ToDomainModel() *Attribute {
	attribute := &Attribute{}
	model.MustCopy(cmd, attribute)
	return attribute
}

func (this DeleteAttributeCommand) ToDomainModel() *Attribute {
	attribute := &Attribute{}
	attribute.Id = &this.Id
	return attribute
}
