package attributevalue

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func (this CreateAttributeValueCommand) ToDomainModel() *domain.AttributeValue {
	attributeValue := &domain.AttributeValue{}
	model.MustCopy(this, attributeValue)
	return attributeValue
}

func (this UpdateAttributeValueCommand) ToDomainModel() *domain.AttributeValue {
	attributeValue := &domain.AttributeValue{}
	model.MustCopy(this, attributeValue)
	return attributeValue
}
