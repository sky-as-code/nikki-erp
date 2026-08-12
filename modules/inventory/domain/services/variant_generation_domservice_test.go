package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// BR §14.3: the combination key is an identity, not a rendering. Everything below pins a
// property that identity depends on — get any of them wrong and two requests for the same
// product resolve to two different variants.

// The order values arrive in must not change the key: a sales configurator picking Size then
// Color must reach the same variant as one picking Color then Size.
func TestCombinationKeyIsOrderIndependent(t *testing.T) {
	colorBlack := itProduct.AttributeSelection{AttributeId: "01ATTRCOLOR", ValueId: "01VALBLACK", Mode: models.VariantCreationModeInstant}
	sizeM := itProduct.AttributeSelection{AttributeId: "01ATTRSIZE", ValueId: "01VALM", Mode: models.VariantCreationModeInstant}

	assert.Equal(t,
		BuildCombinationKey([]itProduct.AttributeSelection{colorBlack, sizeM}),
		BuildCombinationKey([]itProduct.AttributeSelection{sizeM, colorBlack}))
}

// BR §6.5.3 and §14.3 step 2: a NEVER attribute carries information but is not part of what
// makes a variant distinct, so it must not enter the key.
func TestCombinationKeyExcludesNeverAttributes(t *testing.T) {
	generating := itProduct.AttributeSelection{AttributeId: "01ATTRCOLOR", ValueId: "01VALBLACK", Mode: models.VariantCreationModeInstant}
	never := itProduct.AttributeSelection{AttributeId: "01ATTRNOTE", ValueId: "01VALGIFT", Mode: models.VariantCreationModeNever}

	withNever := BuildCombinationKey([]itProduct.AttributeSelection{generating, never})
	withoutNever := BuildCombinationKey([]itProduct.AttributeSelection{generating})

	assert.Equal(t, withoutNever, withNever)
	assert.NotContains(t, withNever, "01ATTRNOTE")
}

// A DYNAMIC attribute does take part in identity; it only differs in when its variant is
// materialized.
func TestCombinationKeyIncludesDynamicAttributes(t *testing.T) {
	dynamic := itProduct.AttributeSelection{AttributeId: "01ATTRSIZE", ValueId: "01VALXL", Mode: models.VariantCreationModeDynamic}

	assert.Contains(t, BuildCombinationKey([]itProduct.AttributeSelection{dynamic}), "01ATTRSIZE")
}

// BR §4.5 and AC-PROD-008: a template with no variant-generating attributes still has one
// concrete variant. Its key is the empty string, which is a key rather than a missing value.
func TestCombinationKeyIsEmptyWithoutGeneratingAttributes(t *testing.T) {
	assert.Equal(t, models.EmptyCombinationKey, BuildCombinationKey(nil))
	assert.Equal(t, models.EmptyCombinationKey, BuildCombinationKey([]itProduct.AttributeSelection{
		{AttributeId: "01ATTRNOTE", ValueId: "01VALGIFT", Mode: models.VariantCreationModeNever},
	}))
}

// Two values for one attribute is not a combination of both: the last choice wins, so a
// configurator that re-picks Color does not produce a two-colour variant.
func TestCombinationKeyCollapsesRepeatedAttribute(t *testing.T) {
	key := BuildCombinationKey([]itProduct.AttributeSelection{
		{AttributeId: "01ATTRCOLOR", ValueId: "01VALBLACK", Mode: models.VariantCreationModeInstant},
		{AttributeId: "01ATTRCOLOR", ValueId: "01VALWHITE", Mode: models.VariantCreationModeInstant},
	})

	assert.Equal(t, "01ATTRCOLOR:01VALWHITE", key)
}

func TestCombinationKeyRoundTrips(t *testing.T) {
	selections := []itProduct.AttributeSelection{
		{AttributeId: "01ATTRCOLOR", ValueId: "01VALBLACK", Mode: models.VariantCreationModeInstant},
		{AttributeId: "01ATTRSIZE", ValueId: "01VALM", Mode: models.VariantCreationModeInstant},
	}

	parsed := ParseCombinationKey(BuildCombinationKey(selections))

	assert.Equal(t, []itProduct.AttributeSelection{
		{AttributeId: "01ATTRCOLOR", ValueId: "01VALBLACK"},
		{AttributeId: "01ATTRSIZE", ValueId: "01VALM"},
	}, parsed)
}

