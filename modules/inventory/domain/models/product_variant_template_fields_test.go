package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
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

// templateComputedFields is the test's own list of the template-owned fields the variant
// exposes. Deliberately a literal rather than derived from the schema: a field accidentally
// dropped from the JSON must fail here, not shrink the loop.
func templateComputedFields() []string {
	return []string{
		ProductVariantFieldTemplateName,
		ProductVariantFieldTemplateShortName,
		ProductVariantFieldTemplateDescription,
		ProductVariantFieldTemplateSalesDescription,
		ProductVariantFieldTemplatePurchaseDescription,
		ProductVariantFieldTemplateCategoryId,
		ProductVariantFieldTemplateBrandId,
		ProductVariantFieldTemplateProductTypeId,
		ProductVariantFieldTemplateStatus,
		ProductVariantFieldTemplateSaleOk,
	}
}

// The variant schema must load with every template_* field computed (and therefore virtual): one
// of them declared as an ordinary field would silently add a column and demand a migration.
func TestProductVariant_TemplateFieldsAreComputed(t *testing.T) {
	schema := buildVariantSchema(t)

	for _, name := range templateComputedFields() {
		field, ok := schema.Field(name)
		require.True(t, ok, "field %q must be declared on the schema", name)
		assert.True(t, field.IsComputed(), "field %q must be computed", name)
		assert.True(t, field.IsVirtual(), "a computed field is virtual: no column, never written")

		def, err := computed.DefOf(field)
		require.NoError(t, err)
		assert.Equal(t, computed.ComputeRelated, def.Kind, "field %q copies from the template edge", name)
		assert.False(t, def.IsStored)
		assert.True(t, strings.HasPrefix(def.Related, ProductVariantEdgeTemplate+"."),
			"field %q must copy through the %q edge, not some other path", name, ProductVariantEdgeTemplate)
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

	for _, name := range templateComputedFields() {
		assert.False(t, columns[name], "field %q must not be a physical column", name)
		assert.True(t, readable[name], "field %q must still be selectable", name)
	}
	assert.True(t, columns[ProductVariantFieldSku], "ordinary fields are unaffected")
}

// The two JSON files are edited independently, so a type on one side can drift from the other
// and only surface as a conversion failure at read time. This pins each computed field's
// declared type to the template leaf its definition names.
func TestProductVariant_TemplateFieldTypesMatchTemplate(t *testing.T) {
	variantSchema := buildVariantSchema(t)
	templateSchema := ProductTemplateSchemaBuilder().Build()

	for _, name := range templateComputedFields() {
		field, ok := variantSchema.Field(name)
		require.True(t, ok, "variant field %q", name)
		def, err := computed.DefOf(field)
		require.NoError(t, err)

		leafName := strings.TrimPrefix(def.Related, ProductVariantEdgeTemplate+".")
		sourceField, ok := templateSchema.Field(leafName)
		require.True(t, ok, "template field %q named by %q's definition", leafName, name)

		assert.Equal(t,
			sourceField.DataType().String(), field.DataType().String(),
			"variant field %q must mirror the type of template field %q", name, leafName)
	}
}

// A computed value must never survive into a write, whatever a client sends.
//
// The assertion is on the result map rather than on the error count: what matters is that
// template_name is absent from what the repository would write. Whether the rest of this payload
// is otherwise valid is a different concern, and asserting on it would make this test fail for
// reasons that have nothing to do with computed fields.
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
