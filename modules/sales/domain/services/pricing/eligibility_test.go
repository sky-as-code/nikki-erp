package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

func facts() BasketFacts {
	return BasketFacts{
		Subtotal:       dec("100000"),
		TotalQuantity:  dec("5"),
		SalesChannelId: "channel-vdmc",
		SalesPointId:   "point-1",
		VariantIds:     []string{"v-cola", "v-crisps"},
		CategoryIds:    []string{"cat-drinks", "cat-snacks"},
		QuantityByVariant: map[string]decimal.Decimal{
			"v-cola": dec("3"), "v-crisps": dec("2"),
		},
		DayOfWeek:        "monday",
		TimeOfDayMinutes: 9 * 60,
		NowUnix:          1_800_000_000,
	}
}

func group(conditions ...Condition) []ConditionGroup {
	return []ConditionGroup{{Sequence: 1, Conditions: conditions}}
}

// A program with no conditions applies to everything.
func TestNoConditionsMeansEligible(t *testing.T) {
	if !IsEligible(nil, facts()) {
		t.Error("a program with no conditions must apply unconditionally")
	}
	if !IsEligible([]ConditionGroup{}, facts()) {
		t.Error("an empty group list must apply unconditionally")
	}
}

// An empty group holds: it expresses no restriction.
func TestEmptyGroupHolds(t *testing.T) {
	if !IsEligible([]ConditionGroup{{Sequence: 1}}, facts()) {
		t.Error("a group with no conditions expresses no restriction and must hold")
	}
}

// Conditions AND within a group.
func TestConditionsAndWithinAGroup(t *testing.T) {
	both := group(
		Condition{Type: "order_subtotal", Operator: "gte", ValueDecimal: dec("50000")},
		Condition{Type: "total_quantity", Operator: "gte", ValueDecimal: dec("3")},
	)
	if !IsEligible(both, facts()) {
		t.Error("both conditions hold, so the group must hold")
	}

	oneFails := group(
		Condition{Type: "order_subtotal", Operator: "gte", ValueDecimal: dec("50000")},
		Condition{Type: "total_quantity", Operator: "gte", ValueDecimal: dec("99")},
	)
	if IsEligible(oneFails, facts()) {
		t.Error("one condition fails, so the AND must fail")
	}
}

// Groups OR with each other.
func TestGroupsOrWithEachOther(t *testing.T) {
	groups := []ConditionGroup{
		{Sequence: 1, Conditions: []Condition{
			{Type: "order_subtotal", Operator: "gte", ValueDecimal: dec("999999")},
		}},
		{Sequence: 2, Conditions: []Condition{
			{Type: "total_quantity", Operator: "gte", ValueDecimal: dec("3")},
		}},
	}
	if !IsEligible(groups, facts()) {
		t.Error("the second group holds, so the OR must hold")
	}

	neither := []ConditionGroup{
		{Sequence: 1, Conditions: []Condition{
			{Type: "order_subtotal", Operator: "gte", ValueDecimal: dec("999999")},
		}},
		{Sequence: 2, Conditions: []Condition{
			{Type: "total_quantity", Operator: "gte", ValueDecimal: dec("99")},
		}},
	}
	if IsEligible(neither, facts()) {
		t.Error("no group holds, so the OR must fail")
	}
}

func TestDecimalOperators(t *testing.T) {
	cases := []struct {
		operator string
		value    string
		from, to string
		want     bool
	}{
		{"eq", "100000", "", "", true},
		{"eq", "99999", "", "", false},
		{"ne", "99999", "", "", true},
		{"gt", "99999", "", "", true},
		{"gt", "100000", "", "", false},
		{"gte", "100000", "", "", true},
		{"lt", "100001", "", "", true},
		{"lte", "100000", "", "", true},
		// between is inclusive of both bounds, unlike the validity windows.
		{"between", "", "100000", "200000", true},
		{"between", "", "50000", "100000", true},
		{"between", "", "100001", "200000", false},
	}
	for _, testCase := range cases {
		condition := Condition{Type: "order_subtotal", Operator: testCase.operator}
		if testCase.value != "" {
			condition.ValueDecimal = dec(testCase.value)
		}
		if testCase.from != "" {
			condition.ValueFrom = dec(testCase.from)
			condition.ValueTo = dec(testCase.to)
		}
		if got := IsEligible(group(condition), facts()); got != testCase.want {
			t.Errorf("subtotal %s %s%s = %v, want %v",
				testCase.operator, testCase.value, testCase.from, got, testCase.want)
		}
	}
}

