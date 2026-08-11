// Package product is the Inventory module's cross-module port: the only Products capability
// other modules bind to.
//
// Consumer modules must not re-derive which product fields come from the template and which from
// the variant. They call GetEffectiveProduct and read the flattened result, so that the
// inheritance rules live in one place. See AC-PROD-032.
package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// EffectiveProduct and AttributeSelection are defined in product_types.go, in this package. They
// deliberately do not live in domain/services: the implementation depends on the contract, so
// this package must never import domain/services.

// The results are wrapped in dyn.OpResult, like every other service here: it carries the
// "succeeded but found nothing" case that a missing product is, and the REST helpers bind to it.

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
	// Products is keyed by variant id, so a caller resolving a batch can look each one up
	// without matching by position.
	Products map[string]EffectiveProduct
}

type GetEffectiveProductsResult = dyn.OpResult[GetEffectiveProductsResultData]

// ResolveProductSelectionQuery asks which concrete variant a template plus a set of chosen
// attribute values identifies.
type ResolveProductSelectionQuery struct {
	TemplateId string
	Selections []AttributeSelection

	// MaterializeIfMissing creates the variant when the combination is valid but has no
	// variant yet, which is what a DYNAMIC-mode template needs. Left false, an unknown
	// combination resolves to nothing rather than silently creating master data.
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

	// ObsoleteVariantIds are variants whose combination is no longer valid. They are reported
	// rather than deleted: one that a transaction already references must be archived instead,
	// and only the caller knows which. See BR §8.5 and AC-PROD-030.
	ObsoleteVariantIds []string

	UnchangedCount int
}

type GenerateVariantsResult = dyn.OpResult[GenerateVariantsResultData]

// ProductService is the Products capability, and the resource service installed on the Product
// Template engine.
//
// It embeds drif.DynamicResourceService so that the engine keeps serving every built-in CRUD
// action through it unchanged; the methods below are the additions a custom action reaches by
// type-asserting ProcessInput.ResourceService to this interface. See the extended-service
// pattern in docs/wiki/05. Dynamic resource engine.md §6.2.
type ProductService interface {
	drif.DynamicResourceService

	GetEffectiveProduct(ctx corectx.Context, query GetEffectiveProductQuery) (*GetEffectiveProductResult, error)
	GetEffectiveProducts(ctx corectx.Context, query GetEffectiveProductsQuery) (*GetEffectiveProductsResult, error)
	ResolveProductSelection(ctx corectx.Context, query ResolveProductSelectionQuery) (*ResolveProductSelectionResult, error)
	GenerateVariants(ctx corectx.Context, query GenerateVariantsQuery) (*GenerateVariantsResult, error)
}
