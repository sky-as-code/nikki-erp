package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The price schema is parsed from JSON at start-up, so a malformed file panics the whole app
// rather than failing here. These tests turn that into a test failure and pin the field set the
// pricing rules depend on.

func TestProductPriceSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := ProductPriceSchemaBuilder().Build()

	assert.Equal(t, ProductPriceSchemaName, schema.Name())
	assert.Equal(t, "inventory_product_prices", schema.TableName())
	assert.Equal(t, ProductPriceFieldPrice, schema.RecordLabelField())

	// BR §6.12: the amount is the one thing a price rule cannot be without.
	assert.True(t, requireField(t, schema, ProductPriceFieldPrice).IsRequiredForCreate())

	// Both targets are individually optional: which one is set is the caller's choice, and the
	// exclusive group below is what constrains it. Declaring either required would forbid the
	// other kind of rule outright.
	assert.False(t, requireField(t, schema, ProductPriceFieldProductTemplateId).IsRequiredForCreate())
	assert.False(t, requireField(t, schema, ProductPriceFieldProductVariantId).IsRequiredForCreate())

	// BR §6.12: a standing price has no start or end date, so neither bound may be mandatory.
	assert.False(t, requireField(t, schema, ProductPriceFieldEffectiveFrom).IsRequiredForCreate())
	assert.False(t, requireField(t, schema, ProductPriceFieldEffectiveTo).IsRequiredForCreate())
}

// BR §6.12 rule 1: "Không đồng thời để product_template_id và product_variant_id vô nghĩa/
// ambiguous" — a price must attach to exactly one target. Neither leaves the rule unreachable;
// both make precedence ambiguous.
//
// The schema's exclusive_required_fields group enforces this, so there is no hand-written
// validator to test. This asserts the framework rule is actually wired up, because losing the one
// line in the JSON would silently allow both invalid shapes.
func TestPriceMustTargetExactlyOneOfTemplateOrVariant(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := ProductPriceSchemaBuilder().Build()

	const templateId = "01M0A1P4QM3P0Y3T60T7RH2KZC"
	const variantId = "01M0A1P4QJ0MFYJY0EQH0NVJ2R"

	newPrice := func() dmodel.DynamicFields {
		return dmodel.DynamicFields{
			ProductPriceFieldPrice:  "1000",
			ProductPriceFieldStatus: string(ProductPriceStatusDraft),
			ProductPriceFieldOrgId:  templateId,
		}
	}

	t.Run("template only is valid", func(t *testing.T) {
		fields := newPrice()
		fields[ProductPriceFieldProductTemplateId] = templateId
		_, errs := schema.Validate(fields)
		assert.Zero(t, errs.Count(), "a template-level base price is the ordinary case: %v", errs)
	})

	t.Run("variant only is valid", func(t *testing.T) {
		fields := newPrice()
		fields[ProductPriceFieldProductVariantId] = variantId
		_, errs := schema.Validate(fields)
		assert.Zero(t, errs.Count(), "a variant override is explicitly allowed by rule 2: %v", errs)
	})

	t.Run("neither is rejected", func(t *testing.T) {
		_, errs := schema.Validate(newPrice())
		assert.NotZero(t, errs.Count(), "a price attached to nothing can never be applied")
	})

	t.Run("both is rejected", func(t *testing.T) {
		fields := newPrice()
		fields[ProductPriceFieldProductTemplateId] = templateId
		fields[ProductPriceFieldProductVariantId] = variantId
		_, errs := schema.Validate(fields)
		assert.NotZero(t, errs.Count(), "targeting both is the ambiguity rule 1 forbids")
	})
}

func TestProductPriceStatusValues(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := ProductPriceSchemaBuilder().Build()

	field := requireField(t, schema, ProductPriceFieldStatus)
	values := field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues]
	assert.ElementsMatch(t,
		[]string{
			string(ProductPriceStatusDraft),
			string(ProductPriceStatusApproved),
			string(ProductPriceStatusExpired),
		},
		values,
		"the Go constants and the schema's enum must not drift apart")

	// A price takes effect only once approved, so anything created without an explicit state must
	// start as draft rather than silently pricing a product. Asserted through validation because
	// the default is applied there; the field's own default holder is unexported.
	sanitized, errs := schema.Validate(dmodel.DynamicFields{
		ProductPriceFieldPrice:             "1000",
		ProductPriceFieldOrgId:             "01M0A1P4QM3P0Y3T60T7RH2KZC",
		ProductPriceFieldProductTemplateId: "01M0A1P4QM3P0Y3T60T7RH2KZC",
	})
	require.Zero(t, errs.Count(), "status is defaulted, so omitting it must not be an error: %v", errs)

	assert.Equal(t, string(ProductPriceStatusDraft), sanitized[ProductPriceFieldStatus])
}

// Several price rows per product is the point of this table: a scheduled change is two rows with
// different effective ranges, and price lists will add more. Every sibling entity has a unique
// tuple, so its absence here is deliberate and worth pinning — adding one would forbid scheduling.
func TestPriceHasNoCompositeUnique(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := ProductPriceSchemaBuilder().Build()

	assert.Empty(t, schema.CompositeUniques(),
		"a product must be allowed more than one price row")
}

// BR §6.12 rule 3 requires historical transaction prices to be snapshotted on the transaction.
// The catalog therefore holds the price in exactly one place; a copy cached onto the template or
// the variant would be a second source of truth that drifts.
func TestPriceIsNotDuplicatedOntoTemplateOrVariant(t *testing.T) {
	requireBaseSchemasRegistered(t)

	template := ProductTemplateSchemaBuilder().Build()
	variant := ProductVariantSchemaBuilder().Build()

	for _, fieldName := range []string{"price", "list_price", "proposed_price", "sale_price"} {
		_, onTemplate := template.Fields()[fieldName]
		assert.Falsef(t, onTemplate, "template must not store %q; price lives on inventory_product_price", fieldName)

		_, onVariant := variant.Fields()[fieldName]
		assert.Falsef(t, onVariant, "variant must not store %q; price lives on inventory_product_price", fieldName)
	}
}

func TestProductPriceAccessorsRoundTrip(t *testing.T) {
	requireBaseSchemasRegistered(t)

	price := NewProductPrice()
	price.SetStatus(WrapProductPriceStatus(string(ProductPriceStatusApproved)))

	status := price.GetStatus()
	require.NotNil(t, status)
	assert.Equal(t, ProductPriceStatusApproved, *status)

	// Nil must clear the field rather than writing the empty string, or a cleared status would
	// read back as an enum value that is not in the schema's allowed set.
	price.SetStatus(nil)
	assert.Nil(t, price.GetStatus())
}
