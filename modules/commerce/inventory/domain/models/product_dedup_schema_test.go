package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The import pilot recognises a record it wrote before by (source_system, external_id). The
// three pilot models must declare both fields and the strict partial unique key over them, or
// a re-import silently duplicates rows.
func TestPilotModelsDeclareTheDeduplicationKey(t *testing.T) {
	requireBaseSchemasRegistered(t)

	for name, builder := range map[string]func() *dmodel.ModelSchemaBuilder{
		ProductCategorySchemaName: ProductCategorySchemaBuilder,
		ProductTemplateSchemaName: ProductTemplateSchemaBuilder,
		ProductVariantSchemaName:  ProductVariantSchemaBuilder,
	} {
		t.Run(name, func(t *testing.T) {
			schema := builder().Build()

			source := requireField(t, schema, "source_system")
			assert.True(t, source.IsRequiredForCreate(), "the not-null side of the key must be required")
			require.NotNil(t, source.Default())
			assert.Equal(t, "manual", *source.Default().Get(), "a UI-created row is a manual one")

			external := requireField(t, schema, "external_id")
			assert.False(t, external.IsRequiredForCreate(), "external_id is optional outside import")

			// Other strict keys may exist (the category's one-root-per-org), so find the dedup one.
			var dedup *dmodel.PartialUniqueParam
			for _, candidate := range schema.PartialUniquesStrict() {
				if candidate.NullableField == "external_id" {
					dedup = &candidate
				}
			}
			require.NotNil(t, dedup, "the strict partial unique key over external_id")
			assert.Equal(t, []string{"source_system"}, dedup.NotNullFields)
			assert.LessOrEqual(t, len(dedup.IndexName)+len("_ukey"), 63, "postgres identifier limit")
		})
	}
}
