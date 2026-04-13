package attribute

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func (this CreateAttributeCommand) ToDomainModel() *domain.Attribute {
	attribute := &domain.Attribute{}
	model.MustCopy(this, attribute)
	return attribute
}

func (this UpdateAttributeCommand) ToDomainModel() *domain.Attribute {
	attribute := &domain.Attribute{}
	model.MustCopy(this, attribute)
	return attribute
}
