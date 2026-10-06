package services

import (
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filefield"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

const (
	productTemplateKeyPrefix = "inventory/product-templates"
	productVariantKeyPrefix  = "inventory/product-variants"
)

func ProductTemplateFileFields() []filefield.FileField {
	return []filefield.FileField{
		{
			Name:         models.ProductTemplateFileDefaultImage,
			KeyField:     models.ProductTemplateFieldDefaultImageFileKey,
			UrlField:     models.ProductTemplateFieldDefaultImageUrl,
			KeyPrefix:    productTemplateKeyPrefix,
			MaxSize:      filefield.MaxImageSize,
			AllowedMimes: &filefield.ImageMimes,
		},
	}
}

func ProductVariantFileFields() []filefield.FileField {
	return []filefield.FileField{
		{
			Name:         models.ProductVariantFileVariantImage,
			KeyField:     models.ProductVariantFieldVariantImageFileKey,
			UrlField:     models.ProductVariantFieldVariantImageUrl,
			KeyPrefix:    productVariantKeyPrefix,
			MaxSize:      filefield.MaxImageSize,
			AllowedMimes: &filefield.ImageMimes,
		},
	}
}
