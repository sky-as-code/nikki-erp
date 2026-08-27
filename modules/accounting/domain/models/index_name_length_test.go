package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// PostgreSQL truncates identifiers at 63 bytes, silently, at CREATE time. The DDL generator refuses
// to emit an over-long name, so an overflow turns "make ent-migration" into a failure a long way
// from the model file that caused it.
//
// The budget below is deliberately stricter than 63. Names here are measured in the nikkierp binary,
// which has no tenant key; the coremart binary prepends one and produces the same names about ten
// bytes longer. Checking against the raw limit would pass here and fail there, which is exactly the
// trap docs/wiki/04 warns about — so the headroom is reserved up front.
const (
	postgresIdentifierLimit = 63

	// The longest suffix the query builder appends: a partial unique emits _ukey_notnull.
	longestIndexSuffix = len("_ukey_notnull")

	// Reserved for coremart's tenant key, which is absent from these schemas but present in the
	// binary that generates the migration.
	tenantKeyHeadroom = 12

	maxIndexNameLength = postgresIdentifierLimit - longestIndexSuffix - tenantKeyHeadroom
)

func TestIndexNamesFitWithinIdentifierLimit(t *testing.T) {
	require.NoError(t, basemodel.RegisterJsonBaseSchemas())

	builders := map[string]*dmodel.ModelSchemaBuilder{
		TaxJurisdictionSchemaName:          TaxJurisdictionSchemaBuilder(),
		TaxGroupSchemaName:                 TaxGroupSchemaBuilder(),
		TaxRoundingPolicySchemaName:        TaxRoundingPolicySchemaBuilder(),
		TaxProductClassificationSchemaName: TaxProductClassificationSchemaBuilder(),
		TaxSchemaName:                      TaxSchemaBuilder(),
		TaxDefinitionVersionSchemaName:     TaxDefinitionVersionSchemaBuilder(),
		TaxRateVersionSchemaName:           TaxRateVersionSchemaBuilder(),
		TaxComponentSchemaName:             TaxComponentSchemaBuilder(),
		TaxMappingSchemaName:               TaxMappingSchemaBuilder(),
		TaxMappingLineSchemaName:           TaxMappingLineSchemaBuilder(),
		TaxRuleSchemaName:                  TaxRuleSchemaBuilder(),
		TaxRuleConditionSchemaName:         TaxRuleConditionSchemaBuilder(),
		TaxRuleResultSchemaName:            TaxRuleResultSchemaBuilder(),
	}

	for schemaName, builder := range builders {
		schema := builder.Build()
		require.NotNil(t, schema, "schema %s failed to build", schemaName)

		for _, index := range schema.CompositeUniques() {
			assertIndexNameFits(t, schemaName, index.IndexName)
		}
		for _, index := range schema.SearchIndexGroups() {
			assertIndexNameFits(t, schemaName, index.IndexName)
		}
	}
}

// assertIndexNameFits also insists the name is stated explicitly.
//
// A derived name takes the table name and every column as its stem, which for a table like
// accounting_tax_definition_versions is over budget before the suffix is appended. Requiring an
// explicit name means the overflow is caught by a reader rather than by a migration run.
func assertIndexNameFits(t *testing.T, schemaName string, indexName string) {
	t.Helper()

	assert.NotEmpty(t, indexName,
		"%s declares an index with no explicit index_name; the derived one is too long", schemaName)
	assert.LessOrEqual(t, len(indexName), maxIndexNameLength,
		"index %q on %s is %d bytes, over the %d budget once the suffix and coremart's tenant key "+
			"are added", indexName, schemaName, len(indexName), maxIndexNameLength)
}
