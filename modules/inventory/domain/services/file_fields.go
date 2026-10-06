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
			KeyField:     models.ProductTemplateFieldDefaultImageKey,
			UrlField:     models.ProductTemplateFieldDefaultImage,
			UploadField:  models.ProductTemplateFieldDefaultImage,
			KeyPrefix:    productTemplateKeyPrefix,
			MaxSize:      filefield.MaxImageSize,
			AllowedMimes: &filefield.ImageMimes,
		},
	}
}

func ProductVariantFileFields() []filefield.FileField {
	return []filefield.FileField{
		{
			KeyField:     models.ProductVariantFieldVariantImageKey,
			UrlField:     models.ProductVariantFieldVariantImage,
			UploadField:  models.ProductVariantFieldVariantImage,
			KeyPrefix:    productVariantKeyPrefix,
			MaxSize:      filefield.MaxImageSize,
			AllowedMimes: &filefield.ImageMimes,
		},
	}
}