func TestParseEmptyCombinationKey(t *testing.T) {
	assert.Empty(t, ParseCombinationKey(models.EmptyCombinationKey))
}

// BR §8.2: Color {Black, White} x Size {M, L} produces exactly four variants.
func TestBuildInstantCombinationsCartesianProduct(t *testing.T) {
	combinations := BuildInstantCombinations([]itProduct.AttributeOptions{
		{AttributeId: "01ATTRCOLOR", Mode: models.VariantCreationModeInstant, ValueIds: []string{"01VALBLACK", "01VALWHITE"}},
		{AttributeId: "01ATTRSIZE", Mode: models.VariantCreationModeInstant, ValueIds: []string{"01VALM", "01VALL"}},
	})

	assert.ElementsMatch(t, []string{
		"01ATTRCOLOR:01VALBLACK|01ATTRSIZE:01VALM",
		"01ATTRCOLOR:01VALBLACK|01ATTRSIZE:01VALL",
		"01ATTRCOLOR:01VALWHITE|01ATTRSIZE:01VALM",
		"01ATTRCOLOR:01VALWHITE|01ATTRSIZE:01VALL",
	}, combinations)
}

// BR §4.7: DYNAMIC combinations are materialized on use, so instant generation must not
// pre-create them.
func TestBuildInstantCombinationsIgnoresDynamicAndNever(t *testing.T) {
	combinations := BuildInstantCombinations([]itProduct.AttributeOptions{
		{AttributeId: "01ATTRCOLOR", Mode: models.VariantCreationModeInstant, ValueIds: []string{"01VALBLACK"}},
		{AttributeId: "01ATTRSIZE", Mode: models.VariantCreationModeDynamic, ValueIds: []string{"01VALM", "01VALL"}},
		{AttributeId: "01ATTRNOTE", Mode: models.VariantCreationModeNever, ValueIds: []string{"01VALGIFT"}},
	})

	assert.Equal(t, []string{"01ATTRCOLOR:01VALBLACK"}, combinations)
}

// A template with no instant attributes yields the single empty combination — one variant, not
// zero. Returning an empty slice here would leave the product with nothing to transact against.
func TestBuildInstantCombinationsWithoutAttributes(t *testing.T) {
	assert.Equal(t, []string{models.EmptyCombinationKey}, BuildInstantCombinations(nil))
}

// An attribute configured with no allowed values would multiply the cartesian product by zero
// and wipe out every combination. It means "not configured yet", not "no products".
func TestBuildInstantCombinationsSkipsValuelessAttribute(t *testing.T) {
	combinations := BuildInstantCombinations([]itProduct.AttributeOptions{
		{AttributeId: "01ATTRCOLOR", Mode: models.VariantCreationModeInstant, ValueIds: []string{"01VALBLACK"}},
		{AttributeId: "01ATTRSIZE", Mode: models.VariantCreationModeInstant, ValueIds: nil},
	})

	assert.Equal(t, []string{"01ATTRCOLOR:01VALBLACK"}, combinations)
}

// BR §8.5 and AC-PROD-030: synchronization reports what changed and leaves the decision to the
// caller, because archiving a used variant and deleting an unused one are different operations
// and only the caller knows which applies.
func TestPlanVariantSync(t *testing.T) {
	plan := PlanVariantSync(
		[]string{"a", "b", "c"},
		[]string{"b", "c", "d"},
	)

	assert.Equal(t, []string{"a"}, plan.ToCreate)
	assert.ElementsMatch(t, []string{"b", "c"}, plan.Unchanged)
	assert.Equal(t, []string{"d"}, plan.Obsolete)
}

func TestPlanVariantSyncWithNothingToDo(t *testing.T) {
	plan := PlanVariantSync([]string{"a"}, []string{"a"})

	assert.Empty(t, plan.ToCreate)
	assert.Empty(t, plan.Obsolete)
	assert.Equal(t, []string{"a"}, plan.Unchanged)
}

// The first generation of a template's variants: everything is new.
func TestPlanVariantSyncFromNothing(t *testing.T) {
	plan := PlanVariantSync([]string{"a", "b"}, nil)

	assert.Equal(t, []string{"a", "b"}, plan.ToCreate)
	assert.Empty(t, plan.Obsolete)
}
