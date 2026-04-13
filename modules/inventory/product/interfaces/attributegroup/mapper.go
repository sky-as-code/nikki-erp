package attributegroup

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func (this CreateAttributeGroupCommand) ToDomainModel() *domain.AttributeGroup {
	attributeGroup := &domain.AttributeGroup{}
	model.MustCopy(this, attributeGroup)
	return attributeGroup
}

func (this UpdateAttributeGroupCommand) ToDomainModel() *domain.AttributeGroup {
	attributeGroup := &domain.AttributeGroup{}
	model.MustCopy(this, attributeGroup)
	return attributeGroup
}
