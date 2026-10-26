package catalog

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// PricingRuleSetExtService hands another module every rule that can change what a selling place
// charges, as rows.
//
// Rows rather than resolved prices: a caller that prices a basket it assembles itself — a vending
// machine showing a catalogue, a POS building an order offline — cannot be served by a per-product
// number, because quantity breaks, bundles and promotions are properties of the basket. It receives
// the rules and runs the same resolution this module runs, so what the caller shows and what the
// server settles at come from one input.
//
// It asserts no permission: a cross-module port never does, and the caller's own transport is what
// authenticates the machine asking. What bounds the answer is the sales point, which the caller
// derives from its own authenticated record rather than from a request body.
type PricingRuleSetExtService interface {
	// PricingRuleSetForPoint answers every rule that could apply at one sales point.
	PricingRuleSetForPoint(
		ctx corectx.Context, query PricingRuleSetQuery,
	) (*PricingRuleSetResult, error)
}

// PricingRuleSetQuery names the selling place to answer for. The channel is NOT a parameter: it is
// read from the point, so a caller cannot ask for one place's prices under another's channel.
type PricingRuleSetQuery struct {
	OrgId        string
	SalesPointId string
}

// PricingRuleSetResult carries the rule set, or the refusal that stopped it being built.
type PricingRuleSetResult struct {
	Data    PricingRuleSet
	HasData bool
}

// PricingRuleSet is the rule tables, flat, one slice per table.
//
// Flat rather than nested, because a caller storing them locally writes each list into a table of
// its own: a nested payload would have to be taken apart before it could be saved and put back
// together before it could be read.
//
// What is deliberately filtered and what is not:
//
//   - SCOPE is filtered. A pricelist scoped to another sales point or another channel can never
//     apply here, and shipping it would invite the caller to rank rules it must never choose.
//   - VALIDITY WINDOWS are NOT filtered. They travel on the rows. A caller pricing offline keeps
//     the set for hours while its clock moves, so a window evaluated once here would be a decision
//     that silently expires; the caller re-evaluates it on every basket instead.
//   - ARCHIVED rows are dropped. Archiving is withdrawal, and a withdrawn rule applies to nothing.
//   - PROMOTION SCOPE is not filtered either: a promotion says where it applies through its own
//     conditions (`sales_channel`, `sales_point`), which the caller evaluates along with the rest
//     of the basket. Pre-filtering on them here would evaluate half a condition tree and leave the
//     other half to somebody else.
//
// It carries no tiers and no totals — both are derived — and no tax, which Accounting owns.
type PricingRuleSet struct {
	// Pricelists carry the scope and window that rank them; PricelistItems carry the rules within.
	// A rule row alone cannot be ranked, which is why the lists travel with their items.
	Pricelists     []models.SalesPricelist
	PricelistItems []models.SalesPricelistItem

	Combos          []models.SalesCombo
	ComboComponents []models.SalesComboComponent

	Programs                []models.SalesPromotionProgram
	ProgramConditionGroups  []models.SalesPromotionConditionGroup
	ProgramConditions       []models.SalesPromotionCondition
	ProgramConditionTargets []models.SalesPromotionConditionTarget
	ProgramRewards          []models.SalesPromotionReward
	ProgramCompatibilities  []models.SalesPromotionCompatibility

	// VoucherCodes name which program a code unlocks, nothing more. Whether a code is still
	// spendable is live state the server answers at settlement: a caller deciding it from the
	// stored usage_count would honour a code somebody else had already spent.
	VoucherCodes []models.SalesVoucherCode
}

// NewPricingRuleSet returns the empty set with every slice allocated, so a caller marshalling it
// sends empty lists rather than nulls and never has to tell "no rule of this kind" from "this
// field was not sent".
func NewPricingRuleSet() PricingRuleSet {
	return PricingRuleSet{
		Pricelists:              []models.SalesPricelist{},
		PricelistItems:          []models.SalesPricelistItem{},
		Combos:                  []models.SalesCombo{},
		ComboComponents:         []models.SalesComboComponent{},
		Programs:                []models.SalesPromotionProgram{},
		ProgramConditionGroups:  []models.SalesPromotionConditionGroup{},
		ProgramConditions:       []models.SalesPromotionCondition{},
		ProgramConditionTargets: []models.SalesPromotionConditionTarget{},
		ProgramRewards:          []models.SalesPromotionReward{},
		ProgramCompatibilities:  []models.SalesPromotionCompatibility{},
		VoucherCodes:            []models.SalesVoucherCode{},
	}
}
