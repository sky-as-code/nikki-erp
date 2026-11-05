// Package promotion declares the layers of the promotion and voucher resources (BR §8): the
// promotion with its scope, rules, rewards and selectors; the vouchers that unlock it; and the
// read-only records of what a confirmed order received.
package promotion

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SalesPromotionRepository reads and writes promotions.
type SalesPromotionRepository interface {
	dynamicresource.CrudRepository
}

// SalesPromotionDomainService owns the lifecycle (BR-L01..L04): the edit lock while active, the
// config_version bump, the immutable fields once used, activation and the delete guard.
type SalesPromotionDomainService interface {
	dynamicresource.CrudDomainService

	Activate(ctx corectx.Context, promotionId string) (*RecordResult, error)
	Suspend(ctx corectx.Context, promotionId string) (*RecordResult, error)
	Duplicate(ctx corectx.Context, promotionId string) (*RecordResult, error)
	ManageScope(ctx corectx.Context, scope PromotionScope, cmd ManageCommand) (*ManageResult, error)
	// BumpConfigVersion records a configuration change of a child row inside ctx's transaction, and
	// refuses it while the promotion is active (BR-L02).
	BumpConfigVersion(ctx corectx.Context, promotionId string) (*ft.ClientErrors, error)
}

// SalesPromotionApplicationService is the authorized promotion surface.
type SalesPromotionApplicationService interface {
	dynamicresource.CrudApplicationService

	Activate(ctx corectx.Context, cmd LifecycleCommand) (*RecordResult, error)
	Suspend(ctx corectx.Context, cmd LifecycleCommand) (*RecordResult, error)
	Duplicate(ctx corectx.Context, cmd LifecycleCommand) (*RecordResult, error)
	ManageChannels(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
	ManagePoints(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
	ManagePricelists(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
}

// PromotionScope names one of the three scope junctions of a promotion.
type PromotionScope string

const (
	PromotionScopeChannels   = PromotionScope("channels")
	PromotionScopePoints     = PromotionScope("points")
	PromotionScopePricelists = PromotionScope("pricelists")
)

// SalesPromotionChildRepository reads and writes a rule, a reward or a selector.
type SalesPromotionChildRepository interface {
	dynamicresource.CrudRepository
}

// SalesPromotionRuleDomainService validates rules against their promotion and manages their
// selectors.
type SalesPromotionRuleDomainService interface {
	dynamicresource.CrudDomainService

	ManageProducts(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
	ManageCategories(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
}

// SalesPromotionRuleApplicationService is the authorized rule surface. Every write also asserts
// update on the parent promotion.
type SalesPromotionRuleApplicationService interface {
	dynamicresource.CrudApplicationService

	ManageProducts(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
	ManageCategories(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
}

// SalesPromotionRewardDomainService validates rewards against their promotion and manages their
// selectors.
type SalesPromotionRewardDomainService interface {
	dynamicresource.CrudDomainService

	ManageProducts(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
	ManageCategories(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
}

// SalesPromotionRewardApplicationService is the authorized reward surface.
type SalesPromotionRewardApplicationService interface {
	dynamicresource.CrudApplicationService

	ManageProducts(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
	ManageCategories(ctx corectx.Context, cmd ManageCommand) (*ManageResult, error)
}

// SalesPromotionReadOnlyApplicationService serves a resource clients may only read: a selector
// (written through its owner's manage_* action) or a confirmed-order record (written by Confirm).
type SalesPromotionReadOnlyApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The record a lifecycle action answers with, and the manage_* contract.
type (
	LifecycleCommand = dynamicresource.UpdateCommand
	RecordResult     = dyn.OpResult[dmodel.DynamicFields]
	ManageCommand    = dmodel.DynamicFields
	ManageResult     = dynamicresource.MutateResult
)
