package models

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The variant's effective base sales price (addendum §5.7, BR-PRICE-VARIANT-001..004).
//
// The field is computed in SQL, so a unit test cannot evaluate it — that needs a database, and the
// migration replay covers it. What CAN be pinned here, and matters just as much, is the SHAPE of
// the definition: which operands it sums, that it is never stored, and that its arithmetic is the
// one the worked example requires. Those are the properties a well-meaning edit would break
// silently, because a computed field that resolves is not thereby a computed field that is right.

// variantField returns one field definition from the variant schema JSON.
func variantField(t *testing.T, name string) map[string]any {
	t.Helper()

	var schema struct {
		Fields []map[string]any `json:"fields"`
	}
	require.NoError(t, json.Unmarshal([]byte(productVariantSchemaJson), &schema))

	for _, field := range schema.Fields {
		if field["name"] == name {
			return field
		}
	}
	t.Fatalf("the variant schema declares no field %q", name)
	return nil
}

func computedOf(t *testing.T, field map[string]any) map[string]any {
	t.Helper()

	computed, ok := field["computed"].(map[string]any)
	require.True(t, ok, "field %v is not computed", field["name"])
	return computed
}

// BR-PRICE-VARIANT-001: the effective price is the template's base plus the variant's extras.
//
// Asserted as structure because that is what the requirement actually constrains: any other pair of
// operands would resolve perfectly well and price every variant wrongly.
func TestEffectiveBaseSalesPriceSumsTheTemplatePriceAndTheExtras(t *testing.T) {
	computed := computedOf(t, variantField(t, "effective_base_sales_price"))

	assert.Equal(t, "expression", computed["kind"])

	expression, ok := computed["expression"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "add", expression["op"], "the two operands are summed, never multiplied")

	args, ok := expression["args"].([]any)
	require.True(t, ok)
	require.Len(t, args, 2, "exactly two operands: the template price and the extras")

	assert.Equal(t, []string{"template_base_sales_price", "sales_price_extra_total"},
		coalescedFieldNames(t, args))
}

// coalescedFieldNames pulls the field name out of each `coalesce(field, 0)` operand.
func coalescedFieldNames(t *testing.T, args []any) []string {
	t.Helper()

	names := make([]string, 0, len(args))
	for _, arg := range args {
		call, ok := arg.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "coalesce", call["function"],
			"each operand is coalesced, or one unpriced side makes the whole sum NULL")

		inner, ok := call["args"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, inner)

		operand, ok := inner[0].(map[string]any)
		require.True(t, ok)
		names = append(names, operand["field"].(string))
	}
	return names
}

// BR-PRICE-VARIANT-004: raising the template's price moves every variant at once.
//
// True by construction only because the field is NOT stored. A stored copy would be correct until
// somebody edited the template, and then wrong with nothing to indicate it — which is precisely the
// failure the requirement names. All three fields in the chain must be virtual, not just the last.
func TestTheWholePricingChainIsComputedOnRead(t *testing.T) {
	for _, name := range []string{
		"template_base_sales_price",
		"sales_price_extra_total",
		"effective_base_sales_price",
	} {
		t.Run(name, func(t *testing.T) {
			computed := computedOf(t, variantField(t, name))

			stored, present := computed["is_stored"]
			require.True(t, present, "is_stored must be stated rather than defaulted")
			assert.Equal(t, false, stored,
				"a stored copy is right until the template changes, then silently wrong")
		})
	}
}

// The sum is over the variant's OWN attribute values, through a direct collection edge.
//
// That edge is load-bearing: an aggregate cannot reach a field two hops away, which is why
// sales_price_extra is denormalised onto the junction row at all (D-06). A source pointing anywhere
// else would fail to compile at boot rather than quietly — but it would fail late, and this says so
// early.
func TestTheExtrasAreSummedOverTheVariantsOwnValues(t *testing.T) {
	computed := computedOf(t, variantField(t, "sales_price_extra_total"))

	assert.Equal(t, "aggregate", computed["kind"])
	assert.Equal(t, "sum", computed["function"])
	assert.Equal(t, "attribute_values", computed["source"])
	assert.Equal(t, "sales_price_extra", computed["field"])
	assert.Equal(t, 0.0, computed["default"],
		"SUM over zero rows is NULL; a variant with no attribute values must read 0, not unknown")
}

// The addendum's worked example, as arithmetic: 100,000 + 10,000 + 20,000 = 130,000, and with a
// 10% pricelist discount, 117,000.
//
// It does not exercise the SQL — it cannot. What it pins is that the numbers in the requirement are
// self-consistent and that the sum is a plain addition of the extras onto the base, so a reader
// comparing this test to §5.7 can see the same figures.
func TestTheAddendumWorkedExample(t *testing.T) {
	templateBase := decimal.RequireFromString("100000")
	sizeExtra := decimal.RequireFromString("10000")
	colourExtra := decimal.RequireFromString("20000")

	effective := templateBase.Add(sizeExtra).Add(colourExtra)
	assert.True(t, effective.Equal(decimal.RequireFromString("130000")),
		"100,000 + 10,000 + 20,000 = %s, want 130,000", effective)

	// BR-PRICE-VARIANT-003: a rule discounting BASE_SALES_PRICE discounts THIS number, not the
	// template's raw 100,000 — which would have given 90,000 and undercharged for the options.
	discounted := effective.Mul(decimal.RequireFromString("0.9"))
	assert.True(t, discounted.Equal(decimal.RequireFromString("117000")),
		"130,000 less 10%% = %s, want 117,000", discounted)
}

// A negative extra is legitimate: a plain colour may subtract. The field's own bounds have to allow
// it, or the rule the addendum states could not be configured.
func TestASalesPriceExtraMayBeNegative(t *testing.T) {
	var schema struct {
		Fields []map[string]any `json:"fields"`
	}
	require.NoError(t, json.Unmarshal([]byte(productTemplateAttributeValueSchemaJson), &schema))

	for _, field := range schema.Fields {
		if field["name"] != "sales_price_extra" {
			continue
		}
		dataType, ok := field["data_type"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "-1000000000000", dataType["min"],
			"the minimum must be negative: XL adds, a plain colour subtracts")
		return
	}
	t.Fatal("the template attribute value schema declares no sales_price_extra")
}
