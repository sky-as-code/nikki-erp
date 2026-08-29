package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// The display name is composed, never stored, so it cannot drift from the combination it describes.
func TestBuildDisplayName(t *testing.T) {
	name := langJson("Classic T-Shirt")

	assert.Equal(t, "Classic T-Shirt / Black / M",
		BuildDisplayName(name, []string{"Black", "M"}))

	// The single-variant case: no attribute values, so the template name stands alone.
	assert.Equal(t, "Classic T-Shirt", BuildDisplayName(name, nil))
}

// A missing label must not leave a dangling separator in the middle of the name.
func TestBuildDisplayNameSkipsEmptyLabels(t *testing.T) {
	assert.Equal(t, "Classic T-Shirt / M",
		BuildDisplayName(langJson("Classic T-Shirt"), []string{"", "M"}))
}

// LangJson is a map, so iterating it directly would pick a different language on each call and
// render the same product under a different name each time.
func TestDisplayNameIsDeterministicAcrossLocales(t *testing.T) {
	name := &model.LangJson{
		"vi-VN":                "Áo thun",
		model.LanguageCodeEnUs: "T-Shirt",
		"fr-FR":                "T-Shirt FR",
	}

	first := BuildDisplayName(name, nil)
	for i := 0; i < 20; i++ {
		assert.Equal(t, first, BuildDisplayName(name, nil))
	}
	// en-US wins when present, rather than an arbitrary map key.
	assert.Equal(t, "T-Shirt", first)
}

func TestDisplayNameFallsBackWhenEnglishAbsent(t *testing.T) {
	name := &model.LangJson{"vi-VN": "Áo thun"}

	assert.Equal(t, "Áo thun", BuildDisplayName(name, nil))
}

// The variant reads name, type, category, brand and the capability flags from its template rather
// than holding copies.
func TestEffectiveProductInheritsTemplateOwnedFields(t *testing.T) {
	template, variant := newTemplate(), newVariant()

	effective := BuildEffectiveProduct(template, variant, []string{"Black"})

	assert.Equal(t, "01TYPE", effective.ProductTypeId)
	assert.Equal(t, "01CATEGORY", effective.CategoryId)
	assert.Equal(t, "01BRAND", effective.BrandId)
	assert.True(t, effective.SaleOk)
	assert.True(t, effective.PurchaseOk)
	assert.Equal(t, "Classic T-Shirt / Black", effective.DisplayName)
}

// SKU and barcode belong to the variant: a template with several variants has nothing they could
// refer to.
func TestEffectiveProductTakesIdentifiersFromVariant(t *testing.T) {
	effective := BuildEffectiveProduct(newTemplate(), newVariant(), nil)

	assert.Equal(t, "TSH-BLK-M", effective.Sku)
	assert.Equal(t, "8931234567890", effective.Barcode)
}

// A variant with no image of its own shows the template's; setting one overrides it, and clearing
// it falls back again. Null means "inherit", not "no image".
func TestEffectiveProductImageFallback(t *testing.T) {
	template := newTemplate()

	withoutOverride := newVariant()
	assert.Equal(t, "01TEMPLATEIMAGE",
		BuildEffectiveProduct(template, withoutOverride, nil).ImageId,
		"a variant with no image shows the template's")

	withOverride := newVariant()
	withOverride.SetVariantImageId(strPtr("01VARIANTIMAGE"))
	assert.Equal(t, "01VARIANTIMAGE",
		BuildEffectiveProduct(template, withOverride, nil).ImageId,
		"a variant image overrides the template's")

	cleared := newVariant()
	cleared.SetVariantImageId(nil)
	assert.Equal(t, "01TEMPLATEIMAGE",
		BuildEffectiveProduct(template, cleared, nil).ImageId,
		"clearing the override returns to the template image")
}

// Weight and dimensions follow the same controlled fallback as the image.
func TestEffectiveProductWeightFallback(t *testing.T) {
	template := newTemplate()

	inherited := BuildEffectiveProduct(template, newVariant(), nil)
	require.NotNil(t, inherited.Weight)
	assert.True(t, inherited.Weight.Equal(decimal.RequireFromString("1.5")))

	overriding := newVariant()
	overridden := decimal.RequireFromString("2.25")
	overriding.SetWeight(&overridden)

	effective := BuildEffectiveProduct(template, overriding, nil)
	require.NotNil(t, effective.Weight)
	assert.True(t, effective.Weight.Equal(overridden))
}

