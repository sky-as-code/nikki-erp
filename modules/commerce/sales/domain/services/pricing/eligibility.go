package pricing

import (
	"strings"

	"github.com/shopspring/decimal"
)

// Promotion eligibility: whether a program's conditions hold for a basket. Conflict resolution is
// deliberately elsewhere (ResolvePromotions), because whether two programs may combine is about the
// programs, not the cart — mixing them would make compatibility depend on basket contents.

// ConditionGroup is a set of conditions ANDed together; groups are ORed with each other.
type ConditionGroup struct {
	Sequence   int32
	Conditions []Condition
}

// Condition is one test.
type Condition struct {
	// Type and Operator are the models.PromotionCondition* / PromotionOperator* values, as strings so
	// this package imports no schema.
	Type     string
	Operator string

	ValueText    string
	ValueDecimal decimal.Decimal
	ValueFrom    decimal.Decimal
	ValueTo      decimal.Decimal

	// TargetIds is the set an `in` or `not_in` operator reads; the caller passes the rows it loaded
	// from the set-valued condition table.
	TargetIds []string
}

// BasketFacts is what a condition can be tested against. Flat facts rather than an order, because
// the engine also prices baskets that are not yet orders.
type BasketFacts struct {
	Subtotal      decimal.Decimal
	TotalQuantity decimal.Decimal

	SalesChannelId string
	SalesPointId   string

	// VariantIds and CategoryIds are what the basket contains.
	VariantIds  []string
	CategoryIds []string

	// QuantityByVariant supports a per-variant quantity condition ("buy 3 of X").
	QuantityByVariant map[string]decimal.Decimal

	// DayOfWeek is lowercase English ("monday"); TimeOfDayMinutes is minutes since midnight. Both are
	// supplied by the caller, never read from a clock, so a historical sale replays reproducibly.
	DayOfWeek        string
	TimeOfDayMinutes int32

	// NowUnix is the evaluation instant, for valid_from / valid_until conditions.
	NowUnix int64
}

// IsEligible reports whether a program's conditions hold: every condition within a group, any group
// across groups. No groups means unconditionally eligible — "10% off everything" is a real campaign.
func IsEligible(groups []ConditionGroup, facts BasketFacts) bool {
	if len(groups) == 0 {
		return true
	}
	for _, group := range groups {
		if groupHolds(group, facts) {
			return true
		}
	}
	return false
}

// groupHolds ANDs a group's conditions. An empty group holds: it expresses no restriction.
func groupHolds(group ConditionGroup, facts BasketFacts) bool {
	for _, condition := range group.Conditions {
		if !conditionHolds(condition, facts) {
			return false
		}
	}
	return true
}

// conditionHolds evaluates one condition. An unrecognised type answers FALSE: a restriction this
// build cannot evaluate must not be treated as satisfied.
func conditionHolds(condition Condition, facts BasketFacts) bool {
	switch condition.Type {
	case "order_subtotal":
		return compareDecimal(condition, facts.Subtotal)

	case "total_quantity":
		return compareDecimal(condition, facts.TotalQuantity)

	case "quantity":
		// Per-variant quantity: holds when ANY named variant reaches the threshold. A program needing
		// several specific quantities expresses that as several conditions.
		if len(condition.TargetIds) == 0 {
			return compareDecimal(condition, facts.TotalQuantity)
		}
		for _, variantId := range condition.TargetIds {
			if quantity, present := facts.QuantityByVariant[variantId]; present {
				if compareDecimal(condition, quantity) {
					return true
				}
			}
		}
		return false

	case "product_variant":
		return compareSet(condition, facts.VariantIds)

	case "product_category":
		return compareSet(condition, facts.CategoryIds)

	case "sales_channel":
		return compareIdentity(condition, facts.SalesChannelId)

	case "sales_point":
		return compareIdentity(condition, facts.SalesPointId)

	case "day_of_week":
		return compareText(condition, strings.ToLower(facts.DayOfWeek))

	case "time_of_day":
		return compareDecimal(condition, decimal.NewFromInt(int64(facts.TimeOfDayMinutes)))

	case "valid_from":
		return decimal.NewFromInt(facts.NowUnix).GreaterThanOrEqual(condition.ValueDecimal)

	case "valid_until":
		// Exclusive, like every other window in this module: a campaign ending at noon does not apply
		// at noon.
		return decimal.NewFromInt(facts.NowUnix).LessThan(condition.ValueDecimal)
	}
	return false
}

