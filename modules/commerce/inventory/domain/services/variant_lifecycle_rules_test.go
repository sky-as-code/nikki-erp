package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func archivedWithSource(source *models.ArchiveSource) *models.ProductVariant {
	variant := models.NewProductVariant()
	variant.SetId(strPtr("01VARIANT"))
	variant.SetIsArchived(boolPtr(true))
	variant.SetArchiveSource(source)
	return variant
}

func sourcePtr(source models.ArchiveSource) *models.ArchiveSource {
	return &source
}

func TestStrongerArchiveSourceFollowsPrecedence(t *testing.T) {
	user, sync, cascade := models.ArchiveSourceUser, models.ArchiveSourceSystemSync, models.ArchiveSourceTemplateCascade
	assert.Equal(t, user, StrongerArchiveSource(sync, user))
	assert.Equal(t, user, StrongerArchiveSource(user, cascade))
	assert.Equal(t, sync, StrongerArchiveSource(cascade, sync))
	assert.Equal(t, sync, StrongerArchiveSource(sync, cascade))
	assert.Equal(t, cascade, StrongerArchiveSource("", cascade))
}

func TestRaiseArchiveFieldsNeverWeakensTheReason(t *testing.T) {
	_, ok := RaiseArchiveFields(archivedWithSource(sourcePtr(models.ArchiveSourceUser)), models.ArchiveSourceSystemSync)
	assert.False(t, ok, "a user archive outranks system_sync")

	_, ok = RaiseArchiveFields(archivedWithSource(nil), models.ArchiveSourceSystemSync)
	assert.False(t, ok, "an archive with no source counts as the user's")

	cascaded := archivedWithSource(sourcePtr(models.ArchiveSourceTemplateCascade))
	fields, ok := RaiseArchiveFields(cascaded, models.ArchiveSourceSystemSync)
	assert.True(t, ok)
	assert.Equal(t, models.ArchiveSourceSystemSync.String(), fields[models.ProductVariantFieldArchiveSource])

	active := models.NewProductVariant()
	active.SetId(strPtr("01VARIANT"))
	active.SetIsArchived(boolPtr(false))
	fields, ok = RaiseArchiveFields(active, models.ArchiveSourceSystemSync)
	assert.True(t, ok)
	assert.Equal(t, true, fields[basemodel.FieldIsArchived])
}

func TestLiftArchiveFieldsLiftsOnlyItsOwnReason(t *testing.T) {
	_, ok := LiftArchiveFields(archivedWithSource(sourcePtr(models.ArchiveSourceUser)), models.ArchiveSourceSystemSync)
	assert.False(t, ok)

	synced := archivedWithSource(sourcePtr(models.ArchiveSourceSystemSync))
	fields, ok := LiftArchiveFields(synced, models.ArchiveSourceSystemSync)
	assert.True(t, ok)
	assert.Equal(t, false, fields[basemodel.FieldIsArchived])
	assert.Nil(t, fields[models.ProductVariantFieldArchiveSource])
}

// The template cascade never overwrites a stronger reason: an archived variant is skipped.
func TestCascadeSkipsAVariantArchivedForAStrongerReason(t *testing.T) {
	for _, source := range []models.ArchiveSource{models.ArchiveSourceUser, models.ArchiveSourceSystemSync} {
		assert.True(t, ShouldSkipCascade(archivedWithSource(sourcePtr(source)), true), source)
		assert.True(t, ShouldSkipCascade(archivedWithSource(sourcePtr(source)), false), source)
	}
}

func mixedConfiguration() TemplateConfiguration {
	return TemplateConfiguration{
		Lines: []TemplateLineConfig{
			{AttributeId: "A1", Mode: models.VariantCreationModeInstant, SelectedValueIds: map[string]bool{"V1": true, "V2": true}},
			{AttributeId: "A2", Mode: models.VariantCreationModeDynamic, SelectedValueIds: map[string]bool{"V3": true}},
			{AttributeId: "A3", Mode: models.VariantCreationModeNever, SelectedValueIds: map[string]bool{"V4": true}},
		},
	}
}

func keyOf(pairs ...[2]string) string {
	selections := make([]itProduct.AttributeSelection, 0, len(pairs))
	for _, pair := range pairs {
		selections = append(selections, itProduct.AttributeSelection{
			AttributeId: pair[0], ValueId: pair[1], Mode: models.VariantCreationModeInstant,
		})
	}
	return BuildCombinationKey(selections)
}

