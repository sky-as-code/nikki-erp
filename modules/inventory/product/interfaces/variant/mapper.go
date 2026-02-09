package variant

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func EntToVariant(entVariant *ent.Variant) *domain.Variant {
	variant := &domain.Variant{}
	model.MustCopy(entVariant, variant)

	// Handle relations if loaded
	// if entVariant.Edges.Product != nil {
	// 	// Handle product relation if needed
	// }

	return variant
}

func EntToVariants(entVariants []*ent.Variant) []domain.Variant {
	if entVariants == nil {
		return nil
	}
	return array.Map(entVariants, func(entVariant *ent.Variant) domain.Variant {
		return *EntToVariant(entVariant)
	})
}

func (cmd CreateVariantCommand) ToDomainModel() *domain.Variant {
	variant := &domain.Variant{}
	model.MustCopy(cmd, variant)
	return variant
}

func (cmd UpdateVariantCommand) ToDomainModel() *domain.Variant {
	variant := &domain.Variant{}
	model.MustCopy(cmd, variant)
	return variant
}

func (this DeleteVariantCommand) ToDomainModel() *domain.Variant {
	variant := &domain.Variant{}
	variant.Id = &this.Id
	return variant
}