// `in` over a product set holds when the basket contains at least one target.
func TestProductVariantInHoldsOnAnyMatch(t *testing.T) {
	condition := Condition{
		Type: "product_variant", Operator: "in",
		TargetIds: []string{"v-cola", "v-water"},
	}
	if !IsEligible(group(condition), facts()) {
		t.Error("the basket contains v-cola, so an `in` over that set must hold")
	}

	condition.TargetIds = []string{"v-water", "v-juice"}
	if IsEligible(group(condition), facts()) {
		t.Error("the basket contains neither target, so `in` must fail")
	}
}

// `not_in` requires the basket to contain NONE of the targets — not the plain inverse of `in` over a
// single element.
func TestProductVariantNotInRequiresNoMatch(t *testing.T) {
	excludesAPresentItem := Condition{
		Type: "product_variant", Operator: "not_in",
		TargetIds: []string{"v-cola"},
	}
	if IsEligible(group(excludesAPresentItem), facts()) {
		t.Error("the basket contains an excluded item, so not_in must fail")
	}

	excludesAbsentItems := Condition{
		Type: "product_variant", Operator: "not_in",
		TargetIds: []string{"v-water", "v-juice"},
	}
	if !IsEligible(group(excludesAbsentItems), facts()) {
		t.Error("the basket contains none of the excluded items, so not_in must hold")
	}
}

// A per-variant quantity condition holds when any named variant reaches the threshold.
func TestPerVariantQuantity(t *testing.T) {
	buyThreeColas := Condition{
		Type: "quantity", Operator: "gte", ValueDecimal: dec("3"),
		TargetIds: []string{"v-cola"},
	}
	if !IsEligible(group(buyThreeColas), facts()) {
		t.Error("the basket has 3 colas, so the condition must hold")
	}

	buyFourColas := Condition{
		Type: "quantity", Operator: "gte", ValueDecimal: dec("4"),
		TargetIds: []string{"v-cola"},
	}
	if IsEligible(group(buyFourColas), facts()) {
		t.Error("the basket has only 3 colas, so a threshold of 4 must fail")
	}

	// A variant not in the basket at all cannot satisfy it.
	buyThreeWaters := Condition{
		Type: "quantity", Operator: "gte", ValueDecimal: dec("3"),
		TargetIds: []string{"v-water"},
	}
	if IsEligible(group(buyThreeWaters), facts()) {
		t.Error("a variant absent from the basket must not satisfy a quantity condition")
	}
}

func TestChannelAndPointConditions(t *testing.T) {
	channelMatches := Condition{
		Type: "sales_channel", Operator: "in", TargetIds: []string{"channel-vdmc", "channel-bo"},
	}
	if !IsEligible(group(channelMatches), facts()) {
		t.Error("the basket is on channel-vdmc, so the condition must hold")
	}

	pointExcluded := Condition{
		Type: "sales_point", Operator: "not_in", TargetIds: []string{"point-1"},
	}
	if IsEligible(group(pointExcluded), facts()) {
		t.Error("the basket is at an excluded point, so not_in must fail")
	}
}

// A day-of-week set may be written as a comma-separated list, so both forms work.
func TestDayOfWeek(t *testing.T) {
	viaTargets := Condition{
		Type: "day_of_week", Operator: "in", TargetIds: []string{"monday", "tuesday"},
	}
	if !IsEligible(group(viaTargets), facts()) {
		t.Error("monday is in the target set")
	}

	viaList := Condition{
		Type: "day_of_week", Operator: "in", ValueText: "Saturday, Sunday",
	}
	if IsEligible(group(viaList), facts()) {
		t.Error("monday is not a weekend day")
	}

	viaListMatching := Condition{
		Type: "day_of_week", Operator: "in", ValueText: "monday,friday",
	}
	if !IsEligible(group(viaListMatching), facts()) {
		t.Error("monday is in the comma-separated list")
	}
}

