package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// A malformed JSON model file panics the app at start-up; these tests turn that into a test
// failure and pin the field set the Products rules depend on.

func TestProductTypeSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductTypeSchemaBuilder().Build()

	assert.Equal(t, ProductTypeSchemaName, schema.Name())
	assert.Equal(t, "inventory_product_types", schema.TableName())
	// Without a record label field the frontend relation picker shows raw ULIDs.
	assert.Equal(t, ProductTypeFieldName, schema.RecordLabelField())
	// Processing logic keys off the code, so it must be present and unique.
	assert.True(t, requireField(t, schema, ProductTypeFieldCode).IsRequiredForCreate())
}

// The capability flags decide which modules may consume a product, so they must exist as booleans
// rather than be inferred from the code.
func TestProductTypeCapabilityFlags(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductTypeSchemaBuilder().Build()

	for _, fieldName := range []string{
		ProductTypeFieldSupportsStock,
		ProductTypeFieldSupportsSale,
		ProductTypeFieldSupportsPurchase,
		ProductTypeFieldSupportsManufacturing,
	} {
		field := requireField(t, schema, fieldName)
		assert.Equal(t, dmodel.FieldDataTypeNameBoolean, field.DataType().String(),
			"field %q must be a boolean", fieldName)
	}
}

func TestProductCategorySchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductCategorySchemaBuilder().Build()

	assert.Equal(t, ProductCategorySchemaName, schema.Name())
	assert.Equal(t, "inventory_product_categories", schema.TableName())
	assert.Equal(t, ProductCategoryFieldName, schema.RecordLabelField())
	// A root category has no parent, so the self-FK must stay optional.
	assert.False(t, requireField(t, schema, ProductCategoryFieldParentCategoryId).IsRequiredForCreate())
}

func TestBrandSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := BrandSchemaBuilder().Build()

	assert.Equal(t, BrandSchemaName, schema.Name())
	assert.Equal(t, "inventory_brands", schema.TableName())
	assert.Equal(t, BrandFieldName, schema.RecordLabelField())
	// Brand is optional on a template, but a brand record itself needs a name.
	assert.True(t, requireField(t, schema, BrandFieldName).IsRequiredForCreate())
}

// Every Products resource carries Archive/Unarchive, so each must extend archivable_model;
// is_archived is never declared by hand.
func TestMasterModelsAreArchivable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	builders := map[string]*dmodel.ModelSchemaBuilder{
		ProductTypeSchemaName:     ProductTypeSchemaBuilder(),
		ProductCategorySchemaName: ProductCategorySchemaBuilder(),
		BrandSchemaName:           BrandSchemaBuilder(),
	}

	for name, builder := range builders {
		schema := builder.Build()
		_, ok := schema.Fields()[basemodel.FieldIsArchived]
		assert.Truef(t, ok, "schema %q must extend core.basemodel.archivable_model", name)
	}
}

func requireBaseSchemasRegistered(t *testing.T) {
	t.Helper()
	// Normally done by CoreModule.RegisterModels during app start-up.
	_ = basemodel.RegisterJsonBaseSchemas()
}

func requireField(t *testing.T, schema *dmodel.ModelSchema, fieldName string) *dmodel.ModelField {
	t.Helper()
	field, ok := schema.Fields()[fieldName]
	require.Truef(t, ok, "schema %q has no field %q", schema.Name(), fieldName)
	return field
}
