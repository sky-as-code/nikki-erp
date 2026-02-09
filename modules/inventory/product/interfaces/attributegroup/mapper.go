package attributegroup

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func EntToAttributeGroup(entAttributeGroup *ent.AttributeGroup) *domain.AttributeGroup {
	attributeGroup := &domain.AttributeGroup{}
	model.MustCopy(entAttributeGroup, attributeGroup)

	return attributeGroup
}

func EntToAttributeGroups(entAttributeGroups []*ent.AttributeGroup) []domain.AttributeGroup {
	if entAttributeGroups == nil {
		return nil
	}
	return array.Map(entAttributeGroups, func(entAttributeGroup *ent.AttributeGroup) domain.AttributeGroup {
		return *EntToAttributeGroup(entAttributeGroup)
	})
}

func (cmd CreateAttributeGroupCommand) ToDomainModel() *domain.AttributeGroup {
	attributeGroup := &domain.AttributeGroup{}
	model.MustCopy(cmd, attributeGroup)
	return attributeGroup
}

func (cmd UpdateAttributeGroupCommand) ToDomainModel() *domain.AttributeGroup {
	attributeGroup := &domain.AttributeGroup{}
	model.MustCopy(cmd, attributeGroup)
	return attributeGroup
}

func (this DeleteAttributeGroupCommand) ToDomainModel() *domain.AttributeGroup {
	attributeGroup := &domain.AttributeGroup{}
	attributeGroup.Id = &this.Id
	return attributeGroup
}
