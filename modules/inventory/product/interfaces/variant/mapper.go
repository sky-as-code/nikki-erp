package variant

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func (this CreateVariantCommand) ToDomainModel() *domain.Variant {
	variant := &domain.Variant{}
	model.MustCopy(this, variant)
	return variant
}

func (this UpdateVariantCommand) ToDomainModel() *domain.Variant {
	variant := &domain.Variant{}
	model.MustCopy(this, variant)
	return variant
}
