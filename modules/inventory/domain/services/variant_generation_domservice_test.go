package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// The combination key is an identity, not a rendering. Everything below pins a property identity
// depends on: get one wrong and two requests for the same product resolve to different variants.

// The order values arrive in must not change the key: a sales configurator picking Size then
// Color must reach the same variant as one picking Color then Size.
func TestCombinationKeyIsOrderIndependent(t *testing.T) {
	colorBlack := itProduct.AttributeSelection{AttributeId: "01ATTRCOLOR", ValueId: "01VALBLACK", Mode: models.VariantCreationModeInstant}
	sizeM := itProduct.AttributeSelection{AttributeId: "01ATTRSIZE", ValueId: "01VALM", Mode: models.VariantCreationModeInstant}

	assert.Equal(t,
		BuildCombinationKey([]itProduct.AttributeSelection{colorBlack, sizeM}),
		BuildCombinationKey([]itProduct.AttributeSelection{sizeM, colorBlack}))
}

// A NEVER attribute carries information but is not part of what makes a variant distinct, so it
// must not enter the key.
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

// A template with no variant-generating attributes still has one concrete variant. Its key is the
// empty string, which is a key rather than a missing value.
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
