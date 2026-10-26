// Package product is the Inventory module's cross-module port: the only Products capability
// other modules bind to. Consumers must not re-derive which fields come from the template and
// which from the variant; they call GetEffectiveProduct so the inheritance rules live in one
// place.
package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
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

// GenerateVariantsQuery reconciles a template's variants with its current configuration.
// DeleteAllowed is the caller's product_variant delete permission, checked by the application
// service: an invalid variant nothing uses is deleted only when it is true.
type GenerateVariantsQuery struct {
	TemplateId    string
	DeleteAllowed bool
}

type GenerateVariantsResultData struct {
	CreatedVariantIds  []string
	ExtendedVariantIds []string
	RestoredVariantIds []string
	ArchivedVariantIds []string
	DeletedVariantIds  []string

	// ObsoleteVariantIds are the variants the configuration no longer allows: archived and
	// deleted together, kept for callers of the earlier shape.
	ObsoleteVariantIds []string

	UnchangedCount int
}

type GenerateVariantsResult = dyn.OpResult[GenerateVariantsResultData]

// ProductService is the Products capability and the resource service installed on the Product
// Template onion. It embeds the composable domain service so the onion keeps serving built-in
// CRUD unchanged; the template's application service reaches the extra methods through it.
type ProductService interface {
	composable.CrudDomainService

	GetEffectiveProduct(ctx corectx.Context, query GetEffectiveProductQuery) (*GetEffectiveProductResult, error)
	GetEffectiveProducts(ctx corectx.Context, query GetEffectiveProductsQuery) (*GetEffectiveProductsResult, error)
	ResolveProductSelection(ctx corectx.Context, query ResolveProductSelectionQuery) (*ResolveProductSelectionResult, error)
	GenerateVariants(ctx corectx.Context, query GenerateVariantsQuery) (*GenerateVariantsResult, error)
}
