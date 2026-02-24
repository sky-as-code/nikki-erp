package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func InitServices() error {
	// Đăng ký các services theo thứ tự để tránh circular dependency
	if err := errors.Join(
		deps.Register(NewAttributeGroupServiceImpl),
		deps.Register(NewAttributeValueServiceImpl),
		deps.Register(NewAttributeServiceImpl), // Đăng ký AttributeService trước VariantService
		deps.Register(NewProductServiceImpl),   // ProductService không depend VariantService
		deps.Register(NewVariantServiceImpl),   // VariantService không depend AttributeService
	); err != nil {
		return err
	}

	// Wire VariantService vào ProductService
	err := deps.Invoke(func(
		productSvc itProduct.ProductService,
		variantSvc itVariant.VariantService,
	) {
		if impl, ok := productSvc.(*ProductServiceImpl); ok {
			impl.SetVariantService(variantSvc)
		}
	})
	if err != nil {
		return err
	}

	// Wire AttributeService vào VariantService
	return deps.Invoke(func(
		attributeSvc itAttribute.AttributeService,
		variantSvc itVariant.VariantService,
	) {
		if impl, ok := variantSvc.(*VariantServiceImpl); ok {
			impl.SetAttributeService(attributeSvc)
		}
	})
}
