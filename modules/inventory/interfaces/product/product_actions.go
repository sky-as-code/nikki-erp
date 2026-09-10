package product

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The custom actions of the product resources, as the bound request.
type (
	GenerateTemplateVariantsCommand = dmodel.DynamicFields
	ResolveSelectionQuery           = dmodel.DynamicFields
	GetEffectiveVariantQuery        = dmodel.DynamicFields
)

type (
	GenerateTemplateVariantsResult = dyn.OpResult[any]
	ResolveSelectionResult         = dyn.OpResult[any]
	GetEffectiveVariantResult      = dyn.OpResult[any]
)

// ProductTemplateActionService is the authorized custom surface of a template: bringing its
// variants in step with its attribute configuration, and turning a chosen attribute combination
// into the concrete variant a transaction line must reference.
type ProductTemplateActionService interface {
	GenerateVariants(ctx corectx.Context, cmd GenerateTemplateVariantsCommand) (*GenerateTemplateVariantsResult, error)
	ResolveSelection(ctx corectx.Context, query ResolveSelectionQuery) (*ResolveSelectionResult, error)
}

// ProductVariantActionService flattens a variant together with its template, so a consumer reads
// the effective product rather than re-deriving which fields come from where.
type ProductVariantActionService interface {
	GetEffective(ctx corectx.Context, query GetEffectiveVariantQuery) (*GetEffectiveVariantResult, error)
}
