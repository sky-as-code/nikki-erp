// Package promotion declares the layers of the promotion and voucher resources: the programs a
// discount is defined by, the conditions and rewards that shape it, and the voucher codes and
// redemption ledger that track its use.
//
// All are plain CRUD -- their rules live in the pricing engine, which reads them rather than
// writing them. Each still gets its own layers so a rule added later has a home.
package promotion

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type SalesPromotionProgramRepository interface {
	dynamicresource.CrudRepository
}

type SalesPromotionProgramDomainService interface {
	dynamicresource.CrudDomainService
}

type SalesPromotionProgramApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesPromotionConditionGroupRepository interface {
	dynamicresource.CrudRepository
}

type SalesPromotionConditionGroupDomainService interface {
	dynamicresource.CrudDomainService
}

type SalesPromotionConditionGroupApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesPromotionConditionRepository interface {
	dynamicresource.CrudRepository
}

type SalesPromotionConditionDomainService interface {
	dynamicresource.CrudDomainService
}

type SalesPromotionConditionApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesPromotionConditionTargetRepository interface {
	dynamicresource.CrudRepository
}

type SalesPromotionConditionTargetDomainService interface {
	dynamicresource.CrudDomainService
}

type SalesPromotionConditionTargetApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesPromotionRewardRepository interface {
	dynamicresource.CrudRepository
}

type SalesPromotionRewardDomainService interface {
	dynamicresource.CrudDomainService
}

type SalesPromotionRewardApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesPromotionCompatibilityRepository interface {
	dynamicresource.CrudRepository
}

type SalesPromotionCompatibilityDomainService interface {
	dynamicresource.CrudDomainService
}

type SalesPromotionCompatibilityApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesVoucherCodeRepository interface {
	dynamicresource.CrudRepository
}

type SalesVoucherCodeDomainService interface {
	dynamicresource.CrudDomainService
}

type SalesVoucherCodeApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesVoucherRedemptionRepository interface {
	dynamicresource.CrudRepository
}

type SalesVoucherRedemptionApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The redemption ledger is read-only: a writable row could forge a discount's provenance or
// release a hold another order relies on. Enforced by CrudActions at the application layer, so it
// holds for an in-process caller as well as for a route.

type (
	CreateSalesPromotionProgramCommand = dynamicresource.CreateCommand
	UpdateSalesPromotionProgramCommand = dynamicresource.UpdateCommand
	SearchSalesPromotionProgramsQuery  = dynamicresource.SearchQuery
	SearchSalesPromotionProgramsResult = dynamicresource.SearchResult
	CreateSalesVoucherCodeCommand      = dynamicresource.CreateCommand
	SearchSalesVoucherCodesQuery       = dynamicresource.SearchQuery
	SearchSalesVoucherCodesResult      = dynamicresource.SearchResult
)
