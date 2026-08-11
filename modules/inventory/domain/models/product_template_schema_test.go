package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestProductTemplateSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductTemplateSchemaBuilder().Build()

	assert.Equal(t, ProductTemplateSchemaName, schema.Name())
	assert.Equal(t, "inventory_product_templates", schema.TableName())
	assert.Equal(t, ProductTemplateFieldName, schema.RecordLabelField())

	// BR §6.1.2: a template must be classifiable, so type and category are mandatory while
	// brand stays optional.
	assert.True(t, requireField(t, schema, ProductTemplateFieldProductTypeId).IsRequiredForCreate())
	assert.True(t, requireField(t, schema, ProductTemplateFieldCategoryId).IsRequiredForCreate())
	assert.False(t, requireField(t, schema, ProductTemplateFieldBrandId).IsRequiredForCreate())
}

func TestProductVariantSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductVariantSchemaBuilder().Build()

	assert.Equal(t, ProductVariantSchemaName, schema.Name())
	assert.Equal(t, "inventory_product_variants", schema.TableName())
	// BR-PROD-VAR-001: every variant belongs to exactly one template.
	assert.True(t, requireField(t, schema, ProductVariantFieldProductTemplateId).IsRequiredForCreate())
}

// AC-PROD-004, BR §5.2 and §7.6: template-owned fields must not be duplicated onto the variant.
// Adding any of these columns would let a variant drift from its template, which is the specific
// failure the Template/Variant split exists to prevent.
func TestVariantDoesNotDuplicateTemplateOwnedFields(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductVariantSchemaBuilder().Build()

	for _, fieldName := range []string{
		ProductTemplateFieldName,
		ProductTemplateFieldCategoryId,
		ProductTemplateFieldProductTypeId,
		ProductTemplateFieldBrandId,
		ProductTemplateFieldSaleOk,
		ProductTemplateFieldPurchaseOk,
	} {
		_, ok := schema.Fields()[fieldName]
		assert.Falsef(t, ok,
			"variant must inherit %q from its template, not store its own copy", fieldName)
	}
}

// BR §5.5 and AC-PROD-013: the display name is computed from template name + attribute values.
// Storing it would let it drift out of step with the combination it describes.
func TestVariantHasNoStoredDisplayName(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductVariantSchemaBuilder().Build()

	for _, fieldName := range []string{"display_name", "variant_name"} {
		_, ok := schema.Fields()[fieldName]
		assert.Falsef(t, ok, "%q must be computed, never stored", fieldName)
	}
}

// BR-PROD-VAR-005 and AC-PROD-010: a single variant is the template's only concrete product by
// definition, so no flag is needed to mark it.
func TestVariantHasNoDefaultFlag(t *testing.T) {
	requireBaseSchemasRegistered(t)

	_, ok := ProductVariantSchemaBuilder().Build().Fields()["is_default_variant"]
	assert.False(t, ok, "the core model must not need an is_default_variant flag")
}

// BR-PROD-VAR-002 and AC-PROD-012. The database unique is unconditional; the engine additionally
// scopes it to non-archived rows, because an archived variant must be allowed to keep its
// combination for history while a replacement takes it over.
func TestVariantCombinationIsUnique(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductVariantSchemaBuilder().Build()

	found := false
	for _, composite := range schema.CompositeUniques() {
		if len(composite.Fields) == 2 &&
			composite.Fields[0] == ProductVariantFieldProductTemplateId &&
			composite.Fields[1] == ProductVariantFieldCombinationKey {
			found = true
		}
	}
	assert.True(t, found,
		"a template must not hold two variants with the same combination")
}

// AC-PROD-018: status and is_archived are different concepts and must both exist as their own
// fields. Collapsing them would make "discontinued but still listed" unrepresentable.
func TestTemplateAndVariantSeparateStatusFromArchive(t *testing.T) {
	requireBaseSchemasRegistered(t)

	for name, builder := range map[string]*dmodel.ModelSchemaBuilder{
		ProductTemplateSchemaName: ProductTemplateSchemaBuilder(),
		ProductVariantSchemaName:  ProductVariantSchemaBuilder(),
	} {
		schema := builder.Build()
		_, hasStatus := schema.Fields()[ProductTemplateFieldStatus]
		_, hasArchived := schema.Fields()[basemodel.FieldIsArchived]
		assert.Truef(t, hasStatus, "schema %q must carry a business status", name)
		assert.Truef(t, hasArchived, "schema %q must carry is_archived", name)
	}
}

func TestProductTemplateStatusEnumValues(t *testing.T) {
	requireBaseSchemasRegistered(t)

	field := requireField(t, ProductTemplateSchemaBuilder().Build(), ProductTemplateFieldStatus)

	assert.ElementsMatch(t,
		[]string{
			ProductTemplateStatusDraft.String(),
			ProductTemplateStatusActive.String(),
			ProductTemplateStatusDiscontinued.String(),
		},
		field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues])
}

// BR §8.9 and BR-PROD-TPL-003: unarchiving a template must restore only what it cascaded to.
func TestArchiveSourceEnumValues(t *testing.T) {
	requireBaseSchemasRegistered(t)

	field := requireField(t, ProductVariantSchemaBuilder().Build(), ProductVariantFieldArchiveSource)

	assert.ElementsMatch(t,
		[]string{
			ArchiveSourceUser.String(),
			ArchiveSourceTemplateCascade.String(),
			ArchiveSourceSystemSync.String(),
		},
		field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues])
}

// AC-PROD-019 / AC-PROD-020: archiving and discontinuing each independently withdraw a variant
// from new transactions.
func TestVariantIsSelectable(t *testing.T) {
	archived, active := true, ProductVariantStatusActive
	discontinued := ProductVariantStatusDiscontinued
	notArchived := false

	selectable := NewProductVariant()
	selectable.SetIsArchived(&notArchived)
	selectable.SetStatus(&active)
	assert.True(t, selectable.IsSelectable())

	archivedVariant := NewProductVariant()
	archivedVariant.SetIsArchived(&archived)
	archivedVariant.SetStatus(&active)
	assert.False(t, archivedVariant.IsSelectable(), "an archived variant is not selectable")

	discontinuedVariant := NewProductVariant()
	discontinuedVariant.SetIsArchived(&notArchived)
	discontinuedVariant.SetStatus(&discontinued)
	assert.False(t, discontinuedVariant.IsSelectable(), "a discontinued variant is not selectable")
}

// BR §6.7: a variant references the template-scoped value, not the global one, so that
// "every value is allowed by the template" is enforceable by the data model.
func TestVariantValueJunctionPointsAtTemplateScopedValue(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductVariantAttributeValueSchemaBuilder().Build()

	requireField(t, schema, ProductVariantAttributeValueFieldTemplateAttributeValueId)
	_, ok := schema.Fields()[ProductAttributeValueFieldId+"_global"]
	assert.False(t, ok)
}