func TestCombinationValidity(t *testing.T) {
	config := mixedConfiguration()
	complete, base := models.VariantKindComplete, models.VariantKindBase

	assert.True(t, config.IsCombinationValid(keyOf([2]string{"A1", "V1"}, [2]string{"A2", "V3"}), complete))
	assert.True(t, config.IsCombinationValid(keyOf([2]string{"A1", "V2"}), base))
	assert.False(t, config.IsCombinationValid(keyOf([2]string{"A1", "V1"}), complete), "a complete needs its dynamic value")
	assert.False(t, config.IsCombinationValid(keyOf([2]string{"A1", "V1"}, [2]string{"A2", "V3"}), base),
		"a base carries the instant values only")
	assert.False(t, config.IsCombinationValid(keyOf([2]string{"A1", "V9"}, [2]string{"A2", "V3"}), complete),
		"an unselected value")
	assert.False(t, config.IsCombinationValid(
		keyOf([2]string{"A1", "V1"}, [2]string{"A2", "V3"}, [2]string{"A3", "V4"}), complete),
		"a never attribute is not part of the key")

	config.Exclusions = [][2]string{{"V1", "V3"}}
	assert.False(t, config.IsCombinationValid(keyOf([2]string{"A1", "V1"}, [2]string{"A2", "V3"}), complete))
	assert.True(t, config.IsCombinationValid(keyOf([2]string{"A1", "V2"}, [2]string{"A2", "V3"}), complete))
}

func TestBaseIsInvalidWithoutADynamicLine(t *testing.T) {
	config := TemplateConfiguration{Lines: []TemplateLineConfig{
		{AttributeId: "A1", Mode: models.VariantCreationModeInstant, SelectedValueIds: map[string]bool{"V1": true}},
	}}
	assert.False(t, config.IsCombinationValid(keyOf([2]string{"A1", "V1"}), models.VariantKindBase))
	assert.True(t, config.IsCombinationValid(keyOf([2]string{"A1", "V1"}), models.VariantKindComplete))
	assert.True(t, TemplateConfiguration{}.IsCombinationValid(models.EmptyCombinationKey, models.VariantKindComplete),
		"the empty default variant of a template with no attribute")
	assert.False(t, config.HasOnDemandLine())
	assert.True(t, mixedConfiguration().HasOnDemandLine())
}

func TestVariantIneligibility(t *testing.T) {
	base := newVariant()
	base.SetVariantKind(variantKindPtr(models.VariantKindBase))
	assert.Equal(t, EligibilityReasonBase, VariantIneligibility(base, newTemplate(), false))

	complete := newVariant()
	complete.SetVariantKind(variantKindPtr(models.VariantKindComplete))
	complete.SetIsArchived(boolPtr(true))
	assert.Equal(t, EligibilityReasonVariantArchived, VariantIneligibility(complete, newTemplate(), true))
	assert.Empty(t, VariantIneligibility(complete, newTemplate(), false), "existing business may finish")

	active := newVariant()
	template := newTemplate()
	template.SetIsArchived(boolPtr(true))
	assert.Equal(t, EligibilityReasonTemplateArchived, VariantIneligibility(active, template, true))
	assert.Empty(t, VariantIneligibility(active, newTemplate(), true))
}

func TestEffectiveProductOfABaseIsNotSelectable(t *testing.T) {
	variant := newVariant()
	variant.SetVariantKind(variantKindPtr(models.VariantKindBase))
	assert.False(t, BuildEffectiveProduct(newTemplate(), variant, nil).IsSelectable())
	assert.True(t, BuildEffectiveProduct(newTemplate(), newVariant(), nil).IsSelectable(),
		"a row without the kind is a complete")
}

func TestNormalizeExclusionPair(t *testing.T) {
	left, right := NormalizeExclusionPair("01B", "01A")
	assert.Equal(t, [2]string{"01A", "01B"}, [2]string{left, right})
	left, right = NormalizeExclusionPair("01A", "01B")
	assert.Equal(t, [2]string{"01A", "01B"}, [2]string{left, right})
}

func TestExclusionPairViolations(t *testing.T) {
	testCases := []struct {
		name  string
		left  ExclusionEnd
		right ExclusionEnd
		want  int
	}{
		{"valid pair", ExclusionEnd{"V1", "T1", "A1"}, ExclusionEnd{"V2", "T1", "A2"}, 0},
		{"same value", ExclusionEnd{"V1", "T1", "A1"}, ExclusionEnd{"V1", "T1", "A1"}, 1},
		{"same attribute", ExclusionEnd{"V1", "T1", "A1"}, ExclusionEnd{"V2", "T1", "A1"}, 1},
		{"other template", ExclusionEnd{"V1", "T1", "A1"}, ExclusionEnd{"V2", "T2", "A2"}, 1},
		{"missing values", ExclusionEnd{ValueId: "V1"}, ExclusionEnd{ValueId: "V2"}, 2},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			vErrs := ft.NewClientErrors()
			AppendExclusionPairViolations("T1", testCase.left, testCase.right, vErrs)
			assert.Equal(t, testCase.want, vErrs.Count())
		})
	}
}

func variantKindPtr(kind models.VariantKind) *models.VariantKind {
	return &kind
}
