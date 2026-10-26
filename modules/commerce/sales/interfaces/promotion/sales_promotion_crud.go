// Package promotion declares the layers of the promotion and voucher resources: the programs a
// discount is defined by, the conditions and rewards that shape it, and the voucher codes and
// redemption ledger that track its use.
//
// All are plain CRUD -- their rules live in the pricing engine, which reads them rather than
// writing them. Each still gets its own layers so a rule added later has a home.
package promotion

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type SalesPromotionProgramRepository interface {
	composable.CrudRepository
}

type SalesPromotionProgramDomainService interface {
	composable.CrudDomainService
}

type SalesPromotionProgramApplicationService interface {
	composable.CrudApplicationService
}

type SalesPromotionConditionGroupRepository interface {
	composable.CrudRepository
}

type SalesPromotionConditionGroupDomainService interface {
	composable.CrudDomainService
}

type SalesPromotionConditionGroupApplicationService interface {
	composable.CrudApplicationService
}

type SalesPromotionConditionRepository interface {
	composable.CrudRepository
}

type SalesPromotionConditionDomainService interface {
	composable.CrudDomainService
}

type SalesPromotionConditionApplicationService interface {
	composable.CrudApplicationService
}

type SalesPromotionConditionTargetRepository interface {
	composable.CrudRepository
}

type SalesPromotionConditionTargetDomainService interface {
	composable.CrudDomainService
}

type SalesPromotionConditionTargetApplicationService interface {
	composable.CrudApplicationService
}

type SalesPromotionRewardRepository interface {
	composable.CrudRepository
}

type SalesPromotionRewardDomainService interface {
	composable.CrudDomainService
}

type SalesPromotionRewardApplicationService interface {
	composable.CrudApplicationService
}

type SalesPromotionCompatibilityRepository interface {
	composable.CrudRepository
}

type SalesPromotionCompatibilityDomainService interface {
	composable.CrudDomainService
}

type SalesPromotionCompatibilityApplicationService interface {
	composable.CrudApplicationService
}

type SalesVoucherCodeRepository interface {
	composable.CrudRepository
}

type SalesVoucherCodeDomainService interface {
	composable.CrudDomainService
}

type SalesVoucherCodeApplicationService interface {
	composable.CrudApplicationService
}

type SalesVoucherRedemptionRepository interface {
	composable.CrudRepository
}

type SalesVoucherRedemptionApplicationService interface {
	composable.CrudApplicationService
}

// The redemption ledger is read-only: a writable row could forge a discount's provenance or
// release a hold another order relies on. Enforced by CrudActions at the application layer, so it
// holds for an in-process caller as well as for a route.

type (
	CreateSalesPromotionProgramCommand = composable.CreateCommand
	UpdateSalesPromotionProgramCommand = composable.UpdateCommand
	SearchSalesPromotionProgramsQuery  = composable.SearchQuery
	SearchSalesPromotionProgramsResult = composable.SearchResult
	CreateSalesVoucherCodeCommand      = composable.CreateCommand
	SearchSalesVoucherCodesQuery       = composable.SearchQuery
	SearchSalesVoucherCodesResult      = composable.SearchResult
)
