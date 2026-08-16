package models

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// finalizedVariantSchema registers the variant with the schemas it depends on and finalizes the
// registry, mirroring InventoryModule.RegisterModels. Registration matters: foreign keys are
// resolved in FinalizeRelations, so a schema that was only Built reports none — and the picker's
// filter keys off exactly that flag. The real builders are used rather than stubs so the flags
// asserted below are the ones production actually serves.
func finalizedVariantSchema(t *testing.T) *dmodel.ModelSchema {
	t.Helper()
	registry := dmodel.GetSchemaRegistry()
	_ = basemodel.RegisterJsonBaseSchemas()
	// Registration is skipped when a sibling test in this package already registered these
	// schemas — the registry is a process-wide singleton. Finalize still has to run either way,
	// because registering alone leaves foreign keys unresolved.
	if registry.Get(ProductVariantSchemaName) == nil {
		require.NoError(t, errors.Join(
			// Master data first: the template points at all of it.
			dmodel.RegisterSchemaB(ProductTypeSchemaBuilder()),
			dmodel.RegisterSchemaB(ProductCategorySchemaBuilder()),
			dmodel.RegisterSchemaB(BrandSchemaBuilder()),
			dmodel.RegisterSchemaB(ProductTemplateSchemaBuilder()),
			dmodel.RegisterSchemaB(ProductVariantSchemaBuilder()),
		))
	}
	require.NoError(t, registry.FinalizeRelations())
	return registry.Get(ProductVariantSchemaName)
}

// simplizedFields marshals the schema exactly as meta/schema serves it, so these tests read the
// same bytes a client does rather than the in-process predicates behind them.
func simplizedFields(t *testing.T, schema *dmodel.ModelSchema) map[string]map[string]any {
	t.Helper()
	raw, err := json.Marshal(schema.ToSimplized())
	require.NoError(t, err)

	var dto struct {
		Fields map[string]map[string]any `json:"fields"`
	}
	require.NoError(t, json.Unmarshal(raw, &dto))
	return dto.Fields
}

// selectableFieldNames mirrors the frontend's getSelectableSchemaFieldNames (DataTable.tsx): the
// column picker offers every field that is neither server-owned nor a relation placeholder.
// Kept in step with that filter deliberately — this is the behaviour the UI actually shows.
func selectableFieldNames(fields map[string]map[string]any) map[string]bool {
	selectable := map[string]bool{}
	for name, field := range fields {
		isSystem, _ := field["is_system_field"].(bool)
		isEdge, _ := field["is_edge_model"].(bool)
		if !isSystem && !isEdge {
			selectable[name] = true
		}
	}
	return selectable
}

// The regression this whole change exists for. Every template_* field is computed, and folding
// "computed" into is_system_field used to strip all of them from the column picker even though
// meta/schema returned them correctly.
func TestProductVariant_ComputedFieldsAreSelectableColumns(t *testing.T) {
	fields := simplizedFields(t, finalizedVariantSchema(t))
	selectable := selectableFieldNames(fields)

	for _, name := range templateComputedFields() {
		require.Contains(t, fields, name, "field %q must be served by meta/schema", name)
		assert.True(t, selectable[name],
			"computed field %q must be offered as a selectable column", name)
	}
}

// The other half of the picker's filter: a field the server owns, or one standing for a relation
// rather than a value, must stay out of the list.
func TestProductVariant_KeysAndEdgesAreNotSelectableColumns(t *testing.T) {
	fields := simplizedFields(t, finalizedVariantSchema(t))
	selectable := selectableFieldNames(fields)

	assert.False(t, selectable[ProductVariantFieldProductTemplateId],
		"a foreign key is server-owned and belongs in a relation picker, not a column list")
	assert.False(t, selectable["id"], "the primary key is not a business column")
	assert.False(t, selectable[ProductVariantEdgeTemplate],
		"an edge stands for a relation and cannot render as a column")

	assert.True(t, selectable[ProductVariantFieldSku],
		"an ordinary business column stays selectable")
}

// The flags a client reads must agree with each other: computed-and-unpersisted is what makes a
// field virtual, and is_system_field must stay orthogonal to all three.
func TestProductVariant_ServedFlagsAreConsistent(t *testing.T) {
	fields := simplizedFields(t, finalizedVariantSchema(t))

	templateName := fields[ProductVariantFieldTemplateName]
	assert.Equal(t, true, templateName["is_computed"])
	assert.Equal(t, false, templateName["is_persisted"])
	assert.Equal(t, true, templateName["is_virtual"])
	assert.Equal(t, false, templateName["is_edge_model"])
	assert.Equal(t, false, templateName["is_system_field"],
		"read-only is not the same as server-owned")

	edge := fields[ProductVariantEdgeTemplate]
	assert.Equal(t, true, edge["is_edge_model"])
	assert.Equal(t, true, edge["is_computed"], "an edge is hydrated, not supplied")
	assert.Equal(t, false, edge["is_persisted"])
	assert.Equal(t, true, edge["is_virtual"])

	sku := fields[ProductVariantFieldSku]
	assert.Equal(t, false, sku["is_computed"])
	assert.Equal(t, true, sku["is_persisted"], "an ordinary column is persisted")
	assert.Equal(t, false, sku["is_virtual"])
	assert.Equal(t, false, sku["is_system_field"])
}
