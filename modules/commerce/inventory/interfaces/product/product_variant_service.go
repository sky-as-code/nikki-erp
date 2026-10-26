package product

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// SearchProductVariantsParams is the org-scoped counterpart of SearchProductVariantsQuery, for a
// caller that reaches this port with no user permission of its own (a device certificate proves
// which machine it is, and nothing about who is operating it) and so must supply the org itself
// rather than have it asserted from the caller's membership.
type SearchProductVariantsParams struct {
	OrgId model.Id

	dyn.SearchQuery
}

// ProductVariantService is the same read as ProductVariantDomainService, published for a caller
// that filters by org itself.
type ProductVariantService interface {
	SearchProductVariants(
		ctx corectx.Context, params SearchProductVariantsParams,
	) (*SearchProductVariantsResult, error)
}
