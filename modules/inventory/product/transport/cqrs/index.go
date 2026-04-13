package cqrs

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func InitCqrsHandlers() error {
	err := errors.Join(
		initAttributeHandler(),
		initAttributeGroupHandler(),
		initAttributeValueHandler(),
		initProductHandler(),
		initVariantHandler(),
		initProductCategoryHandler(),
	)
	return err
}

func initAttributeHandler() error {
	deps.Register(NewAttributeHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *AttributeHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateAttribute),
			cqrs.NewHandler(handler.UpdateAttribute),
			cqrs.NewHandler(handler.DeleteAttribute),
			cqrs.NewHandler(handler.GetAttribute),
			cqrs.NewHandler(handler.SearchAttributes),
		)
	})
}

func initAttributeGroupHandler() error {
	deps.Register(NewAttributeGroupHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *AttributeGroupHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateAttributeGroup),
			cqrs.NewHandler(handler.UpdateAttributeGroup),
			cqrs.NewHandler(handler.DeleteAttributeGroup),
			cqrs.NewHandler(handler.GetAttributeGroup),
			cqrs.NewHandler(handler.SearchAttributeGroups),
		)
	})
}

func initAttributeValueHandler() error {
	deps.Register(NewAttributeValueHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *AttributeValueHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateAttributeValue),
			cqrs.NewHandler(handler.UpdateAttributeValue),
			cqrs.NewHandler(handler.DeleteAttributeValue),
			cqrs.NewHandler(handler.GetAttributeValue),
			cqrs.NewHandler(handler.SearchAttributeValues),
		)
	})
}

func initProductHandler() error {
	deps.Register(NewProductHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *ProductHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateProduct),
			cqrs.NewHandler(handler.UpdateProduct),
			cqrs.NewHandler(handler.DeleteProduct),
			cqrs.NewHandler(handler.GetProduct),
			cqrs.NewHandler(handler.ProductExists),
			cqrs.NewHandler(handler.SetProductIsArchived),
			cqrs.NewHandler(handler.SearchProducts),
		)
	})
}

func initVariantHandler() error {
	deps.Register(NewVariantHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *VariantHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateVariant),
			cqrs.NewHandler(handler.UpdateVariant),
			cqrs.NewHandler(handler.DeleteVariant),
			cqrs.NewHandler(handler.GetVariant),
			cqrs.NewHandler(handler.SearchVariants),
		)
	})
}

func initProductCategoryHandler() error {
	deps.Register(NewProductCategoryHandler)

	return deps.Invoke(func(cqrsBus cqrs.CqrsBus, handler *ProductCategoryHandler) error {
		ctx := context.Background()
		return cqrsBus.SubscribeRequests(
			ctx,
			cqrs.NewHandler(handler.CreateProductCategory),
			cqrs.NewHandler(handler.UpdateProductCategory),
			cqrs.NewHandler(handler.DeleteProductCategory),
			cqrs.NewHandler(handler.GetProductCategory),
			cqrs.NewHandler(handler.SearchProductCategories),
		)
	})
}
