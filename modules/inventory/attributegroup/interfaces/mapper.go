package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
)

func EntToAttributeGroup(entAttributeGroup *ent.AttributeGroup) *AttributeGroup {
	attributeGroup := &AttributeGroup{}
	model.MustCopy(entAttributeGroup, attributeGroup)

	return attributeGroup
}

func EntToAttributeGroups(entAttributeGroups []*ent.AttributeGroup) []AttributeGroup {
	if entAttributeGroups == nil {
		return nil
	}
	return array.Map(entAttributeGroups, func(entAttributeGroup *ent.AttributeGroup) AttributeGroup {
		return *EntToAttributeGroup(entAttributeGroup)
	})
}

func (cmd CreateAttributeGroupCommand) ToDomainModel() *AttributeGroup {
	attributeGroup := &AttributeGroup{}
	model.MustCopy(cmd, attributeGroup)
	return attributeGroup
}

func (cmd UpdateAttributeGroupCommand) ToDomainModel() *AttributeGroup {
	attributeGroup := &AttributeGroup{}
	model.MustCopy(cmd, attributeGroup)
	return attributeGroup
}

func (this DeleteAttributeGroupCommand) ToDomainModel() *AttributeGroup {
	attributeGroup := &AttributeGroup{}
	attributeGroup.Id = &this.Id
	return attributeGroup
}
