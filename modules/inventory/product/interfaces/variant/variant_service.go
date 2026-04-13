package variant

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type VariantService interface {
	CreateVariant(ctx corectx.Context, cmd CreateVariantCommand) (*CreateVariantResult, error)
	DeleteVariant(ctx corectx.Context, cmd DeleteVariantCommand) (*DeleteVariantResult, error)
	VariantExists(ctx corectx.Context, query VariantExistsQuery) (*VariantExistsResult, error)
	GetVariant(ctx corectx.Context, query GetVariantQuery) (*GetVariantResult, error)
	SearchVariants(ctx corectx.Context, query SearchVariantsQuery) (*SearchVariantsResult, error)
	UpdateVariant(ctx corectx.Context, cmd UpdateVariantCommand) (*dyn.OpResult[dyn.MutateResultData], error)
}