// Happy hour, as minutes since midnight.
func TestTimeOfDay(t *testing.T) {
	morning := Condition{
		Type: "time_of_day", Operator: "between",
		ValueFrom: dec("480"), ValueTo: dec("600"), // 08:00–10:00
	}
	if !IsEligible(group(morning), facts()) {
		t.Error("09:00 falls inside 08:00-10:00")
	}

	evening := Condition{
		Type: "time_of_day", Operator: "between",
		ValueFrom: dec("1020"), ValueTo: dec("1200"), // 17:00–20:00
	}
	if IsEligible(group(evening), facts()) {
		t.Error("09:00 does not fall inside 17:00-20:00")
	}
}

// valid_until is EXCLUSIVE: a campaign ending at an instant does not apply at that instant.
func TestValidityWindowIsExclusiveAtTheEnd(t *testing.T) {
	basketFacts := facts()

	endsNow := Condition{
		Type: "valid_until", Operator: "lt",
		ValueDecimal: decimal.NewFromInt(basketFacts.NowUnix),
	}
	if IsEligible(group(endsNow), basketFacts) {
		t.Error("a campaign ending exactly now must not apply now")
	}

	endsLater := Condition{
		Type: "valid_until", Operator: "lt",
		ValueDecimal: decimal.NewFromInt(basketFacts.NowUnix + 1),
	}
	if !IsEligible(group(endsLater), basketFacts) {
		t.Error("a campaign ending a second from now still applies")
	}

	startedAlready := Condition{
		Type: "valid_from", Operator: "gte",
		ValueDecimal: decimal.NewFromInt(basketFacts.NowUnix - 1),
	}
	if !IsEligible(group(startedAlready), basketFacts) {
		t.Error("a campaign that started a second ago applies")
	}
}

// A condition type this build does not recognise answers FALSE, rather than silently applying a
// discount whose restriction nobody could evaluate.
func TestUnknownConditionTypeIsNotEligible(t *testing.T) {
	unknown := Condition{Type: "customer_loyalty_tier", Operator: "eq", ValueText: "gold"}
	if IsEligible(group(unknown), facts()) {
		t.Error("an unrecognised condition type must not be treated as satisfied")
	}
}

func TestUnknownOperatorIsNotEligible(t *testing.T) {
	unknown := Condition{Type: "order_subtotal", Operator: "approximately", ValueDecimal: dec("1")}
	if IsEligible(group(unknown), facts()) {
		t.Error("an unrecognised operator must not be treated as satisfied")
	}
}

// FactsFromLines must agree with what the engine computed and exclude giveaway lines, or one
// promotion would qualify a basket for another it did not earn.
func TestFactsFromLinesExcludesGiveaways(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "2", "50000")},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Free gift",
			Rewards: []RewardInput{{
				RewardId: "free-1", Sequence: 1, Type: "free_quantity", Value: dec("1"),
				FreeProductVariantId: "v-gift", FreeUomId: "uom-each",
			}},
		}},
		Context: vndContext(),
	})

	derived := FactsFromLines(result.Lines)
	if !derived.Subtotal.Equal(dec("100000")) {
		t.Errorf("subtotal = %s, want 100000: the giveaway must not count", derived.Subtotal)
	}
	if !derived.TotalQuantity.Equal(dec("2")) {
		t.Errorf("total quantity = %s, want 2: the giveaway must not count", derived.TotalQuantity)
	}
	for _, variantId := range derived.VariantIds {
		if variantId == "v-gift" {
			t.Error("the giveaway variant must not appear in the basket facts")
		}
	}
}

// The same variant on two lines contributes its combined quantity.
func TestFactsAggregateQuantityPerVariant(t *testing.T) {
	lines := []LineResult{
		{Key: "a", ProductVariantId: "v-cola", Quantity: dec("2"), GrossAmount: dec("20000")},
		{Key: "b", ProductVariantId: "v-cola", Quantity: dec("3"), GrossAmount: dec("30000")},
	}
	derived := FactsFromLines(lines)

	if got := derived.QuantityByVariant["v-cola"]; !got.Equal(dec("5")) {
		t.Errorf("v-cola quantity = %s, want 5", got)
	}
	if !derived.Subtotal.Equal(dec("50000")) {
		t.Errorf("subtotal = %s, want 50000", derived.Subtotal)
	}
}
