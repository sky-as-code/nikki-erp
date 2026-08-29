// Package product is the Inventory module's cross-module port: the only Products capability
// other modules bind to. Consumers must not re-derive which fields come from the template and
// which from the variant; they call GetEffectiveProduct so the inheritance rules live in one
// place.
package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// EffectiveProduct and AttributeSelection live in this package, not domain/services: the
// implementation depends on the contract, so this package must never import domain/services.

// dyn.OpResult carries the "succeeded but found nothing" case a missing product is.

type GetEffectiveProductQuery struct {
	VariantId string
}

type GetEffectiveProductResultData struct {
	Product EffectiveProduct
}

type GetEffectiveProductResult = dyn.OpResult[GetEffectiveProductResultData]

type GetEffectiveProductsQuery struct {
	VariantIds []string
}

type GetEffectiveProductsResultData struct {
	// Products is keyed by variant id, not matched by position.
	Products map[string]EffectiveProduct
}

type GetEffectiveProductsResult = dyn.OpResult[GetEffectiveProductsResultData]

// ResolveProductSelectionQuery asks which concrete variant a template plus chosen attribute
// values identifies.
type ResolveProductSelectionQuery struct {
	TemplateId string
	Selections []AttributeSelection

	// MaterializeIfMissing creates the variant when the combination is valid but has none yet, as
	// a DYNAMIC-mode template needs. Left false, an unknown combination resolves to nothing rather
	// than silently creating master data.
	MaterializeIfMissing bool
}

type ResolveProductSelectionResultData struct {
	VariantId      string
	CombinationKey string

	// Materialized reports whether this call created the variant.
	Materialized bool
}

type ResolveProductSelectionResult = dyn.OpResult[ResolveProductSelectionResultData]

type GenerateVariantsQuery struct {
	TemplateId string
}

type GenerateVariantsResultData struct {
	CreatedVariantIds []string

	// ObsoleteVariantIds are variants whose combination is no longer valid. Reported rather than
	// deleted: one a transaction already references must be archived, and only the caller knows
	// which.
	ObsoleteVariantIds []string

	UnchangedCount int
}

type GenerateVariantsResult = dyn.OpResult[GenerateVariantsResultData]

// ProductService is the Products capability and the resource service installed on the Product
// Template engine. It embeds drif.DynamicResourceService so the engine keeps serving built-in
// CRUD unchanged; a custom action reaches the extra methods by type-asserting
// ProcessInput.ResourceService to this interface.
type ProductService interface {
	drif.DynamicResourceService

	GetEffectiveProduct(ctx corectx.Context, query GetEffectiveProductQuery) (*GetEffectiveProductResult, error)
	GetEffectiveProducts(ctx corectx.Context, query GetEffectiveProductsQuery) (*GetEffectiveProductsResult, error)
	ResolveProductSelection(ctx corectx.Context, query ResolveProductSelectionQuery) (*ResolveProductSelectionResult, error)
	GenerateVariants(ctx corectx.Context, query GenerateVariantsQuery) (*GenerateVariantsResult, error)
}
