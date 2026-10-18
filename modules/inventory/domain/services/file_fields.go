package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable/filefield"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

const (
	productTemplateKeyPrefix = "inventory/product-templates"
	productVariantKeyPrefix  = "inventory/product-variants"
)

func ProductTemplateFileFields() []filefield.FileField {
	return []filefield.FileField{
		filefield.ImageField(models.ProductTemplateFileDefaultImage, models.ProductTemplateFieldDefaultImageFileKey,
			models.ProductTemplateFieldDefaultImageUrl, productTemplateKeyPrefix),
	}
}

func ProductVariantFileFields() []filefield.FileField {
	return []filefield.FileField{
		filefield.ImageField(models.ProductVariantFileVariantImage, models.ProductVariantFieldVariantImageFileKey,
			models.ProductVariantFieldVariantImageUrl, productVariantKeyPrefix),
	}
}
