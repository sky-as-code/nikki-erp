package variant

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type VariantService interface {
	CreateVariant(ctx crud.Context, cmd CreateVariantCommand) (*CreateVariantResult, error)
	UpdateVariant(ctx crud.Context, cmd UpdateVariantCommand) (*UpdateVariantResult, error)
	DeleteVariant(ctx crud.Context, cmd DeleteVariantCommand) (*DeleteVariantResult, error)
	GetVariantById(ctx crud.Context, query GetVariantByIdQuery) (*GetVariantByIdResult, error)
	SearchVariants(ctx crud.Context, query SearchVariantsQuery) (*SearchVariantsResult, error)
}
