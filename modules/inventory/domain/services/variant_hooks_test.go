package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

const hookTemplateId = "01TEMPLATE0000000000000000"

// The create check runs on the raw request, ahead of the schema's generic composite unique, so the
// clash is reported on combination_key with the variant's own key.
func TestCheckCreateCombinationReportsTheDuplicateOnItsField(t *testing.T) {
	repo := &fakeSearcher{items: []dmodel.DynamicFields{{models.ProductVariantFieldId: "01OTHER00000000000000000000"}}}
	entity := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldProductTemplateId: hookTemplateId,
		models.ProductVariantFieldCombinationKey:    "12:101",
	})

	vErrs := &ft.ClientErrors{}
	_, err := checkCreateCombination(repo)(testContext(t), entity, vErrs)

	require.NoError(t, err)
	assertViolation(t, vErrs, "product_variant.duplicate_combination")
	assert.Equal(t, models.ProductVariantFieldCombinationKey, (*vErrs)[0].Field)
}

// A value of the wrong type is the schema validator's to report; the hook must neither panic on it
// nor search with it.
func TestCheckCreateCombinationLeavesMalformedInputToTheSchema(t *testing.T) {
	repo := &fakeSearcher{}
	entity := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldProductTemplateId: 42,
		models.ProductVariantFieldCombinationKey:    "12:101",
	})

	vErrs := &ft.ClientErrors{}
	_, err := checkCreateCombination(repo)(testContext(t), entity, vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
	assert.Zero(t, repo.calls)
}

// product_template_id is no_update, so the schema validator drops it before ValidateExtra. The
// value captured beforehand is what lets the update refuse a re-parenting instead of answering 200.
func TestVariantUpdateRefusesReparentingCapturedBeforeValidation(t *testing.T) {
	submitted := ""
	_, err := captureSubmittedTemplateId(&submitted)(testContext(t), composable.NewDynamicEntityFrom(
		dmodel.DynamicFields{models.ProductVariantFieldProductTemplateId: "01OTHERTEMPLATE00000000000"},
	), &ft.ClientErrors{})
	require.NoError(t, err)

	// What ValidateExtra sees: the field already dropped from the input.
	input := composable.NewDynamicEntityFrom(dmodel.DynamicFields{})
	stored := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldId:                "01VARIANT00000000000000000",
		models.ProductVariantFieldProductTemplateId: hookTemplateId,
	})
	vErrs := &ft.ClientErrors{}
	require.NoError(t, validateVariantUpdate(&fakeSearcher{}, &submitted)(testContext(t), input, stored, vErrs))

	assertViolation(t, vErrs, "product_variant.immutable_template")
}

func TestVariantUpdateAcceptsTheSameTemplate(t *testing.T) {
	submitted := hookTemplateId
	stored := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldId:                "01VARIANT00000000000000000",
		models.ProductVariantFieldProductTemplateId: hookTemplateId,
	})

	vErrs := &ft.ClientErrors{}
	require.NoError(t, validateVariantUpdate(&fakeSearcher{}, &submitted)(
		testContext(t), composable.NewDynamicEntityFrom(dmodel.DynamicFields{}), stored, vErrs))

	assert.Zero(t, vErrs.Count())
}
