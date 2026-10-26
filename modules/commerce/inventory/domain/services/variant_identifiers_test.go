package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const identifierOrgId = "01ORG000000000000000000000"

// The partial uniques skip only NULL, so a cleared SKU stored as "" would collide with every other
// cleared one. A value with stray spaces is trimmed so " A1" and "A1" are the same SKU.
func TestNormalizeVariantIdentifiersStoresBlanksAsNull(t *testing.T) {
	entity := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldSku:            "   ",
		models.ProductVariantFieldPrimaryBarcode: " 8935001000011 ",
	})

	_, err := normalizeVariantIdentifiers()(testContext(t), entity, &ft.ClientErrors{})

	require.NoError(t, err)
	data := entity.GetFieldData()
	value, present := data[models.ProductVariantFieldSku]
	assert.True(t, present, "the blank must be written as NULL, not dropped from an update")
	assert.Nil(t, value)
	assert.Equal(t, "8935001000011", data[models.ProductVariantFieldPrimaryBarcode])
}

func TestCreateReportsADuplicateSkuOnItsField(t *testing.T) {
	repo := &fakeSearcher{items: []dmodel.DynamicFields{{models.ProductVariantFieldId: "01OTHER"}}}
	entity := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldOrgId: identifierOrgId,
		models.ProductVariantFieldSku:   "INV-SEED-ED-BERRY",
	})

	vErrs := &ft.ClientErrors{}
	err := validateCreateIdentifiers(repo)(testContext(t), entity, nil, vErrs)

	require.NoError(t, err)
	assertViolation(t, vErrs, "inventory:product_variant.sku_duplicate")
	assert.Equal(t, models.ProductVariantFieldSku, (*vErrs)[0].Field)
	assert.Equal(t, 1, repo.calls, "only the submitted identifier is searched")
}

// The record being updated finds itself under its own barcode, which is no conflict.
func TestUpdateKeepingItsOwnBarcodeIsNoConflict(t *testing.T) {
	repo := &fakeSearcher{items: []dmodel.DynamicFields{{models.ProductVariantFieldId: "01SELF"}}}
	input := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldPrimaryBarcode: "8935001000011",
	})
	stored := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldId: "01SELF", models.ProductVariantFieldOrgId: identifierOrgId,
	})

	vErrs := &ft.ClientErrors{}
	err := validateUpdateIdentifiers(repo)(testContext(t), input, stored, vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}

func TestUpdateReportsABarcodeAnotherVariantHolds(t *testing.T) {
	repo := &fakeSearcher{items: []dmodel.DynamicFields{{models.ProductVariantFieldId: "01OTHER"}}}
	input := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldPrimaryBarcode: "8935001000011",
	})
	stored := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldId: "01SELF", models.ProductVariantFieldOrgId: identifierOrgId,
	})

	vErrs := &ft.ClientErrors{}
	err := validateUpdateIdentifiers(repo)(testContext(t), input, stored, vErrs)

	require.NoError(t, err)
	assertViolation(t, vErrs, "inventory:product_variant.primary_barcode_duplicate")
}

// A cleared identifier (NULL after normalization) never collides, so it is not searched.
func TestNullIdentifiersAreNotSearched(t *testing.T) {
	repo := &fakeSearcher{items: []dmodel.DynamicFields{{models.ProductVariantFieldId: "01OTHER"}}}
	entity := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.ProductVariantFieldOrgId: identifierOrgId,
		models.ProductVariantFieldSku:   nil,
	})

	vErrs := &ft.ClientErrors{}
	err := validateCreateIdentifiers(repo)(testContext(t), entity, nil, vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
	assert.Zero(t, repo.calls)
}

// Archiving a template's last selectable variant archives the template too; stamping that
// variant template_cascade is what lets the template's unarchive bring it back.
func TestArchiveStampFollowsTheTemplateAutoArchive(t *testing.T) {
	assert.Nil(t, archiveStamp(false, false))
	require.NotNil(t, archiveStamp(true, false))
	assert.Equal(t, models.ArchiveSourceUser, *archiveStamp(true, false))
	assert.Equal(t, models.ArchiveSourceTemplateCascade, *archiveStamp(true, true))
}