// Either level can withdraw a product from new transactions.
func TestEffectiveProductSelectability(t *testing.T) {
	testCases := []struct {
		name             string
		templateArchived bool
		variantArchived  bool
		templateStatus   models.ProductTemplateStatus
		variantStatus    models.ProductVariantStatus
		want             bool
	}{
		{"active on both levels", false, false,
			models.ProductTemplateStatusActive, models.ProductVariantStatusActive, true},
		{"archived template hides its variant", true, false,
			models.ProductTemplateStatusActive, models.ProductVariantStatusActive, false},
		{"archived variant", false, true,
			models.ProductTemplateStatusActive, models.ProductVariantStatusActive, false},
		{"discontinued template", false, false,
			models.ProductTemplateStatusDiscontinued, models.ProductVariantStatusActive, false},
		{"discontinued variant", false, false,
			models.ProductTemplateStatusActive, models.ProductVariantStatusDiscontinued, false},
		{"draft template is still selectable", false, false,
			models.ProductTemplateStatusDraft, models.ProductVariantStatusActive, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			effective := itProduct.EffectiveProduct{
				IsTemplateArchived: testCase.templateArchived,
				IsVariantArchived:  testCase.variantArchived,
				TemplateStatus:     testCase.templateStatus.String(),
				VariantStatus:      testCase.variantStatus.String(),
			}

			assert.Equal(t, testCase.want, effective.IsSelectable())
		})
	}
}

// The vending machine snapshots the product as an opaque map, so the key set must keep carrying
// what its consumers already read.
func TestEffectiveProductFieldMapKeepsConsumerKeys(t *testing.T) {
	fields := BuildEffectiveProduct(newTemplate(), newVariant(), []string{"Black"}).ToFieldMap()

	for _, key := range []string{"id", "name", "sku", "barcode", "image_id", "status", "is_archived"} {
		assert.Containsf(t, fields, key, "consumers read %q from the product snapshot", key)
	}
	// The variant is the concrete product, so "id" must be the variant's.
	assert.Equal(t, "01VARIANT", fields["id"])
	assert.Equal(t, "Classic T-Shirt / Black", fields["name"])
}

func TestVariantValueIds(t *testing.T) {
	variant := models.NewProductVariant()
	variant.SetCombinationKey(strPtr("01ATTRCOLOR:01VALBLACK|01ATTRSIZE:01VALM"))

	assert.Equal(t, []string{"01VALBLACK", "01VALM"}, VariantValueIds(variant))
}

// The single-variant case carries no values, and must not be mistaken for malformed data.
func TestVariantValueIdsOfEmptyCombination(t *testing.T) {
	variant := models.NewProductVariant()
	variant.SetCombinationKey(strPtr(models.EmptyCombinationKey))

	assert.Empty(t, VariantValueIds(variant))
}

func newTemplate() *models.ProductTemplate {
	template := models.NewProductTemplate()
	template.SetId(strPtr("01TEMPLATE"))
	template.SetName(langJson("Classic T-Shirt"))
	template.SetProductTypeId(strPtr("01TYPE"))
	template.SetCategoryId(strPtr("01CATEGORY"))
	template.SetBrandId(strPtr("01BRAND"))
	template.SetSaleOk(boolPtr(true))
	template.SetPurchaseOk(boolPtr(true))
	template.SetDefaultImageId(strPtr("01TEMPLATEIMAGE"))
	weight := decimal.RequireFromString("1.5")
	template.SetDefaultWeight(&weight)
	status := models.ProductTemplateStatusActive
	template.SetStatus(&status)
	return template
}

func newVariant() *models.ProductVariant {
	variant := models.NewProductVariant()
	variant.SetId(strPtr("01VARIANT"))
	variant.SetProductTemplateId(strPtr("01TEMPLATE"))
	variant.SetSku(strPtr("TSH-BLK-M"))
	variant.SetPrimaryBarcode(strPtr("8931234567890"))
	variant.SetCombinationKey(strPtr("01ATTRCOLOR:01VALBLACK"))
	status := models.ProductVariantStatusActive
	variant.SetStatus(&status)
	return variant
}

func langJson(text string) *model.LangJson {
	value := model.LangJson{model.LanguageCodeEnUs: text}
	return &value
}

func strPtr(v string) *string {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}