func compareDecimal(condition Condition, actual decimal.Decimal) bool {
	switch condition.Operator {
	case "eq":
		return actual.Equal(condition.ValueDecimal)
	case "ne":
		return !actual.Equal(condition.ValueDecimal)
	case "gt":
		return actual.GreaterThan(condition.ValueDecimal)
	case "gte":
		return actual.GreaterThanOrEqual(condition.ValueDecimal)
	case "lt":
		return actual.LessThan(condition.ValueDecimal)
	case "lte":
		return actual.LessThanOrEqual(condition.ValueDecimal)
	case "between":
		// Inclusive of both bounds, unlike the validity windows above.
		return actual.GreaterThanOrEqual(condition.ValueFrom) &&
			actual.LessThanOrEqual(condition.ValueTo)
	}
	return false
}

func compareText(condition Condition, actual string) bool {
	switch condition.Operator {
	case "eq":
		return actual == strings.ToLower(condition.ValueText)
	case "ne":
		return actual != strings.ToLower(condition.ValueText)
	case "in":
		for _, target := range condition.TargetIds {
			if actual == strings.ToLower(target) {
				return true
			}
		}
		// A text `in` may also carry a comma-separated list, e.g. a hand-written day-of-week set.
		for _, target := range strings.Split(condition.ValueText, ",") {
			if actual == strings.ToLower(strings.TrimSpace(target)) {
				return true
			}
		}
		return false
	case "not_in":
		inverted := condition
		inverted.Operator = "in"
		return !compareText(inverted, actual)
	}
	return false
}

// compareIdentity tests a single id against the condition's target set.
func compareIdentity(condition Condition, actual string) bool {
	switch condition.Operator {
	case "eq":
		return actual == condition.ValueText
	case "ne":
		return actual != condition.ValueText
	case "in":
		return containsId(condition.TargetIds, actual)
	case "not_in":
		return !containsId(condition.TargetIds, actual)
	}
	return false
}

// compareSet tests the condition's targets against what the basket contains: `in` holds when the
// basket contains at least one target, `not_in` when it contains none.
func compareSet(condition Condition, basketIds []string) bool {
	switch condition.Operator {
	case "in", "eq":
		for _, target := range condition.TargetIds {
			if containsId(basketIds, target) {
				return true
			}
		}
		if condition.ValueText != "" {
			return containsId(basketIds, condition.ValueText)
		}
		return false
	case "not_in", "ne":
		for _, target := range condition.TargetIds {
			if containsId(basketIds, target) {
				return false
			}
		}
		if condition.ValueText != "" {
			return !containsId(basketIds, condition.ValueText)
		}
		return true
	}
	return false
}

func containsId(ids []string, wanted string) bool {
	for _, id := range ids {
		if id == wanted {
			return true
		}
	}
	return false
}

// FactsFromLines derives basket facts from the priced lines, so the subtotal a condition tests is
// the number the engine computed rather than one the caller re-derived.
func FactsFromLines(lines []LineResult) BasketFacts {
	facts := BasketFacts{
		Subtotal:          decimal.Zero,
		TotalQuantity:     decimal.Zero,
		QuantityByVariant: make(map[string]decimal.Decimal, len(lines)),
	}
	for _, line := range lines {
		// A giveaway line contributes nothing: counting a free item toward a spend threshold would let
		// one promotion qualify a basket for another it did not earn.
		if line.LineType == "promotion_reward" {
			continue
		}
		facts.Subtotal = facts.Subtotal.Add(line.GrossAmount)
		facts.TotalQuantity = facts.TotalQuantity.Add(line.Quantity)
		facts.VariantIds = append(facts.VariantIds, line.ProductVariantId)

		existing := facts.QuantityByVariant[line.ProductVariantId]
		facts.QuantityByVariant[line.ProductVariantId] = existing.Add(line.Quantity)
	}
	return facts
}
