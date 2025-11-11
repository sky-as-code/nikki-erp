package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
)

func EntToVariant(entVariant *ent.Variant) *Variant {
	variant := &Variant{}
	model.MustCopy(entVariant, variant)

	// Handle relations if loaded
	// if entVariant.Edges.Product != nil {
	// 	// Handle product relation if needed
	// }

	return variant
}

func EntToVariants(entVariants []*ent.Variant) []Variant {
	if entVariants == nil {
		return nil
	}
	return array.Map(entVariants, func(entVariant *ent.Variant) Variant {
		return *EntToVariant(entVariant)
	})
}

func (cmd CreateVariantCommand) ToDomainModel() *Variant {
	variant := &Variant{}
	model.MustCopy(cmd, variant)
	return variant
}

func (cmd UpdateVariantCommand) ToDomainModel() *Variant {
	variant := &Variant{}
	model.MustCopy(cmd, variant)
	return variant
}

func (this DeleteVariantCommand) ToDomainModel() *Variant {
	variant := &Variant{}
	variant.Id = &this.Id
	return variant
}
