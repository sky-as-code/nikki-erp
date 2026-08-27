package pricing

import (
	"strings"

	"github.com/shopspring/decimal"
)

// Promotion eligibility: deciding which programs apply to a basket (BR §23, D-06, D-07).
//
// It lives here rather than in the parent package for the same reason the engine does — it is pure,
// and the caller needs it and the engine together. The split of responsibility is:
//
//	this file          — does this program's conditions hold for this basket?
//	ResolvePromotions  — of the programs that hold, which may apply together? (BR §29)
//	Calculate          — apply the survivors, in order
//
// Keeping eligibility separate from conflict resolution is deliberate: whether a program's
// conditions hold is about the cart, while whether two programs may combine is about the programs.
// Mixing them would make a compatibility rule depend on basket contents, which is not what BR §27
// means.

// ConditionGroup is a set of conditions ANDed together. Groups are ORed with each other (D-06) —
// the standard sum-of-products shape, which expresses every BR §23 example and needs no parser.
type ConditionGroup struct {
	Sequence   int32
	Conditions []Condition
}

// Condition is one test.
type Condition struct {
	// Type and Operator are the models.PromotionCondition* / PromotionOperator* values, as strings
	// so this package imports no schema.
	Type     string
	Operator string

	ValueText    string
	ValueDecimal decimal.Decimal
	ValueFrom    decimal.Decimal
	ValueTo      decimal.Decimal

	// TargetIds is the set an `in` or `not_in` operator reads. Set-valued conditions live in their
	// own table rather than a column (D-07), so the caller passes the rows it loaded.
	TargetIds []string
}

// BasketFacts is what a condition can be tested against.
//
// A flat struct rather than the order itself: the engine prices baskets that are not yet orders —
// a quote preview has no order id — and passing facts rather than a record keeps eligibility
// answerable for both.
type BasketFacts struct {
	Subtotal      decimal.Decimal
	TotalQuantity decimal.Decimal

	SalesChannelId string
	SalesPointId   string

	// VariantIds and CategoryIds are what the basket contains. Slices rather than sets because they
	// are small and the caller has them in this shape already.
	VariantIds  []string
	CategoryIds []string

	// QuantityByVariant supports a per-variant quantity condition ("buy 3 of X").
	QuantityByVariant map[string]decimal.Decimal

	// DayOfWeek is lowercase English ("monday"), and TimeOfDayMinutes is minutes since midnight.
	// Both are supplied by the caller rather than read from a clock, because the engine takes no
	// clock — that is what makes a replay of a historical sale reproducible.
	DayOfWeek        string
	TimeOfDayMinutes int32

	// NowUnix is the evaluation instant, for valid_from / valid_until conditions.
	NowUnix int64
}

// IsEligible reports whether a program's conditions hold.
//
// No groups means unconditionally eligible: a program with no conditions is one the operator wants
// applied to everything, which is a legitimate campaign and not a misconfiguration. Refusing it
// would make "10% off everything" inexpressible.
//
// Within a group every condition must hold; across groups any group holding is enough (D-06).
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

// groupHolds ANDs a group's conditions.
//
// An EMPTY group holds. It contributes nothing to the OR across groups except "yes", which is the
// same answer as having no groups at all — and an operator who created a group and added no
// conditions to it has expressed no restriction.
func groupHolds(group ConditionGroup, facts BasketFacts) bool {
	for _, condition := range group.Conditions {
		if !conditionHolds(condition, facts) {
			return false
		}
	}
	return true
}

// conditionHolds evaluates one condition.
//
// An unrecognised condition type answers FALSE. The value came from a database row, and a type this
// build does not understand must not be treated as satisfied: silently applying a discount whose
// restriction nobody could evaluate is the expensive direction of that mistake.
func conditionHolds(condition Condition, facts BasketFacts) bool {
	switch condition.Type {
	case "order_subtotal":
		return compareDecimal(condition, facts.Subtotal)

	case "total_quantity":
		return compareDecimal(condition, facts.TotalQuantity)

	case "quantity":
		// Per-variant quantity: holds when ANY named variant reaches the threshold. Any rather than
		// all, because "buy 3 of anything in this range" is what BR §23's examples describe; a
		// program needing several specific quantities expresses that as several conditions.
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
		// Exclusive, matching every other window in this module: a campaign ending at noon does not
		// apply at noon.
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
		// Inclusive of both bounds, unlike the validity windows: an operator writing "between 5 and
		// 10" means a customer buying exactly 10 qualifies.
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
		// A text `in` may also carry a comma-separated list, which is how a day-of-week set is
		// naturally written by hand.
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

// compareSet tests the condition's targets against what the basket contains.
//
// `in` holds when the basket contains AT LEAST ONE of the targets — "buy any of these" — while
// `not_in` requires the basket to contain NONE of them. They are not each other's inverse over a
// set: the negation of "contains at least one" is "contains none", which is what not_in means, so
// the asymmetry is only apparent.
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

// FactsFromLines derives the basket facts the conditions test, from the priced lines.
//
// A helper rather than a requirement: a caller may build BasketFacts itself. It exists so that the
// subtotal a condition tests is the same number the engine computed, rather than one the caller
// re-derived and could get subtly different.
func FactsFromLines(lines []LineResult) BasketFacts {
	facts := BasketFacts{
		Subtotal:          decimal.Zero,
		TotalQuantity:     decimal.Zero,
		QuantityByVariant: make(map[string]decimal.Decimal, len(lines)),
	}
	for _, line := range lines {
		// A giveaway line contributes nothing to the facts: counting a free item toward a spend
		// threshold would let one promotion qualify a basket for another it did not earn.
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
