package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// buildVariantSchema registers the base schemas the variant extends, which CoreModule normally
// does at start-up. Registering here rather than relying on another test having run first keeps
// each test in this file independent of ordering.
func buildVariantSchema(t *testing.T) *dmodel.ModelSchema {
	t.Helper()
	_ = basemodel.RegisterJsonBaseSchemas()
	return ProductVariantSchemaBuilder().Build()
}

// The variant schema must load with every template_* field virtual: one of them declared as an
// ordinary field would silently add a column and demand a migration.
func TestProductVariant_TemplateFieldsAreVirtual(t *testing.T) {
	schema := buildVariantSchema(t)

	for _, name := range TemplateVirtualFields {
		field, ok := schema.Field(name)
		require.True(t, ok, "field %q must be declared on the schema", name)
		assert.True(t, field.IsVirtual(), "field %q must be virtual", name)
	}
}

func TestProductVariant_TemplateFieldsHaveNoColumn(t *testing.T) {
	schema := buildVariantSchema(t)

	columns := map[string]bool{}
	for _, col := range schema.Columns() {
		columns[col.Name()] = true
	}
	readable := map[string]bool{}
	for _, field := range schema.ReadableFields() {
		readable[field.Name()] = true
	}

	for _, name := range TemplateVirtualFields {
		assert.False(t, columns[name], "field %q must not be a physical column", name)
		assert.True(t, readable[name], "field %q must still be selectable", name)
	}
	assert.True(t, columns[ProductVariantFieldSku], "ordinary fields are unaffected")
}

// The two JSON files are edited independently, so a type on one side can drift from the other
// and only surface as a conversion failure at read time. This pins them together.
func TestProductVariant_TemplateFieldTypesMatchTemplate(t *testing.T) {
	variantSchema := buildVariantSchema(t)
	templateSchema := ProductTemplateSchemaBuilder().Build()

	for virtualName, sourceName := range TemplateSourceField {
		virtualField, ok := variantSchema.Field(virtualName)
		require.True(t, ok, "variant field %q", virtualName)
		sourceField, ok := templateSchema.Field(sourceName)
		require.True(t, ok, "template field %q", sourceName)

		assert.Equal(t,
			sourceField.DataType().String(), virtualField.DataType().String(),
			"variant field %q must mirror the type of template field %q", virtualName, sourceName)
	}
}

func TestProductVariant_FillFromTemplateCopiesValues(t *testing.T) {
	template := NewProductTemplate()
	name := model.LangJson{"en-US": "Classic T-Shirt"}
	template.SetName(&name)
	template.SetStatus(WrapProductTemplateStatus(string(ProductTemplateStatusDiscontinued)))
	categoryId := model.Id("01J00000000000000000000C")
	template.SetCategoryId(&categoryId)
	saleOk := true
	template.SetSaleOk(&saleOk)

	variant := NewProductVariant()
	variant.FillFromTemplate(template)

	require.NotNil(t, variant.GetTemplateName())
	assert.Equal(t, "Classic T-Shirt", (*variant.GetTemplateName())["en-US"])
	require.NotNil(t, variant.GetTemplateStatus())
	assert.Equal(t, ProductTemplateStatusDiscontinued, *variant.GetTemplateStatus())
	require.NotNil(t, variant.GetTemplateCategoryId())
	assert.Equal(t, categoryId, *variant.GetTemplateCategoryId())
	require.NotNil(t, variant.GetTemplateSaleOk())
	assert.True(t, *variant.GetTemplateSaleOk())
}

// A variant whose template was deleted must read as "unknown", not as a product with an empty
// name -- the caller has to be able to tell the two apart.
func TestProductVariant_FillFromNilTemplateLeavesFieldsNil(t *testing.T) {
	variant := NewProductVariant()

	variant.FillFromTemplate(nil)

	assert.Nil(t, variant.GetTemplateName())
	assert.Nil(t, variant.GetTemplateStatus())
	assert.Nil(t, variant.GetTemplateCategoryId())
	assert.Nil(t, variant.GetTemplateSaleOk())
}

// A virtual value must never survive into a write, whatever a client sends.
//
// The assertion is on the result map rather than on the error count: what matters is that
// template_name is absent from what the repository would write. Whether the rest of this payload
// is otherwise valid is a different concern, and asserting on it would make this test fail for
// reasons that have nothing to do with virtual fields.
func TestProductVariant_TemplateFieldsDroppedByValidate(t *testing.T) {
	schema := buildVariantSchema(t)

	result, _ := schema.Validate(dmodel.DynamicFields{
		ProductVariantFieldProductTemplateId: "01JBXR8ZQ0YT4W9F6K2M3N5P7Q",
		ProductVariantFieldCombinationKey:    "black-m",
		ProductVariantFieldStatus:            string(ProductVariantStatusActive),
		ProductVariantFieldOrgId:             "01JBXR8ZQ0YT4W9F6K2M3N5P8R",
		ProductVariantFieldTemplateName:      map[string]any{"en-US": "Injected"},
	})

	assert.NotContains(t, result, ProductVariantFieldTemplateName)
}

// The same must hold on update, where a client typically PUTs back the whole object it just read.
// Validate returns a nil map alongside any error, so the etag is supplied here to get a populated
// result -- otherwise both assertions would pass vacuously against nil.
func TestProductVariant_TemplateFieldsDroppedByValidateOnEdit(t *testing.T) {
	schema := buildVariantSchema(t)

	result, errs := schema.Validate(dmodel.DynamicFields{
		ProductVariantFieldSku:          "SKU-1",
		basemodel.FieldEtag:             "01JBXR8ZQ0YT4W9F6K2M3N5P9S",
		ProductVariantFieldTemplateName: map[string]any{"en-US": "Injected"},
	}, true)

	require.Zero(t, errs.Count(), "%v", errs)
	require.NotNil(t, result)
	assert.NotContains(t, result, ProductVariantFieldTemplateName)
	assert.Contains(t, result, ProductVariantFieldSku, "ordinary fields still pass through")
}
